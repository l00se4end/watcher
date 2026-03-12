# 🔭 Watcher

A self-hosted uptime monitoring service. Track whether your websites and services are up, get notified when they go down, and share public status pages — all behind your own authentication.

Built as a portfolio project to demonstrate a production-style backend architecture using Go, PostgreSQL, Redis, and Docker.

---

## Architecture

```
Internet
   │
   ▼
┌─────────────────────────────────────────────┐
│  Caddy (Reverse Proxy)                      │
│  - Terminates TLS                           │
│  - Forward Auth: every request is verified  │
│    against Pocket ID before routing         │
│  - Routes /api/* → Go backend               │
│  - Routes /* → Vue.js SPA                   │
└────────────┬──────────────────┬─────────────┘
             │                  │
             ▼                  ▼
     ┌───────────────┐   ┌─────────────┐
     │  Go Backend   │   │  Vue.js SPA │
     │  (Gin)        │   │  (Vite)     │
     └───────┬───────┘   └─────────────┘
             │
     ┌───────┴────────┐
     │                │
     ▼                ▼
┌──────────┐    ┌──────────┐
│ Postgres │    │  Redis   │
│ (data)   │    │ (cache / │
│          │    │  queues) │
└──────────┘    └──────────┘

┌──────────────┐
│  Pocket ID   │  ← Identity Provider (OIDC)
│  (auth.*)    │    Runs on separate subdomain
└──────────────┘
```

### Request lifecycle

1. User visits `watcher.example.com`
2. Caddy intercepts the request and calls `pocket-id:1411/api/verify` (forward auth)
3. If the session is invalid → Pocket ID redirects to login
4. If valid → Pocket ID returns `200` with headers `X-Pocketid-Uid`, `X-Pocketid-User`, `X-Pocketid-Email`
5. Caddy copies those headers and forwards the request to either the Go backend or the Vue SPA
6. The Go backend reads `X-Pocketid-Uid` to identify the user — it never handles passwords or sessions itself

> **Security note:** The Go backend only accepts the `X-Pocketid-Uid` header from requests originating within Docker's internal network (`172.16.0.0/12`). Any attempt to spoof this header from the public internet is rejected at the middleware level before any handler runs.

---

## Tech Stack

| Layer | Technology | Why |
|---|---|---|
| Reverse proxy | [Caddy v2](https://caddyserver.com/) | Automatic HTTPS, clean forward auth support |
| Authentication | [Pocket ID](https://github.com/pocket-id/pocket-id) | Self-hosted OIDC identity provider |
| Backend | Go + [Gin](https://github.com/gin-gonic/gin) | Fast, lightweight, statically compiled |
| Frontend | Vue 3 + Vite | Reactive SPA, served behind the same origin as the API (no CORS) |
| Database | PostgreSQL 16 | Relational data, uptime history |
| Cache / Queue | Redis 7 | Rate limiting, check queues, fast status reads |
| Containers | Docker + Docker Compose | Reproducible local and production environments |

---

## Project Structure

```
watcher/
├── backend/
│   ├── main.go
│   ├── go.mod
│   ├── go.sum
│   ├── infra/             # Docker, Caddy config
│   └── internal/
│       ├── database/
│       │   └── db.go
│       └── server/
│           └── net.go
├── client/                # Vue 3 + Vite (in progress)
├── .dockerignore
├── .gitignore
└── Dockerfile
```

---

## Getting Started

### Prerequisites

- [Docker](https://docs.docker.com/get-docker/) and Docker Compose
- A domain name (or use `localhost` for local development)

### 1. Clone the repository

```bash
git clone https://github.com/yourusername/watcher.git
cd watcher
```

### 2. Configure environment variables

```bash
cp .env.example .env
```

Edit `.env`:

```env
POSTGRES_USER=watcher
POSTGRES_PASSWORD=changeme
POSTGRES_DB=watcher_db

# Pocket ID — see https://pocket-id.org/docs for setup
APP_URL=https://watcher.example.com
```

### 3. Start the stack

```bash
docker compose -f docker/docker-compose.yaml up -d
```

### 4. Set up Pocket ID

Visit `https://auth.example.com` (or `auth.localhost`) and complete the initial setup. Create an OIDC application pointing back to your Watcher instance.

### 5. Run database migrations

```bash
docker exec -it watcher_app ./watcher migrate
```

---

## API Reference

All endpoints require authentication via Caddy's forward auth. Unauthenticated requests never reach the backend.

The authenticated user's ID is injected by Caddy as `X-Pocketid-Uid` on every request.

### `GET /`

Returns all websites the authenticated user has access to.

**Response `200 OK`:**
```json
[
  {
    "id": 1,
    "name": "My Blog",
    "url": "https://example.com",
    "status": "up",
    "last_check": "2025-01-15T10:30:00Z"
  }
]
```

---

### `POST /api/websites`

Add a new website to monitor.

**Request body:**
```json
{
  "name": "My Blog",
  "url": "https://example.com",
  "description": "Personal blog",
  "is_public": false
}
```

**Responses:**

| Status | Meaning |
|---|---|
| `201 Created` | Website added successfully |
| `400 Bad Request` | Missing required fields (`name`, `url`) |
| `409 Conflict` | You are already monitoring this URL |
| `500 Internal Server Error` | Database error |

---

## Docker Network Design

The stack uses two Docker networks to enforce a security boundary:

```
public_net   — Caddy and Pocket ID only (internet-facing)
internal_net — Caddy, Go backend, Postgres, Redis (never exposed to the internet)
```

`watcher-app` is **not** attached to `public_net`. The only way to reach the Go backend is through Caddy, which has already validated the session before forwarding.

---

## Security Considerations

- **No direct backend exposure** — `watcher-app` has no published ports and lives only on `internal_net`
- **Header spoofing prevention** — The Go backend rejects `X-Pocketid-Uid` headers from IPs outside the Docker bridge range (`172.16.0.0/12`), enforced as a Gin middleware before any route handler runs
- **Parameterized queries** — All database operations use `$1, $2...` placeholders, never string interpolation
- **Duplicate protection** — `ON CONFLICT DO NOTHING` with unique constraints prevents race conditions on concurrent inserts
- **Atomic conflict detection** — `RowsAffected()` is checked after every insert so the caller knows whether a row was actually written

---

## Roadmap

- [ ] Uptime check worker (background goroutine polling URLs)
- [ ] Redis-backed check queue and rate limiting
- [ ] Email / webhook notifications on status change
- [ ] Vue.js frontend dashboard
- [ ] Public status pages per website
- [ ] Uptime history and response time graphs
- [ ] Docker health checks for the Go service

---

## License

MIT