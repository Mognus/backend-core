# backend-core

Go Fiber API Gateway. Single HTTP entry point — validates JWTs, proxies to gRPC microservices, serves the admin panel.

**No own database.** All data lives in the gRPC services. Stateless except for Redis (rate limiting, session tokens).

---

## Architecture

### Auth as shared config

`auth-service` is the only required service. Its `*Config` is created once and passed everywhere that needs middleware:

```
authclient.New(addr, secret, redis)
  └─ *Config{jwtSecret, *Middleware}
       ├─ adminconf.New(config)     → admin panel uses JWT + admin check
       └─ blogclient.New(addr, config) → blog service uses same middleware on its routes
```

This means there is one JWT secret, one middleware instance, shared across all services and the admin panel.

### Request flow

```
POST /auth/login
  → authHandler → gRPC → auth-service → JWT cookie set

GET /auth/me
  → JWTMiddleware()       validates token → sets c.Locals("user")
  → RequireAuth           checks c.Locals("user") != nil
  → authHandler.me()      reads userID via GetUserIDFromContext(c)

GET /admin/api/users
  → JWTMiddleware()       (registered on /admin/* group by adminconf)
  → RequireAdmin          checks claims["role"] == "admin"
  → UserProvider.HandleList() → gRPC → auth-service
```

### Service Registry

The service registry mounts service routes when services are registered and keeps lifecycle handling in one place. Admin wiring stays in `adminconf`:

```
serviceregistry.New(router)
  authSvc, _ := authclient.New(...)
  adminconf.New(authSvc.Config)    → admin panel initialized with auth config
  admin.Mount(router)              → /admin/* routes registered
  admin.RegisterProviders(...svc)  → CRUD providers mount their own /admin/api/<model> routes
  services.RegisterServices(...svc) → one or many services get routes + lifecycle + health names

  admin.RegisterProviders(authSvc, blogSvc, ...)
  services.RegisterServices(authSvc, blogSvc, ...)
```

Admin CRUD routes are mounted per model namespace during provider registration, while the service registry handles service routes and lifecycle.

### Protecting routes

Services receive `*authclient.Config` in their constructor and use it directly:

```go
// single route
router.Get("/me", h.config.JWTMiddleware(), h.config.RequireAuth, h.me)

// group
protected := router.Group("/posts")
protected.Use(h.config.JWTMiddleware(), h.config.RequireAuth)
protected.Post("/", h.create)
```

Reading user data from a handler after JWT middleware ran:

```go
userID, err := authclient.GetUserIDFromContext(c)
role, err   := authclient.GetUserRoleFromContext(c)
```

---

## Structure

```
cmd/server/
  main.go      ← startup: Redis, service registry, health route, listen
  server.go    ← newApp(), errorHandler(), getEnv(), isEnabled()
internal/
  adminconf/   ← admin panel: JWT + admin middleware, provider-mounted CRUD routes
  registry/    ← service registry: RegisterServices, Close, Names
```

## Adding a new service

1. Implement `Name()` and `Close()` — satisfies `registry.Service`
2. Optionally implement `RegisterRoutes()` if the service exposes Fiber routes
3. Optionally implement `Providers()` if the service contributes admin CRUD providers
4. Add to `loadServices()` in `cmd/server/main.go`:
```go
admin := adminconf.New(authSvc.Config)
admin.Mount(api)

if isEnabled("blog") {
    blogSvc, err := blogclient.New(getEnv("BLOG_SERVICE_ADDR", "localhost:50052"), authSvc.Config)
    if err != nil { log.Fatalf(...) }
    admin.RegisterProviders(blogSvc)
    services.RegisterServices(blogSvc)
}
```
5. Set `ENABLED_SERVICES=blog` in `.env`

The service's providers are auto-registered on the admin panel — no further changes needed.

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `8080` | HTTP port |
| `HOST` | `0.0.0.0` | Bind address |
| `JWT_SECRET` | — | Shared JWT secret with auth-service |
| `AUTH_SERVICE_ADDR` | `localhost:50051` | auth-service gRPC address |
| `ENABLED_SERVICES` | — | Comma-separated optional services, e.g. `blog,gallery` |
| `REDIS_URL` | `redis://localhost:6379` | Redis connection URL |
| `CORS_ALLOW_ORIGINS` | `http://localhost:3000` | Allowed CORS origins |
