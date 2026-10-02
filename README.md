# Foodeli

Foodeli is a production-ready, full-stack food delivery and marketplace platform built with a Go Modular Monolith backend and a Next.js (App Router) TypeScript frontend. The system enforces domain isolation, minor-unit financial precision, transactional dual-write safety (Outbox Pattern), real-time event streaming via Server-Sent Events (SSE), and seamless B2B payment gateway integration with TyrFintech.

---

## Architectural Principles and Core Highlights

- Modular Monolith Architecture: Each business domain (`auth`, `store`, `product`, `order`, `payment`, `cart`, `notification`) lives in a self-contained package under `services/`, with isolated models, repositories, business logic, and HTTP handlers. Ready for clean microservices extraction when needed.
- Decoupled Inter-Service Communication: Services interact exclusively through consumer-driven interfaces (such as `StoreChecker`, `ProductChecker`, `OrderManager`) rather than direct database joins across domain boundaries.
- Transactional Outbox Pattern: Solves the dual-write problem across PostgreSQL and message brokers. Domain events (`order.created`, `payment.succeeded`) are committed atomically within the same database transaction as the business operation, and relayed asynchronously to RabbitMQ via a dedicated worker using `SELECT ... FOR UPDATE SKIP LOCKED`.
- Real-Time Event Streaming (RabbitMQ to SSE): Event consumers ingest broker messages and route them through an in-memory, thread-safe `NotificationHub`. Clients receive instant status updates and push notifications over Server-Sent Events (`/api/v1/notifications/stream`) without long-polling overhead.
- Financial Precision: Monetary values are strictly stored and calculated in minor currency units (`int64` cents, EUR) across database columns, Go structs, and API contracts, eliminating IEEE 754 floating-point inaccuracies.
- Pluggable Dual Payment Gateways: Supports both direct credit card settlement and external hosted checkout sessions through the `PaymentGateway` strategy interface. Integrated with the TyrFintech B2B platform via hosted checkout URLs and idempotent server-to-server webhook receivers (`/api/v1/webhooks/tyrfintech`).
- Distributed Sliding-Window Rate Limiting: Atomic Redis Lua scripts guard public endpoints (`/auth/login`, `/auth/register`, `/webhooks/tyrfintech`) against credential stuffing and replay attacks, with standard `X-RateLimit-*` and `Retry-After` HTTP headers.
- Ephemeral Multi-Store Cart Management: Redis-backed cart store with 7-day TTL, strict single-store validation (with user-confirmed replacement), and atomic checkout-to-order conversion.
- Keyset-Friendly Pagination: Standardized query parameters (`page`, `limit`) and response headers (`X-Total-Count`, `X-Total-Pages`, `X-Page`, `X-Limit`) across order histories and store catalogs, paired with client-side pagination navigators.
- Recruiter and Demonstration Ready: Built-in 1-click test profiles (Customer, Merchant, Admin) and pre-seeded database with 8 Finnish restaurants and 40 menu items.

---

## Technology Stack

### Backend
- Language: Go 1.24+
- HTTP Framework: Gin Web Framework
- Database: PostgreSQL 16 (connection pooling via `pgxpool` v5)
- Cache and Rate Limiting: Redis 7 (`go-redis` v9)
- Message Broker: RabbitMQ 3 (Topic Exchange `foodeli.events`)
- Authentication: JWT (HMAC-SHA256) with Bearer token and HttpOnly cookie fallback
- API Documentation: Swagger / OpenAPI 2.0 (`swag`, `gin-swagger`)

### Frontend
- Framework: Next.js 14 (App Router, Standalone build)
- Language: TypeScript 5
- Styling: Tailwind CSS
- State Management: Zustand (with persistence middleware)
- Component Library: shadcn/ui primitives, Lucide icons, Sonner toast engine
- Real-Time Protocol: Native Browser EventSource (Server-Sent Events)

### Containerization and Deployment
- Go API: Multi-stage Alpine container with non-root unprivileged execution (`appuser`)
- Web Client: Multi-stage standalone Next.js container (optimized ~100MB footprint)
- Orchestration: Docker Compose with automated health checks and database seed provisioning

---

## Service Overview

### Platform Core (`platform/`)
- `config`: 12-factor environment loader supporting local and production runtime settings.
- `database`: PostgreSQL connection pool management via `pgxpool`.
- `redis`: Redis connection manager with health verification and Lua script execution.
- `rabbitmq`: RabbitMQ channel manager, durable topic exchange declarations, and persistent publishing.
- `outbox`: Transactional outbox storage and background relay worker with exponential backoff.
- `events`: Centralized domain event schemas (`OrderCreatedPayload`, `PaymentSucceededPayload`).
- `middleware`: JWT authentication, RBAC authorization, and distributed sliding-window rate limiters.
- `response`: Standardized JSON envelope (`success`, `data`, `error`).

