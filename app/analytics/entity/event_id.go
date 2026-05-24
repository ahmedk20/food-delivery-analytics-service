package entity

import "time"

type EventID struct {
	EventID    string    `bson:"event_id"`
	ReceivedAt time.Time `bson:"received_at"`
}
