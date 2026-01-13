package queue

import (
	"context"
	"fmt"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

// DLQRequeuer moves messages from DLQ back to primary queue with age/limit checks.
type DLQRequeuer struct {
	Client   *Client
	Backoffs []time.Duration
}

// Requeue moves up to batch messages older than minAge from DLQ to main queue.
func (r *DLQRequeuer) Requeue(ctx context.Context, qs QueueSet, minAge time.Duration, batch int) (int, error) {
	if r.Client == nil {
		return 0, fmt.Errorf("client is nil")
	}
	if batch <= 0 {
		return 0, fmt.Errorf("batch must be > 0")
	}

	ch, err := r.Client.OpenChannel()
	if err != nil {
		return 0, err
	}
	defer func() {
		_ = ch.Close()
	}()

	// Ensure DLQ exists
	if _, err := ch.QueueDeclarePassive(qs.DLQ, true, false, false, false, nil); err != nil {
		return 0, fmt.Errorf("inspect dlq %s: %w", qs.DLQ, err)
	}

	requeued := 0
	for requeued < batch {
		msg, ok, err := ch.Get(qs.DLQ, false)
		if err != nil {
			return requeued, fmt.Errorf("get from dlq: %w", err)
		}
		if !ok {
			return requeued, nil
		}

		// Age check
		if minAge > 0 && time.Since(msg.Timestamp) < minAge {
			_ = msg.Nack(false, true) // requeue back to DLQ
			break
		}

		// Send back to main exchange with routing key
		if err := ch.PublishWithContext(ctx, ExchangeRequests, qs.RoutingKey, false, false, amqp.Publishing{
			ContentType:  msg.ContentType,
			Body:         msg.Body,
			DeliveryMode: amqp.Persistent,
			Headers:      msg.Headers,
			Timestamp:    time.Now(),
			MessageId:    msg.MessageId,
		}); err != nil {
			_ = msg.Nack(false, true)
			return requeued, fmt.Errorf("publish requeue: %w", err)
		}

		_ = msg.Ack(false)
		requeued++
	}

	return requeued, nil
}
