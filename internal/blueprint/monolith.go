package blueprint

import "github.com/indalyadav56/gogen/internal/cli"

// buildMonolith produces a DDD modular monolith: each entity is an independent
// bounded context under internal/<entity>/ with its own clean layering, all
// wired into a single binary in cmd/server/main.go.
func buildMonolith(cfg *cli.Config) []FileSpec {
	files := commonFiles()
	files = append(files, richServerFiles()...)
	files = append(files,
		FileSpec{Path: "cmd/server/main.go", Template: "main_server.tmpl", Package: "main"},
		FileSpec{Path: "internal/di/container.go", Template: "di_container_monolith.tmpl", Package: "di"},
		FileSpec{Path: "internal/server/http/router.go", Template: "router_monolith.tmpl", Package: "http"},
		FileSpec{Path: "internal/shared/shared.go", Template: "", Package: "shared"},
	)

	for _, e := range cfg.Entities {
		el := lower(e)
		base := "internal/" + el
		dirs := contextDirs{
			entity:  base + "/domain/entity",
			repo:    base + "/domain/repository",
			app:     base + "/application",
			infra:   base + "/infrastructure/postgres",
			handler: base + "/transport/http/v1/handlers",
			routes:  base + "/transport/http/v1/routes",
			dto:     base + "/transport/http/v1/dto",
		}
		files = append(files, contextFiles(el, e, monolithImports(cfg.ModuleName, el), dirs)...)
	}
	if cfg.Auth {
		files = append(files, authFiles(cfg.ModuleName, "migrations/postgres")...)
	}
	files = append(files, frontendFiles(cfg.Frontend)...)
	return files
}
