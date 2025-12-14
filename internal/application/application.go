package application

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"calendar/internal/config"
	"calendar/internal/databases"
	"calendar/internal/handlers"
	"calendar/internal/kafka"
	"calendar/internal/logger"
	"calendar/internal/repos"
	"calendar/internal/services"
)

type App struct {
	cfg    *config.Config
	log    logger.Logger
	db     *databases.Postgres
	server *http.Server

	events   services.EventsService
	producer *kafka.Producer
	consumer *kafka.Consumer
}

// NewApp собирает все зависимости: логгер, БД, storage, HTTP‑хендлеры и сервер.
func NewApp(cfg *config.Config) (*App, error) {
	// 1. Логгер
	log := logger.InitLogger(cfg.Logging.Level)

	// 2. База данных
	db, err := databases.NewPostgres(cfg)
	if err != nil {
		log.Error("failed to connect to postgres", "error", err)
		return nil, err
	}
	log.Info("connected to postgres")

	// 3. Репозиторий событий
	eventsRepo := repos.NewPGEventStorage(db.DB)

	// 4. Сервис событий
	eventsService := services.NewEventsService(eventsRepo)

	// 5. HTTP‑хендлеры
	h := handlers.NewHandlers(log, eventsService)

	// 6. HTTP‑роутер
	mux := http.NewServeMux()

	// healthcheck
	mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("OK"))
	})

	// маршруты /api/events..., внутри RegisterRoutes — CRUD
	h.RegisterRoutes(mux)

	// 7. HTTP‑сервер
	addr := fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)
	srv := &http.Server{
		Addr:    addr,
		Handler: mux,
	}

	// 8. Kafka producer
	producer := kafka.NewProducer(cfg, log, eventsRepo)

	// 9. Kafka consumer
	consumer := kafka.NewConsumer(cfg, log)

	return &App{
		cfg:      cfg,
		log:      log,
		db:       db,
		server:   srv,
		events:   eventsService,
		producer: producer,
		consumer: consumer,
	}, nil
}

// Run запускает HTTP‑сервер, Kafka producer и consumer, и делает graceful shutdown.
func (a *App) Run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	a.log.Info("starting http server", "addr", a.server.Addr)

	// сервер в отдельной горутине
	go func() {
		if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			a.log.Error("http server error", "error", err)
		}
	}()

	// запускаем Kafka producer
	if err := a.producer.Start(ctx); err != nil {
		a.log.Error("failed to start kafka producer", "error", err)
		return err
	}
	a.log.Info("kafka producer started")

	// запускаем Kafka consumer
	if err := a.consumer.Start(ctx); err != nil {
		a.log.Error("failed to start kafka consumer", "error", err)
		return err
	}
	a.log.Info("kafka consumer started")

	// ждём сигнала ОС
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

	<-stop
	a.log.Info("shutdown signal received")

	// отменяем контекст для остановки producer и consumer
	cancel()

	// контекст для graceful shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutdownCancel()

	// останавливаем Kafka consumer
	if err := a.consumer.Stop(); err != nil {
		a.log.Error("kafka consumer stop error", "error", err)
	}

	// останавливаем Kafka producer
	if err := a.producer.Stop(); err != nil {
		a.log.Error("kafka producer stop error", "error", err)
	}

	// останавливаем HTTP‑сервер
	if err := a.server.Shutdown(shutdownCtx); err != nil {
		a.log.Error("http server shutdown error", "error", err)
	}

	// закрываем БД
	if err := a.db.Close(); err != nil {
		a.log.Error("postgres close error", "error", err)
	}

	a.log.Info("application stopped gracefully")
	return nil
}
