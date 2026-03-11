package consumer

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type DLQConsumer struct {
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
}

func NewDLQConsumer(url, queueName string) (*DLQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("failed to open a channel: %w", err)
	}

	err = ch.Qos(1, 0, false)
	if err != nil {
		ch.Close()
		conn.Close()
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	return &DLQConsumer{
		conn:      conn,
		ch:        ch,
		queueName: queueName,
	}, nil
}

func getFailCount(d amqp.Delivery) int {
	deaths, ok := d.Headers["x-death"].([]interface{})
	if !ok {
		return 1
	}

	var total int64
	for _, death := range deaths {
		deathTable, ok := death.(amqp.Table)
		if !ok {
			continue
		}
		if count, ok := deathTable["count"].(int64); ok {
			total += count
		}
	}
	return int(total)
}

func (c *DLQConsumer) Start(ctx context.Context, handler func(ctx context.Context, job TranscodeJob, failCount int) error) error {
	msgs, err := c.ch.Consume(
		c.queueName,
		"dlq-consumer",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		return fmt.Errorf("failed to register dlq consumer: %w", err)
	}

	slog.Info("Started listening for dead-letter messages", "queue", c.queueName)

	for {
		select {
		case <-ctx.Done():
			slog.Info("Context cancelled, stopping DLQ consumer...")
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				slog.Warn("DLQ RabbitMQ channel closed")
				return errors.New("dlq rabbitmq channel closed")
			}

			var job TranscodeJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				slog.Error("Failed to unmarshal DLQ message body", "error", err)
				_ = d.Ack(false) // Ack to drop unparsable junk permanently
				continue
			}

			failCount := getFailCount(d)
			logger := slog.With("video_id", job.VideoID, "fail_count", failCount)

			if failCount >= 3 {
				logger.Error("Job permanently failed after retries")
				_ = d.Ack(false)
				continue
			}

			logger.Info("Processing job from DLQ")
			err := handler(ctx, job, failCount)
			if err == nil {
				_ = d.Ack(false)
			} else {
				logger.Error("Job failed again in DLQ", "error", err)
				// Requeue into DLQ to wait for manual intervention or custom DLX router retry
				_ = d.Nack(false, true) 
			}
		}
	}
}

func (c *DLQConsumer) Close() error {
	slog.Info("Closing DLQ RabbitMQ consumer...")
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
