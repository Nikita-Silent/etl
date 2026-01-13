package queue

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Message represents a queue payload for webhook load requests.
type Message struct {
	RequestID     string `json:"request_id"`
	Date          string `json:"date"`
	OperationType string `json:"operation_type"`
	SourceFolder  string `json:"source_folder"`
	CreatedAt     string `json:"created_at"`
}

// Handler processes a message from a queue.
type Handler func(ctx context.Context, msg Message) error

// Consumer consumes queues for one operation across cashboxes.
type Consumer struct {
	url              string
	prefetch         int
	backoffs         []time.Duration
	maxRetries       int
	declareOnPublish bool
	conn             *amqp.Connection
}

// ConsumerConfig holds parameters for creating a Consumer.
type ConsumerConfig struct {
	URL              string
	Prefetch         int
	Backoffs         []time.Duration
	MaxRetries       int
	DeclareOnPublish bool
}

// NewConsumer builds a Consumer (connection opened on Start).
func NewConsumer(cfg ConsumerConfig) *Consumer {
	return &Consumer{
		url:              cfg.URL,
		prefetch:         cfg.Prefetch,
		backoffs:         cfg.Backoffs,
		maxRetries:       cfg.MaxRetries,
		declareOnPublish: cfg.DeclareOnPublish,
	}
}

// Start begins consuming load queues for provided cashboxes.
func (c *Consumer) Start(ctx context.Context, cashboxes []string, handler Handler) error {
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}
	c.conn = conn

	for _, cashbox := range cashboxes {
		qs := BuildQueueSet("load", cashbox)
		go c.consumeOne(ctx, qs, handler)
	}
	return nil
}

// Stop closes the AMQP connection.
func (c *Consumer) Stop() error {
	if c.conn != nil && !c.conn.IsClosed() {
		return c.conn.Close()
	}
	return nil
}

func (c *Consumer) consumeOne(ctx context.Context, qs QueueSet, handler Handler) {
	for {
		if err := c.consumeLoop(ctx, qs, handler); err != nil {
			// Sleep briefly before reconnecting
			time.Sleep(2 * time.Second)
		}
		// Exit if context cancelled
		if ctx.Err() != nil {
			return
		}
	}
}

func (c *Consumer) consumeLoop(ctx context.Context, qs QueueSet, handler Handler) error {
	ch, err := c.conn.Channel()
	if err != nil {
		return err
	}
	defer ch.Close()

	if c.prefetch > 0 {
		if err := ch.Qos(c.prefetch, 0, false); err != nil {
			return err
		}
	}

	backoff := time.Minute
	if len(c.backoffs) > 0 {
		backoff = c.backoffs[0]
	}

	if err := DeclareTopology(ch, qs, backoff, c.declareOnPublish); err != nil {
		return err
	}

	deliveries, err := ch.Consume(qs.PrimaryQueue, "", false, false, false, false, nil)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-deliveries:
			if !ok {
				return fmt.Errorf("delivery channel closed for queue %s", qs.PrimaryQueue)
			}
			c.handleDelivery(ctx, ch, qs, msg, handler)
		}
	}
}

func (c *Consumer) handleDelivery(ctx context.Context, ch *amqp.Channel, qs QueueSet, d amqp.Delivery, handler Handler) {
	retryCount := int32(0)
	if val, ok := d.Headers["x-retry-count"]; ok {
		if v, ok2 := val.(int32); ok2 {
			retryCount = v
		}
	}
	firstSeen := time.Now().UTC().Format(time.RFC3339)
	if val, ok := d.Headers["x-first-seen"]; ok {
		if s, ok2 := val.(string); ok2 {
			firstSeen = s
		}
	}

	var payload Message
	if err := json.Unmarshal(d.Body, &payload); err != nil {
		_ = d.Nack(false, false)
		return
	}

	if err := handler(ctx, payload); err != nil {
		if retryCount < int32(c.maxRetries) {
			headers := amqp.Table{
				"x-retry-count": retryCount + 1,
				"x-first-seen":  firstSeen,
			}
			_ = ch.PublishWithContext(ctx, ExchangeRetry, qs.RetryRoutingKey, false, false, amqp.Publishing{
				ContentType:  "application/json",
				Body:         d.Body,
				DeliveryMode: amqp.Persistent,
				Headers:      headers,
				Timestamp:    time.Now(),
				MessageId:    d.MessageId,
			})
			_ = d.Ack(false)
			return
		}

		// Send to DLQ
		_ = ch.PublishWithContext(ctx, ExchangeDLX, qs.DLQRoutingKey, false, false, amqp.Publishing{
			ContentType:  "application/json",
			Body:         d.Body,
			DeliveryMode: amqp.Persistent,
			Headers:      d.Headers,
			Timestamp:    time.Now(),
			MessageId:    d.MessageId,
		})
		_ = d.Ack(false)
		return
	}

	_ = d.Ack(false)
}
