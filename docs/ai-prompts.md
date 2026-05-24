# AI Prompt Templates for Homework

The skill you're learning isn't Go — it's **directing an AI through a non-trivial implementation in a new language by anchoring it to a codebase you already know.**

---

## Ground Rule: Make the AI Prove Understanding First

Before asking the AI to write code, ask it to explain the existing pattern:

> "Read `app/analytics/repository/restaurant_day_repo.go` and `app/analytics/repository/indexes.go`. Explain how the upsert works, why we use `$inc` instead of `$set` for `orders_count`, and how the unique index prevents duplicates. Then I'll ask you to build a similar repo for `agg_branch_day`."

If the AI can't explain the existing code correctly, it won't write good new code.

---

## Homework 1: Add `agg_branch_day` Collection

### Step 1 — Understand the pattern
> "Read `app/analytics/repository/restaurant_day_repo.go` and `app/analytics/types.go`. List every type and function involved in the restaurant-day aggregate. I want a complete dependency map before we write anything."

### Step 2 — Create the types
> "Following the exact same pattern as `RestaurantDayRow` and `RestaurantDayResponse` in `types.go`, add `BranchDayRow` and `BranchDayResponse`. The key is `(branch_id, date)` instead of `(restaurant_id, date)`. Show me the diff only — don't rewrite the whole file."

### Step 3 — Create the repo
> "Mirror `restaurant_day_repo.go` to create `branch_day_repo.go`. Same Upsert + FindByRange pattern, but keyed on `branch_id`. Add indexes in `indexes.go` following the same unique + range pattern."

### Step 4 — Wire it up
> "The event payload from `order.placed` already contains `branchId`. Update `analytics_service.go` to also call `branchDayRepo.Upsert` inside `HandleOrderPlaced`. Then update `boot.go` to construct and inject the new repo."

---

## Homework 2: Add `order.delivered` Event Handler

### Step 1 — Understand the consumer
> "Read `lib/coreevents/consumer.go` and `app/analytics/eventhandlers/handlers.go`. Explain the full path an event takes from RabbitMQ message to service method call. Include the dedup step."

### Step 2 — Design the input
> "The `order.delivered` event has a payload with `{orderId, restaurantId, branchId, deliveryDurationMs, deliveredAt}`. Add an `OnOrderDeliveredInput` struct to `types.go`. What's the Go analogue of the TypeScript type I'd write for this?"

### Step 3 — Update the service
> "Add a `HandleOrderDelivered` method to `AnalyticsService` that updates `delivery_ms_sum` and `delivery_ms_count` on the existing `agg_restaurant_day` row. Use `$inc` — the same pattern as `HandleOrderPlaced`."

### Step 4 — Register the handler
> "In `eventhandlers/handlers.go`, register `order.delivered` → `handleOrderDelivered`. Follow the exact pattern of `handleOrderPlaced`."

---

## Homework 3: Add RBAC Permission Invalidation via Event

### Step 1 — Understand the current approach
> "Read `lib/rbac/cache.go`. How does the TTL-based invalidation work? What's the failure mode if a permission is removed but the cache hasn't expired?"

### Step 2 — Plan the event-driven approach
> "In the Node order-service, `rbac.permissions_changed` events arrive on `core.events` exchange and call `permissionCacheService.invalidate(roleName)`. How would we add the same pattern here? The exchange is `core-service.events`, not `order-service.events` — so we need a second consumer or to bind our queue to both exchanges. Which approach is better, and why?"

### Step 3 — Implement
> "Add a second consumer in `boot.go` that subscribes to `core-service.events` with binding key `rbac.#`. When `rbac.permissions_changed` arrives, extract `roleName` from the payload and call `permCache.Invalidate(roleName)`. Show me the full diff."

---

## Homework 4: Add a New GET Endpoint

> "Add `GET /api/v1/analytics/platform/days?from=&to=` that returns platform-wide aggregates (all restaurants). Follow the exact pattern of the restaurant endpoint:
> 1. DTO in `dto/` for request parsing
> 2. Service method that queries a new `agg_platform_day` collection
> 3. Controller method that calls service and sends response
> 4. Route registration in `controller/routes.go`
>
> Before writing any code, show me the Node analogue — if I had this endpoint in the order-service, what would the controller, service, and repo look like? Then translate to Go."

---

## Meta-Prompts (Process, Not Content)

### When stuck on Go syntax
> "I know how to do X in TypeScript: [paste TS code]. What's the idiomatic Go equivalent? Show me both side by side, and explain any differences in error handling or concurrency."

### When the AI writes non-idiomatic code
> "Compare your code to `app/analytics/service/analytics_service.go` lines 20-35. Your version uses [pattern]. The existing code uses [different pattern]. Which matches the project's conventions? Rewrite to match."

### When the AI tries to skip steps
> "Don't write the implementation yet. First: read the three files I listed, then explain the pattern in 3 sentences. I want to verify you understood before we write code."

### When adding a new module
> "List every file I'd need to create for a new `billing` module, following the `analytics` module structure. Just the file paths and one-line descriptions — no code yet."
