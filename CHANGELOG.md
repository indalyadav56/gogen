# Changelog

All notable changes to this project are documented here. The format is based on
[Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## [Unreleased]

### Changed
- **Complete rewrite into a declarative blueprint engine.** Each architecture is
  now defined as data (`internal/blueprint`), rendered from a small shared set of
  templates with architecture-aware import paths.
- New CLI: `--arch`, `--router`, `--db`, repeatable `--entity`. An optional
  `new` subcommand is accepted.
- Generated projects are guaranteed to compile (`go build ./...`); a generator
  test parses every produced `.go` file across all architectures and routers.

### Added
- **Production-grade DDD scaffolding** for `clean`, `microservice` and `monolith`
  (modelled on a real-world reference monolith):
  - `internal/di` composition root + a thin `cmd/server/main` with graceful
    shutdown; an `internal/server/http` router package with `request_id`,
    `logger` and `recovery` middleware.
  - `pkg/response` (uniform JSON envelopes) and `pkg/validator`
    (go-playground/validator), wired into the generated handlers.
  - Per-context `domain` errors and `application` use-case DTOs.
  - Extra binaries `cmd/migrator` (applies `migrations/postgres/*.sql`) and
    `cmd/seed`; layered configs (`configs/{base,local,prod}.yaml` via `APP_ENV`).
- Optional **frontend layer** via `--frontend html|htmx|react` (or the web UI
  dropdown): a `web/` layer the Go binary serves itself (single fullstack
  binary). `html`/`htmx` embed static assets (htmx adds a server-rendered
  fragment endpoint); `react` is a Vite app whose built `dist/` is embedded and
  served with SPA fallback, plus a Node build stage in the Dockerfile. Available
  for every architecture.
- **Structured logging in the gogen tool itself** (`internal/logx`, zap): logs to
  stderr (results stay on stdout), level via `GOGEN_LOG`, with per-request
  logging under `gogen serve`.
- **`gogen serve`** — a local web UI (stdlib `net/http`, no new deps) to pick
  architecture/router/db/entities, preview the structure live, and download the
  project as a `.zip`. Binds to `127.0.0.1`, defaults to an uncommon port
  (`7720`) and falls back to a free OS-assigned port if it's taken.
- Optional **JWT auth + RBAC** via `--auth` (or the web UI toggle): a
  self-contained `internal/auth` module (user/role entities, repositories,
  service, handlers, routes, middleware) plus `pkg/auth` (access+refresh JWT via
  golang-jwt, bcrypt) and an auth migration. Roles map to permissions;
  middleware exposes `RequireAuth` + `RequirePermission`. Available for
  layered/clean/microservice/monolith (not `simple`).
- Five architectures via `--arch`: `simple`, `layered`, `clean`, `microservice`,
  `monolith`.
- Selectable HTTP router (`gin` | `chi`) across all generated code.
- `microservice` extras: `/healthz`, graceful shutdown, gRPC server stub, and
  `docker-compose.yml`.
- File-based configuration: generated projects ship `configs/config.yaml` loaded
  with [Viper](https://github.com/spf13/viper), with `SERVER_PORT`/`DATABASE_URL`/
  `LOG_LEVEL`/`LOG_FILE` environment overrides.
- Structured logging: generated `pkg/logger` is a [zap](https://github.com/uber-go/zap)
  logger with [lumberjack](https://github.com/natefinch/lumberjack) file rotation
  (coloured stdout + rotating JSON file), wired through main, services and
  handlers. The `simple` arch stays dependency-light and is unaffected.
- Documentation under `docs/` (architectures, CLI reference, development guide),
  `LICENSE`, and this changelog.

### Removed
- Legacy `--monolith` / `--gin` / `--auth` flags and the hardcoded scaffold
  engine.
- JWT/RBAC auth templates (to be reintroduced later as an optional module).
- Dead `utils` package, the broken `internal/generator/project` stub, and the
  misnamed `.golintci.yml` (replaced by `.golangci.yml`).
