/*
Copyright © 2026 Ilango Rajagopal <hey@i4o.dev>
*/
package main

import (
	"embed"

	"github.com/aureliushq/ink/cmd"
)

//go:embed themes/*
var themes embed.FS

// Build information. Populated at build time via -ldflags -X.
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	cmd.Execute(themes, cmd.BuildInfo{Version: version, Commit: commit, Date: date})
}
