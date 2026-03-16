package shell

import "slices"

const (
	Ash    = "ash"
	Bash   = "bash"
	Dash   = "dash"
	Elvish = "elvish"
	Fish   = "fish"
	Ksh    = "ksh"
	Nu     = "nu"
	Pwsh   = "pwsh"
	Sh     = "sh"
	Tcsh   = "tcsh"
	Zsh    = "zsh"
)

var knownShells = []string{
	Ash,
	Bash,
	Dash,
	Elvish,
	Fish,
	Ksh,
	Nu,
	Pwsh,
	Sh,
	Tcsh,
	Zsh,
}

var knownShellSet = map[string]struct{}{
	Ash:    {},
	Bash:   {},
	Dash:   {},
	Elvish: {},
	Fish:   {},
	Ksh:    {},
	Nu:     {},
	Pwsh:   {},
	Sh:     {},
	Tcsh:   {},
	Zsh:    {},
}

// Known returns the set of recognized shell names.
func Known() []string {
	return slices.Clone(knownShells)
}

// IsKnown reports whether name matches a known shell.
func IsKnown(name string) bool {
	_, ok := knownShellSet[name]
	return ok
}
