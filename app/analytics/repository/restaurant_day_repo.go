package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/quickbite/analytics-service/app/analytics"
	"github.com/quickbite/analytics-service/app/analytics/entity"
)

type RestaurantDayRepo struct {
	coll *mongo.Collection
}

func NewRestaurantDayRepo(db *mongo.Database) *RestaurantDayRepo {
	return &RestaurantDayRepo{coll: db.Collection(analytics.CollectionAggRestaurantDay)}
}

func (r *RestaurantDayRepo) Upsert(ctx context.Context, restaurantID int, date, currency string, revenue int64) error {
	filter := bson.M{"restaurant_id": restaurantID, "date": date}
	update := bson.M{
		"$inc": bson.M{
			"orders_count": 1,
			"revenue_sum":  revenue,
		},
		"$set": bson.M{
			"currency":   currency,
			"updated_at": time.Now().UTC(),
		},
		"$setOnInsert": bson.M{
			"restaurant_id":    restaurantID,
			"date":             date,
			"delivery_ms_sum":  int64(0),
			"delivery_ms_count": 0,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *RestaurantDayRepo) FindByRestaurantAndRange(
	ctx context.Context,
	restaurantID int,
	from, to string,
) ([]entity.RestaurantDay, error) {
	filter := bson.M{
		"restaurant_id": restaurantID,
		"date":          bson.M{"$gte": from, "$lte": to},
	}

	cursor, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []entity.RestaurantDay
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
