package zsh

import (
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/shell"
)

//nolint:gochecknoinits // shell subpackages register themselves via init by design
func init() {
	complete.RegisterShell(shell.Zsh, complete.GenerateZsh)
}

// Generate generates a zsh shell completion script.
func Generate(g *complete.Generator) (string, error) {
	return complete.GenerateZsh(g)
}
