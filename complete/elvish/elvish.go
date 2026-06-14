package elvish

import (
	"github.com/gechr/clib/complete"
	"github.com/gechr/x/shell"
)

//nolint:gochecknoinits // shell subpackages register themselves via init by design
func init() {
	complete.RegisterShell(shell.Elvish, complete.GenerateElvish)
}

// Generate generates an Elvish completion script.
func Generate(g *complete.Generator) (string, error) {
	return complete.GenerateElvish(g)
}
