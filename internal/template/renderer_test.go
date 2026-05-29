package template

import (
	"embed"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

//go:embed testdata/sample.tmpl
var testFS embed.FS

func TestRenderToFile(t *testing.T) {
	r := NewRenderer(testFS)
	out := filepath.Join(t.TempDir(), "out.go")

	err := r.RenderToFile("testdata/sample.tmpl", out, Data{
		Package: "demo",
		Module:  "github.com/test/project",
		Entity:  "category",
	})
	if err != nil {
		t.Fatalf("RenderToFile() error = %v", err)
	}

	content, err := os.ReadFile(out)
	if err != nil {
		t.Fatal(err)
	}
	got := string(content)
	for _, want := range []string{"package demo", "module=github.com/test/project", "entity=Category", "plural=categories"} {
		if !strings.Contains(got, want) {
			t.Errorf("output missing %q; got:\n%s", want, got)
		}
	}
}

func TestRenderToFileMissingTemplate(t *testing.T) {
	r := NewRenderer(testFS)
	out := filepath.Join(t.TempDir(), "out.go")
	if err := r.RenderToFile("testdata/nope.tmpl", out, Data{}); err == nil {
		t.Error("expected error for missing template, got nil")
	}
}

func TestToPascal(t *testing.T) {
	cases := map[string]string{"user": "User", "User": "User", "": ""}
	for in, want := range cases {
		if got := toPascal(in); got != want {
			t.Errorf("toPascal(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestToCamel(t *testing.T) {
	cases := map[string]string{"User": "user", "user": "user", "": ""}
	for in, want := range cases {
		if got := toCamel(in); got != want {
			t.Errorf("toCamel(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestPlural(t *testing.T) {
	cases := map[string]string{
		"user":     "users",
		"category": "categories",
		"box":      "boxes",
		"class":    "classes",
		"day":      "days",
		"":         "",
	}
	for in, want := range cases {
		if got := plural(in); got != want {
			t.Errorf("plural(%q) = %q, want %q", in, got, want)
		}
	}
}
