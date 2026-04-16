package bash

import (
	"github.com/gechr/clib/complete"
	"github.com/gechr/x/shell"
)

//nolint:gochecknoinits // shell subpackages register themselves via init by design
func init() {
	complete.RegisterShell(shell.Bash, complete.GenerateBash)
}

// Generate generates a bash shell completion script.
func Generate(g *complete.Generator) (string, error) {
	return complete.GenerateBash(g)
}
