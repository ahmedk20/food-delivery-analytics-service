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

type PlatformDayRepo struct {
	coll *mongo.Collection
}

func NewPlatformDayRepo(db *mongo.Database) *PlatformDayRepo {
	return &PlatformDayRepo{coll: db.Collection(analytics.CollectionAggPlatformDay)}
}

func (r *PlatformDayRepo) Upsert(ctx context.Context, date, currency string, revenue int64) error {
	filter := bson.M{"date": date, "currency": currency}
	update := bson.M{
		"$inc": bson.M{
			"orders_count": 1,
			"revenue_sum":  revenue,
		},
		"$set": bson.M{
			"updated_at": time.Now().UTC(),
		},
		"$setOnInsert": bson.M{
			"date":             date,
			"currency":         currency,
			"delivery_ms_sum":  int64(0),
			"delivery_ms_count": 0,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *PlatformDayRepo) IncrRevenue(ctx context.Context, date string, amount int64) error {
	filter := bson.M{"date": date}
	update := bson.M{
		"$inc": bson.M{"revenue_sum": amount},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}
	_, err := r.coll.UpdateOne(ctx, filter, update)
	return err
}

func (r *PlatformDayRepo) DecrOrder(ctx context.Context, date string) error {
	filter := bson.M{"date": date}
	update := bson.M{
		"$inc": bson.M{"orders_count": -1},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}
	_, err := r.coll.UpdateOne(ctx, filter, update)
	return err
}

func (r *PlatformDayRepo) FindByRange(
	ctx context.Context,
	from, to string,
) ([]entity.PlatformDay, error) {
	filter := bson.M{
		"date": bson.M{"$gte": from, "$lte": to},
	}

	cursor, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []entity.PlatformDay
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
