package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
	"streamforge/api/internal/metrics"
)

type Publisher struct {
	url  string
	conn *amqp.Connection
	ch   *amqp.Channel
}

type TranscodeJob struct {
	VideoID    string `json:"video_id"`
	RawBlobURL string `json:"raw_blob_url"`
	UserID     string `json:"user_id"`
	UploadedAt string `json:"uploaded_at"`
}

func NewPublisher(url string) (*Publisher, error) {
	p := &Publisher{url: url}
	
	if err := p.connect(); err != nil {
		return nil, fmt.Errorf("failed to initially connect to rabbitmq: %w", err)
	}

	return p, nil
}

func (p *Publisher) connect() error {
	logger := slog.With("component", "rabbitmq", "operation", "connect")
	logger.Info("connecting to RabbitMQ")

	conn, err := amqp.Dial(p.url)
	if err != nil {
		logger.Error("failed to dial RabbitMQ", "error", err.Error())
		return fmt.Errorf("dial: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		logger.Error("failed to open channel", "error", err.Error())
		return fmt.Errorf("channel: %w", err)
	}

	// 1. Declare Main Exchange
	err = ch.ExchangeDeclare(
		"transcoder", // name
		"direct",     // type
		true,         // durable
		false,        // auto-deleted
		false,        // internal
		false,        // no-wait
		nil,          // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		logger.Error("failed to declare exchange", "exchange", "transcoder", "error", err.Error())
		return fmt.Errorf("exchange declare: %w", err)
	}

	// 2. Declare Dead Letter Exchange (DLX)
	err = ch.ExchangeDeclare(
		"transcoder.dlx",
		"direct",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		logger.Error("failed to declare DLX exchange", "exchange", "transcoder.dlx", "error", err.Error())
		return fmt.Errorf("dlx exchange declare: %w", err)
	}

	// 3. Declare Main Queue with DLX argument
	args := amqp.Table{
		"x-dead-letter-exchange":    "transcoder.dlx",
		"x-dead-letter-routing-key": "dead", // Route explicitly for dead messages if needed
	}
	mainQueue, err := ch.QueueDeclare(
		"transcoder.jobs", // name
		true,              // durable
		false,             // delete when unused
		false,             // exclusive
		false,             // no-wait
		args,              // arguments
	)
	if err != nil {
		ch.Close()
		conn.Close()
		logger.Error("failed to declare main queue", "queue", "transcoder.jobs", "error", err.Error())
		return fmt.Errorf("queue declare: %w", err)
	}

	// 4. Bind Main Queue
	err = ch.QueueBind(
		mainQueue.Name, // queue name
		"transcode",    // routing key
		"transcoder",   // exchange
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		logger.Error("failed to bind main queue", "queue", mainQueue.Name, "error", err.Error())
		return fmt.Errorf("queue bind: %w", err)
	}

	// 5. Declare Dead Letter Queue (DLQ)
	dlq, err := ch.QueueDeclare(
		"transcoder.jobs.dlq",
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		logger.Error("failed to declare DLQ", "queue", "transcoder.jobs.dlq", "error", err.Error())
		return fmt.Errorf("dlq declare: %w", err)
	}

	// 6. Bind DLQ
	// Binding # explicitly so any routing key sent to the DLX falls back correctly.
	err = ch.QueueBind(
		dlq.Name,
		"dead", // Match the dead-letter routing key defined above
		"transcoder.dlx",
		false,
		nil,
	)
	if err != nil {
		ch.Close()
		conn.Close()
		logger.Error("failed to bind DLQ", "queue", dlq.Name, "error", err.Error())
		return fmt.Errorf("dlq bind: %w", err)
	}

	p.conn = conn
	p.ch = ch
	logger.Info("connected to RabbitMQ",
		"exchange", "transcoder",
		"queue", "transcoder.jobs",
		"dlq", "transcoder.jobs.dlq",
	)
	return nil
}

func (p *Publisher) publish(ctx context.Context, job TranscodeJob) error {
	body, err := json.Marshal(job)
	if err != nil {
		return fmt.Errorf("json marshal: %w", err)
	}

	err = p.ch.PublishWithContext(ctx,
		"transcoder", // exchange
		"transcode",  // routing key
		false,        // mandatory
		false,        // immediate
		amqp.Publishing{
			DeliveryMode: amqp.Persistent,
			ContentType:  "application/json",
			Body:         body,
		})
	return err
}

func (p *Publisher) PublishTranscodeJob(ctx context.Context, videoID, rawBlobURL, userID string) error {
	start := time.Now()
	logger := slog.With("component", "rabbitmq", "operation", "publish",
		"video_id", videoID, "queue", "transcoder.jobs")

	job := TranscodeJob{
		VideoID:    videoID,
		RawBlobURL: rawBlobURL,
		UserID:     userID,
		UploadedAt: time.Now().UTC().Format(time.RFC3339),
	}

	body, _ := json.Marshal(job)
	logger.Info("publishing transcode job", "message_size_bytes", len(body), "user_id", userID)

	err := p.publish(ctx, job)
	if err == nil {
		metrics.RabbitMQPublishedTotal.WithLabelValues("transcoder.jobs", "success").Inc()
		logger.Info("publish succeeded", "duration_ms", time.Since(start).Milliseconds())
		return nil
	}

	// If publish failed, try reconnecting once automatically
	logger.Warn("publish failed, attempting reconnect", "error", err.Error(), "duration_ms", time.Since(start).Milliseconds())
	if reconnErr := p.connect(); reconnErr != nil {
		metrics.RabbitMQPublishedTotal.WithLabelValues("transcoder.jobs", "error").Inc()
		logger.Error("reconnect failed", "original_error", err.Error(), "reconnect_error", reconnErr.Error())
		return fmt.Errorf("publish failed and reconnect failed: original=%v, reconn=%w", err, reconnErr)
	}

	logger.Info("reconnected, retrying publish")

	// Retry publish
	if retryErr := p.publish(ctx, job); retryErr != nil {
		metrics.RabbitMQPublishedTotal.WithLabelValues("transcoder.jobs", "error").Inc()
		logger.Error("publish failed after reconnect", "error", retryErr.Error(), "duration_ms", time.Since(start).Milliseconds())
		return fmt.Errorf("publish failed after reconnect: %w", retryErr)
	}

	metrics.RabbitMQPublishedTotal.WithLabelValues("transcoder.jobs", "success").Inc()
	logger.Info("publish succeeded after reconnect", "duration_ms", time.Since(start).Milliseconds())
	return nil
}

func (p *Publisher) IsHealthy() error {
	if p.ch == nil || p.ch.IsClosed() {
		return fmt.Errorf("channel is closed")
	}
	return nil
}

func (p *Publisher) Close() error {
	logger := slog.With("component", "rabbitmq", "operation", "close")

	if p.ch != nil {
		p.ch.Close()
	}
	if p.conn != nil {
		logger.Info("disconnected from RabbitMQ")
		return p.conn.Close()
	}
	return nil
}
