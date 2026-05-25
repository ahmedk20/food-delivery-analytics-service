package coreevents

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/quickbite/analytics-service/pkg/messaging"
)

type EventDeduper interface {
	MarkSeen(ctx context.Context, eventID string) (firstTime bool, err error)
}

type EventHandler func(ctx context.Context, payload json.RawMessage) error

type Consumer struct {
	broker   messaging.Broker
	deduper  EventDeduper
	handlers map[string]EventHandler
	log      *slog.Logger
}

func NewConsumer(broker messaging.Broker, deduper EventDeduper, log *slog.Logger) *Consumer {
	return &Consumer{
		broker:   broker,
		deduper:  deduper,
		handlers: make(map[string]EventHandler),
		log:      log,
	}
}

func (c *Consumer) Register(eventType string, handler EventHandler) {
	c.handlers[eventType] = handler
}

func (c *Consumer) Start(ctx context.Context, opts messaging.ConsumerOptions) error {
	return c.broker.Consume(ctx, opts, func(ctx context.Context, msg messaging.Message) error {
		var envelope Envelope
		if err := json.Unmarshal(msg.Body, &envelope); err != nil {
			c.log.Warn("malformed event envelope, acking to skip", "routing_key", msg.RoutingKey, "error", err)
			return msg.Ack()
		}

		eventID := envelope.EventID
		if eventID == "" {
			c.log.Warn("event missing event_id, acking to skip", "routing_key", msg.RoutingKey)
			return msg.Ack()
		}

		firstTime, err := c.deduper.MarkSeen(ctx, eventID)
		if err != nil {
			c.log.Error("dedup check failed", "event_id", eventID, "error", err)
			return err
		}
		if !firstTime {
			c.log.Debug("duplicate event, acking", "event_id", eventID, "routing_key", msg.RoutingKey)
			return msg.Ack()
		}

		handler, ok := c.handlers[msg.RoutingKey]
		if !ok {
			c.log.Debug("unknown event type, acking to skip", "routing_key", msg.RoutingKey)
			return msg.Ack()
		}

		if err := handler(ctx, envelope.Payload); err != nil {
			c.log.Error("event handler failed", "routing_key", msg.RoutingKey, "event_id", eventID, "error", err)
			return err
		}

		c.log.Info("event processed", "routing_key", msg.RoutingKey, "event_id", eventID)
		return msg.Ack()
	})
}
