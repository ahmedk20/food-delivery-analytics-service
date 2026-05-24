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
	platformDayRepo   *repository.PlatformDayRepo
	productDayRepo    *repository.ProductDayRepo
	log               *slog.Logger
}

func NewAnalyticsService(
	restaurantDayRepo *repository.RestaurantDayRepo,
	branchDayRepo *repository.BranchDayRepo,
	platformDayRepo *repository.PlatformDayRepo,
	productDayRepo *repository.ProductDayRepo,
	log *slog.Logger,
) *AnalyticsService {
	return &AnalyticsService{
		restaurantDayRepo: restaurantDayRepo,
		branchDayRepo:     branchDayRepo,
		platformDayRepo:   platformDayRepo,
		productDayRepo:    productDayRepo,
		log:               log,
	}
}

func parseEventDate(raw string) string {
	if raw == "" {
		return time.Now().UTC().Format("2006-01-02")
	}
	t, err := time.Parse(time.RFC3339, raw)
	if err != nil {
		return time.Now().UTC().Format("2006-01-02")
	}
	return t.UTC().Format("2006-01-02")
}

func (s *AnalyticsService) HandleOrderPlaced(ctx context.Context, input analytics.OnOrderPlacedInput) error {
	date := parseEventDate(input.PlacedAt)

	currency := input.Currency
	if currency == "" {
		currency = "EGP"
	}

	if err := s.restaurantDayRepo.Upsert(ctx, input.RestaurantID, date, currency, int64(input.TotalAmount)); err != nil {
		return err
	}
	if err := s.branchDayRepo.Upsert(ctx, input.BranchID, date, currency, int64(input.TotalAmount)); err != nil {
		return err
	}
	if err := s.platformDayRepo.Upsert(ctx, date, currency, int64(input.TotalAmount)); err != nil {
		return err
	}
	for _, item := range input.Items {
		if item.ProductID <= 0 || item.Quantity <= 0 {
			continue
		}
		if err := s.productDayRepo.Upsert(ctx, item.ProductID, input.RestaurantID, date, item.Quantity); err != nil {
			return err
		}
	}
	return nil
}

func (s *AnalyticsService) HandleOrderDelivered(ctx context.Context, input analytics.OnOrderDeliveredInput) error {
	date := parseEventDate(input.DeliveredAt)
	if err := s.restaurantDayRepo.IncrDelivery(ctx, input.RestaurantID, date, input.DeliveryDurationMs); err != nil {
		return err
	}
	return s.branchDayRepo.IncrDelivery(ctx, input.BranchID, date, input.DeliveryDurationMs)
}

func (s *AnalyticsService) HandlePaymentSucceeded(ctx context.Context, input analytics.OnPaymentSucceededInput) error {
	date := parseEventDate(input.PaidAt)
	if err := s.restaurantDayRepo.IncrRevenue(ctx, input.RestaurantID, date, input.Amount); err != nil {
		return err
	}
	if err := s.branchDayRepo.IncrRevenue(ctx, input.BranchID, date, input.Amount); err != nil {
		return err
	}
	return s.platformDayRepo.IncrRevenue(ctx, date, input.Amount)
}

func (s *AnalyticsService) HandleOrderCancelled(ctx context.Context, input analytics.OnOrderCancelledInput) error {
	date := parseEventDate(input.CancelledAt)
	if err := s.restaurantDayRepo.DecrOrder(ctx, input.RestaurantID, date); err != nil {
		return err
	}
	if err := s.branchDayRepo.DecrOrder(ctx, input.BranchID, date); err != nil {
		return err
	}
	return s.platformDayRepo.DecrOrder(ctx, date)
}

func (s *AnalyticsService) GetProductDays(
	ctx context.Context,
	restaurantID int,
	dateRange analytics.DateRange,
) ([]analytics.ProductDayResponse, error) {
	if dateRange.From > dateRange.To {
		return nil, analytics.ErrInvalidDateRange
	}

	rows, err := s.productDayRepo.FindByRestaurantAndRange(ctx, restaurantID, dateRange.From, dateRange.To)
	if err != nil {
		return nil, err
	}

	result := make([]analytics.ProductDayResponse, len(rows))
	for i, row := range rows {
		result[i] = analytics.ProductDayResponse{
			ProductID:    row.ProductID,
			Date:         row.Date,
			OrdersCount:  row.OrdersCount,
			QuantitySold: row.QuantitySold,
		}
	}
	return result, nil
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

func (s *AnalyticsService) GetPlatformDays(
	ctx context.Context,
	dateRange analytics.DateRange,
) ([]analytics.PlatformDayResponse, error) {
	if dateRange.From > dateRange.To {
		return nil, analytics.ErrInvalidDateRange
	}

	rows, err := s.platformDayRepo.FindByRange(ctx, dateRange.From, dateRange.To)
	if err != nil {
		return nil, err
	}

	result := make([]analytics.PlatformDayResponse, len(rows))
	for i, row := range rows {
		var avgOrder int64
		if row.OrdersCount > 0 {
			avgOrder = row.RevenueSum / int64(row.OrdersCount)
		}
		result[i] = analytics.PlatformDayResponse{
			Date:          row.Date,
			OrdersCount:   row.OrdersCount,
			RevenueMinor:  row.RevenueSum,
			Currency:      row.Currency,
			AvgOrderMinor: avgOrder,
		}
	}

	return result, nil
}
