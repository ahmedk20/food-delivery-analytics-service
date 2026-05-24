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

type BranchDayRepo struct {
	coll *mongo.Collection
}

func NewBranchDayRepo(db *mongo.Database) *BranchDayRepo {
	return &BranchDayRepo{coll: db.Collection(analytics.CollectionAggBranchDay)}
}

func (r *BranchDayRepo) Upsert(ctx context.Context, branchID int, date, currency string, revenue int64) error {
	filter := bson.M{"branch_id": branchID, "date": date}
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
			"branch_id":         branchID,
			"date":              date,
			"delivery_ms_sum":   int64(0),
			"delivery_ms_count": 0,
		},
	}

	opts := options.Update().SetUpsert(true)
	_, err := r.coll.UpdateOne(ctx, filter, update, opts)
	return err
}

func (r *BranchDayRepo) IncrDelivery(ctx context.Context, branchID int, date string, deliveryMs int64) error {
	filter := bson.M{"branch_id": branchID, "date": date}
	update := bson.M{
		"$inc": bson.M{
			"delivery_ms_sum":   deliveryMs,
			"delivery_ms_count": 1,
		},
		"$set": bson.M{"updated_at": time.Now().UTC()},
	}
	_, err := r.coll.UpdateOne(ctx, filter, update)
	return err
}

func (r *BranchDayRepo) FindByBranchAndRange(
	ctx context.Context,
	branchID int,
	from, to string,
) ([]entity.BranchDay, error) {
	filter := bson.M{
		"branch_id": branchID,
		"date":      bson.M{"$gte": from, "$lte": to},
	}

	cursor, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date", Value: 1}}))
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var results []entity.BranchDay
	if err := cursor.All(ctx, &results); err != nil {
		return nil, err
	}
	return results, nil
}
