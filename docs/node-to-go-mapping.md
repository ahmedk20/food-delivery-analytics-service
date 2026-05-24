# Node.js → Go Mapping Guide

This is the reference document for porting patterns from the existing Node.js services (core-service, order-service) to Go (analytics-service). Each section shows the TypeScript idiom on the left and the Go equivalent on the right.

---

## 1. Config / Environment

### TypeScript (Zod schema)
```typescript
const schema = z.object({
  PORT: z.coerce.number().positive().default(3001),
  ACCESS_SECRET: z.string().min(1),
  MONGO_URI: z.string().url(),
});
const env = schema.parse(process.env);
```

### Go (env struct tags)
```go
type Config struct {
    Port         int    `env:"PORT" envDefault:"3002"`
    AccessSecret string `env:"ACCESS_SECRET,required"`
    MongoURI     string `env:"MONGO_URI" envDefault:"mongodb://localhost:27017"`
}
cfg, err := env.Parse(&Config{})
```

**Key difference**: Go uses struct tags parsed at startup. No runtime schema object — the struct IS the schema. `required` tag replaces `z.string().min(1)`. Defaults via `envDefault` tag.

---

## 2. Logger

### TypeScript (winston)
```typescript
import logger from '../logger/logger.js';
logger.info('Order placed', { orderId: 42 });
logger.error('Payment failed', { err, region });
```

### Go (slog)
```go
log := slog.Default()
log.Info("order placed", "order_id", 42)
log.Error("payment failed", "error", err, "region", region)
```

**Key difference**: `slog` is stdlib — no dependency. Structured key-value pairs instead of an object. JSON output in production, text in dev. Logger threaded via `context.Context`, not imported as a singleton.

---

## 3. Error Handling

### TypeScript (throw + catch)
```typescript
// Define
export const OrderNotFoundError = new AppError('Order not found', 404);

// Throw
throw OrderNotFoundError;

// Catch (middleware)
app.use((err, req, res, next) => {
  res.status(err.statusCode).json({ error: err.message });
});
```

### Go (return error + middleware wrapper)
```go
// Define
var ErrInvalidDateRange = apperr.New("ANALYTICS_INVALID_DATE_RANGE", 400, "'from' must be before 'to'")

// Return
return nil, ErrInvalidDateRange

// Wrap (middleware)
func Wrap(log *slog.Logger, h func(w, r) error) http.HandlerFunc {
    return func(w, r) {
        if err := h(w, r); err != nil {
            var appErr *AppError
            if errors.As(err, &appErr) {
                SendError(w, appErr.StatusCode, appErr.Code, appErr.Message)
                return
            }
            SendError(w, 500, "INTERNAL_ERROR", "Something went wrong")
        }
    }
}
```

**Key difference**: Go doesn't throw — every function returns `(result, error)`. The `Wrap` middleware is the Go analogue of Express's error handler middleware. `errors.As` replaces `instanceof`.

---

## 4. DI / Dependency Injection

### TypeScript (TSyringe)
```typescript
@injectable()
export class OrderService {
  constructor(
    @inject(TOKENS.CacheProvider) private readonly cache: ICacheProvider,
    @inject(TOKENS.CoreServiceClient) private readonly coreClient: ICoreServiceClient,
  ) {}
}
// In container.ts
container.register(TOKENS.OrderService, { useClass: OrderService });
```

### Go (explicit constructors)
```go
// In service package
type AnalyticsService struct {
    repo *repository.RestaurantDayRepo
    log  *slog.Logger
}

func NewAnalyticsService(repo *repository.RestaurantDayRepo, log *slog.Logger) *AnalyticsService {
    return &AnalyticsService{repo: repo, log: log}
}

// In boot.go (the wiring file)
repo := repository.NewRestaurantDayRepo(db)
svc := service.NewAnalyticsService(repo, log)
ctrl := controller.NewAnalyticsController(svc)
```

**Key difference**: No framework, no decorators, no tokens. Constructor functions take explicit dependencies. All wiring in one file (`boot.go`). Compile-time safety — if you forget a dependency, it won't compile.

---

## 5. Classes → Structs + Methods

### TypeScript
```typescript
export class PermissionCacheService {
  private cache: Map<string, { permissions: string[]; cachedAt: number }>;
  private readonly TTL = toMs(1, 'h');

  constructor(private readonly rbacClient: RbacClient) {
    this.cache = new Map();
  }

  async getPermissions(roleName: string): Promise<string[]> {
    // ...
  }
}
```

