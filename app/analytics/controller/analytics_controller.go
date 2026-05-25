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

func (c *AnalyticsController) GetBranchDays(w http.ResponseWriter, r *http.Request) error {
	req, err := dto.ParseGetBranchDaysRequest(r)
	if err != nil {
		return err
	}

	rows, err := c.svc.GetBranchDays(r.Context(), req.BranchID, analytics.DateRange{
		From: req.From,
		To:   req.To,
	})
	if err != nil {
		return err
	}

	apphttp.SendSuccess(w, http.StatusOK, dto.ToBranchDayDTOs(rows))
	return nil
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

func (c *AnalyticsController) GetProductDays(w http.ResponseWriter, r *http.Request) error {
	req, err := dto.ParseGetProductDaysRequest(r)
	if err != nil {
		return err
	}

	rows, err := c.svc.GetProductDays(r.Context(), req.RestaurantID, analytics.DateRange{
		From: req.From,
		To:   req.To,
	})
	if err != nil {
		return err
	}

	apphttp.SendSuccess(w, http.StatusOK, dto.ToProductDayDTOs(rows))
	return nil
}

func (c *AnalyticsController) GetPlatformDays(w http.ResponseWriter, r *http.Request) error {
	req, err := dto.ParseGetPlatformDaysRequest(r)
	if err != nil {
		return err
	}

	rows, err := c.svc.GetPlatformDays(r.Context(), analytics.DateRange{
		From: req.From,
		To:   req.To,
	})
	if err != nil {
		return err
	}

	apphttp.SendSuccess(w, http.StatusOK, dto.ToPlatformDayDTOs(rows))
	return nil
}
