package nu

import (
	"github.com/gechr/clib/complete"
	"github.com/gechr/x/shell"
)

//nolint:gochecknoinits // shell subpackages register themselves via init by design
func init() {
	complete.RegisterShell(shell.Nu, complete.GenerateNu)
}

// Generate generates a Nushell completion script.
func Generate(g *complete.Generator) (string, error) {
	return complete.GenerateNu(g)
}