### Go
```go
type PermissionCache struct {
    client *coreclient.Client
    ttl    time.Duration
    mu     sync.RWMutex
    cache  map[string]cacheEntry
}

func NewPermissionCache(client *coreclient.Client, ttl time.Duration) *PermissionCache {
    return &PermissionCache{client: client, ttl: ttl, cache: make(map[string]cacheEntry)}
}

func (pc *PermissionCache) GetPermissions(ctx context.Context, roleName string) ([]string, error) {
    // ...
}
```

**Key differences**:
- No `class` keyword — use `struct` + methods with receiver
- No `this` — use explicit receiver (`pc`)
- No `private` keyword — lowercase = unexported, uppercase = exported
- Constructor is just a function that returns a pointer
- Concurrency safety requires explicit `sync.RWMutex` (JS is single-threaded)

---

## 6. Promises → (T, error)

### TypeScript
```typescript
async function findOrderById(id: number, region: string): Promise<OrderEntity | undefined> {
  const row = await db(region)('orders').where({ id }).first();
  return row ? toEntity(row) : undefined;
}
```

### Go
```go
func (r *RestaurantDayRepo) FindByRestaurantAndRange(
    ctx context.Context,
    restaurantID int,
    from, to string,
) ([]entity.RestaurantDay, error) {
    cursor, err := r.coll.Find(ctx, filter)
    if err != nil {
        return nil, err
    }
    defer cursor.Close(ctx)
    // ...
}
```

**Key differences**:
- No `async/await` — Go uses goroutines + channels (or just sequential code)
- Return `(value, error)` instead of throwing
- Caller must check `err != nil` — no implicit catch
- `context.Context` is the first parameter (cancellation, deadlines, values)

---

## 7. Express Middleware → Chi Middleware

### TypeScript
```typescript
export async function authenticate(req: Request, res: Response, next: NextFunction) {
  const token = req.cookies?.access_token;
  if (!token) throw NotAuthenticated;
  const { payload } = await jwtVerify(token, secretKey);
  req.user = { userId: payload.userId, ... };
  next();
}

// Usage
router.use(authenticate);
```

### Go
```go
func Authenticate(secret string) func(http.Handler) http.Handler {
    return func(next http.Handler) http.Handler {
        return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
            token := extractToken(r)
            if token == "" {
                SendError(w, 401, "UNAUTHENTICATED", "...")
                return
            }
            claims, err := VerifyToken(token, secret)
            if err != nil {
                SendError(w, 401, "UNAUTHENTICATED", "...")
                return
            }
            ctx := appcontext.SetClaims(r.Context(), claims)
            next.ServeHTTP(w, r.WithContext(ctx))
        })
    }
}

// Usage
r.Use(auth.Authenticate(cfg.AccessSecret))
```

**Key differences**:
- Express: `(req, res, next)` with `next()` to continue
- Chi: `func(http.Handler) http.Handler` — wraps the next handler
- Express puts data on `req.user`; Go puts it in `context.Context`
- No `next(err)` in Go — either write the error response or call `next.ServeHTTP`

---

## 8. Route Registration

### TypeScript
```typescript
const router = express.Router();
router.get('/api/orders', authenticate, requireRole('customer'), orderCtrl.listOrders);
```

### Go
```go
r.Route("/api/v1/analytics", func(r chi.Router) {
    r.With(rbac.Require(pc, "analytics:read")).
        Get("/restaurants/{restaurantId}/days", apperr.Wrap(log, ctrl.GetRestaurantDays))
})
```

**Key differences**:
- Express: middleware as arguments before handler
- Chi: `r.With(middleware)` chains middleware; `r.Route` groups paths
- Path params: Express `:id`, Chi `{id}`
- Controller methods return `error` in Go (caught by `Wrap`)

---

## 9. Response DTOs

### TypeScript
```typescript
export class OrderResponseDTO {
  id: string;
  status: OrderStatus;
  static fromEntity(entity: OrderEntity): OrderResponseDTO {
    const dto = new OrderResponseDTO();
    dto.id = entity.publicId;
    return dto;
  }
}
sendSuccess(res, OrderResponseDTO.fromEntity(order), 201);
```

