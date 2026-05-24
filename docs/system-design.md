# System Design

## Platform Architecture

```
┌──────────────┐       ┌──────────────┐       ┌──────────────────┐
│  Frontend    │──────▶│ Core Service  │       │ Analytics Service │
│  (React)     │       │ (Node + PG)  │       │ (Go + MongoDB)   │
└──────────────┘       └──────┬───────┘       └────────▲─────────┘
                              │                        │
                    sync HTTP │               async consume
                              │                        │
                       ┌──────▼───────┐       ┌───────┴────────┐
                       │ Order Service │──────▶│   RabbitMQ     │
                       │ (Node + PG)  │ outbox │ order.events   │
                       └──────────────┘ drain  └────────────────┘
```

## Data Flow

### Sync (HTTP)
1. Frontend → Core Service: auth, users, restaurants, products, RBAC
2. Frontend → Order Service: place order, list orders, payments
3. Frontend → Analytics Service: GET aggregated data
4. Analytics Service → Core Service: RBAC permission lookup (role → permissions)

### Async (RabbitMQ)
1. Order Service publishes `order.placed` via transactional outbox → `order-service.events` exchange
2. Analytics Service consumes from `analytics-service.order-events` queue (bound to `order.#`, `payment.#`)
3. Each event is deduplicated via `event_ids` collection (unique index)
4. Aggregates upserted into `agg_restaurant_day`

## Failure Modes

| Failure                      | Impact                               | Mitigation                                              |
| ---------------------------- | ------------------------------------ | ------------------------------------------------------- |
| MongoDB down                 | API returns 500, events nack→requeue | Rabbit redelivers; health check reports unhealthy        |
| RabbitMQ down                | Events not consumed                  | Outbox rows stay in order-service; drain retries         |
| Core service down            | RBAC lookup fails                    | In-process cache serves stale perms until TTL expires    |
| Duplicate event delivery     | Double-counting risk                 | `event_ids` unique index → dup-key → ack-and-skip       |
| Unknown event type           | —                                    | Log + ack-skip (don't DLQ)                              |
| Handler error                | Message nacked to DLQ                | DLQ for manual inspection; no infinite retry loop        |
| Analytics service restarts   | Brief gap in consumption             | Unacked messages redeliver; dedupe prevents double-count |

## Why MongoDB Dedup, Not Redis SETNX

Redis SETNX would require Redis as a dependency just for dedup. MongoDB is already the primary store,
and the `event_ids` collection with a unique index + TTL provides the same guarantee:
- Insert succeeds → first time → process
- Insert returns dup-key → replay → ack-and-skip
- TTL auto-cleans old entries (7 days)

No additional infrastructure. No cache invalidation complexity. The trade-off is slightly higher
latency per dedup check (disk vs. memory), which is fine for async event processing.

## Aggregate Design

Aggregates use sum+count pattern for associative merging:

```
{
  restaurant_id: 42,
  date: "2026-05-24",
  currency: "EGP",
  orders_count: 5,       // count
  revenue_sum: 12500,    // sum in minor units (piastres)
  delivery_ms_sum: 0,    // sum of delivery durations
  delivery_ms_count: 0,  // count of deliveries (for avg)
  updated_at: ISODate(...)
}
```

Averages are computed at read time: `avgOrderMinor = revenue_sum / orders_count`.
This ensures per-day rows merge correctly even if events replay out of order.
