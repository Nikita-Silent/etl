package queue

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	ExchangeRequests = "etl.requests"
	ExchangeRetry    = "etl.retry"
	ExchangeDLX      = "etl.dlx"
)

var (
	cashboxSanitizer = regexp.MustCompile(`[^a-z0-9_-]+`)
)

// NormalizeCashbox converts a folder name to queue-safe format.
func NormalizeCashbox(folder string) string {
	lower := strings.ToLower(folder)
	return cashboxSanitizer.ReplaceAllString(lower, "_")
}

// QueueSet holds names for primary/retry/DLQ queues and routing keys.
type QueueSet struct {
	Operation       string
	Cashbox         string
	PrimaryQueue    string
	RetryQueue      string
	DLQ             string
	RoutingKey      string
	RetryRoutingKey string
	DLQRoutingKey   string
}

// BuildQueueSet constructs queue names and routing keys per design doc.
func BuildQueueSet(operation, cashbox string) QueueSet {
	normalized := NormalizeCashbox(cashbox)
	base := fmt.Sprintf("etl.%s.%s", operation, normalized)
	return QueueSet{
		Operation:       operation,
		Cashbox:         normalized,
		PrimaryQueue:    base,
		RetryQueue:      base + ".retry",
		DLQ:             base + ".dlq",
		RoutingKey:      fmt.Sprintf("%s.%s", operation, normalized),
		RetryRoutingKey: fmt.Sprintf("%s.%s.retry", operation, normalized),
		DLQRoutingKey:   fmt.Sprintf("%s.%s.dlq", operation, normalized),
	}
}

// DeclareTopology declares exchanges and queues for a single cashbox/operation.
// Backoff controls TTL for retry queue; declareOnPublish is a hint to upstream callers.
func DeclareTopology(ch *amqp.Channel, qs QueueSet, backoff time.Duration, declareOnPublish bool) error {
	if ch == nil {
		return fmt.Errorf("channel is nil")
	}

	// Exchanges
	if err := ch.ExchangeDeclare(ExchangeRequests, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", ExchangeRequests, err)
	}
	if err := ch.ExchangeDeclare(ExchangeRetry, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", ExchangeRetry, err)
	}
	if err := ch.ExchangeDeclare(ExchangeDLX, "topic", true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare exchange %s: %w", ExchangeDLX, err)
	}

	args := amqp.Table{
		"x-dead-letter-exchange":    ExchangeDLX,
		"x-dead-letter-routing-key": qs.DLQRoutingKey,
	}

	// Primary queue
	if _, err := ch.QueueDeclare(qs.PrimaryQueue, true, false, false, false, args); err != nil {
		return fmt.Errorf("declare queue %s: %w", qs.PrimaryQueue, err)
	}
	if err := ch.QueueBind(qs.PrimaryQueue, qs.RoutingKey, ExchangeRequests, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", qs.PrimaryQueue, err)
	}

	// Retry queue with TTL and DLX back to main exchange
	retryArgs := amqp.Table{
		"x-message-ttl":             int64(backoff / time.Millisecond),
		"x-dead-letter-exchange":    ExchangeRequests,
		"x-dead-letter-routing-key": qs.RoutingKey,
	}
	if _, err := ch.QueueDeclare(qs.RetryQueue, true, false, false, false, retryArgs); err != nil {
		return fmt.Errorf("declare queue %s: %w", qs.RetryQueue, err)
	}
	if err := ch.QueueBind(qs.RetryQueue, qs.RetryRoutingKey, ExchangeRetry, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", qs.RetryQueue, err)
	}

	// DLQ
	if _, err := ch.QueueDeclare(qs.DLQ, true, false, false, false, nil); err != nil {
		return fmt.Errorf("declare queue %s: %w", qs.DLQ, err)
	}
	if err := ch.QueueBind(qs.DLQ, qs.DLQRoutingKey, ExchangeDLX, false, nil); err != nil {
		return fmt.Errorf("bind queue %s: %w", qs.DLQ, err)
	}

	// Prefetch is set by consumer creation, not here. declareOnPublish is a hint to upstream callers.
	_ = declareOnPublish
	return nil
}
