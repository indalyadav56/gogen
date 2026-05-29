package main

import (
	"fmt"
	"os"

	goembed "github.com/indalyadav56/gogen"
	"github.com/indalyadav56/gogen/internal/cli"
	"github.com/indalyadav56/gogen/internal/generator"
)

func main() {
	config, err := cli.ParseFlags()
	if err != nil {
		exitWithError(err)
	}

	projectGen := generator.NewProjectGenerator(config, goembed.TemplateFS)
	if err := projectGen.Generate(); err != nil {
		exitWithError(err)
	}

	fmt.Printf("✅ Scaffolded %q (%s architecture, %s router) at ./%s\n",
		config.ModuleName, config.Arch, config.Router, config.GetProjectRoot())
}

func exitWithError(err error) {
	fmt.Fprintf(os.Stderr, "❌ Error: %v\n", err)
	os.Exit(1)
}