### Go
```go
type RestaurantDayDTO struct {
    Date          string `json:"date"`
    OrdersCount   int    `json:"ordersCount"`
    RevenueMinor  int64  `json:"revenueMinor"`
    Currency      string `json:"currency"`
    AvgOrderMinor int64  `json:"avgOrderMinor"`
}

func ToRestaurantDayDTOs(rows []analytics.RestaurantDayResponse) []RestaurantDayDTO { ... }

apphttp.SendSuccess(w, http.StatusOK, dto.ToRestaurantDayDTOs(rows))
```

**Key difference**: No `static` methods — use package-level functions. JSON field names via `json` struct tags (Go's equivalent of `class-transformer`).

---

## 10. Messaging / Event Consumer

### TypeScript
```typescript
await broker.consume(opts, async (msg) => {
  const isFirst = await cache.trySet(dedupKey, '1', 3600);
  if (!isFirst) { msg.ack(); return; }
  const payload = JSON.parse(msg.body.toString());
  await handlePayload(msg.routingKey, payload);
  msg.ack();
});
```

### Go
```go
consumer := coreevents.NewConsumer(broker, eventIDsRepo, log)
consumer.Register("order.placed", handleOrderPlaced(svc))
consumer.Start(ctx, consumerOpts)

// Inside consumer.Start:
// 1. Parse envelope → extract event_id
// 2. deduper.MarkSeen(eventID) — Mongo unique index, not Redis
// 3. Look up handler by routing key
// 4. Call handler(ctx, payload)
// 5. msg.Ack()
```

**Key differences**:
- Node uses Redis SETNX for dedup; Go uses MongoDB unique index (no Redis dependency)
- Node parses body inline; Go uses typed Envelope struct
- Handler registration is explicit map, not a switch statement

---

## 11. Repository / Database

### TypeScript (Knex)
```typescript
const [row] = await knex('orders').insert({ ... }).returning(COLUMNS);
const rows = await knex('orders').where({ customer_id: id }).select(COLUMNS);
```

### Go (mongo-driver)
```go
_, err := r.coll.UpdateOne(ctx, filter, update, options.Update().SetUpsert(true))

cursor, err := r.coll.Find(ctx, filter, options.Find().SetSort(bson.D{{Key: "date", Value: 1}}))
defer cursor.Close(ctx)
var results []entity.RestaurantDay
cursor.All(ctx, &results)
```

**Key differences**:
- Knex is SQL query builder; mongo-driver is document operations
- Both avoid ORMs — explicit queries, typed results
- Go uses `bson` struct tags (like Knex column arrays)
- `context.Context` for timeouts/cancellation (Knex uses connection pool)

---

## Common Gotchas When Porting

1. **No implicit type coercion**: `strconv.Atoi("42")` returns `(int, error)`, not `42`.
2. **Nil vs zero values**: Go has zero values (`0`, `""`, `nil`). A `*int` can be `nil`; a plain `int` is always `0`.
3. **No optional parameters**: Use pointer fields (`*int`) or option structs, not `param?: number`.
4. **Error handling is verbose**: Every call that can fail needs `if err != nil`. This is by design.
5. **Goroutines are not promises**: Don't `go func()` casually. Use `context.Context` for cancellation.
6. **Exported = uppercase**: `HandleOrderPlaced` is public, `handleOrderPlaced` is private.
7. **No constructor overloading**: One `New*` function. Use functional options for complex cases.
8. **Slices are nil-safe**: `len(nil)` is `0`, not a panic. But `json.Marshal(nil)` → `null`, not `[]`.
9. **Defer runs LIFO at function exit**: `defer cursor.Close(ctx)` — not at block exit.
10. **Maps are not concurrency-safe**: Use `sync.RWMutex` or `sync.Map`.

## When to Break the Mapping

- **Generics**: Go has generics (1.18+). Use `httpclient.Get[T]` instead of casting `any`.
- **Channels**: For fan-out/fan-in patterns, use channels instead of `Promise.all`.
- **Embedded structs**: Go composition replaces inheritance — don't force class hierarchies.
- **Interface satisfaction is implicit**: No `implements` keyword. If the struct has the methods, it satisfies the interface.
- **Package = directory**: No index.ts barrel exports. Each `.go` file in a directory shares the package.
