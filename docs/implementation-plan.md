# Implementation Plan

## Phase 1 — Project Scaffold ✅ (Video)
- Initialize Go module
- Create directory structure matching the spec
- Add `.gitignore`, `.env.example`, `.env.dev`
- **Check**: `go mod init` succeeds, directories exist

## Phase 2 — pkg/ Layer ✅ (Video)
- `pkg/mongo/client.go` — Connect, Disconnect, Database, Collection
- `pkg/messaging/types.go` — Broker interface, Message, ConsumerOptions
- `pkg/messaging/amqp.go` — AMQP implementation
- `pkg/httpclient/client.go` — HTTP wrapper with retry
- **Check**: `go build ./pkg/...` clean

## Phase 3 — lib/ Layer ✅ (Video)
- `lib/config/env.go` — Config struct with env tags
- `lib/logger/logger.go` — slog wrapper
- `lib/appcontext/context.go` — Claims, CorrelationID context keys
- `lib/errors/apperror.go` — AppError type + common errors
- `lib/errors/handler.go` — Wrap middleware
- `lib/http/response.go` — SendSuccess, SendError
- `lib/middleware/correlation.go` — Correlation ID + AccessLog
- `lib/auth/jwt.go` — VerifyToken
- `lib/auth/middleware.go` — Authenticate middleware
- `lib/rbac/cache.go` — PermissionCache (in-process, TTL)
- `lib/rbac/middleware.go` — Require middleware
- `lib/coreclient/client.go` — Core service HTTP client
- `lib/coreevents/consumer.go` — Generic consumer with dedup
- `lib/coreevents/payloads.go` — Envelope struct
- **Check**: `go build ./lib/...` clean

## Phase 4 — app/analytics Module ✅ (Video)
- `app/analytics/types.go` — Input/output types
- `app/analytics/errors.go` — Domain errors
- `app/analytics/enums.go` — Permission + collection constants
- `app/analytics/entity/` — RestaurantDay, EventID structs
- `app/analytics/repository/indexes.go` — EnsureIndexes
- `app/analytics/repository/restaurant_day_repo.go` — Upsert + Find
- `app/analytics/repository/event_ids_repo.go` — MarkSeen (dedup)
- `app/analytics/service/analytics_service.go` — HandleOrderPlaced, GetRestaurantDays
- `app/analytics/controller/analytics_controller.go` — GetRestaurantDays handler
- `app/analytics/controller/routes.go` — RegisterRoutes
- `app/analytics/dto/days_request.go` — Request parsing
- `app/analytics/dto/days_response.go` — Response DTO
- `app/analytics/eventhandlers/handlers.go` — order.placed → service
- **Check**: `go build ./app/...` clean

## Phase 5 — Boot + Main ✅ (Video)
- `lib/boot/boot.go` — Wire all singletons, start consumer, start HTTP
- `cmd/api/main.go` — Calls boot.Run()
- **Check**: `go build ./...` and `go vet ./...` clean

## Phase 6 — Play Helpers + E2E ✅ (Video)
- `play/mint-jwt/main.go` — Mint test JWT
- `play/mock-core/main.go` — Fake RBAC endpoint
- `play/publish-test/main.go` — Publish order.placed to RabbitMQ
- `play/check-mongo/main.go` — Dump collections
- **Check**: All 11 acceptance criteria pass

---

## Phase 7 — agg_branch_day (Homework)
- New entity, repo, indexes
- Update service to upsert branch day on order.placed
- New GET endpoint + DTO
- **Check**: Same acceptance pattern as restaurant days

## Phase 8 — agg_product_day (Homework)
- Product-level aggregation: orders_count + quantity_sold
- Requires parsing items array from event payload
- **Check**: Publish event with items, verify product-day rows

## Phase 9 — agg_platform_day (Homework)
- Platform-wide rollup
- New GET endpoint (no restaurantId param)
- **Check**: Query returns sum across all restaurants

## Phase 10 — Additional Events (Homework)
- `order.delivered` — delivery_ms tracking
- `payment.succeeded` — revenue confirmation
- `order.cancelled` — decrement counts
- **Check**: Each event updates the correct aggregate fields

## Phase 11 — RBAC Event Consumer (Homework)
- Subscribe to `core-service.events` exchange
- Handle `rbac.permissions_changed` → invalidate cache
- **Check**: Change permission in core, verify analytics serves/denies immediately

## Phase 12 — Integration Tests (Homework)
- testcontainers for MongoDB + RabbitMQ
- Test: publish event → verify aggregate → query API
- Test: dedup (same event twice → count=1)
- Test: auth + RBAC (401, 403, 200)
- **Check**: `go test ./...` green with 80%+ coverage
