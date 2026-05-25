package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/golang-jwt/jwt/v5"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
	mongodrvr "go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/quickbite/analytics-service/app/analytics/controller"
	"github.com/quickbite/analytics-service/app/analytics/repository"
	"github.com/quickbite/analytics-service/app/analytics/service"
	"github.com/quickbite/analytics-service/lib/auth"
	"github.com/quickbite/analytics-service/lib/coreclient"
	"github.com/quickbite/analytics-service/lib/rbac"
)

const jwtSecret = "integration-test-secret-key"

var (
	testDB        *mongodrvr.Database
	testRabbitURI string

	restaurantRepo *repository.RestaurantDayRepo
	branchRepo     *repository.BranchDayRepo
	platformRepo   *repository.PlatformDayRepo
	productRepo    *repository.ProductDayRepo
	eventIDsRepo   *repository.EventIDsRepo
	svc            *service.AnalyticsService

	apiServer *httptest.Server
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	// ── MongoDB ──────────────────────────────────────────────────────────────────
	mongoContainer, err := mongodb.Run(ctx, "mongo:7")
	if err != nil {
		log.Fatalf("start mongo container: %v", err)
	}
	defer func() { _ = mongoContainer.Terminate(ctx) }()

	mongoURI, err := mongoContainer.ConnectionString(ctx)
	if err != nil {
		log.Fatalf("mongo connection string: %v", err)
	}

	mongoClient, err := mongodrvr.Connect(ctx, options.Client().ApplyURI(mongoURI))
	if err != nil {
		log.Fatalf("connect mongo: %v", err)
	}
	defer func() { _ = mongoClient.Disconnect(context.Background()) }()

	testDB = mongoClient.Database("analytics_integration_test")
	if err := repository.EnsureIndexes(ctx, testDB, slog.Default()); err != nil {
		log.Fatalf("ensure indexes: %v", err)
	}

	// ── RabbitMQ ─────────────────────────────────────────────────────────────────
	rabbitContainer, err := rabbitmq.Run(ctx, "rabbitmq:3")
	if err != nil {
		log.Fatalf("start rabbitmq container: %v", err)
	}
	defer func() { _ = rabbitContainer.Terminate(ctx) }()

	testRabbitURI, err = rabbitContainer.AmqpURL(ctx)
	if err != nil {
		log.Fatalf("rabbitmq amqp url: %v", err)
	}

	// ── Repos + Service ───────────────────────────────────────────────────────────
	restaurantRepo = repository.NewRestaurantDayRepo(testDB)
	branchRepo = repository.NewBranchDayRepo(testDB)
	platformRepo = repository.NewPlatformDayRepo(testDB)
	productRepo = repository.NewProductDayRepo(testDB)
	eventIDsRepo = repository.NewEventIDsRepo(testDB)
	svc = service.NewAnalyticsService(restaurantRepo, branchRepo, platformRepo, productRepo, slog.Default())

	// ── Mock RBAC core server ─────────────────────────────────────────────────────
	// Response must match httpclient.APIResponse envelope: {"success":true,"data":{...}}.
	// Returns analytics:read for role "manager", empty permissions for all others.
	mockCore := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		role := r.URL.Query().Get("role")
		var permsJSON string
		if role == "manager" {
			permsJSON = `{"permissions":[{"permission":"analytics:read"}]}`
		} else {
			permsJSON = `{"permissions":[]}`
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success":true,"data":%s}`, permsJSON)
	}))
	defer mockCore.Close()

	// ── HTTP test server ──────────────────────────────────────────────────────────
	coreClient := coreclient.New(mockCore.URL, "test-key")
	permCache := rbac.NewPermissionCache(coreClient, 5*time.Second)
	ctrl := controller.NewAnalyticsController(svc)

	r := chi.NewRouter()
	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(jwtSecret))
		controller.RegisterRoutes(r, ctrl, permCache, slog.Default())
	})
	apiServer = httptest.NewServer(r)
	defer apiServer.Close()

	os.Exit(m.Run())
}

// mintJWT creates a signed HS256 JWT with the given role claim.
func mintJWT(role string) string {
	claims := jwt.MapClaims{
		"userId":      1,
		"role":        role,
		"countryCode": "EG",
		"exp":         time.Now().Add(time.Hour).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, _ := token.SignedString([]byte(jwtSecret))
	return signed
}

// mustJSON marshals v and panics on error.
func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		panic(err)
	}
	return b
}
