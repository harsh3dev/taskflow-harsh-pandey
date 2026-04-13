# ── Build stage ──────────────────────────────────────────────────────────────
FROM golang:1.22-alpine AS build

WORKDIR /app

# Fetch dependencies first so this layer is cached when only source changes.
COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-w -s" \
    -o /bin/api \
    ./cmd/api

# ── Runtime stage ─────────────────────────────────────────────────────────────
FROM alpine:3.19

# ca-certificates: needed for TLS; tzdata: timezone support.
RUN apk add --no-cache ca-certificates tzdata

WORKDIR /app

COPY --from=build /bin/api /app/api

EXPOSE 8080

ENTRYPOINT ["/app/api"]
