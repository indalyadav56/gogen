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
├── cmd/server/main.go
├── internal/
│   ├── domain/
│   │   ├── entity/         # User, Product
│   │   └── repository/     # UserRepository interface, ProductRepository interface
│   ├── application/        # UserService, ProductService (use cases)
│   ├── infrastructure/
│   │   └── postgres/       # interface implementations
│   └── transport/http/v1/
│       ├── handlers/       # HTTP handlers
│       ├── routes/         # RegisterUserRoutes, RegisterProductRoutes
│       └── dto/            # request/response DTOs
├── config/ · configs/ · pkg/{db,logger}/ · migrations/
└── Dockerfile · Taskfile.yaml · .env.example · README.md · go.mod
```

**Dependency rule:** `transport → application → domain`. Infrastructure depends
inward on the domain interface; nothing depends on infrastructure except the
composition root (`main.go`).

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

## Configuration

Every layout ships a `configs/config.yaml` loaded with
[Viper](https://github.com/spf13/viper):

- db-backed layouts load `config/config.go` → `config.Load()` returning a typed
  `Config{ Server, Database }`; `simple` uses an inline `loadConfig()`.
- Values fall back to built-in defaults, then `configs/config.yaml`, then
  environment variables (`SERVER_PORT`, `DATABASE_URL`) which take precedence.
- The config file is optional — delete it and defaults + env still work.

```yaml
# configs/config.yaml
server:
  port: "8080"
database:
  url: "postgres://postgres:postgres@localhost:5432/demo?sslmode=disable"
```

---

## Choosing one

| Need | Use |
|------|-----|
| Throwaway prototype, no DB | `simple` |
| Straightforward CRUD service | `layered` |
| Testable core, swappable infrastructure | `clean` |
| Deployable service with gRPC + ops glue | `microservice` |
| Several domains, one deploy, room to split later | `monolith` |
