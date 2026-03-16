package shell

// SetParentProcessName replaces the process name lookup for testing.
// Returns a cleanup function that restores the original.
func SetParentProcessName(fn func() string) func() {
	old := parentProcessName
	parentProcessName = fn
	return func() { parentProcessName = old }
}
