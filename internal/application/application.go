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
    "calendar/internal/logger"
)

type App struct {
    cfg    *config.Config
    log    logger.Logger
    db     *databases.Postgres
    server *http.Server
}

// NewApp собирает все зависимости: логгер, БД, HTTP‑сервер.
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

    // 3. HTTP‑роутер 
    mux := http.NewServeMux()
    mux.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
        _, _ = w.Write([]byte("OK"))
    })

    // 4. HTTP‑сервер
    addr := fmt.Sprintf("%s:%d", cfg.HTTPServer.Host, cfg.HTTPServer.Port)
    srv := &http.Server{
        Addr:    addr,
        Handler: mux,
    }

    return &App{
        cfg:    cfg,
        log:    log,
        db:     db,
        server: srv,
    }, nil
}

// Run запускает HTTP‑сервер и делает graceful shutdown 
func (a *App) Run() error {
    a.log.Info("starting http server", "addr", a.server.Addr)

    // запускаем сервер в отдельной горутине
    go func() {
        if err := a.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
            a.log.Error("http server error", "error", err)
        }
    }()

    // ждём сигналов ОС
    stop := make(chan os.Signal, 1)
    signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)

    <-stop
    a.log.Info("shutdown signal received")

    // контекст для graceful shutdown
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    // останавливаем HTTP‑сервер
    if err := a.server.Shutdown(ctx); err != nil {
        a.log.Error("http server shutdown error", "error", err)
    }

    // закрываем БД
    if err := a.db.Close(); err != nil {
        a.log.Error("postgres close error", "error", err)
    }

    a.log.Info("application stopped gracefully")
    return nil
}
