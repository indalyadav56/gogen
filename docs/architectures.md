# Architectures

`gogen` generates five layouts. Pick one with `--arch`. Every layout compiles
out of the box, serves its HTTP API under `/api/v1`, and supports multiple
entities (repeat `--entity`).

The examples below were generated with:

```bash
gogen --module github.com/acme/demo --arch <arch> --entity User --entity Product
```

---

## `simple`

Flat, single `package main`. One file per entity holding the model, an
in-memory store, and HTTP handlers. No database, no layers — ideal for
prototypes, spikes, and small tools.

```
demo/
├── main.go            # router + server bootstrap
├── config.go          # loads configs/config.yaml via Viper
├── configs/config.yaml
├── user.go            # User model + in-memory store + handlers
├── product.go         # Product model + in-memory store + handlers
├── Dockerfile
├── .gitignore
├── README.md
└── go.mod
```

**Wiring:** `main.go` calls `registerUserRoutes(group)` for each entity.

---

## `layered`

Classic N-tier. Dependencies flow `handler → service → repository → model`.
The repository is a concrete Postgres type (no domain interface); route
registration lives on the handler (`RegisterRoutes`).

```
demo/
├── cmd/server/main.go
├── internal/
│   ├── model/         # User, Product structs
│   ├── repository/    # *UserRepository (Postgres), *ProductRepository
│   ├── service/       # UserService, ProductService
│   └── handler/       # UserHandler.RegisterRoutes, ProductHandler.RegisterRoutes
├── config/config.go   # Viper loader
├── configs/config.yaml
├── pkg/
│   ├── db/db.go       # Postgres connection
│   └── logger/logger.go
├── migrations/
├── Dockerfile · Taskfile.yaml · .env.example · README.md · go.mod
```

---

## `clean`

Clean / Hexagonal architecture for a single service. The domain owns the
repository **interface**; infrastructure implements it; a dedicated `routes`
package keeps transport wiring out of the handlers.

```
demo/
├── cmd/
│   ├── server/main.go     # thin: config → logger → di.New → router → serve
│   ├── migrator/main.go   # applies migrations/postgres/*.sql
│   └── seed/main.go       # dev seed data
├── internal/
│   ├── di/container.go    # composition root (wires repos → services → handlers)
│   ├── server/http/
│   │   ├── router.go      # NewRouter(c *di.Container)
│   │   └── middleware/    # request_id, logger, recovery
│   ├── db/migrate.go      # migration runner
│   ├── domain/
│   │   ├── entity/         # User/Product + <entity>_errors.go
│   │   └── repository/     # repository interfaces
│   ├── application/        # services + <entity>_dto.go (use-case inputs)
│   ├── infrastructure/postgres/
│   └── transport/http/v1/{handlers,routes,dto}/
├── configs/{base,local,prod}.yaml      # layered config (APP_ENV)
├── migrations/postgres/000001_init.sql
├── pkg/{db,logger,response,validator}/
└── Dockerfile · Taskfile.yaml · .env.example · README.md · go.mod
```

**Dependency rule:** `transport → application → domain`. Infrastructure depends
inward on the domain interface; nothing depends on infrastructure except the DI
container (`internal/di`), which `main` and the router consume.

---

## `microservice`

Everything in `clean`, plus production extras for an independently deployable
service:

- `/healthz` liveness endpoint
- graceful shutdown on `SIGINT`/`SIGTERM` via `http.Server.Shutdown`
- a minimal gRPC server (`:9090`) ready for your generated stubs
- `docker-compose.yml` (app + Postgres)

```
demo/
├── cmd/server/main.go              # HTTP + gRPC, graceful shutdown
├── internal/
│   ├── domain/ · application/ · infrastructure/ · transport/http/v1/   # as clean
│   ├── health/health.go            # /healthz
│   └── transport/grpc/server.go    # gRPC server wrapper
├── proto/                          # put your .proto files here
├── docker-compose.yml
├── config/ · configs/ · pkg/ · migrations/
└── Dockerfile · Taskfile.yaml · .env.example · README.md · go.mod
```

---

## `monolith`

A DDD modular monolith: each entity is an **independent bounded context** under
`internal/<entity>/` with its own clean layering, all wired into one binary.
Add a context by adding an `--entity`.

```
demo/
├── cmd/server/main.go              # wires every context onto /api/v1
├── internal/
│   ├── shared/                     # cross-context helpers
│   ├── user/
│   │   ├── domain/{entity,repository}/
│   │   ├── application/
│   │   ├── infrastructure/postgres/
│   │   └── transport/http/v1/{handlers,routes,dto}/
│   └── product/
│       └── … (same structure)
├── config/ · configs/ · pkg/{db,logger}/ · migrations/
└── Dockerfile · Taskfile.yaml · .env.example · README.md · go.mod
```

**Boundary rule:** contexts never import each other's internals. Cross-context
collaboration belongs in `internal/shared/` or via ports wired in `main.go`.

---

## Production scaffolding (clean · microservice · monolith)

The DDD architectures ship extra structure modelled on real-world services:

