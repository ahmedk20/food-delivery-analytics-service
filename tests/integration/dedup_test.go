package integration_test

import (
	"context"
	"testing"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/quickbite/analytics-service/app/analytics"
)

// TestDedup_FirstTimeSeen verifies that MarkSeen returns true for a new event ID.
func TestDedup_FirstTimeSeen(t *testing.T) {
	ctx := context.Background()
	firstTime, err := eventIDsRepo.MarkSeen(ctx, uuid.NewString())
	if err != nil {
		t.Fatalf("MarkSeen: %v", err)
	}
	if !firstTime {
		t.Error("expected firstTime=true for a new event ID, got false")
	}
}

// TestDedup_DuplicateSeen verifies that MarkSeen returns false when the same
// event ID is submitted twice.
func TestDedup_DuplicateSeen(t *testing.T) {
	ctx := context.Background()
	eventID := uuid.NewString()

	if _, err := eventIDsRepo.MarkSeen(ctx, eventID); err != nil {
		t.Fatalf("first MarkSeen: %v", err)
	}

	secondTime, err := eventIDsRepo.MarkSeen(ctx, eventID)
	if err != nil {
		t.Fatalf("second MarkSeen: %v", err)
	}
	if secondTime {
		t.Error("expected firstTime=false on duplicate event ID, got true")
	}
}

// TestDedup_SameEventTwice_CountStaysOne simulates the consumer dedup path:
// the same order.placed event is submitted twice but HandleOrderPlaced is
// only called for the first occurrence, so orders_count remains 1.
func TestDedup_SameEventTwice_CountStaysOne(t *testing.T) {
	ctx := context.Background()
	restaurantID := 30001
	date := "2024-07-01"
	eventID := uuid.NewString()

	input := analytics.OnOrderPlacedInput{
		OrderID:      50,
		RestaurantID: restaurantID,
		BranchID:     40001,
		TotalAmount:  1000,
		Currency:     "EGP",
		PlacedAt:     date + "T08:00:00Z",
	}

	// First occurrence: marked as new → call handler
	firstTime, err := eventIDsRepo.MarkSeen(ctx, eventID)
	if err != nil {
		t.Fatalf("MarkSeen (1st): %v", err)
	}
	if firstTime {
		if err := svc.HandleOrderPlaced(ctx, input); err != nil {
			t.Fatalf("HandleOrderPlaced (1st): %v", err)
		}
	}

	// Second occurrence: duplicate → skip handler
	firstTime, err = eventIDsRepo.MarkSeen(ctx, eventID)
	if err != nil {
		t.Fatalf("MarkSeen (2nd): %v", err)
	}
	if firstTime {
		if err := svc.HandleOrderPlaced(ctx, input); err != nil {
			t.Fatalf("HandleOrderPlaced (2nd, should not reach here): %v", err)
		}
	}

	var doc struct {
		OrdersCount int `bson:"orders_count"`
	}
	if err := testDB.Collection(analytics.CollectionAggRestaurantDay).
		FindOne(ctx, bson.M{"restaurant_id": restaurantID, "date": date}).
		Decode(&doc); err != nil {
		t.Fatalf("find restaurant day: %v", err)
	}
	if doc.OrdersCount != 1 {
		t.Errorf("orders_count = %d, want 1 (dedup must prevent double-counting)", doc.OrdersCount)
	}
}

// TestDedup_InvalidDateRange verifies that GetRestaurantDays returns 400
// when 'from' is after 'to'.
func TestDedup_InvalidDateRange(t *testing.T) {
	_, err := svc.GetRestaurantDays(context.Background(), 1, analytics.DateRange{
		From: "2024-12-31",
		To:   "2024-01-01",
	})
	if err == nil {
		t.Fatal("expected error for invalid date range, got nil")
	}
	if err != analytics.ErrInvalidDateRange {
		t.Errorf("error = %v, want ErrInvalidDateRange", err)
	}
}
