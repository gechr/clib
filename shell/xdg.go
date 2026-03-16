package shell

import (
	"os"
	"path/filepath"
)

// XDGDataHome returns the XDG data home directory, defaulting to
// ~/.local/share.
func XDGDataHome() (string, error) {
	if dataDir := os.Getenv("XDG_DATA_HOME"); dataDir != "" {
		return dataDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".local", "share"), nil
}

// XDGConfigHome returns the XDG config home directory, defaulting to
// ~/.config.
func XDGConfigHome() (string, error) {
	if configDir := os.Getenv("XDG_CONFIG_HOME"); configDir != "" {
		return configDir, nil
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".config"), nil
}
