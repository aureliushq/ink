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

func main() {
	cmd.Execute(themes)
}
