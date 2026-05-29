package generator

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/indalyadav56/gogen/internal/blueprint"
	"github.com/indalyadav56/gogen/internal/cli"
	"github.com/indalyadav56/gogen/internal/gomod"
	"github.com/indalyadav56/gogen/internal/template"
)

// ProjectGenerator drives the end-to-end generation of a project.
type ProjectGenerator struct {
	config      *cli.Config
	renderer    *template.Renderer
	gomod       *gomod.Manager
	projectRoot string
}

// NewProjectGenerator creates a generator bound to the embedded template FS.
func NewProjectGenerator(config *cli.Config, templateFS embed.FS) *ProjectGenerator {
	root := config.GetProjectRoot()
	return &ProjectGenerator{
		config:      config,
		renderer:    template.NewRenderer(templateFS),
		gomod:       gomod.NewManager(root),
		projectRoot: root,
	}
}

// Generate scaffolds the project files, then initializes and tidies the Go
// module (the latter requires the `go` toolchain and network access).
func (g *ProjectGenerator) Generate() error {
	if err := g.Scaffold(); err != nil {
		return err
	}
	if err := g.gomod.Init(g.config.ModuleName); err != nil {
		return err
	}
	return g.gomod.Tidy()
}

// Scaffold resolves the selected architecture's blueprint and writes every
// file. It performs no toolchain or network operations.
func (g *ProjectGenerator) Scaffold() error {
	bp, ok := blueprint.Get(g.config.Arch)
	if !ok {
		return fmt.Errorf("unknown architecture %q", g.config.Arch)
	}

	repoIsInterface := g.config.Arch != "simple" && g.config.Arch != "layered"
	embedRoutes := g.config.Arch == "simple" || g.config.Arch == "layered"

	for _, spec := range bp.Build(g.config) {
		data := template.Data{
			Module:          g.config.ModuleName,
			Project:         g.projectRoot,
			Arch:            g.config.Arch,
			Router:          g.config.Router,
			DB:              g.config.DB,
			Package:         spec.Package,
			Entity:          spec.Entity,
			Entities:        g.config.Entities,
			Imports:         spec.Imports,
			RepoIsInterface: repoIsInterface,
			EmbedRoutes:     embedRoutes,
		}
		if err := g.writeFile(spec, data); err != nil {
			return fmt.Errorf("generating %s: %w", spec.Path, err)
		}
	}
	return nil
}

// writeFile renders a spec's template, or writes a bare stub when no template
// is set (empty file when Package is also empty).
func (g *ProjectGenerator) writeFile(spec blueprint.FileSpec, data template.Data) error {
	fullPath := filepath.Join(g.projectRoot, spec.Path)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		return err
	}

	if spec.Template != "" {
		return g.renderer.RenderToFile("templates/"+spec.Template, fullPath, data)
	}
	if spec.Package == "" {
		return os.WriteFile(fullPath, nil, 0o644)
	}
	return os.WriteFile(fullPath, []byte("package "+spec.Package+"\n"), 0o644)
}
