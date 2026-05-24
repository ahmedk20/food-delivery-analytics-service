package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/quickbite/analytics-service/app/analytics"
	"github.com/quickbite/analytics-service/app/analytics/repository"
)

type AnalyticsService struct {
	restaurantDayRepo *repository.RestaurantDayRepo
	branchDayRepo     *repository.BranchDayRepo
	log               *slog.Logger
}

func NewAnalyticsService(restaurantDayRepo *repository.RestaurantDayRepo, branchDayRepo *repository.BranchDayRepo, log *slog.Logger) *AnalyticsService {
	return &AnalyticsService{restaurantDayRepo: restaurantDayRepo, branchDayRepo: branchDayRepo, log: log}
}

func (s *AnalyticsService) HandleOrderPlaced(ctx context.Context, input analytics.OnOrderPlacedInput) error {
	date := input.PlacedAt
	if date == "" {
		date = time.Now().UTC().Format("2006-01-02")
	} else {
		t, err := time.Parse(time.RFC3339, date)
		if err != nil {
			date = time.Now().UTC().Format("2006-01-02")
		} else {
			date = t.UTC().Format("2006-01-02")
		}
	}

	currency := input.Currency
	if currency == "" {
		currency = "EGP"
	}

	if err := s.restaurantDayRepo.Upsert(ctx, input.RestaurantID, date, currency, int64(input.TotalAmount)); err != nil {
		return err
	}
	return s.branchDayRepo.Upsert(ctx, input.BranchID, date, currency, int64(input.TotalAmount))
}

func (s *AnalyticsService) GetRestaurantDays(
	ctx context.Context,
	restaurantID int,
	dateRange analytics.DateRange,
) ([]analytics.RestaurantDayResponse, error) {
	if dateRange.From > dateRange.To {
		return nil, analytics.ErrInvalidDateRange
	}

	rows, err := s.restaurantDayRepo.FindByRestaurantAndRange(ctx, restaurantID, dateRange.From, dateRange.To)
	if err != nil {
		return nil, err
	}

	result := make([]analytics.RestaurantDayResponse, len(rows))
	for i, row := range rows {
		var avgOrder int64
		if row.OrdersCount > 0 {
			avgOrder = row.RevenueSum / int64(row.OrdersCount)
		}
		result[i] = analytics.RestaurantDayResponse{
			Date:          row.Date,
			OrdersCount:   row.OrdersCount,
			RevenueMinor:  row.RevenueSum,
			Currency:      row.Currency,
			AvgOrderMinor: avgOrder,
		}
	}

	return result, nil
}