- **DI composition root** (`internal/di`) wires repositories → services →
  handlers; `cmd/server/main` stays thin and does graceful shutdown.
- **Router package** (`internal/server/http`) builds the router from the
  container and applies `request_id`, `logger` and `recovery` middleware.
- **`pkg/response`** — uniform `{success,data,error}` JSON envelopes — and
  **`pkg/validator`** (go-playground/validator) are wired into the handlers.
- **Domain errors** (`<entity>_errors.go`) and **use-case DTOs**
  (`application/<entity>_dto.go`) per context.
- **Extra binaries**: `cmd/migrator` applies `migrations/postgres/*.sql` (in
  lexical order); `cmd/seed` is a stub seeder.
- **Layered config**: `configs/base.yaml` merged with `configs/<APP_ENV>.yaml`
  (default `local`), then environment variables.

`simple` and `layered` stay deliberately lean (no DI/router package).

---

## Configuration

Config is loaded with [Viper](https://github.com/spf13/viper) into a typed
`Config{ Server, Database, Log }` by `config/config.go` (`simple` uses an inline
`loadConfig()`). Precedence is defaults → YAML → environment variables
(`SERVER_PORT`, `DATABASE_URL`, `LOG_LEVEL`, …).

- **simple / layered** read a single `configs/config.yaml`.
- **clean / microservice / monolith** read `configs/base.yaml` then merge
  `configs/<APP_ENV>.yaml` (`APP_ENV` default `local`).

```yaml
# configs/base.yaml (or config.yaml)
server:
  port: "8080"
database:
  url: "postgres://postgres:postgres@localhost:5432/demo?sslmode=disable"
log:
  level: "info"          # debug | info | warn | error
  file: "logs/app.log"   # rotated by lumberjack
```

Logging is handled by `pkg/logger` — a [zap](https://github.com/uber-go/zap)
logger that prints coloured logs to stdout and writes rotating JSON logs to
`log.file` via [lumberjack](https://github.com/natefinch/lumberjack). It's
installed as zap's global logger and used across `main`, services and handlers.

---

## Authentication (`--auth`)

Adding `--auth` (or ticking the toggle in the web UI) scaffolds a self-contained
JWT + RBAC module. It uses the same DDD layout in every structured architecture
(not available for `simple`):

```
internal/auth/
├── domain/{entity,repository}/   # User, Role + repository interfaces
├── application/auth_service.go   # Register / Login / Refresh
├── infrastructure/postgres/      # repository adapters (stubs)
└── transport/http/
    ├── v1/{handlers,routes,dto}/ # /auth endpoints
    └── middleware/               # RequireAuth + RequirePermission
pkg/auth/{jwt.go,password.go}      # access+refresh JWT (golang-jwt), bcrypt
migrations/000001_create_auth_tables.up.sql
```

- **Endpoints** (mounted under `/api/v1/auth`): `POST /register`, `POST /login`,
  `POST /refresh` (public); `GET /me` (any authenticated user); `GET /admin`
  (requires the `users:read` permission — an RBAC example).
- **RBAC**: users have a role; roles map to permissions. `Login` embeds the
  caller's permissions in the access token; `RequirePermission("…")` enforces them.
- **Config**: `auth.secret` / `auth.access_ttl` / `auth.refresh_ttl`
  (env: `AUTH_SECRET`, `AUTH_ACCESS_TTL`, `AUTH_REFRESH_TTL`).
- The repository methods are stubs (return `nil`) — implement the SQL against the
  generated migration to make login persistent.

---

## Frontend (`--frontend`)

`--frontend html|htmx|react` adds a `web/` layer that the **Go binary serves
itself** (single fullstack binary). Available for every architecture.

| Value | What you get | How it's served |
|-------|--------------|-----------------|
| `none` | Backend only (default) | — |
| `html` | `web/index.html` + `web/static/` + a "Ping the API" demo | embedded static files via `web.Handler()` |
| `htmx` | Same, plus htmx (CDN) and a server-rendered fragment endpoint (`/ui/hello`) | embedded static + Go fragment handler |
| `react` | A Vite + React app (`web/`) with `npm` scripts | embedded `web/dist` with SPA fallback |

The frontend is mounted on the router for non-API paths (gin `NoRoute`, chi
`Handle("/*")`), so `/api/v1/...` keeps working and everything else hits the UI.

```bash
# html / htmx — nothing to build, just run:
go run ./cmd/server          # open http://localhost:8080

# react — build once, then the Go server serves it:
cd web && npm install && npm run build && cd ..
go run ./cmd/server
# during development: cd web && npm run dev   (Vite proxies /api to :8080)
```

For `react`, a placeholder `web/dist/index.html` ships so the binary always
compiles before the first `npm run build`; the Dockerfile gains a Node stage that
builds the frontend automatically.

---

## Choosing one

| Need | Use |
|------|-----|
| Throwaway prototype, no DB | `simple` |
| Straightforward CRUD service | `layered` |
| Testable core, swappable infrastructure | `clean` |
| Deployable service with gRPC + ops glue | `microservice` |
| Several domains, one deploy, room to split later | `monolith` |
