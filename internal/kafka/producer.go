package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"calendar/internal/config"
	"calendar/internal/logger"
	"calendar/internal/repos"

	"github.com/segmentio/kafka-go"
)

// Producer отправляет сообщения в Kafka.
type Producer struct {
	writer  *kafka.Writer
	log     logger.Logger
	repo    repos.EventsRepo
	cfg     *config.Config
	running bool
	stopCh  chan struct{}
}

// NewProducer создаёт новый producer.
func NewProducer(cfg *config.Config, log logger.Logger, repo repos.EventsRepo) *Producer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(cfg.Kafka.Brokers...),
		Topic:        cfg.Kafka.Topic,
		Balancer:     &kafka.LeastBytes{},
		WriteTimeout: 10 * time.Second,
		RequiredAcks: kafka.RequireOne,
	}

	return &Producer{
		writer: writer,
		log:    log,
		repo:   repo,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// Start запускает producer, который периодически читает события из БД и отправляет их в Kafka.
func (p *Producer) Start(ctx context.Context) error {
	if p.running {
		return fmt.Errorf("producer is already running")
	}

	p.running = true
	p.log.Info("starting kafka producer")

	// Запускаем горутину для периодической отправки событий
	go p.run(ctx)

	return nil
}

// run периодически читает события из БД и отправляет их в Kafka.
func (p *Producer) run(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second) // Проверяем каждые 10 секунд
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			p.log.Info("producer context cancelled")
			return
		case <-p.stopCh:
			p.log.Info("producer stopped")
			return
		case <-ticker.C:
			if err := p.sendEventsFromDB(ctx); err != nil {
				p.log.Error("failed to send events from DB", "error", err)
			}
		}
	}
}

// sendEventsFromDB читает события из БД и отправляет их в Kafka.
func (p *Producer) sendEventsFromDB(ctx context.Context) error {
	events, err := p.repo.GetAllEvents(ctx)
	if err != nil {
		return fmt.Errorf("failed to get events from DB: %w", err)
	}

	if len(events) == 0 {
		p.log.Debug("no events to send")
		return nil
	}

	p.log.Info("sending events to kafka", "count", len(events))

	for _, event := range events {
		msg := EventMessage{
			ID:          event.ID,
			Title:       event.Title,
			Description: event.Description,
			StartTime:   event.StartTime,
			EndTime:     event.EndTime,
			OwnerID:     event.OwnerID,
			CreatedAt:   event.CreatedAt,
			UpdatedAt:   event.UpdatedAt,
			SentAt:      time.Now(),
		}

		if err := p.sendMessage(ctx, msg); err != nil {
			p.log.Error("failed to send event message", "event_id", event.ID, "error", err)
			continue
		}

		p.log.Debug("event sent to kafka", "event_id", event.ID, "title", event.Title)
	}

	return nil
}

// sendMessage отправляет одно сообщение в Kafka.
func (p *Producer) sendMessage(ctx context.Context, msg EventMessage) error {
	body, err := json.Marshal(msg)
	if err != nil {
		return fmt.Errorf("failed to marshal message: %w", err)
	}

	kafkaMsg := kafka.Message{
		Key:   []byte(msg.ID),
		Value: body,
		Time:  time.Now(),
	}

	if err := p.writer.WriteMessages(ctx, kafkaMsg); err != nil {
		return fmt.Errorf("failed to write message to kafka: %w", err)
	}

	return nil
}

// Stop останавливает producer.
func (p *Producer) Stop() error {
	if !p.running {
		return nil
	}

	p.log.Info("stopping kafka producer")
	close(p.stopCh)
	p.running = false

	if err := p.writer.Close(); err != nil {
		return fmt.Errorf("failed to close kafka writer: %w", err)
	}

	return nil
}
