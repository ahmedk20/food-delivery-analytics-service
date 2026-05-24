package eventhandlers

import (
	"context"
	"encoding/json"

	"github.com/quickbite/analytics-service/app/analytics"
	"github.com/quickbite/analytics-service/app/analytics/service"
	"github.com/quickbite/analytics-service/lib/coreevents"
)

func Register(consumer *coreevents.Consumer, svc *service.AnalyticsService) {
	consumer.Register("order.placed", handleOrderPlaced(svc))
}

func handleOrderPlaced(svc *service.AnalyticsService) coreevents.EventHandler {
	return func(ctx context.Context, payload json.RawMessage) error {
		var input analytics.OnOrderPlacedInput
		if err := json.Unmarshal(payload, &input); err != nil {
			return err
		}
		return svc.HandleOrderPlaced(ctx, input)
	}
}
