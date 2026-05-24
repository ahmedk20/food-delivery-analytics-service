package messaging

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type AMQPBroker struct {
	url  string
	conn *amqp.Connection
	ch   *amqp.Channel
	log  *slog.Logger
}

func NewAMQPBroker(url string, log *slog.Logger) *AMQPBroker {
	return &AMQPBroker{url: url, log: log}
}

func (b *AMQPBroker) Connect(_ context.Context) error {
	conn, err := amqp.Dial(b.url)
	if err != nil {
		return fmt.Errorf("amqp dial: %w", err)
	}
	b.conn = conn

	ch, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("amqp channel: %w", err)
	}
	b.ch = ch
	return nil
}

func (b *AMQPBroker) Close() error {
	if b.ch != nil {
		_ = b.ch.Close()
	}
	if b.conn != nil {
		return b.conn.Close()
	}
	return nil
}

func (b *AMQPBroker) Publish(ctx context.Context, exchange, routingKey string, body []byte) error {
	return b.ch.PublishWithContext(ctx, exchange, routingKey, false, false, amqp.Publishing{
		ContentType:  "application/json",
		DeliveryMode: amqp.Persistent,
		Body:         body,
	})
}

func (b *AMQPBroker) Consume(ctx context.Context, opts ConsumerOptions, handler Handler) error {
	if err := b.assertTopology(opts); err != nil {
		return fmt.Errorf("assert topology: %w", err)
	}

	if err := b.ch.Qos(opts.Prefetch, 0, false); err != nil {
		return fmt.Errorf("qos: %w", err)
	}

	deliveries, err := b.ch.Consume(opts.Queue, "", false, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("consume: %w", err)
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case d, ok := <-deliveries:
				if !ok {
					return
				}
				msg := Message{
					RoutingKey: d.RoutingKey,
					Body:       d.Body,
					Ack:        func() error { return d.Ack(false) },
					Nack:       func(requeue bool) error { return d.Nack(false, requeue) },
				}
				if err := handler(ctx, msg); err != nil {
					b.log.Error("message handler failed, nacking to DLQ", "routing_key", d.RoutingKey, "error", err)
					_ = d.Nack(false, false)
				}
			}
		}
	}()

	return nil
}

func (b *AMQPBroker) assertTopology(opts ConsumerOptions) error {
	if err := b.ch.ExchangeDeclare(opts.Exchange, "topic", true, false, false, false, nil); err != nil {
		return err
	}

	if opts.DeadLetterExchange != "" {
		if err := b.ch.ExchangeDeclare(opts.DeadLetterExchange, "topic", true, false, false, false, nil); err != nil {
			return err
		}
		if _, err := b.ch.QueueDeclare(opts.DeadLetterQueue, true, false, false, false, nil); err != nil {
			return err
		}
		if err := b.ch.QueueBind(opts.DeadLetterQueue, "#", opts.DeadLetterExchange, false, nil); err != nil {
			return err
		}
	}

	args := amqp.Table{}
	if opts.DeadLetterExchange != "" {
		args["x-dead-letter-exchange"] = opts.DeadLetterExchange
	}

	if _, err := b.ch.QueueDeclare(opts.Queue, true, false, false, false, args); err != nil {
		return err
	}

	for _, key := range opts.BindingKeys {
		if err := b.ch.QueueBind(opts.Queue, key, opts.Exchange, false, nil); err != nil {
			return err
		}
	}

	return nil
}
