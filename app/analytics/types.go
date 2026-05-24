package analytics

import "time"

type OnOrderPlacedInput struct {
	OrderID      int    `json:"orderId"`
	RestaurantID int    `json:"restaurantId"`
	BranchID     int    `json:"branchId"`
	CustomerID   int    `json:"customerId"`
	TotalAmount  int    `json:"totalAmount"`
	ItemsCount   int    `json:"itemsCount"`
	Currency     string `json:"currency"`
	Region       string `json:"region"`
	PlacedAt     string `json:"placedAt"`
}

type RestaurantDayRow struct {
	RestaurantID    int       `bson:"restaurant_id" json:"restaurantId"`
	Date            string    `bson:"date" json:"date"`
	Currency        string    `bson:"currency" json:"currency"`
	OrdersCount     int       `bson:"orders_count" json:"ordersCount"`
	RevenueSum      int64     `bson:"revenue_sum" json:"revenueSum"`
	DeliveryMsSum   int64     `bson:"delivery_ms_sum" json:"deliveryMsSum"`
	DeliveryMsCount int       `bson:"delivery_ms_count" json:"deliveryMsCount"`
	UpdatedAt       time.Time `bson:"updated_at" json:"updatedAt"`
}

type RestaurantDayResponse struct {
	Date          string `json:"date"`
	OrdersCount   int    `json:"ordersCount"`
	RevenueMinor  int64  `json:"revenueMinor"`
	Currency      string `json:"currency"`
	AvgOrderMinor int64  `json:"avgOrderMinor"`
}

type BranchDayRow struct {
	BranchID        int       `bson:"branch_id" json:"branchId"`
	Date            string    `bson:"date" json:"date"`
	Currency        string    `bson:"currency" json:"currency"`
	OrdersCount     int       `bson:"orders_count" json:"ordersCount"`
	RevenueSum      int64     `bson:"revenue_sum" json:"revenueSum"`
	DeliveryMsSum   int64     `bson:"delivery_ms_sum" json:"deliveryMsSum"`
	DeliveryMsCount int       `bson:"delivery_ms_count" json:"deliveryMsCount"`
	UpdatedAt       time.Time `bson:"updated_at" json:"updatedAt"`
}

type BranchDayResponse struct {
	Date          string `json:"date"`
	OrdersCount   int    `json:"ordersCount"`
	RevenueMinor  int64  `json:"revenueMinor"`
	Currency      string `json:"currency"`
	AvgOrderMinor int64  `json:"avgOrderMinor"`
}

type DateRange struct {
	From string
	To   string
}
