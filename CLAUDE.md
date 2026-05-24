# QuickBite — Analytics Service: Project Guidelines

This file is the authoritative reference for Claude when working in this codebase.

---

## 1. Mission

Read-only analytics aggregation service for the QuickBite food-delivery platform.

- **Consumes** events from `order-service` via RabbitMQ.
- **Upserts** day-grained aggregates into MongoDB.
- **Serves** read-only HTTP endpoints under `/api/v1/analytics/...`.

Does **not** own users, restaurants, products, orders, or payments.
Does **not** write to any operational data store.
Does **not** emit events.

---

## 2. Stack

| Concern        | Choice                                    |
| -------------- | ----------------------------------------- |
| Runtime        | Go 1.21+                                  |
| HTTP router    | `github.com/go-chi/chi/v5`                |
| DB             | `go.mongodb.org/mongo-driver` (v1)        |
| Messaging      | `github.com/rabbitmq/amqp091-go`          |
| Config         | `github.com/caarlos0/env/v11`             |
| Logger         | stdlib `log/slog`                         |
| Validation     | `github.com/go-playground/validator/v10`  |
| JWT            | `github.com/golang-jwt/jwt/v5`            |
| UUID           | `github.com/google/uuid`                  |
| DI             | None — explicit constructor wiring in `lib/boot/boot.go` |

**Forbidden:** GORM, Ent, any ODM. Repos use mongo-driver directly with `bson` tags.
**Forbidden:** DDL migrations. Indexes in `app/analytics/repository/indexes.go`, idempotent on boot.

---

## 3. Folder Structure

```
cmd/api/main.go              # Entry point — calls lib/boot.Run()
pkg/                          # Framework-free, NO imports from lib/ or app/
  mongo/client.go             # Connect/Disconnect only
  messaging/{types,amqp}.go   # Broker interface + AMQP impl
  httpclient/client.go        # net/http wrapper
lib/                          # App-aware glue
  boot/boot.go                # Wires every singleton
  config/env.go               # Struct + Load()
  logger/logger.go            # slog wrapper + FromContext
  appcontext/context.go       # ctx keys: claims, correlation_id
  errors/{apperror,handler}.go
  http/response.go            # SendSuccess, SendError
  middleware/correlation.go   # Correlation + AccessLog
  auth/{jwt,middleware}.go    # JWT verify + Authenticate middleware
  rbac/{cache,middleware}.go  # In-process cache + Require()
  coreclient/{client,types}.go
  coreevents/{consumer,payloads}.go
app/analytics/                # Business module
  types.go, errors.go, enums.go
  entity/, repository/, service/, controller/, dto/, eventhandlers/
play/                         # GITIGNORED dev aids
docs/
```

---

## 4. Layering Rules (STRICT)

| Layer  | May import from          | Must NOT import from |
| ------ | ------------------------ | -------------------- |
| `pkg/` | Go stdlib, npm packages  | `lib/`, `app/`       |
| `lib/` | `pkg/`, Go stdlib        | `app/`               |
| `app/` | `lib/`, `pkg/`, stdlib   | —                    |

If `lib/` needs something from `app/`, define an interface in `lib/` (duck typing).

---

## 5. Naming Conventions

### Go
- Package names: lowercase, single word (`analytics`, `repository`, `dto`)
- Exported types: PascalCase (`AnalyticsService`, `RestaurantDayRepo`)
- Files: snake_case (`analytics_service.go`, `restaurant_day_repo.go`)
- Errors: `Err` prefix (`ErrInvalidDateRange`)
- Constants: PascalCase (`PermAnalyticsRead`, `CollectionAggRestaurantDay`)

### MongoDB
- Collection names: snake_case (`agg_restaurant_day`, `event_ids`)
- Field names: snake_case in BSON (`restaurant_id`, `orders_count`)

### Routes
- kebab-case paths: `/api/v1/analytics/restaurants/:restaurantId/days`

---

## 6. Response Envelopes

- Success: `{"success": true, "data": ...}`
- Error: `{"success": false, "error": {"code": "...", "message": "..."}}`
- Every response through DTO structs in `dto/*.go`. Never leak entity fields.

---

## 7. Database Rules

- Money in integer minor units (piastres). No floats.
- Dates as `YYYY-MM-DD` strings (UTC day grain).
- Averages stored as sum + count (associative for replays).
- No DDL migrations — schemaless MongoDB. Indexes idempotent on boot.

---

## 8. Cross-Cutting

### Auth
- JWT verified with `ACCESS_SECRET` (HS256). Claims: userId, role, countryCode, restaurantId?, restaurantRole?, branchIds?.
- Token from cookie first (`access_token`), then `Authorization: Bearer`.

### RBAC
- No local permissions catalog. Fetches from core-service `GET /api/internal/rbac/permissions?role=...`.
- In-process cache by role with configurable TTL.

### Idempotency
- `event_ids` collection with unique index on `event_id`.
- Duplicate insert → dup-key error → ack-and-skip.
- Unknown event types → log + ack-skip (don't DLQ).

### Error Handling
- Module errors: `var Err… = apperr.New(...)` in `app/analytics/errors.go`.
- Controller handlers return `error`; `errors.Wrap(log, handler)` renders envelope.

---

## 9. What NOT To Do

- Do NOT import `app/` from `lib/` or `pkg/`.
- Do NOT put collection names in `pkg/mongo/`.
- Do NOT declare inline structs in service/controller/repo files.
- Do NOT store floats for money.
- Do NOT use an ODM/ORM.
- Do NOT emit events from this service.
- Do NOT write to operational databases.
- Do NOT add indexes speculatively — query-driven only.

---

## 10. Reference Docs

- `docs/folder-structure.md` — annotated tree
- `docs/system-design.md` — architecture diagram + failure modes
- `docs/api-contracts.md` — request/response shapes + error codes
- `docs/node-to-go-mapping.md` — TS→Go idiom tables
- `docs/ai-prompts.md` — prompt templates for homework
- `docs/implementation-plan.md` — phases + acceptance checks
