# CLI reference

```
gogen [new] --module <path> [--arch <arch>] [--router <router>] [--db <db>] [--entity <Name> ...]
```

The optional `new` subcommand is accepted and ignored, so both
`gogen --module …` and `gogen new --module …` work.

## Flags

| Flag | Default | Allowed | Description |
|------|---------|---------|-------------|
| `--module` | `github.com/username/project` | any module path | Go module path. The project directory is the last path segment (`…/shop` → `shop/`). |
| `--arch` | `clean` | `simple`, `layered`, `clean`, `microservice`, `monolith` | Project architecture. See [architectures.md](architectures.md). |
| `--router` | `gin` | `gin`, `chi` | HTTP router used in generated handlers, routes, and `main`. |
| `--db` | `postgres` | `postgres` | Database driver wired into `pkg/db` and config. |
| `--entity` | `Item` | any identifier; repeatable | Resource to scaffold. Repeat for several: `--entity User --entity Order`. Comma-separated values are also accepted. |

Invalid values for `--arch`, `--router`, or `--db` fail fast with a clear error.

## Examples

```bash
# Clean architecture (defaults: gin + postgres), two entities
gogen --module github.com/acme/shop --entity User --entity Product

# Layered CRUD service with chi
gogen --module github.com/acme/billing --arch layered --router chi --entity Invoice

# Deployable microservice (gRPC, /healthz, docker-compose)
gogen --module github.com/acme/payments --arch microservice --entity Payment

# DDD modular monolith with three bounded contexts
gogen --module github.com/acme/erp --arch monolith \
  --entity Customer --entity Order --entity Product

# Throwaway prototype, in-memory
gogen --module github.com/acme/spike --arch simple --entity Note
```

## What happens

1. The blueprint for `--arch` produces the file list.
2. Each file is rendered from a shared template with architecture-aware imports.
3. `go mod init <module>` then `go mod tidy` run in the new project directory
   (requires the Go toolchain and network access to fetch dependencies).

After generation:

```bash
cd <project>
# Edit configs/config.yaml, or set SERVER_PORT / DATABASE_URL to override.
go run ./cmd/server      # or `go run .` for simple
```

Configuration is loaded with [Viper](https://github.com/spf13/viper) from
`configs/config.yaml`, with environment variables (`SERVER_PORT`,
`DATABASE_URL`) taking precedence. See
[architectures.md → Configuration](architectures.md#configuration).
