package entity

import "time"

type BranchDay struct {
	BranchID        int       `bson:"branch_id"`
	Date            string    `bson:"date"`
	Currency        string    `bson:"currency"`
	OrdersCount     int       `bson:"orders_count"`
	RevenueSum      int64     `bson:"revenue_sum"`
	DeliveryMsSum   int64     `bson:"delivery_ms_sum"`
	DeliveryMsCount int       `bson:"delivery_ms_count"`
	UpdatedAt       time.Time `bson:"updated_at"`
}
