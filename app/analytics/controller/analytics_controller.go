package controller

import (
	"net/http"

	"github.com/quickbite/analytics-service/app/analytics"
	"github.com/quickbite/analytics-service/app/analytics/dto"
	"github.com/quickbite/analytics-service/app/analytics/service"
	apphttp "github.com/quickbite/analytics-service/lib/http"
)

type AnalyticsController struct {
	svc *service.AnalyticsService
}

func NewAnalyticsController(svc *service.AnalyticsService) *AnalyticsController {
	return &AnalyticsController{svc: svc}
}

func (c *AnalyticsController) GetRestaurantDays(w http.ResponseWriter, r *http.Request) error {
	req, err := dto.ParseGetRestaurantDaysRequest(r)
	if err != nil {
		return err
	}

	rows, err := c.svc.GetRestaurantDays(r.Context(), req.RestaurantID, analytics.DateRange{
		From: req.From,
		To:   req.To,
	})
	if err != nil {
		return err
	}

	apphttp.SendSuccess(w, http.StatusOK, dto.ToRestaurantDayDTOs(rows))
	return nil
}
