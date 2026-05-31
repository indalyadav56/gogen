package blueprint

import (
	"github.com/indalyadav56/gogen/internal/cli"
	"github.com/indalyadav56/gogen/internal/template"
)

// contextDirs holds the directory each layer of a clean-style context lives in.
type contextDirs struct {
	entity, repo, app, infra, handler, routes, dto string
}

// contextFiles emits the per-entity files shared by the clean, microservice and
// monolith layouts. Only the directories and import paths differ between them.
func contextFiles(el, e string, imp template.Imports, d contextDirs) []FileSpec {
	return []FileSpec{
		{Path: d.entity + "/" + el + ".go", Template: "entity.tmpl", Package: "entity", Entity: e, Imports: imp},
		{Path: d.entity + "/" + el + "_errors.go", Template: "domain_errors.tmpl", Package: "entity", Entity: e, Imports: imp},
		{Path: d.repo + "/" + el + "_repository.go", Template: "repository.tmpl", Package: "repository", Entity: e, Imports: imp},
		{Path: d.app + "/" + el + "_service.go", Template: "service.tmpl", Package: "application", Entity: e, Imports: imp},
		{Path: d.app + "/" + el + "_dto.go", Template: "application_dto.tmpl", Package: "application", Entity: e, Imports: imp},
		{Path: d.infra + "/" + el + "_postgres.go", Template: "repository_postgres.tmpl", Package: "postgres", Entity: e, Imports: imp},
		{Path: d.handler + "/" + el + "_handler.go", Template: "handler_rich.tmpl", Package: "handlers", Entity: e, Imports: imp},
		{Path: d.routes + "/" + el + "_routes.go", Template: "routes.tmpl", Package: "routes", Entity: e, Imports: imp},
		{Path: d.dto + "/" + el + "_dto.go", Template: "dto.tmpl", Package: "dto", Entity: e, Imports: imp},
	}
}

// buildClean produces a Clean/Hexagonal service with a DI composition root, a
// dedicated router package, standardized response/validation, domain errors,
// and migrator/seeder binaries.
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

	files := commonFiles()
	files = append(files, richServerFiles()...)
	files = append(files,
		FileSpec{Path: "cmd/server/main.go", Template: "main_server.tmpl", Package: "main"},
		FileSpec{Path: "internal/di/container.go", Template: "di_container.tmpl", Package: "di"},
		FileSpec{Path: "internal/server/http/router.go", Template: "router.tmpl", Package: "http"},
	)
	for _, e := range cfg.Entities {
		files = append(files, contextFiles(lower(e), e, imp, dirs)...)
	}
	if cfg.Auth {
		files = append(files, authFiles(cfg.ModuleName, "migrations/postgres")...)
	}
	files = append(files, frontendFiles(cfg.Frontend)...)
	return files
}
