package complete

// Hidden flag names used by the completion system.
// The @ prefix avoids clashing with user-defined flags.
const (
	FlagComplete = "@complete" // dynamic completion requests
	FlagShell    = "@shell"    // shell type for completions
)
