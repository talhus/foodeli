# Multi-stage build for Foodeli Core API
FROM golang:1.24-alpine AS builder

RUN apk add --no-cache git ca-certificates tzdata

WORKDIR /app

# Cache dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source code (respecting .dockerignore)
COPY . .

# Build statically-linked, stripped binary
RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w" -o /app/foodeli-api ./cmd/api

# Minimal unprivileged runtime image
FROM alpine:3.20 AS runner

RUN apk --no-cache add ca-certificates tzdata wget \
    && addgroup -S appgroup && adduser -S appuser -G appgroup

WORKDIR /app

COPY --from=builder /app/foodeli-api .
COPY --from=builder /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/

USER appuser

EXPOSE 8080

HEALTHCHECK --interval=20s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

CMD ["./foodeli-api"]
