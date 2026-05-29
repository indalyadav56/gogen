// Package blueprint declares each supported architecture as data: a Build
// function that turns a CLI config into a flat list of files to render. Adding
// or changing an architecture means editing one file here plus its templates —
// no engine changes required.
package blueprint

import (
	"strings"

	"github.com/indalyadav56/gogen/internal/cli"
	"github.com/indalyadav56/gogen/internal/template"
)

// FileSpec describes one file to generate.
type FileSpec struct {
	Path     string           // target path relative to the project root
	Template string           // template name under templates/; "" => bare stub
	Package  string           // Go package name (for stubs / package declarations)
	Entity   string           // entity this file belongs to ("" for shared files)
	Imports  template.Imports // import paths for the file's architecture/entity
}

// Blueprint is a named architecture able to build its file list from a config.
type Blueprint struct {
	Name  string
	Build func(cfg *cli.Config) []FileSpec
}

// Registry maps architecture names to their blueprints.
var Registry = map[string]Blueprint{
	"simple":       {Name: "simple", Build: buildSimple},
	"layered":      {Name: "layered", Build: buildLayered},
	"clean":        {Name: "clean", Build: buildClean},
	"microservice": {Name: "microservice", Build: buildMicroservice},
	"monolith":     {Name: "monolith", Build: buildMonolith},
}

// Get returns the blueprint for an architecture name.
func Get(name string) (Blueprint, bool) {
	bp, ok := Registry[name]
	return bp, ok
}

func lower(s string) string { return strings.ToLower(s) }

// commonFiles are shared scaffolding files used by every non-simple layout.
func commonFiles() []FileSpec {
	return []FileSpec{
		{Path: "config/config.go", Template: "config.tmpl", Package: "config"},
		{Path: "configs/config.yaml", Template: "config_yaml.tmpl"},
		{Path: "pkg/db/db.go", Template: "pkg_db.tmpl", Package: "db"},
		{Path: "pkg/logger/logger.go", Template: "pkg_logger.tmpl", Package: "logger"},
		{Path: "Dockerfile", Template: "dockerfile.tmpl"},
		{Path: "Taskfile.yaml", Template: "taskfile.tmpl"},
		{Path: ".gitignore", Template: "gitignore.tmpl"},
		{Path: ".env.example", Template: "env.tmpl"},
		{Path: "README.md", Template: "readme.tmpl"},
		{Path: "migrations/.gitkeep", Template: ""},
	}
}

// Import-path sets per architecture. Shared business templates reference these
// via aliases, so the same template renders correctly in any layout.

func cleanImports(mod string) template.Imports {
	return template.Imports{
		Entity:     mod + "/internal/domain/entity",
		Repository: mod + "/internal/domain/repository",
		Service:    mod + "/internal/application",
		Infra:      mod + "/internal/infrastructure/postgres",
		Handler:    mod + "/internal/transport/http/v1/handlers",
		Routes:     mod + "/internal/transport/http/v1/routes",
		DTO:        mod + "/internal/transport/http/v1/dto",
	}
}

func layeredImports(mod string) template.Imports {
	return template.Imports{
		Entity:     mod + "/internal/model",
		Repository: mod + "/internal/repository",
		Service:    mod + "/internal/service",
		Infra:      mod + "/internal/repository",
		Handler:    mod + "/internal/handler",
		Routes:     mod + "/internal/handler",
	}
}

func monolithImports(mod, entityLower string) template.Imports {
	base := mod + "/internal/" + entityLower
	return template.Imports{
		Entity:     base + "/domain/entity",
		Repository: base + "/domain/repository",
		Service:    base + "/application",
		Infra:      base + "/infrastructure/postgres",
		Handler:    base + "/transport/http/v1/handlers",
		Routes:     base + "/transport/http/v1/routes",
		DTO:        base + "/transport/http/v1/dto",
	}
}
