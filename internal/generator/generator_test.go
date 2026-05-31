package generator

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strings"
	"testing"

	goembed "github.com/indalyadav56/gogen"
	"github.com/indalyadav56/gogen/internal/cli"
)

// TestScaffoldArchitectures generates each architecture into a temp dir and
// asserts every produced .go file parses, then checks a marker file exists.
func TestScaffoldArchitectures(t *testing.T) {
	markers := map[string]string{
		"simple":       "user.go",
		"layered":      "internal/service/user_service.go",
		"clean":        "internal/application/user_service.go",
		"microservice": "internal/transport/grpc/server.go",
		"monolith":     "internal/user/application/user_service.go",
	}

	for _, arch := range cli.Architectures {
		for _, router := range cli.Routers {
			t.Run(arch+"_"+router, func(t *testing.T) {
				root := t.TempDir()
				cfg := &cli.Config{
					ModuleName: "github.com/test/demo",
					Arch:       arch,
					Router:     router,
					DB:         "postgres",
					Entities:   []string{"User", "Product"},
				}

				g := NewProjectGenerator(cfg, goembed.TemplateFS)
				g.projectRoot = root // write directly into the temp dir

				if err := g.Scaffold(); err != nil {
					t.Fatalf("Scaffold() error = %v", err)
				}

				parseGoFiles(t, root)

				marker := filepath.Join(root, markers[arch])
				if _, err := os.Stat(marker); err != nil {
					t.Errorf("expected marker file %s: %v", markers[arch], err)
				}
			})
		}
	}
}

// TestScaffoldAuth generates each structured architecture with --auth and
// asserts the auth module's Go files all parse.
func TestScaffoldAuth(t *testing.T) {
	for _, arch := range []string{"layered", "clean", "microservice", "monolith"} {
		for _, router := range cli.Routers {
			t.Run(arch+"_"+router, func(t *testing.T) {
				root := t.TempDir()
				cfg := &cli.Config{
					ModuleName: "github.com/test/demo",
					Arch:       arch, Router: router, DB: "postgres", Auth: true,
					Entities: []string{"User"},
				}
				g := NewProjectGenerator(cfg, goembed.TemplateFS)
				g.projectRoot = root
				if err := g.Scaffold(); err != nil {
					t.Fatalf("Scaffold() error = %v", err)
				}
				parseGoFiles(t, root)

				migration := "migrations/postgres/000002_create_auth_tables.up.sql"
				if arch == "layered" {
					migration = "migrations/000002_create_auth_tables.up.sql"
				}
				for _, marker := range []string{
					"pkg/auth/jwt.go",
					"internal/auth/application/auth_service.go",
					"internal/auth/transport/http/middleware/auth_middleware.go",
					migration,
				} {
					if _, err := os.Stat(filepath.Join(root, marker)); err != nil {
						t.Errorf("expected auth file %s: %v", marker, err)
					}
				}
			})
		}
	}
}

// TestScaffoldFrontend generates each frontend variant and asserts the web
// package's Go files parse and the entry asset exists.
func TestScaffoldFrontend(t *testing.T) {
	for _, fe := range []string{"html", "htmx", "react"} {
		for _, arch := range []string{"simple", "clean"} {
			t.Run(fe+"_"+arch, func(t *testing.T) {
				root := t.TempDir()
				cfg := &cli.Config{
					ModuleName: "github.com/test/demo",
					Arch:       arch, Router: "gin", DB: "postgres", Frontend: fe,
					Entities: []string{"User"},
				}
				g := NewProjectGenerator(cfg, goembed.TemplateFS)
				g.projectRoot = root
				if err := g.Scaffold(); err != nil {
					t.Fatalf("Scaffold() error = %v", err)
				}
				parseGoFiles(t, root)

				if _, err := os.Stat(filepath.Join(root, "web/web.go")); err != nil {
					t.Errorf("expected web/web.go: %v", err)
				}
			})
		}
	}
}

func parseGoFiles(t *testing.T, root string) {
	t.Helper()
	fset := token.NewFileSet()
	count := 0
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() || !strings.HasSuffix(path, ".go") {
			return err
		}
		count++
		if _, perr := parser.ParseFile(fset, path, nil, parser.AllErrors); perr != nil {
			t.Errorf("parse %s: %v", path, perr)
		}
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
	if count == 0 {
		t.Error("no .go files generated")
	}
}
