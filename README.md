# backend-core

Go Fiber API Gateway for the NextJS-Go Template. Acts as the single HTTP entry point — authenticates requests, proxies to gRPC microservices, and serves the admin panel.

**No own database.** All data lives in the gRPC services. The backend is stateless except for Redis (rate limiting, session tokens).

---

## What it does

- **Client Registry** (`internal/registry`) — wires all gRPC service clients at startup. Each service registers its HTTP routes and admin CRUD providers automatically.
- **Admin Panel** (`internal/adminconf`) — schema-driven CRUD UI, populated by whatever providers the services register. No frontend changes needed when models are added.
- JWT middleware (validates access tokens, injects user context)
- Redis-backed rate limiting (via auth-service config)
- RFC 7807 error responses

## Structure

```
cmd/server/main.go       ← entry point: env, Redis, registry, Fiber
internal/
  adminconf/             ← admin panel module (JWT + admin middleware)
  registry/              ← client registry: AddAuth, AddGallery, AddCMS, Mount
pkg/
  config/                ← env helpers
```

## Adding a new service

1. Implement the service client with `Name()`, `Providers()`, `RegisterRoutes()`, `Close()`
2. Add `r.AddXxx(addr)` in `internal/registry/registry.go`
3. Call it in `cmd/server/main.go`

The service's providers are auto-registered on the admin panel.

## Environment variables

| Variable | Description |
|---|---|
| `PORT` | HTTP port (default `8080`) |
| `JWT_SECRET` | Shared JWT secret with auth-service |
| `AUTH_SERVICE_ADDR` | auth-service gRPC address |
| `GALLERY_SERVICE_ADDR` | gallery-service gRPC address |
| `CMS_SERVICE_ADDR` | cms-service gRPC address |
| `REDIS_URL` | Redis connection URL |
| `CORS_ALLOW_ORIGINS` | Comma-separated allowed origins |
