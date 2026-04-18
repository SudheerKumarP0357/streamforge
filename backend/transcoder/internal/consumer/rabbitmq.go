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
	url       string
	conn      *amqp.Connection
	ch        *amqp.Channel
	queueName string
}

func NewConsumer(url, queueName string) (*Consumer, error) {
	logger := slog.With("component", "rabbitmq_consumer", "queue", queueName)
	logger.Info("connecting to RabbitMQ")

	conn, err := amqp.Dial(url)
	if err != nil {
		logger.Error("failed to connect to RabbitMQ", "error", err.Error())
		return nil, fmt.Errorf("failed to connect to rabbitmq: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error("failed to open channel", "error", err.Error())
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
		logger.Error("failed to set QoS", "error", err.Error())
		return nil, fmt.Errorf("failed to set QoS: %w", err)
	}

	logger.Info("connected to RabbitMQ")

	return &Consumer{
		url:       url,
		conn:      conn,
		ch:        ch,
		queueName: queueName,
	}, nil
}

// reconnect attempts to re-establish the RabbitMQ connection with exponential backoff.
func (c *Consumer) reconnect(ctx context.Context) error {
	logger := slog.With("component", "rabbitmq_consumer", "queue", c.queueName)
	backoff := 1 * time.Second
	maxBackoff := 60 * time.Second

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		logger.Warn("connection lost, attempting reconnect", "backoff_ms", backoff.Milliseconds())
		time.Sleep(backoff)

		conn, err := amqp.Dial(c.url)
		if err != nil {
			logger.Error("reconnect dial failed", "error", err.Error(), "backoff_ms", backoff.Milliseconds())
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		ch, err := conn.Channel()
		if err != nil {
			conn.Close()
			logger.Error("reconnect channel failed", "error", err.Error())
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		if err := ch.Qos(1, 0, false); err != nil {
			ch.Close()
			conn.Close()
			logger.Error("reconnect QoS failed", "error", err.Error())
			backoff *= 2
			if backoff > maxBackoff {
				backoff = maxBackoff
			}
			continue
		}

		c.conn = conn
		c.ch = ch
		logger.Info("reconnected to RabbitMQ")
		return nil
	}
}

func (c *Consumer) Start(ctx context.Context, handler func(ctx context.Context, job TranscodeJob) error) error {
	logger := slog.With("component", "rabbitmq_consumer", "queue", c.queueName)

	for {
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
			logger.Error("failed to register consumer, attempting reconnect", "error", err.Error())
			if reconnErr := c.reconnect(ctx); reconnErr != nil {
				return fmt.Errorf("failed to reconnect: %w", reconnErr)
			}
			continue
		}

		logger.Info("started listening for messages")

		if err := c.consumeLoop(ctx, logger, msgs, handler); err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			// Channel closed — try to reconnect
			logger.Warn("message channel closed, attempting reconnect")
			if reconnErr := c.reconnect(ctx); reconnErr != nil {
				return fmt.Errorf("failed to reconnect after channel close: %w", reconnErr)
			}
			continue
		}
		return nil
	}
}

// consumeLoop processes messages from the delivery channel until it closes or context is cancelled.
func (c *Consumer) consumeLoop(ctx context.Context, logger *slog.Logger, msgs <-chan amqp.Delivery, handler func(ctx context.Context, job TranscodeJob) error) error {
	for {
		select {
		case <-ctx.Done():
			logger.Info("context cancelled, stopping consumer")
			return ctx.Err()
		case d, ok := <-msgs:
			if !ok {
				return errors.New("rabbitmq channel closed")
			}

			var job TranscodeJob
			if err := json.Unmarshal(d.Body, &job); err != nil {
				logger.Error("failed to unmarshal message body", "error", err.Error())
				_ = d.Nack(false, false)
				logger.Warn("message nacked",
					"reason", "unmarshal_error",
					"requeue", false,
				)
				continue
			}

			// Inspect queue depth at receive time
			queueDepth := -1
			if q, inspectErr := c.ch.QueueInspect(c.queueName); inspectErr == nil {
				queueDepth = q.Messages
			}
			logger.Info("message received",
				"video_id", job.VideoID,
				"queue_depth", queueDepth,
			)
			metrics.RabbitMQConsumedTotal.WithLabelValues(c.queueName).Inc()

			err := handler(ctx, job)

			if err == nil {
				if ackErr := d.Ack(false); ackErr != nil {
					logger.Error("failed to ack message", "error", ackErr.Error(), "video_id", job.VideoID)
				} else {
					logger.Info("message acked", "video_id", job.VideoID)
				}
			} else {
				logger.Error("failed to process job", "error", err.Error(), "video_id", job.VideoID)
				requeue := false
				if nackErr := d.Nack(false, requeue); nackErr != nil {
					logger.Error("failed to nack message", "error", nackErr.Error(), "video_id", job.VideoID)
				} else {
					logger.Warn("message nacked",
						"video_id", job.VideoID,
						"reason", "handler_error",
						"requeue", requeue,
					)
				}
			}
		}
	}
}

// MonitorQueues routinely logs the queue depth of the main and dlq queues
func (c *Consumer) MonitorQueues(ctx context.Context, dlqName string) {
	logger := slog.With("component", "rabbitmq_consumer")
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Inspect Main Queue
			if q, err := c.ch.QueueInspect(c.queueName); err == nil {
				logger.Info("queue depth report", "queue", c.queueName, "messages", q.Messages, "consumers", q.Consumers)
			} else {
				logger.Warn("failed to inspect main queue", "queue", c.queueName, "error", err.Error())
			}

			// Inspect DLQ
			if q, err := c.ch.QueueInspect(dlqName); err == nil {
				logger.Info("queue depth report", "queue", dlqName, "messages", q.Messages, "consumers", q.Consumers)
			} else {
				logger.Warn("failed to inspect DLQ", "queue", dlqName, "error", err.Error())
			}
		}
	}
}

func (c *Consumer) Close() error {
	logger := slog.With("component", "rabbitmq_consumer", "queue", c.queueName)
	logger.Info("closing RabbitMQ consumer")

	if c.ch != nil {
		if err := c.ch.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			logger.Error("failed to close channel", "error", err.Error())
		}
	}
	if c.conn != nil {
		if err := c.conn.Close(); err != nil && !errors.Is(err, amqp.ErrClosed) {
			return fmt.Errorf("error closing connection: %w", err)
		}
	}

	logger.Info("disconnected from RabbitMQ")
	return nil
}
