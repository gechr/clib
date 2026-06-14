package pwsh

import (
	"github.com/gechr/clib/complete"
	"github.com/gechr/x/shell"
)

//nolint:gochecknoinits // shell subpackages register themselves via init by design
func init() {
	complete.RegisterShell(shell.Pwsh, complete.GeneratePwsh)
}

// Generate generates a PowerShell completion script.
func Generate(g *complete.Generator) (string, error) {
	return complete.GeneratePwsh(g)
}
