package queue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// Client provides lightweight helpers around AMQP connection/channel lifecycle.
type Client struct {
	url       string
	prefetch  int
	conn      *amqp.Connection
	channel   *amqp.Channel
	reconnect time.Duration
}

// Config holds initialization params.
type Config struct {
	URL       string
	Prefetch  int
	Reconnect time.Duration
}

// NewClient creates but does not connect.
func NewClient(cfg Config) *Client {
	return &Client{
		url:       cfg.URL,
		prefetch:  cfg.Prefetch,
		reconnect: cfg.Reconnect,
	}
}

// Connect establishes connection and channel.
func (c *Client) Connect() error {
	if c.conn != nil && !c.conn.IsClosed() && c.channel != nil && !c.channel.IsClosed() {
		return nil
	}
	// Try to close old channel/conn if half-open
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil && !c.conn.IsClosed() {
		_ = c.conn.Close()
	}
	conn, err := amqp.Dial(c.url)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}
	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return fmt.Errorf("amqp channel: %w", err)
	}
	if c.prefetch > 0 {
		if err := ch.Qos(c.prefetch, 0, false); err != nil {
			_ = ch.Close()
			_ = conn.Close()
			return fmt.Errorf("amqp qos: %w", err)
		}
	}
	c.conn = conn
	c.channel = ch
	return nil
}

// Close closes channel and connection.
func (c *Client) Close() error {
	if c.channel != nil {
		_ = c.channel.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}

// Channel returns a live channel, reconnecting if needed.
func (c *Client) Channel() (*amqp.Channel, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}
	return c.channel, nil
}

// OpenChannel creates a fresh channel (callers must close).
func (c *Client) OpenChannel() (*amqp.Channel, error) {
	if err := c.Connect(); err != nil {
		return nil, err
	}
	return c.conn.Channel()
}

// Publish sends a message to the requests exchange with routing key.
func (c *Client) Publish(ctx context.Context, routingKey string, body []byte, headers amqp.Table) error {
	if c.channel == nil {
		return fmt.Errorf("channel is nil (not connected)")
	}
	return c.channel.PublishWithContext(ctx, ExchangeRequests, routingKey, false, false, amqp.Publishing{
		ContentType:   "application/json",
		Body:          body,
		DeliveryMode:  amqp.Persistent,
		Timestamp:     time.Now(),
		MessageId:     fmt.Sprintf("msg-%d", time.Now().UnixNano()),
		Headers:       headers,
		CorrelationId: "",
	})
}

// DeclareQueues ensures topology for a single operation/cashbox.
func (c *Client) DeclareQueues(qs QueueSet, backoff time.Duration, declareOnPublish bool) error {
	if err := c.Connect(); err != nil {
		return err
	}
	return DeclareTopology(c.channel, qs, backoff, declareOnPublish)
}
