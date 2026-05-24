# Analytics Service — Video vs. Homework

## Video Scope (Built)

1. **One aggregate collection**: `agg_restaurant_day`
2. **One inbound event**: `order.placed` via RabbitMQ
3. **One GET endpoint**: `GET /api/v1/analytics/restaurants/:restaurantId/days?from=&to=`
4. **Idempotency**: `event_ids` collection with unique index + TTL
5. **Cross-cutting infra**: JWT auth, RBAC (read-through from core), correlation IDs, structured logging, error envelope, graceful shutdown
6. **Order-service publisher**: Transactional outbox + drain for `order.placed`

## Homework

1. **`agg_branch_day`** — same pattern as restaurant, keyed by `(branch_id, date)`
2. **`agg_product_day`** — orders_count + quantity_sold per product per day
3. **`agg_platform_day`** — platform-wide rollup (all restaurants)
4. **Additional events**: `order.delivered` (delivery_ms tracking), `payment.succeeded`, `order.cancelled`
5. **`rbac.permissions_changed` consumer** — invalidate RBAC cache on event instead of TTL
6. **Date-range validation max span** — reject queries wider than 90 days
7. **Pagination** for endpoints returning many rows
8. **Integration tests** with testcontainers for Mongo + Rabbit
