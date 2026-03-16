package shell

import (
	"fmt"
	"path/filepath"
)

// CompletionFile returns the standard completion file path for the given
// command and shell.
func CompletionFile(command, sh string) (string, error) {
	switch sh {
	case Bash:
		dir, err := XDGDataHome()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "bash-completion", "completions", command), nil
	case Zsh:
		dir, err := XDGDataHome()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "zsh", "site-functions", "_"+command), nil
	case Fish:
		dir, err := XDGConfigHome()
		if err != nil {
			return "", err
		}
		return filepath.Join(dir, "fish", "completions", command+".fish"), nil
	default:
		return "", fmt.Errorf("unsupported shell %q", sh)
	}
}
