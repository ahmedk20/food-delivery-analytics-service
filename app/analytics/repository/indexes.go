package repository

import (
	"context"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/quickbite/analytics-service/app/analytics"
)

func EnsureIndexes(ctx context.Context, db *mongo.Database, log *slog.Logger) error {
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	aggColl := db.Collection(analytics.CollectionAggRestaurantDay)
	aggIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "restaurant_id", Value: 1}, {Key: "date", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uq_restaurant_date"),
		},
		{
			Keys:    bson.D{{Key: "date", Value: 1}, {Key: "restaurant_id", Value: 1}},
			Options: options.Index().SetName("idx_date_restaurant"),
		},
	}

	if _, err := aggColl.Indexes().CreateMany(ctx, aggIndexes); err != nil {
		return err
	}
	log.Info("indexes ensured", "collection", analytics.CollectionAggRestaurantDay)

	eventColl := db.Collection(analytics.CollectionEventIDs)
	ttlSeconds := int32(7 * 24 * 60 * 60)
	eventIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "event_id", Value: 1}},
			Options: options.Index().SetUnique(true).SetName("uq_event_id"),
		},
		{
			Keys:    bson.D{{Key: "received_at", Value: 1}},
			Options: options.Index().SetExpireAfterSeconds(ttlSeconds).SetName("ttl_received_at"),
		},
	}

	if _, err := eventColl.Indexes().CreateMany(ctx, eventIndexes); err != nil {
		return err
	}
	log.Info("indexes ensured", "collection", analytics.CollectionEventIDs)

	return nil
}
