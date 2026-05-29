package blueprint

import (
	"github.com/indalyadav56/gogen/internal/cli"
	"github.com/indalyadav56/gogen/internal/template"
)

// contextDirs holds the directory each layer of a clean-style context lives in.
type contextDirs struct {
	entity, repo, app, infra, handler, routes, dto string
}

// contextFiles emits the seven per-entity files shared by the clean,
// microservice, and monolith layouts. Only the directories and import paths
// differ between them.
func contextFiles(el, e string, imp template.Imports, d contextDirs) []FileSpec {
	return []FileSpec{
		{Path: d.entity + "/" + el + ".go", Template: "entity.tmpl", Package: "entity", Entity: e, Imports: imp},
		{Path: d.repo + "/" + el + "_repository.go", Template: "repository.tmpl", Package: "repository", Entity: e, Imports: imp},
		{Path: d.app + "/" + el + "_service.go", Template: "service.tmpl", Package: "application", Entity: e, Imports: imp},
		{Path: d.infra + "/" + el + "_postgres.go", Template: "repository_postgres.tmpl", Package: "postgres", Entity: e, Imports: imp},
		{Path: d.handler + "/" + el + "_handler.go", Template: "handler.tmpl", Package: "handlers", Entity: e, Imports: imp},
		{Path: d.routes + "/" + el + "_routes.go", Template: "routes.tmpl", Package: "routes", Entity: e, Imports: imp},
		{Path: d.dto + "/" + el + "_dto.go", Template: "dto.tmpl", Package: "dto", Entity: e, Imports: imp},
	}
}

// buildClean produces a Clean/Hexagonal single service: domain (entity +
// repository interface), application (services), infrastructure (postgres
// adapter), and interface (http handlers/routes/dto). A dedicated routes
// package keeps transport wiring out of the handlers.
func buildClean(cfg *cli.Config) []FileSpec {
	imp := cleanImports(cfg.ModuleName)
	dirs := contextDirs{
		entity:  "internal/domain/entity",
		repo:    "internal/domain/repository",
		app:     "internal/application",
		infra:   "internal/infrastructure/postgres",
		handler: "internal/transport/http/v1/handlers",
		routes:  "internal/transport/http/v1/routes",
		dto:     "internal/transport/http/v1/dto",
	}
	files := append(commonFiles(), FileSpec{
		Path: "cmd/server/main.go", Template: "main_clean.tmpl", Package: "main", Imports: imp,
	})
	for _, e := range cfg.Entities {
		files = append(files, contextFiles(lower(e), e, imp, dirs)...)
	}
	return files
}
