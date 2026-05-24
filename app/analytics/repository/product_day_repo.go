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

type ProductDayRepo struct {
	coll *mongo.Collection
}

func NewProductDayRepo(db *mongo.Database) *ProductDayRepo {
	return &ProductDayRepo{coll: db.Collection(analytics.CollectionAggProductDay)}
}

func (r *ProductDayRepo) Upsert(ctx context.Context, productID, restaurantID int, date string, quantity int) error {
	filter := bson.M{"product_id": productID, "restaurant_id": restaurantID, "date": date}
	update := bson.M{
		"$inc": bson.M{
			"orders_count":  1,
			"quantity_sold": quantity,
		},
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
		"$setOnInsert": bson.M{
			"product_id":    productID,
			"restaurant_id": restaurantID,
			"date":          date,
		},
	}
	opts := options.Update().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *ProductDayRepo) FindByRestaurantAndRange(
	ctx context.Context,
	restaurantID int,
	from, to string,
) ([]entity.ProductDay, error) {
	filter := bson.M{
		"restaurant_id": restaurantID,
		"date":          bson.M{"$gte": from, "$lte": to},
	}
	sort := bson.D{{Key: "date", Value: 1}, {Key: "product_id", Value: 1}}

	cursor, err := r.coll.Find(ctx, filter, options.Find().SetSort(sort))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []entity.ProductDay
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
