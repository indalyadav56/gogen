package blueprint

import "github.com/indalyadav56/gogen/internal/cli"

// buildSimple produces a flat, single-package (package main) project: one file
// per entity holding the model, an in-memory store, and HTTP handlers. No
// layers, no database — ideal for prototypes and small tools.
func buildSimple(cfg *cli.Config) []FileSpec {
	files := []FileSpec{
		{Path: "main.go", Template: "main_simple.tmpl", Package: "main"},
		{Path: "config.go", Template: "config_simple.tmpl", Package: "main"},
		{Path: "configs/config.yaml", Template: "config_yaml.tmpl"},
		{Path: "Dockerfile", Template: "dockerfile.tmpl"},
		{Path: ".gitignore", Template: "gitignore.tmpl"},
		{Path: "README.md", Template: "readme.tmpl"},
	}
	for _, e := range cfg.Entities {
		files = append(files, FileSpec{
			Path:     lower(e) + ".go",
			Template: "simple_resource.tmpl",
			Package:  "main",
			Entity:   e,
		})
	}
	files = append(files, frontendFiles(cfg.Frontend)...)
	return files
}
