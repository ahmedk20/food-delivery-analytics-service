# QuickBite Analytics Service

Read-only analytics aggregation service. Consumes events from `order-service` via RabbitMQ, upserts day-grained aggregates into MongoDB, and serves read-only HTTP endpoints.

## Prerequisites

- Go 1.21+
- MongoDB running on `localhost:27017`
- RabbitMQ running on `localhost:5672`

## Quick Start

### 1. Environment

```bash
cp .env.example .env.dev
# Edit ACCESS_SECRET to match your core/order services
```

### 2. Three terminals

**Terminal 1 — Mock core service** (RBAC permission lookup):
```bash
go run play/mock-core/main.go
# Listens on :3000, returns analytics:read for all roles
```

**Terminal 2 — Analytics service**:
```bash
go run cmd/api/main.go
# Listens on :3002
# Logs: mongo connected, rabbit connected, event consumer started, http listening
```

**Terminal 3 — Publish test events**:
```bash
# Mint a JWT
export ACCESS_SECRET=dev-secret-change-me
TOKEN=$(go run play/mint-jwt/main.go)

# Publish an order.placed event
go run play/publish-test/main.go

# Check MongoDB
go run play/check-mongo/main.go

# Query the API
curl -s -b "access_token=$TOKEN" \
  "http://localhost:3002/api/v1/analytics/restaurants/42/days?from=2026-01-01&to=2026-12-31" | jq .
```

### Expected output

After publishing one event with `restaurant=42, total=2500`:
```json
{
  "success": true,
  "data": [
    {
      "date": "2026-05-24",
      "ordersCount": 1,
      "revenueMinor": 2500,
      "currency": "EGP",
      "avgOrderMinor": 2500
    }
  ]
}
```

## Health Check

```bash
curl http://localhost:3002/health
# {"success":true,"data":{"status":"ok"}}
```

## Project Structure

See `docs/folder-structure.md` for the annotated tree.
