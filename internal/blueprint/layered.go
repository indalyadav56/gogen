package blueprint

import "github.com/indalyadav56/gogen/internal/cli"

// buildLayered produces a classic N-tier layout: model / repository / service /
// handler packages under internal/, wired together in cmd/server/main.go. The
// repository is a concrete Postgres type (no domain interface); route
// registration lives on the handler (EmbedRoutes).
func buildLayered(cfg *cli.Config) []FileSpec {
	imp := layeredImports(cfg.ModuleName)
	files := append(commonFiles(),
		FileSpec{Path: "configs/config.yaml", Template: "config_yaml.tmpl"},
		FileSpec{Path: "migrations/.gitkeep", Template: ""},
		FileSpec{Path: "cmd/server/main.go", Template: "main_layered.tmpl", Package: "main", Imports: imp},
	)

	for _, e := range cfg.Entities {
		el := lower(e)
		files = append(files,
			FileSpec{Path: "internal/model/" + el + ".go", Template: "entity.tmpl", Package: "model", Entity: e, Imports: imp},
			FileSpec{Path: "internal/repository/" + el + "_repository.go", Template: "repository_postgres.tmpl", Package: "repository", Entity: e, Imports: imp},
			FileSpec{Path: "internal/service/" + el + "_service.go", Template: "service.tmpl", Package: "service", Entity: e, Imports: imp},
			FileSpec{Path: "internal/handler/" + el + "_handler.go", Template: "handler.tmpl", Package: "handler", Entity: e, Imports: imp},
		)
	}
	if cfg.Auth {
		files = append(files, authFiles(cfg.ModuleName, "migrations")...)
	}
	files = append(files, frontendFiles(cfg.Frontend)...)
	return files
}
