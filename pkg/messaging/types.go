package messaging

import "context"

type Message struct {
	RoutingKey string
	Body       []byte
	Ack        func() error
	Nack       func(requeue bool) error
}

type ConsumerOptions struct {
	Exchange           string
	Queue              string
	BindingKeys        []string
	Prefetch           int
	DeadLetterExchange string
	DeadLetterQueue    string
}

type Handler func(ctx context.Context, msg Message) error

type Broker interface {
	Connect(ctx context.Context) error
	Close() error
	Publish(ctx context.Context, exchange, routingKey string, body []byte) error
	Consume(ctx context.Context, opts ConsumerOptions, handler Handler) error
}
