package shell

import (
	"os"
	"path/filepath"
)

const EnvShell = "SHELL"

// DetectFromEnv returns the base name of env if it names a recognized shell.
func DetectFromEnv(env string) string {
	return normalizeShellName(os.Getenv(env))
}

// parentProcessName returns the parent process name.
// Replaced in tests to control DetectFromProcess behavior.
var parentProcessName = func() string {
	if ppid := os.Getppid(); ppid > 0 {
		return processName(ppid)
	}
	return ""
}

// DetectFromProcess returns the parent process name if it is a known shell,
// or empty if unavailable or not recognized.
func DetectFromProcess() string {
	if name := parentProcessName(); IsKnown(name) {
		return name
	}
	return ""
}

// Detect returns the shell to use for completions.
// Priority: COMPLETE_SHELL env var, parent process name, SHELL env var.
func Detect() string {
	if shell := DetectFromEnv("COMPLETE_SHELL"); shell != "" {
		return shell
	}
	if shell := DetectFromProcess(); shell != "" {
		return shell
	}
	return normalizeShellName(os.Getenv(EnvShell))
}

func normalizeShellName(raw string) string {
	if raw == "" {
		return ""
	}
	name := filepath.Base(raw)
	if !IsKnown(name) {
		return ""
	}
	return name
}
