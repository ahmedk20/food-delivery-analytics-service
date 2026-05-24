package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/quickbite/analytics-service/app/analytics"
)

type EventIDsRepo struct {
	coll *mongo.Collection
}

func NewEventIDsRepo(db *mongo.Database) *EventIDsRepo {
	return &EventIDsRepo{coll: db.Collection(analytics.CollectionEventIDs)}
}

// MarkSeen inserts the event_id. Returns (true, nil) if first time, (false, nil) if duplicate.
func (r *EventIDsRepo) MarkSeen(ctx context.Context, eventID string) (bool, error) {
	_, err := r.coll.InsertOne(ctx, bson.M{
		"event_id":    eventID,
		"received_at": time.Now().UTC(),
	})
	if err != nil {
		if mongo.IsDuplicateKeyError(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}
