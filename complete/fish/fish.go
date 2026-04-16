package fish

import (
	"github.com/gechr/clib/complete"
	"github.com/gechr/x/shell"
)

//nolint:gochecknoinits // shell subpackages register themselves via init by design
func init() {
	complete.RegisterShell(shell.Fish, complete.GenerateFish)
}

// Generate generates a fish shell completion script.
func Generate(g *complete.Generator) (string, error) {
	return complete.GenerateFish(g)
}
