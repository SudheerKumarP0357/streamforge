package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"streamforge/transcoder/internal/metrics"
)

type TranscodeJob struct {
	VideoID    string `json:"video_id"`
	RawBlobURL string `json:"raw_blob_url"`
	UserID     string `json:"user_id"`
	UploadedAt string `json:"uploaded_at"`
}

type Consumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
}

func NewConsumer(url, queueName string) (*Consumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	// Fair dispatch - don't dispatch a new message to a worker until it has processed and acknowledged the previous one.
	err = ch.Qos(
		1,     // prefetch count
		0,     // prefetch size
		false, // global
	)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &Consumer{
		conn:      conn,
		ch:        ch,
		queueName: queueName,
	}, nil
}

func (c *Consumer) Start(ctx context.Context, handler func(ctx context.Context, job TranscodeJob) error) error {
	msgs, err := c.ch.Consume(
		c.queueName,
		"",    // consumer tag
		false, // auto-ack (set to false to manually ack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("failed to register a consumer: %w", err)
	}

	slog.Info("Started listening for messages", "queue", c.queueName)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping consumer...")
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				slog.Warn("RabbitMQ message channel closed")
				return errors.New("rabbitmq channel closed")
			}

			var job TranscodeJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				slog.Error("Failed to unmarshal message body", "error", err)
				_ = d.Nack(false, false)
				continue
			}

			slog.Info("Received job for video", "video_id", job.VideoID)
			metrics.RabbitMQConsumedTotal.WithLabelValues(c.queueName).Inc()
			
			err := handler(ctx, job)
			
			if err == nil {
				if ackErr := d.Ack(false); ackErr != nil {
					slog.Error("Failed to ack message", "error", ackErr, "video_id", job.VideoID)
				}
			} else {
				slog.Error("Failed to process job", "error", err, "video_id", job.VideoID)
				if nackErr := d.Nack(false, false); nackErr != nil {
					slog.Error("Failed to nack message", "error", nackErr, "video_id", job.VideoID)
				}
			}
		}
	}
}

// MonitorQueues routinely logs the queue depth of the main and dlq queues
func (c *Consumer) MonitorQueues(ctx context.Context, dlqName string) {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Inspect Main Queue
			if q, err := c.ch.QueueInspect(c.queueName); err == nil {
				slog.Info("Queue depth report", "queue", c.queueName, "messages", q.Messages, "consumers", q.Consumers)
			} else {
				slog.Warn("Failed to inspect main queue", "error", err)
			}

			// Inspect DLQ
			if q, err := c.ch.QueueInspect(dlqName); err == nil {
				slog.Info("Queue depth report", "queue", dlqName, "messages", q.Messages, "consumers", q.Consumers)
			} else {
				slog.Warn("Failed to inspect DLQ", "error", err)
			}
		}
	}
}

func (c *Consumer) Close() error {
	slog.Info("Closing RabbitMQ consumer...")
	if c.ch != nil {
		if err := c.ch.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			slog.Error("Failed to close channel", "error", err)
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			return fmt.Errorf("error closing connection: %w", err)
		}
	}
	return nil
}
