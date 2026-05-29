package cli

import (
	"reflect"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name     string
		args     []string
		expected Config
		wantErr  bool
	}{
		{
			name: "defaults",
			args: []string{"--module", "github.com/x/y"},
			expected: Config{
				ModuleName: "github.com/x/y", Arch: "clean", Router: "gin", DB: "postgres",
				Entities: []string{"Item"},
			},
		},
		{
			name: "new subcommand ignored",
			args: []string{"new", "--module", "github.com/x/y", "--arch", "monolith"},
			expected: Config{
				ModuleName: "github.com/x/y", Arch: "monolith", Router: "gin", DB: "postgres",
				Entities: []string{"Item"},
			},
		},
		{
			name: "all flags",
			args: []string{
				"--module", "github.com/c/api", "--arch", "layered", "--router", "chi",
				"--db", "postgres", "--entity", "User", "--entity", "Order",
			},
			expected: Config{
				ModuleName: "github.com/c/api", Arch: "layered", Router: "chi", DB: "postgres",
				Entities: []string{"User", "Order"},
			},
		},
		{
			name:    "invalid arch errors",
			args:    []string{"--module", "github.com/x/y", "--arch", "bogus"},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, err := parseArgs(tt.args)
			if (err != nil) != tt.wantErr {
				t.Fatalf("parseArgs() err = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if !reflect.DeepEqual(*cfg, tt.expected) {
				t.Errorf("parseArgs() = %+v, want %+v", *cfg, tt.expected)
			}
		})
	}
}

func TestGetProjectRoot(t *testing.T) {
	cases := map[string]string{
		"github.com/user/myapp": "myapp",
		"example.com/a/b/c":     "c",
		"single":                "single",
	}
	for module, want := range cases {
		if got := (&Config{ModuleName: module}).GetProjectRoot(); got != want {
			t.Errorf("GetProjectRoot(%q) = %q, want %q", module, got, want)
		}
	}
}

func TestStringSliceSet(t *testing.T) {
	var s stringSlice
	_ = s.Set("User, Product")
	_ = s.Set("Order")
	if !reflect.DeepEqual([]string(s), []string{"User", "Product", "Order"}) {
		t.Errorf("unexpected entities: %#v", s)
	}
	if s.String() != "User,Product,Order" {
		t.Errorf("String() = %q", s.String())
	}
}
