package cli

import (
	"flag"
	"fmt"
	"os"
	"strings"
)

// stringSlice implements flag.Value for repeatable string flags.
type stringSlice []string

func (s *stringSlice) String() string { return strings.Join(*s, ",") }

func (s *stringSlice) Set(value string) error {
	for _, v := range strings.Split(value, ",") {
		if v = strings.TrimSpace(v); v != "" {
			*s = append(*s, v)
		}
	}
	return nil
}

// Allowed values for the validated flags.
var (
	Architectures = []string{"simple", "layered", "clean", "microservice", "monolith"}
	Routers       = []string{"gin", "chi"}
	Databases     = []string{"postgres"}
	Frontends     = []string{"none", "html", "htmx", "react"}
)

// Config holds all CLI configuration for a generation run.
type Config struct {
	ModuleName string
	Arch       string
	Entities   []string
	Router     string
	DB         string
	Auth       bool
	Frontend   string
}

// ParseFlags parses os.Args and returns a validated Config.
func ParseFlags() (*Config, error) {
	return parseArgs(os.Args[1:])
}

// parseArgs parses an explicit argument slice. A leading "new" subcommand is
// accepted (and ignored) so both `gogen --module ...` and `gogen new --module
// ...` work.
func parseArgs(args []string) (*Config, error) {
	if len(args) > 0 && args[0] == "new" {
		args = args[1:]
	}

	fs := flag.NewFlagSet("gogen", flag.ContinueOnError)
	module := fs.String("module", "github.com/username/project", "Go module path (e.g. github.com/you/project)")
	arch := fs.String("arch", "clean", "architecture: "+strings.Join(Architectures, "|"))
	router := fs.String("router", "gin", "http router: "+strings.Join(Routers, "|"))
	db := fs.String("db", "postgres", "database: "+strings.Join(Databases, "|"))
	auth := fs.Bool("auth", false, "include JWT auth + RBAC module (not for --arch simple)")
	frontend := fs.String("frontend", "none", "frontend layer: "+strings.Join(Frontends, "|"))

	var entities stringSlice
	fs.Var(&entities, "entity", "entity name; repeatable (e.g. --entity User --entity Product)")

	if err := fs.Parse(args); err != nil {
		return nil, err
	}

	if len(entities) == 0 {
		entities = stringSlice{"Item"}
	}

	cfg := &Config{
		ModuleName: *module,
		Arch:       strings.ToLower(*arch),
		Entities:   entities,
		Router:     strings.ToLower(*router),
		DB:         strings.ToLower(*db),
		Auth:       *auth,
		Frontend:   strings.ToLower(*frontend),
	}

	return cfg, cfg.Validate()
}

// Validate checks the configuration against the allowed values.
func (c *Config) Validate() error {
	if c.ModuleName == "" {
		return fmt.Errorf("--module is required")
	}
	if !contains(Architectures, c.Arch) {
		return fmt.Errorf("invalid --arch %q (allowed: %s)", c.Arch, strings.Join(Architectures, ", "))
	}
	if !contains(Routers, c.Router) {
		return fmt.Errorf("invalid --router %q (allowed: %s)", c.Router, strings.Join(Routers, ", "))
	}
	if !contains(Databases, c.DB) {
		return fmt.Errorf("invalid --db %q (allowed: %s)", c.DB, strings.Join(Databases, ", "))
	}
	if c.Auth && c.Arch == "simple" {
		return fmt.Errorf("--auth is not supported for --arch simple")
	}
	if c.Frontend == "" {
		c.Frontend = "none"
	}
	if !contains(Frontends, c.Frontend) {
		return fmt.Errorf("invalid --frontend %q (allowed: %s)", c.Frontend, strings.Join(Frontends, ", "))
	}
	return nil
}

// GetProjectRoot extracts the project root directory name from the module path.
func (c *Config) GetProjectRoot() string {
	parts := strings.Split(c.ModuleName, "/")
	return parts[len(parts)-1]
}

func contains(set []string, v string) bool {
	for _, s := range set {
		if s == v {
			return true
		}
	}
	return false
}
