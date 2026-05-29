package blueprint

import (
	"testing"

	"github.com/indalyadav56/gogen/internal/cli"
)

func newConfig(arch string) *cli.Config {
	return &cli.Config{
		ModuleName: "github.com/test/demo",
		Arch:       arch,
		Router:     "gin",
		DB:         "postgres",
		Entities:   []string{"User", "Product"},
	}
}

func TestRegistryCoversAllArchitectures(t *testing.T) {
	for _, arch := range cli.Architectures {
		if _, ok := Get(arch); !ok {
			t.Errorf("no blueprint registered for architecture %q", arch)
		}
	}
}

func TestBuildProducesUniqueNonEmptySpecs(t *testing.T) {
	for arch := range Registry {
		t.Run(arch, func(t *testing.T) {
			bp, _ := Get(arch)
			files := bp.Build(newConfig(arch))
			if len(files) == 0 {
				t.Fatal("Build() returned no files")
			}
			seen := map[string]bool{}
			for _, f := range files {
				if f.Path == "" {
					t.Error("file spec with empty path")
				}
				if seen[f.Path] {
					t.Errorf("duplicate target path: %s", f.Path)
				}
				seen[f.Path] = true
			}
		})
	}
}

func TestBuildMarkerFiles(t *testing.T) {
	markers := map[string]string{
		"simple":       "user.go",
		"layered":      "internal/repository/user_repository.go",
		"clean":        "internal/domain/repository/user_repository.go",
		"microservice": "docker-compose.yml",
		"monolith":     "internal/product/application/product_service.go",
	}
	for arch, marker := range markers {
		t.Run(arch, func(t *testing.T) {
			bp, _ := Get(arch)
			files := bp.Build(newConfig(arch))
			found := false
			for _, f := range files {
				if f.Path == marker {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("%s: expected marker file %q not produced", arch, marker)
			}
		})
	}
}
