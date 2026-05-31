package template

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

// Renderer handles template rendering operations against an embedded FS.
type Renderer struct {
	templateFS embed.FS
}

// NewRenderer creates a new template renderer backed by the given embedded FS.
func NewRenderer(templateFS embed.FS) *Renderer {
	return &Renderer{templateFS: templateFS}
}

// Imports holds the per-architecture package import paths consumed by the
// shared business templates (handler/service/repository/...). Populating these
// in the blueprint keeps the templates architecture-agnostic.
type Imports struct {
	Domain     string // entities + repository interfaces (or just entities)
	Entity     string // entity package import path
	Repository string // repository interface import path
	Service    string // application/service import path
	Infra      string // infrastructure (postgres) import path
	Handler    string // http handler import path
	Routes     string // http routes import path
	DTO        string // dto import path
	Middleware string // http middleware import path (auth)
}

// Data is the single, architecture-agnostic payload passed to every template.
type Data struct {
	Module   string // go module path, e.g. github.com/you/proj
	Project  string // project root / binary name
	Arch     string // simple|layered|clean|microservice|monolith
	Router   string // gin|chi
	DB       string // postgres
	Package  string // Go package name of the file being rendered
	Entity   string // current entity (when rendering a per-entity file)
	Entities []string
	Imports  Imports

	// RepoIsInterface is true for architectures with a domain repository
	// interface (clean/microservice/monolith) and false where the postgres
	// repository is used concretely (layered).
	RepoIsInterface bool
	// EmbedRoutes is true when route registration lives on the handler
	// (simple/layered) instead of a dedicated routes package.
	EmbedRoutes bool
	// Auth is true when the JWT + RBAC auth module is included.
	Auth bool
	// Frontend is the frontend layer: none|html|htmx|react.
	Frontend string
	// Rich is true for the production-grade DDD architectures (clean,
	// microservice, monolith) that get DI containers, a router package,
	// pkg/response + pkg/validator, domain errors and extra binaries.
	Rich bool
}

// FuncMap returns the template helper functions shared by all templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"ToLower":      strings.ToLower,
		"ToUpper":      strings.ToUpper,
		"ToCamelCase":  toCamel,
		"ToPascalCase": toPascal,
		"Plural":       plural,
	}
}

func toCamel(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func toPascal(s string) string {
	if s == "" {
		return s
	}
	return strings.ToUpper(s[:1]) + s[1:]
}

// plural returns a naive English plural — good enough for generated identifiers.
func plural(s string) string {
	if s == "" {
		return s
	}
	switch {
	case strings.HasSuffix(s, "y") && !endsInVowelY(s):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"),
		strings.HasSuffix(s, "z"), strings.HasSuffix(s, "ch"),
		strings.HasSuffix(s, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

func endsInVowelY(s string) bool {
	if len(s) < 2 {
		return false
	}
	switch s[len(s)-2] {
	case 'a', 'e', 'i', 'o', 'u', 'A', 'E', 'I', 'O', 'U':
		return true
	}
	return false
}

// RenderToFile renders the named embedded template to outputPath.
func (r *Renderer) RenderToFile(templatePath, outputPath string, data Data) error {
	templatePath = strings.ReplaceAll(templatePath, "\\", "/")

	content, err := r.templateFS.ReadFile(templatePath)
	if err != nil {
		return fmt.Errorf("failed to read embedded template %s: %w", templatePath, err)
	}

	tmpl, err := template.New(filepath.Base(templatePath)).Funcs(FuncMap()).Parse(string(content))
	if err != nil {
		return fmt.Errorf("failed to parse template %s: %w", templatePath, err)
	}

	file, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file %s: %w", outputPath, err)
	}
	defer file.Close()

	if err := tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("failed to execute template %s: %w", templatePath, err)
	}
	return nil
}
