# API Contracts

## Health Check

### `GET /health`

No authentication required.

**Response** `200 OK`:
```json
{
  "success": true,
  "data": { "status": "ok" }
}
```

---

## Analytics — Restaurant Days

### `GET /api/v1/analytics/restaurants/:restaurantId/days`

**Auth**: JWT required + RBAC `analytics:read` permission.

**Path params**:
| Param          | Type   | Required | Description                    |
| -------------- | ------ | -------- | ------------------------------ |
| `restaurantId` | int    | yes      | Positive integer restaurant ID |

**Query params**:
| Param  | Type   | Required | Format       | Description       |
| ------ | ------ | -------- | ------------ | ----------------- |
| `from` | string | yes      | `YYYY-MM-DD` | Start date (incl) |
| `to`   | string | yes      | `YYYY-MM-DD` | End date (incl)   |

**Response** `200 OK`:
```json
{
  "success": true,
  "data": [
    {
      "date": "2026-05-24",
      "ordersCount": 5,
      "revenueMinor": 12500,
      "currency": "EGP",
      "avgOrderMinor": 2500
    }
  ]
}
```

`data` is an empty array `[]` if no data exists for the restaurant in the range.

---

## Error Codes

| Code                          | HTTP | When                                    |
| ----------------------------- | ---- | --------------------------------------- |
| `UNAUTHENTICATED`             | 401  | No token, expired token, invalid token  |
| `FORBIDDEN`                   | 403  | Role lacks `analytics:read` permission  |
| `VALIDATION_ERROR`            | 400  | Missing/malformed query params or path  |
| `ANALYTICS_INVALID_DATE_RANGE`| 400  | `from` is after `to`                    |
| `RBAC_LOOKUP_FAILED`          | 500  | Core service RBAC call failed           |
| `INTERNAL_ERROR`              | 500  | Unhandled server error                  |

**Error envelope**:
```json
{
  "success": false,
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Query parameter 'from' is required (YYYY-MM-DD)"
  }
}
```
