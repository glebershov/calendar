package kafka

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"calendar/internal/config"
	"calendar/internal/logger"

	"github.com/segmentio/kafka-go"
)

// Consumer читает сообщения из Kafka и логирует их.
type Consumer struct {
	reader  *kafka.Reader
	log     logger.Logger
	cfg     *config.Config
	running bool
	stopCh  chan struct{}
}

// NewConsumer создаёт новый consumer.
func NewConsumer(cfg *config.Config, log logger.Logger) *Consumer {
	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:        cfg.Kafka.Brokers,
		Topic:          cfg.Kafka.Topic,
		GroupID:        "calendar-consumer-group",
		MinBytes:       10e3, // 10KB
		MaxBytes:       10e6, // 10MB
		CommitInterval: time.Second,
		StartOffset:    kafka.LastOffset,
	})

	return &Consumer{
		reader: reader,
		log:    log,
		cfg:    cfg,
		stopCh: make(chan struct{}),
	}
}

// Start запускает consumer, который читает сообщения из Kafka и логирует их.
func (c *Consumer) Start(ctx context.Context) error {
	if c.running {
		return fmt.Errorf("consumer is already running")
	}

	c.running = true
	c.log.Info("starting kafka consumer", "topic", c.cfg.Kafka.Topic, "group_id", "calendar-consumer-group")

	// Запускаем горутину для чтения сообщений
	go c.run(ctx)

	return nil
}

// run читает сообщения из Kafka и логирует их.
func (c *Consumer) run(ctx context.Context) {
	for {
		select {
		case <-ctx.Done():
			c.log.Info("consumer context cancelled")
			return
		case <-c.stopCh:
			c.log.Info("consumer stopped")
			return
		default:
			// Читаем сообщение с таймаутом
			msgCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			msg, err := c.reader.ReadMessage(msgCtx)
			cancel()

			if err != nil {
				if err == context.DeadlineExceeded || err == context.Canceled {
					continue
				}
				c.log.Error("failed to read message from kafka", "error", err)
				time.Sleep(1 * time.Second) // Небольшая задержка перед повтором
				continue
			}

			// Обрабатываем сообщение
			if err := c.processMessage(msg); err != nil {
				c.log.Error("failed to process message", "error", err, "offset", msg.Offset)
			}
		}
	}
}

// processMessage обрабатывает одно сообщение из Kafka.
func (c *Consumer) processMessage(msg kafka.Message) error {
	var eventMsg EventMessage
	if err := json.Unmarshal(msg.Value, &eventMsg); err != nil {
		return fmt.Errorf("failed to unmarshal message: %w", err)
	}

	// Устанавливаем дату отправки (если её еще нет)
	if eventMsg.SentAt.IsZero() {
		eventMsg.SentAt = time.Now()
	}

	// Логируем сообщение
	c.log.Info("received event from kafka",
		"event_id", eventMsg.ID,
		"title", eventMsg.Title,
		"description", eventMsg.Description,
		"start_time", eventMsg.StartTime,
		"end_time", eventMsg.EndTime,
		"owner_id", eventMsg.OwnerID,
		"sent_at", eventMsg.SentAt,
		"offset", msg.Offset,
		"partition", msg.Partition,
	)

	// Также выводим текст сообщения в лог
	messageText := fmt.Sprintf(
		"Event: %s (ID: %s) - %s. Starts: %s, Ends: %s. Sent at: %s",
		eventMsg.Title,
		eventMsg.ID,
		eventMsg.Description,
		eventMsg.StartTime.Format(time.RFC3339),
		eventMsg.EndTime.Format(time.RFC3339),
		eventMsg.SentAt.Format(time.RFC3339),
	)

	c.log.Info("event message text", "message", messageText)

	return nil
}

// Stop останавливает consumer.
func (c *Consumer) Stop() error {
	if !c.running {
		return nil
	}

	c.log.Info("stopping kafka consumer")
	close(c.stopCh)
	c.running = false

	if err := c.reader.Close(); err != nil {
		return fmt.Errorf("failed to close kafka reader: %w", err)
	}

	return nil
}
