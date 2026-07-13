# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.25-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /analytics-service ./cmd/api

# ── Run stage ─────────────────────────────────────────────────────────────────
FROM alpine:3.21

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app
COPY --from=builder /analytics-service .

# Amazon DocumentDB requires TLS. MONGO_URI references
# `tls=true&tlsCAFile=global-bundle.pem` (path relative to /app), so bundle the
# RDS/DocumentDB CA into the image and make it world-readable.
ADD https://truststore.pki.rds.amazonaws.com/global/global-bundle.pem /app/global-bundle.pem
RUN chmod 0644 /app/global-bundle.pem

EXPOSE 3002

CMD ["./analytics-service"]
