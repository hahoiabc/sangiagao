# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

**Sàn Giao Gạo (Rice Marketplace)** — A Vietnamese rice marketplace platform with 4 services in a monorepo:

| Service | Tech | Port | Directory |
|---------|------|------|-----------|
| Backend API | Go 1.25 / Gin | :8080 | `backend/` |
| Admin Panel | Next.js 16 / React 19 | :3000 | `admin/` |
| Mobile App | Flutter / Riverpod | — | `mobile/` |
| Chat Service | Elixir / Phoenix | :4000 | `chat/` |

Infrastructure: PostgreSQL 16, Redis 7, MongoDB 7, MinIO (S3-compatible).

## Common Commands

### Infrastructure
```bash
cd infras/docker && docker compose --env-file ../.env up -d   # Start all services
docker compose --env-file ../.env down                         # Stop
docker compose --env-file ../.env down -v                      # Stop + delete data
bash infras/scripts/start.sh                                   # Convenience wrapper
```

### Backend (Go)
```bash
cd backend
go run ./cmd/api              # Run API server
go test ./internal/service/...  # Run service tests
go test ./...                 # Run all tests
go test -run TestSubscription ./internal/service/...  # Single test pattern
```

### Admin (Next.js)
```bash
cd admin
npm run dev       # Dev server (localhost:3000)
npm run build     # Production build
npx tsc --noEmit  # Type check without building
```

### Mobile (Flutter)
```bash
cd mobile
flutter run
flutter build apk
```

## Architecture

### Backend — Layered Architecture (handler → service → repository)

```
cmd/api/main.go          — Entry point, DI wiring, route registration
internal/
  config/                — Env config loading
  middleware/            — JWT auth, CORS, rate limiting, role checks
  handler/              — HTTP handlers (Gin), request/response parsing
  service/              — Business logic, no DB access
  service/deps.go       — Consumer-defined interfaces for all repositories
  repository/           — PostgreSQL queries (pgx/v5)
  model/                — Domain structs with JSON tags
  ws/                   — WebSocket hub for real-time chat
pkg/
  jwt/                  — JWT token manager
  sms/                  — SMS sender (mock in dev)
  storage/              — MinIO S3 client
  cache/                — Redis cache abstraction
```

**Key patterns:**
- Interfaces defined by consumers in `service/deps.go`, not by repositories
- Tests use `testify/mock` — mock structs in `*_test.go` files implement `deps.go` interfaces
- All API routes under `/api/v1/`, admin routes require `middleware.RequireRole("admin")`
- Subscription expiry cron runs every hour via goroutine in `main.go`

### Admin — Next.js App Router

```
admin/src/
  app/
    (admin)/             — Protected admin layout group
      layout.tsx         — Sidebar + header shell
      dashboard/         — Stats overview
      users/             — User list + [id] detail
      listings/          — Listing list + [id] detail
      reports/           — Report management
      subscriptions/     — Subscription management
    login/               — Login page (separate layout)
  components/
    ui/                  — shadcn/ui components (Radix-based)
    admin-sidebar.tsx    — Navigation sidebar
    admin-header.tsx     — Top header bar
  services/api.ts        — All API calls + TypeScript interfaces
  lib/auth.ts            — Auth context (JWT token management)
```

**Key patterns:**
- All pages are `"use client"` with `useAuth()` hook for token
- API calls go through `services/api.ts` — single source of truth for types and endpoints
- UI uses shadcn/ui (Tailwind CSS v4 + Radix UI), indigo-based SaaS color palette
- CSS custom properties in `globals.css` using oklch color space

### Database Schema (PostgreSQL)

Core tables: `users`, `listings`, `subscriptions`, `ratings`, `reports`, `notifications`, `conversations`, `otp_codes`

- `subscriptions` has `status` ('active', 'expired') and `expires_at` — cron expires overdue subs and hides associated listings
- `subscription_expires_at` on User model is a computed field via subquery, not a real column
- Listings have `status` ('active', 'hidden', 'deleted') — hidden when subscription expires, restored when renewed

## Language

The application UI is in Vietnamese. Error messages, labels, and API response messages use Vietnamese text.
