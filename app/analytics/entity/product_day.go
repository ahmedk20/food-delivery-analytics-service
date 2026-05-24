package entity

import "time"

type ProductDay struct {
	ProductID    int       `bson:"product_id"`
	RestaurantID int       `bson:"restaurant_id"`
	Date         string    `bson:"date"`
	OrdersCount  int       `bson:"orders_count"`
	QuantitySold int       `bson:"quantity_sold"`
	UpdatedAt    time.Time `bson:"updated_at"`
}
