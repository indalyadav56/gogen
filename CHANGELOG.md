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
- Five architectures via `--arch`: `simple`, `layered`, `clean`, `microservice`,
  `monolith`.
- Selectable HTTP router (`gin` | `chi`) across all generated code.
- `microservice` extras: `/healthz`, graceful shutdown, gRPC server stub, and
  `docker-compose.yml`.
- File-based configuration: generated projects ship `configs/config.yaml` loaded
  with [Viper](https://github.com/spf13/viper), with `SERVER_PORT`/`DATABASE_URL`
  environment overrides.
- Documentation under `docs/` (architectures, CLI reference, development guide),
  `LICENSE`, and this changelog.

### Removed
- Legacy `--monolith` / `--gin` / `--auth` flags and the hardcoded scaffold
  engine.
- JWT/RBAC auth templates (to be reintroduced later as an optional module).
- Dead `utils` package, the broken `internal/generator/project` stub, and the
  misnamed `.golintci.yml` (replaced by `.golangci.yml`).
