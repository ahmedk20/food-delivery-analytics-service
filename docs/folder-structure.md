# Folder Structure

```
analytics-service/
├── cmd/
│   └── api/
│       └── main.go                          # Entry point — calls lib/boot.Run()
│
├── pkg/                                      # Framework-free. NO imports from lib/ or app/.
│   ├── mongo/
│   │   └── client.go                        # Connect/Disconnect/Database/Collection
│   ├── messaging/
│   │   ├── types.go                         # Broker interface, Message, ConsumerOptions, Handler
│   │   └── amqp.go                          # AMQP implementation of Broker
│   └── httpclient/
│       └── client.go                        # net/http wrapper: timeout, JSON decode, retry-on-5xx
│
├── lib/                                      # App-aware glue. May import pkg/.
│   ├── boot/
│   │   └── boot.go                          # Wires every singleton. Adding a module = edit one function.
│   ├── config/
│   │   └── env.go                           # Config struct + Load() via env tags
│   ├── logger/
│   │   └── logger.go                        # slog New + WithContext/FromContext
│   ├── appcontext/
│   │   └── context.go                       # ctx keys: Claims, CorrelationID
│   ├── errors/
│   │   ├── apperror.go                      # AppError struct + New() + common vars
│   │   └── handler.go                       # Wrap(log, handler) middleware
│   ├── http/
│   │   └── response.go                      # SendSuccess, SendError envelope helpers
│   ├── middleware/
│   │   └── correlation.go                   # Correlation ID + AccessLog middleware
│   ├── auth/
│   │   ├── jwt.go                           # VerifyToken(tokenStr, secret) → Claims
│   │   └── middleware.go                    # Authenticate(secret) chi middleware
│   ├── rbac/
│   │   ├── cache.go                         # PermissionCache: in-process by role + TTL
│   │   └── middleware.go                    # Require(cache, perm) chi middleware
│   ├── coreclient/
│   │   ├── client.go                        # Core service HTTP client (RBAC lookups)
│   │   └── types.go                         # Response types from core service
│   └── coreevents/
│       ├── consumer.go                      # Generic consumer: dedup → dispatch by routing key
│       └── payloads.go                      # Envelope struct (event_id, event_type, payload)
│
├── app/
│   └── analytics/                            # Package analytics — the one business module
│       ├── types.go                         # OnOrderPlacedInput, RestaurantDayResponse, DateRange
│       ├── errors.go                        # var Err… = apperr.New(...)
│       ├── enums.go                         # PermAnalyticsRead, collection name constants
│       ├── entity/
│       │   ├── restaurant_day.go            # RestaurantDay struct + bson tags
│       │   └── event_id.go                  # EventID struct + bson tags
│       ├── repository/
│       │   ├── indexes.go                   # EnsureIndexes — idempotent on boot
│       │   ├── restaurant_day_repo.go       # Upsert, FindByRestaurantAndRange
│       │   └── event_ids_repo.go            # MarkSeen (dedup via dup-key error)
│       ├── service/
│       │   └── analytics_service.go         # HandleOrderPlaced, GetRestaurantDays
│       ├── controller/
│       │   ├── analytics_controller.go      # GetRestaurantDays handler
│       │   └── routes.go                    # RegisterRoutes on chi.Router
│       ├── dto/
│       │   ├── days_request.go              # ParseGetRestaurantDaysRequest
│       │   └── days_response.go             # RestaurantDayDTO + ToRestaurantDayDTOs
│       └── eventhandlers/
│           └── handlers.go                  # Register(consumer, svc) — order.placed → service
│
├── play/                                     # GITIGNORED — dev-only helpers
│   ├── mock-core/main.go                    # Fake core service (RBAC endpoint)
│   ├── mint-jwt/main.go                     # Mint a test JWT
│   ├── publish-test/main.go                 # Publish order.placed to RabbitMQ
│   └── check-mongo/main.go                 # Dump agg_restaurant_day + event_ids
│
├── docs/                                     # Architecture + teaching docs
│   ├── folder-structure.md                  # This file
│   ├── system-design.md
│   ├── api-contracts.md
│   ├── node-to-go-mapping.md
│   ├── ai-prompts.md
│   └── implementation-plan.md
│
├── cmd/api/main.go
├── go.mod, go.sum
├── .env.example, .env.dev
├── .gitignore
├── CLAUDE.md
├── README.md
└── plan.md
```
