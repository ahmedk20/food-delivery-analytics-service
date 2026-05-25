package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"

	"github.com/quickbite/analytics-service/app/analytics"
	"github.com/quickbite/analytics-service/app/analytics/eventhandlers"
	"github.com/quickbite/analytics-service/lib/coreevents"
	"github.com/quickbite/analytics-service/pkg/messaging"
)

// TestOrderPlaced_DirectService verifies that HandleOrderPlaced upserts all
// aggregate collections (restaurant, branch, platform, product) and that the
// REST API returns the correct aggregated values.
func TestOrderPlaced_DirectService(t *testing.T) {
	ctx := context.Background()
	restaurantID := 10001
	branchID := 20001
	date := "2024-06-01"

	err := svc.HandleOrderPlaced(ctx, analytics.OnOrderPlacedInput{
		OrderID:      1,
		RestaurantID: restaurantID,
		BranchID:     branchID,
		TotalAmount:  5000,
		Currency:     "EGP",
		PlacedAt:     date + "T10:00:00Z",
		Items: []analytics.OrderItem{
			{ProductID: 101, Quantity: 2},
			{ProductID: 102, Quantity: 1},
		},
	})
	if err != nil {
		t.Fatalf("HandleOrderPlaced: %v", err)
	}

	// Check agg_restaurant_day
	var restDay struct {
		OrdersCount int   `bson:"orders_count"`
		RevenueSum  int64 `bson:"revenue_sum"`
		Currency    string `bson:"currency"`
	}
	if err := testDB.Collection(analytics.CollectionAggRestaurantDay).
		FindOne(ctx, bson.M{"restaurant_id": restaurantID, "date": date}).
		Decode(&restDay); err != nil {
		t.Fatalf("find restaurant day: %v", err)
	}
	if restDay.OrdersCount != 1 {
		t.Errorf("restaurant orders_count = %d, want 1", restDay.OrdersCount)
	}
	if restDay.RevenueSum != 5000 {
		t.Errorf("restaurant revenue_sum = %d, want 5000", restDay.RevenueSum)
	}
	if restDay.Currency != "EGP" {
		t.Errorf("restaurant currency = %q, want EGP", restDay.Currency)
	}

	// Check agg_branch_day
	var branchDay struct {
		OrdersCount int `bson:"orders_count"`
	}
	if err := testDB.Collection(analytics.CollectionAggBranchDay).
		FindOne(ctx, bson.M{"branch_id": branchID, "date": date}).
		Decode(&branchDay); err != nil {
		t.Fatalf("find branch day: %v", err)
	}
	if branchDay.OrdersCount != 1 {
		t.Errorf("branch orders_count = %d, want 1", branchDay.OrdersCount)
	}

	// Check agg_product_day for product 101 (quantity 2)
	var prodDay struct {
		OrdersCount  int `bson:"orders_count"`
		QuantitySold int `bson:"quantity_sold"`
	}
	if err := testDB.Collection(analytics.CollectionAggProductDay).
		FindOne(ctx, bson.M{"product_id": 101, "restaurant_id": restaurantID, "date": date}).
		Decode(&prodDay); err != nil {
		t.Fatalf("find product day: %v", err)
	}
	if prodDay.OrdersCount != 1 {
		t.Errorf("product 101 orders_count = %d, want 1", prodDay.OrdersCount)
	}
	if prodDay.QuantitySold != 2 {
		t.Errorf("product 101 quantity_sold = %d, want 2", prodDay.QuantitySold)
	}

	// Query REST API: GET /restaurants/{id}/days
	token := mintJWT("system_admin")
	url := fmt.Sprintf("%s/api/v1/analytics/restaurants/%d/days?from=%s&to=%s",
		apiServer.URL, restaurantID, date, date)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("API request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("API status = %d, want 200", resp.StatusCode)
	}

	var envelope struct {
		Success bool `json:"success"`
		Data    []struct {
			Date          string `json:"date"`
			OrdersCount   int    `json:"ordersCount"`
			RevenueMinor  int64  `json:"revenueMinor"`
			Currency      string `json:"currency"`
			AvgOrderMinor int64  `json:"avgOrderMinor"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode API response: %v", err)
	}
	if !envelope.Success {
		t.Fatal("API success = false")
	}
	if len(envelope.Data) != 1 {
		t.Fatalf("API data length = %d, want 1", len(envelope.Data))
	}
	row := envelope.Data[0]
	if row.Date != date {
		t.Errorf("date = %q, want %q", row.Date, date)
	}
	if row.OrdersCount != 1 {
		t.Errorf("ordersCount = %d, want 1", row.OrdersCount)
	}
	if row.RevenueMinor != 5000 {
		t.Errorf("revenueMinor = %d, want 5000", row.RevenueMinor)
	}
	if row.AvgOrderMinor != 5000 {
		t.Errorf("avgOrderMinor = %d, want 5000", row.AvgOrderMinor)
	}
}

// TestOrderDelivered_UpdatesDeliveryStats verifies that HandleOrderDelivered
// increments delivery_ms_sum and delivery_ms_count on the restaurant day.
func TestOrderDelivered_UpdatesDeliveryStats(t *testing.T) {
	ctx := context.Background()
	restaurantID := 10002
	branchID := 20002
	date := "2024-06-02"

	// Seed a restaurant day first
	if err := svc.HandleOrderPlaced(ctx, analytics.OnOrderPlacedInput{
		OrderID:      2,
		RestaurantID: restaurantID,
		BranchID:     branchID,
		TotalAmount:  2000,
		Currency:     "EGP",
		PlacedAt:     date + "T09:00:00Z",
	}); err != nil {
		t.Fatalf("HandleOrderPlaced: %v", err)
	}

	if err := svc.HandleOrderDelivered(ctx, analytics.OnOrderDeliveredInput{
		OrderID:            2,
		RestaurantID:       restaurantID,
		BranchID:           branchID,
		DeliveryDurationMs: 1800000, // 30 minutes
		DeliveredAt:        date + "T09:30:00Z",
	}); err != nil {
		t.Fatalf("HandleOrderDelivered: %v", err)
	}

	var restDay struct {
		DeliveryMsSum   int64 `bson:"delivery_ms_sum"`
		DeliveryMsCount int   `bson:"delivery_ms_count"`
	}
	if err := testDB.Collection(analytics.CollectionAggRestaurantDay).
		FindOne(ctx, bson.M{"restaurant_id": restaurantID, "date": date}).
		Decode(&restDay); err != nil {
		t.Fatalf("find restaurant day: %v", err)
	}
	if restDay.DeliveryMsSum != 1800000 {
		t.Errorf("delivery_ms_sum = %d, want 1800000", restDay.DeliveryMsSum)
	}
	if restDay.DeliveryMsCount != 1 {
		t.Errorf("delivery_ms_count = %d, want 1", restDay.DeliveryMsCount)
	}
}

// TestPaymentSucceeded_UpdatesRevenue verifies that HandlePaymentSucceeded
// adds the confirmed revenue on top of the placed-order revenue.
func TestPaymentSucceeded_UpdatesRevenue(t *testing.T) {
	ctx := context.Background()
	restaurantID := 10003
	branchID := 20003
	date := "2024-06-03"

	if err := svc.HandleOrderPlaced(ctx, analytics.OnOrderPlacedInput{
		OrderID:      3,
		RestaurantID: restaurantID,
		BranchID:     branchID,
		TotalAmount:  3000,
		Currency:     "EGP",
		PlacedAt:     date + "T11:00:00Z",
	}); err != nil {
		t.Fatalf("HandleOrderPlaced: %v", err)
	}

	if err := svc.HandlePaymentSucceeded(ctx, analytics.OnPaymentSucceededInput{
		OrderID:      3,
		RestaurantID: restaurantID,
		BranchID:     branchID,
		Amount:       3000,
		Currency:     "EGP",
		PaidAt:       date + "T11:05:00Z",
	}); err != nil {
		t.Fatalf("HandlePaymentSucceeded: %v", err)
	}

	var restDay struct {
		RevenueSum int64 `bson:"revenue_sum"`
	}
	if err := testDB.Collection(analytics.CollectionAggRestaurantDay).
		FindOne(ctx, bson.M{"restaurant_id": restaurantID, "date": date}).
		Decode(&restDay); err != nil {
		t.Fatalf("find restaurant day: %v", err)
	}
	// Initial placed revenue 3000 + payment confirmation 3000 = 6000
	if restDay.RevenueSum != 6000 {
		t.Errorf("revenue_sum = %d, want 6000", restDay.RevenueSum)
	}
}

// TestOrderCancelled_DecrementsCount verifies that HandleOrderCancelled
// decrements orders_count by one.
func TestOrderCancelled_DecrementsCount(t *testing.T) {
	ctx := context.Background()
	restaurantID := 10004
	branchID := 20004
	date := "2024-06-04"

	// Place two orders first
	for i := range 2 {
		if err := svc.HandleOrderPlaced(ctx, analytics.OnOrderPlacedInput{
			OrderID:      4 + i,
			RestaurantID: restaurantID,
			BranchID:     branchID,
			TotalAmount:  1000,
			Currency:     "EGP",
			PlacedAt:     date + "T12:00:00Z",
		}); err != nil {
			t.Fatalf("HandleOrderPlaced %d: %v", i, err)
		}
	}

	// Cancel one
	if err := svc.HandleOrderCancelled(ctx, analytics.OnOrderCancelledInput{
		OrderID:      4,
		RestaurantID: restaurantID,
		BranchID:     branchID,
		CancelledAt:  date + "T12:10:00Z",
	}); err != nil {
		t.Fatalf("HandleOrderCancelled: %v", err)
	}

	var restDay struct {
		OrdersCount int `bson:"orders_count"`
	}
	if err := testDB.Collection(analytics.CollectionAggRestaurantDay).
		FindOne(ctx, bson.M{"restaurant_id": restaurantID, "date": date}).
		Decode(&restDay); err != nil {
		t.Fatalf("find restaurant day: %v", err)
	}
	if restDay.OrdersCount != 1 {
		t.Errorf("orders_count = %d, want 1 after cancellation", restDay.OrdersCount)
	}
}

// TestOrderPlaced_ViaRabbitMQ publishes an order.placed event to the RabbitMQ
// testcontainer and polls until the aggregate appears in MongoDB, proving the
// full consumer pipeline works end-to-end.
func TestOrderPlaced_ViaRabbitMQ(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	restaurantID := 10005
	branchID := 20005
	date := time.Now().UTC().Format("2006-01-02")
	eventID := uuid.NewString()
	exchange := "test.order.events." + eventID[:8]
	queue := "test-queue-" + eventID[:8]

	// Wire up a fresh broker + consumer for this test
	broker := messaging.NewAMQPBroker(testRabbitURI, slog.Default())
	if err := broker.Connect(ctx); err != nil {
		t.Fatalf("broker connect: %v", err)
	}
	defer broker.Close()

	consumer := coreevents.NewConsumer(broker, eventIDsRepo, slog.Default())
	eventhandlers.Register(consumer, svc)

	opts := messaging.ConsumerOptions{
		Exchange:    exchange,
		Queue:       queue,
		BindingKeys: []string{"order.#"},
		Prefetch:    1,
	}
	if err := consumer.Start(ctx, opts); err != nil {
		t.Fatalf("consumer start: %v", err)
	}

	// Publish the event
	payload := mustJSON(analytics.OnOrderPlacedInput{
		OrderID:      99,
		RestaurantID: restaurantID,
		BranchID:     branchID,
		TotalAmount:  7500,
		Currency:     "EGP",
		PlacedAt:     date + "T14:00:00Z",
	})
	envelope := coreevents.Envelope{
		EventID:   eventID,
		EventType: "order.placed",
		Payload:   payload,
	}
	if err := broker.Publish(ctx, exchange, "order.placed", mustJSON(envelope)); err != nil {
		t.Fatalf("publish: %v", err)
	}

	// Poll until the aggregate document appears (max 15 s)
	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		var doc struct {
			OrdersCount int `bson:"orders_count"`
		}
		err := testDB.Collection(analytics.CollectionAggRestaurantDay).
			FindOne(ctx, bson.M{"restaurant_id": restaurantID, "date": date}).
			Decode(&doc)
		if err == nil && doc.OrdersCount == 1 {
			return
		}
		time.Sleep(200 * time.Millisecond)
	}
	t.Fatal("aggregate not updated within 15 seconds after RabbitMQ publish")
}

// TestGetPlatformDays verifies the platform days API endpoint.
func TestGetPlatformDays(t *testing.T) {
	ctx := context.Background()
	date := "2099-01-01" // far-future date to avoid conflicts

	if err := svc.HandleOrderPlaced(ctx, analytics.OnOrderPlacedInput{
		OrderID:      100,
		RestaurantID: 10006,
		BranchID:     20006,
		TotalAmount:  8000,
		Currency:     "EGP",
		PlacedAt:     date + "T08:00:00Z",
	}); err != nil {
		t.Fatalf("HandleOrderPlaced: %v", err)
	}

	token := mintJWT("system_admin")
	url := fmt.Sprintf("%s/api/v1/analytics/platform/days?from=%s&to=%s", apiServer.URL, date, date)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var envelope struct {
		Success bool `json:"success"`
		Data    []struct {
			OrdersCount int `json:"ordersCount"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success {
		t.Fatal("success = false")
	}
	if len(envelope.Data) == 0 {
		t.Fatal("no platform day rows returned")
	}
	if envelope.Data[0].OrdersCount < 1 {
		t.Errorf("ordersCount = %d, want >= 1", envelope.Data[0].OrdersCount)
	}
}

// TestGetProductDays verifies the product days API endpoint.
func TestGetProductDays(t *testing.T) {
	ctx := context.Background()
	restaurantID := 10007
	date := "2024-06-07"

	if err := svc.HandleOrderPlaced(ctx, analytics.OnOrderPlacedInput{
		OrderID:      101,
		RestaurantID: restaurantID,
		BranchID:     20007,
		TotalAmount:  1500,
		Currency:     "EGP",
		PlacedAt:     date + "T09:00:00Z",
		Items: []analytics.OrderItem{
			{ProductID: 201, Quantity: 3},
		},
	}); err != nil {
		t.Fatalf("HandleOrderPlaced: %v", err)
	}

	token := mintJWT("system_admin")
	url := fmt.Sprintf("%s/api/v1/analytics/restaurants/%d/products/days?from=%s&to=%s",
		apiServer.URL, restaurantID, date, date)
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	req.Header.Set("Authorization", "Bearer "+token)

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("status = %d, want 200", resp.StatusCode)
	}

	var envelope struct {
		Success bool `json:"success"`
		Data    []struct {
			ProductID    int `json:"productId"`
			QuantitySold int `json:"quantitySold"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&envelope); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if !envelope.Success || len(envelope.Data) == 0 {
		t.Fatal("expected at least one product day row")
	}
	if envelope.Data[0].ProductID != 201 {
		t.Errorf("productId = %d, want 201", envelope.Data[0].ProductID)
	}
	if envelope.Data[0].QuantitySold != 3 {
		t.Errorf("quantitySold = %d, want 3", envelope.Data[0].QuantitySold)
	}
}
