package goembed

import (
	"embed"
)

// TemplateFS holds the project templates rendered into generated projects.
//
//go:embed all:templates/*
var TemplateFS embed.FS

// WebFS holds the static assets for the `gogen serve` web UI.
//
//go:embed all:web
var WebFS embed.FS
