// Package broker is a thin RabbitMQ helper for publishing and consuming events
// over a durable topic exchange.
package broker

import (
	"context"
	"encoding/json"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"

	"swiftmind/pkg/events"
)

// Broker holds a connection and channel to RabbitMQ.
type Broker struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// Connect dials RabbitMQ (retrying for up to ~30s while it boots), opens a
// channel, and declares the shared topic exchange.
func Connect(url string) (*Broker, error) {
	var (
		conn *amqp.Connection
		err  error
	)
	// RabbitMQ's Erlang VM can take a while to accept connections on a cold
	// `docker compose up`, so retry generously before giving up.
	deadline := time.Now().Add(60 * time.Second)
	for {
		conn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		if time.Now().After(deadline) {
			return nil, err
		}
		time.Sleep(2 * time.Second)
	}

	ch, err := conn.Channel()
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	if err := ch.ExchangeDeclare(events.Exchange, "topic", true, false, false, false, nil); err != nil {
		_ = conn.Close()
		return nil, err
	}
	return &Broker{conn: conn, ch: ch}, nil
}

// Publish marshals payload to JSON and publishes it with the given routing key.
func (b *Broker) Publish(ctx context.Context, routingKey string, payload any) error {
	body, err := json.Marshal(payload)
	if err != nil {
		return err
	}
	return b.ch.PublishWithContext(ctx, events.Exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		Body:         body,
		DeliveryMode: amqp.Persistent,
		Timestamp:    time.Now(),
	})
}

// Consume binds a durable queue to the exchange for routingKey and invokes
// handler for each message. Messages are acked on success, nacked (requeued) on
// error. It runs in a background goroutine and returns immediately.
func (b *Broker) Consume(queue, routingKey string, handler func([]byte) error) error {
	q, err := b.ch.QueueDeclare(queue, true, false, false, false, nil)
	if err != nil {
		return err
	}
	if err := b.ch.QueueBind(q.Name, routingKey, events.Exchange, false, nil); err != nil {
		return err
	}
	deliveries, err := b.ch.Consume(q.Name, "", false, false, false, false, nil)
	if err != nil {
		return err
	}
	go func() {
		for d := range deliveries {
			if err := handler(d.Body); err != nil {
				_ = d.Nack(false, true)
				continue
			}
			_ = d.Ack(false)
		}
	}()
	return nil
}

// Close tears down the channel and connection.
func (b *Broker) Close() {
	if b.ch != nil {
		_ = b.ch.Close()
	}
	if b.conn != nil {
		_ = b.conn.Close()
	}
}
