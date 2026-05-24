package dto

import "github.com/quickbite/analytics-service/app/analytics"

type RestaurantDayDTO struct {
	Date          string `json:"date"`
	OrdersCount   int    `json:"ordersCount"`
	RevenueMinor  int64  `json:"revenueMinor"`
	Currency      string `json:"currency"`
	AvgOrderMinor int64  `json:"avgOrderMinor"`
}

func ToRestaurantDayDTOs(rows []analytics.RestaurantDayResponse) []RestaurantDayDTO {
	dtos := make([]RestaurantDayDTO, len(rows))
	for i, row := range rows {
		dtos[i] = RestaurantDayDTO{
			Date:          row.Date,
			OrdersCount:   row.OrdersCount,
			RevenueMinor:  row.RevenueMinor,
			Currency:      row.Currency,
			AvgOrderMinor: row.AvgOrderMinor,
		}
	}
	return dtos
}

type ProductDayDTO struct {
	ProductID    int    `json:"productId"`
	Date         string `json:"date"`
	OrdersCount  int    `json:"ordersCount"`
	QuantitySold int    `json:"quantitySold"`
}

func ToProductDayDTOs(rows []analytics.ProductDayResponse) []ProductDayDTO {
	dtos := make([]ProductDayDTO, len(rows))
	for i, row := range rows {
		dtos[i] = ProductDayDTO{
			ProductID:    row.ProductID,
			Date:         row.Date,
			OrdersCount:  row.OrdersCount,
			QuantitySold: row.QuantitySold,
		}
	}
	return dtos
}

type PlatformDayDTO struct {
	Date          string `json:"date"`
	OrdersCount   int    `json:"ordersCount"`
	RevenueMinor  int64  `json:"revenueMinor"`
	Currency      string `json:"currency"`
	AvgOrderMinor int64  `json:"avgOrderMinor"`
}

func ToPlatformDayDTOs(rows []analytics.PlatformDayResponse) []PlatformDayDTO {
	dtos := make([]PlatformDayDTO, len(rows))
	for i, row := range rows {
		dtos[i] = PlatformDayDTO{
			Date:          row.Date,
			OrdersCount:   row.OrdersCount,
			RevenueMinor:  row.RevenueMinor,
			Currency:      row.Currency,
			AvgOrderMinor: row.AvgOrderMinor,
		}
	}
	return dtos
}
