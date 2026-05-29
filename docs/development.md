# Development & internals

## Project layout

```
gogen/
├── cmd/gogen/main.go        # entry point: parse flags → generate
├── embed.go                 # //go:embed templates/* (must stay at module root)
├── internal/
│   ├── cli/                 # flag parsing + validation (Config, allowed values)
│   ├── blueprint/           # one file per architecture; the "what to generate"
│   ├── template/            # renderer, shared Data/Imports, helper funcs
│   ├── generator/           # orchestration: blueprint → render → go mod
│   └── gomod/               # `go mod init` / `go mod tidy` wrappers
├── templates/               # ~20 reusable text/templates
└── docs/
```

## How a generation flows

```
cli.ParseFlags() ──▶ Config
                       │
generator.Generate() ──┤
                       ├─ blueprint.Get(arch).Build(cfg) ──▶ []FileSpec
                       ├─ for each FileSpec: renderer.RenderToFile(tmpl, data)
                       └─ gomod.Init + gomod.Tidy
```

- **`FileSpec`** (in [`internal/blueprint/blueprint.go`](../internal/blueprint/blueprint.go))
  describes one output file: target `Path`, `Template`, `Package`, owning
  `Entity`, and the `Imports` for that file.
- **`template.Data`** is the single, architecture-agnostic payload passed to
  every template. Templates reference packages through **import aliases**
  (`svc`, `repo`, `ent`, `hdl`) populated from `Data.Imports`, so the same
  `handler.tmpl`/`service.tmpl` render correctly in any layout.
- Two booleans switch template behaviour: `RepoIsInterface` (clean/micro/
  monolith use a domain interface; layered uses a concrete repo) and
  `EmbedRoutes` (simple/layered put route registration on the handler; the
  others use a dedicated `routes` package).

## Adding an architecture

1. **Templates** — add `templates/main_<name>.tmpl` and any layout-specific
   templates. Reuse the shared business templates (`entity`, `repository`,
   `service`, `handler`, `routes`, `dto`) where possible.
2. **Blueprint** — add `internal/blueprint/<name>.go` with a
   `build<Name>(cfg *cli.Config) []FileSpec`. Use `commonFiles()` for shared
   scaffolding and `contextFiles(...)` for clean-style per-entity files.
3. **Register** — add the entry to `Registry` in `blueprint.go` and add
   `"<name>"` to `cli.Architectures` in `internal/cli/flags.go`.
4. **Tests** — add a marker file to the maps in `blueprint_test.go` and
   `generator_test.go`; the generator test already parses every generated
   `.go` file for the new arch across both routers.

That's it — no engine changes are needed.

## Helper functions in templates

| Func | Example | Result |
|------|---------|--------|
| `ToPascalCase` | `{{.Entity \| ToPascalCase}}` | `User` |
| `ToCamelCase` | `{{.Entity \| ToCamelCase}}` | `user` |
| `ToLower` / `ToUpper` | `{{.Entity \| ToLower}}` | `user` |
| `Plural` | `{{.Entity \| ToLower \| Plural}}` | `users`, `categories`, `boxes` |

## Building & testing

```bash
go build ./...        # build the tool
go test ./...         # unit tests + per-architecture parse checks
go vet ./...
golangci-lint run     # uses .golangci.yml

task run              # generate an example project
```

The generator test (`internal/generator`) scaffolds **every architecture × both
routers** into a temp dir and parses each generated `.go` file with `go/parser`,
so a template that produces invalid Go fails the suite immediately.