### Business Services (`services/`)
- `auth`: User registration, bcrypt hashing, JWT issuance, profile retrieval, and cookie invalidation.
- `store`: Restaurant storefront provisioning, 1-to-1 merchant ownership guards, and catalog queries.
- `product`: Inventory management, soft deletes (`deleted_at`), and price tracking in minor units.
- `order`: Multi-item checkout, price snapshots, concurrency-safe inventory deduction, and state machine validations (`PENDING` -> `CONFIRMED` -> `PREPARING` -> `DELIVERED` / `CANCELLED`).
- `payment`: Multi-gateway dispatcher (`MockGateway`, `TyrFintechGateway`), hosted checkout sessions, and webhook processing.
- `cart`: Redis-based cart store enforcing single-restaurant integrity.
- `notification`: In-memory SSE client registry (`NotificationHub`) and RabbitMQ queue consumer.

---

## Directory Structure

```text
foodeli/
├── cmd/
│   └── api/
│       ├── app.go              # Infrastructure initialization & lifecycle management
│       ├── main.go             # Application entrypoint
│       └── router.go           # Route registration and middleware pipeline
├── migrations/
│   ├── 000001_init_schema.up.sql   # PostgreSQL relational schema
│   ├── seed_finland_data.sql       # Seed data for Finnish restaurants & menus
│   └── create_admin_user.sql       # Administrative test account
├── platform/
│   ├── config/                 # Environment variables configuration
│   ├── database/               # PostgreSQL pgxpool driver
│   ├── events/                 # Domain event payload definitions
│   ├── middleware/             # JWT, RBAC, and Redis Lua rate limiting
│   ├── outbox/                 # Transactional Outbox pattern & relay worker
│   ├── rabbitmq/               # RabbitMQ AMQP client wrapper
│   ├── redis/                  # Redis client pool
│   └── response/               # Standardized HTTP JSON responses
├── services/
│   ├── auth/                   # Authentication & user identity
│   ├── cart/                   # Redis ephemeral shopping cart
│   ├── notification/           # NotificationHub & RabbitMQ SSE consumer
│   ├── order/                  # Order processing & state machine
│   ├── payment/                # Dual payment gateways (Mock & TyrFintech)
│   ├── product/                # Product catalog & stock tracking
│   └── store/                  # Storefront management & menu retrieval
├── web/
│   ├── src/
│   │   ├── app/                # Next.js App Router pages
│   │   ├── components/         # Reusable UI components & feature pages
│   │   ├── hooks/              # Custom React hooks (useNotifications SSE)
│   │   ├── lib/                # API client & currency formatting utilities
│   │   ├── store/              # Zustand state stores (auth, cart)
│   │   └── types/              # TypeScript API type definitions
│   ├── Dockerfile              # Next.js standalone multi-stage build
│   └── next.config.mjs         # Standalone build configuration
├── Dockerfile                  # Go API multi-stage build
├── docker-compose.prod.yml     # Complete production orchestration
├── docker-compose.yml          # Local infrastructure orchestration
├── go.mod
├── go.sum
└── README.md
```

---

## Running with Docker (Recommended for Production / VPS)

The production configuration runs the entire stack (PostgreSQL, Redis, RabbitMQ, Go API, Next.js Web) in isolated containers with automated health checks and initial database migrations.

### 1. Clone the Repository
```bash
git clone <repository-url>
cd foodeli
```

### 2. Configure Environment
```bash
cp .env.example .env
```
Ensure database credentials, JWT secrets, and TyrFintech connection URLs match your deployment target.

### 3. Build and Launch
```bash
docker compose -f docker-compose.prod.yml up -d --build
```

### 4. Verify Services
```bash
docker compose -f docker-compose.prod.yml ps
```

Default exposed ports:
- Web Application: `http://localhost:3000`
- API Server: `http://localhost:8080`
- API Health Check: `http://localhost:8080/health`
- Swagger UI: `http://localhost:8080/swagger/index.html`
- RabbitMQ Management: `http://localhost:15672` (guest / guest)

---

## Local Development Setup

### 1. Start Infrastructure Dependencies
```bash
docker compose up -d postgres redis rabbitmq
```

### 2. Run Database Migrations and Seed Data
```bash
docker exec -i foodeli-postgres psql -U foodeli -d foodeli_db < migrations/000001_init_schema.up.sql
docker exec -i foodeli-postgres psql -U foodeli -d foodeli_db < migrations/seed_finland_data.sql
docker exec -i foodeli-postgres psql -U foodeli -d foodeli_db < migrations/create_admin_user.sql
```

### 3. Run Backend API
```bash
go run ./cmd/api
```

### 4. Run Frontend Client
```bash
cd web
npm install
npm run dev
```

The web application will be available at `http://localhost:3000`.

---

## Running Automated Tests

Run the complete backend test suite across all services:
```bash
go test -v ./...
```

Verify frontend TypeScript type safety:
```bash
cd web && npx tsc --noEmit
```

---

## API Documentation

Interactive OpenAPI / Swagger documentation is auto-generated and served at:
`http://localhost:8080/swagger/index.html`

To regenerate Swagger documentation after modifying handler annotations:
```bash
swag init -g cmd/api/main.go -o docs
```
