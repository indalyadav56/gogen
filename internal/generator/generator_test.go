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
