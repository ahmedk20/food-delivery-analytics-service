package controller

import (
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/quickbite/analytics-service/app/analytics"
	apperr "github.com/quickbite/analytics-service/lib/errors"
	"github.com/quickbite/analytics-service/lib/rbac"
)

func RegisterRoutes(r chi.Router, ctrl *AnalyticsController, pc *rbac.PermissionCache, log *slog.Logger) {
	r.Route("/api/v1/analytics", func(r chi.Router) {
		r.With(rbac.Require(pc, analytics.PermAnalyticsRead)).
			Get("/restaurants/{restaurantId}/days", apperr.Wrap(log, ctrl.GetRestaurantDays))
		r.With(rbac.Require(pc, analytics.PermAnalyticsRead)).
			Get("/restaurants/{restaurantId}/products/days", apperr.Wrap(log, ctrl.GetProductDays))
		r.With(rbac.Require(pc, analytics.PermAnalyticsRead)).
			Get("/platform/days", apperr.Wrap(log, ctrl.GetPlatformDays))
	})
}
