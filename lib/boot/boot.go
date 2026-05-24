package boot

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/quickbite/analytics-service/app/analytics/controller"
	"github.com/quickbite/analytics-service/app/analytics/eventhandlers"
	"github.com/quickbite/analytics-service/app/analytics/repository"
	"github.com/quickbite/analytics-service/app/analytics/service"
	"github.com/quickbite/analytics-service/lib/auth"
	"github.com/quickbite/analytics-service/lib/config"
	"github.com/quickbite/analytics-service/lib/coreclient"
	"github.com/quickbite/analytics-service/lib/coreevents"
	apphttp "github.com/quickbite/analytics-service/lib/http"
	"github.com/quickbite/analytics-service/lib/logger"
	"github.com/quickbite/analytics-service/lib/middleware"
	"github.com/quickbite/analytics-service/lib/rbac"
	mongopkg "github.com/quickbite/analytics-service/pkg/messaging"
	pkgmongo "github.com/quickbite/analytics-service/pkg/mongo"
)

func Run() error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	log := logger.New(cfg.AppStage)
	slog.SetDefault(log)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	// ── MongoDB ─────────────────────────────────────────────────────────────
	mongoClient, err := pkgmongo.Connect(ctx, cfg.MongoURI, cfg.MongoDatabase)
	if err != nil {
		return fmt.Errorf("mongo connect: %w", err)
	}
	defer mongoClient.Disconnect(context.Background())
	log.Info("mongo connected", "database", cfg.MongoDatabase)

	db := mongoClient.Database()

	if err := repository.EnsureIndexes(ctx, db, log); err != nil {
		return fmt.Errorf("ensure indexes: %w", err)
	}

	// ── RabbitMQ ────────────────────────────────────────────────────────────
	broker := mongopkg.NewAMQPBroker(cfg.RabbitMQURL, log)
	if err := broker.Connect(ctx); err != nil {
		return fmt.Errorf("rabbit connect: %w", err)
	}
	defer broker.Close()
	log.Info("rabbit connected")

	// ── Repositories ────────────────────────────────────────────────────────
	restaurantDayRepo := repository.NewRestaurantDayRepo(db)
	branchDayRepo := repository.NewBranchDayRepo(db)
	platformDayRepo := repository.NewPlatformDayRepo(db)
	eventIDsRepo := repository.NewEventIDsRepo(db)

	// ── Core client + RBAC cache ────────────────────────────────────────────
	coreClient := coreclient.New(cfg.CoreServiceURL, cfg.CoreServiceAPIKey)
	permCache := rbac.NewPermissionCache(coreClient, time.Duration(cfg.RBACCacheTTLSec)*time.Second)

	// ── Service ─────────────────────────────────────────────────────────────
	analyticsSvc := service.NewAnalyticsService(restaurantDayRepo, branchDayRepo, platformDayRepo, log)

	// ── Event consumer ──────────────────────────────────────────────────────
	consumer := coreevents.NewConsumer(broker, eventIDsRepo, log)
	eventhandlers.Register(consumer, analyticsSvc)

	consumerOpts := mongopkg.ConsumerOptions{
		Exchange:           cfg.RabbitMQExchange,
		Queue:              cfg.RabbitMQQueue,
		BindingKeys:        []string{"order.#", "payment.#"},
		Prefetch:           cfg.RabbitMQPrefetch,
		DeadLetterExchange: cfg.RabbitMQDLX,
		DeadLetterQueue:    cfg.RabbitMQDLQ,
	}
	if err := consumer.Start(ctx, consumerOpts); err != nil {
		return fmt.Errorf("start consumer: %w", err)
	}
	log.Info("event consumer started", "queue", cfg.RabbitMQQueue)

	// ── Core-service event consumer (RBAC cache invalidation) ───────────────
	coreBroker := mongopkg.NewAMQPBroker(cfg.RabbitMQURL, log)
	if err := coreBroker.Connect(ctx); err != nil {
		return fmt.Errorf("core-service rabbit connect: %w", err)
	}
	defer coreBroker.Close()

	coreConsumerOpts := mongopkg.ConsumerOptions{
		Exchange:    cfg.CoreEventsExchange,
		Queue:       cfg.CoreEventsQueue,
		BindingKeys: []string{"rbac.#"},
		Prefetch:    1,
	}
	if err := coreBroker.Consume(ctx, coreConsumerOpts, func(ctx context.Context, msg mongopkg.Message) error {
		var env coreevents.Envelope
		if err := json.Unmarshal(msg.Body, &env); err != nil {
			log.Warn("malformed core-service event, skipping", "error", err)
			return msg.Ack()
		}

		var payload coreevents.RBACPermissionsChangedPayload
		if err := json.Unmarshal(env.Payload, &payload); err != nil || payload.RoleName == "" {
			log.Warn("malformed rbac payload, skipping", "routing_key", msg.RoutingKey)
			return msg.Ack()
		}

		permCache.Invalidate(payload.RoleName)
		log.Info("rbac cache invalidated via event", "role", payload.RoleName)
		return msg.Ack()
	}); err != nil {
		return fmt.Errorf("start core-events consumer: %w", err)
	}
	log.Info("core-events consumer started", "queue", cfg.CoreEventsQueue)

	// ── HTTP ────────────────────────────────────────────────────────────────
	analyticsCtrl := controller.NewAnalyticsController(analyticsSvc)

	r := chi.NewRouter()
	r.Use(middleware.Correlation)
	r.Use(middleware.AccessLog(log))

	r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
		apphttp.SendSuccess(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Group(func(r chi.Router) {
		r.Use(auth.Authenticate(cfg.AccessSecret))
		controller.RegisterRoutes(r, analyticsCtrl, permCache, log)
	})

	addr := fmt.Sprintf("%s:%d", cfg.Host, cfg.Port)
	srv := &http.Server{Addr: addr, Handler: r}

	go func() {
		log.Info("http listening", "addr", addr)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("http server error", "error", err)
		}
	}()

	// ── Graceful shutdown ───────────────────────────────────────────────────
	<-ctx.Done()
	log.Info("shutting down…")

	shutCtx, shutCancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer shutCancel()
	return srv.Shutdown(shutCtx)
}
