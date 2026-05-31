package main

import (
	"flag"
	"fmt"
	"os"

	goembed "github.com/indalyadav56/gogen"
	"github.com/indalyadav56/gogen/internal/cli"
	"github.com/indalyadav56/gogen/internal/generator"
	"github.com/indalyadav56/gogen/internal/logx"
	"github.com/indalyadav56/gogen/internal/server"
)

func main() {
	log := logx.Init()
	defer func() { _ = log.Sync() }()

	// `gogen serve [--port N]` starts the local web UI.
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		fs := flag.NewFlagSet("serve", flag.ContinueOnError)
		port := fs.Int("port", server.DefaultPort, "preferred port (falls back to a free one if taken)")
		if err := fs.Parse(os.Args[2:]); err != nil {
			exitWithError(err)
		}
		if err := server.Serve(goembed.TemplateFS, goembed.WebFS, *port); err != nil {
			exitWithError(err)
		}
		return
	}

	// Otherwise, generate a project from CLI flags.
	config, err := cli.ParseFlags()
	if err != nil {
		exitWithError(err)
	}

	logx.S().Infow("generating project",
		"module", config.ModuleName, "arch", config.Arch, "router", config.Router,
		"auth", config.Auth, "frontend", config.Frontend, "entities", config.Entities)

	projectGen := generator.NewProjectGenerator(config, goembed.TemplateFS)
	if err := projectGen.Generate(); err != nil {
		exitWithError(err)
	}

	fmt.Printf("✅ Scaffolded %q (%s architecture, %s router) at ./%s\n",
		config.ModuleName, config.Arch, config.Router, config.GetProjectRoot())
}

func exitWithError(err error) {
	logx.S().Errorw("command failed", "error", err)
	os.Exit(1)
}
