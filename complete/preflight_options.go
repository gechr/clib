package complete

type preflightConfig struct {
	quiet bool
	args  []string
}

// PreflightOption configures [CompletionFlags.Handle] behavior.
type PreflightOption func(*preflightConfig)

// WithQuiet suppresses output during install/uninstall.
func WithQuiet(quiet bool) PreflightOption {
	return func(c *preflightConfig) {
		c.quiet = quiet
	}
}

// WithArgs passes preceding positional args to the completion handler.
func WithArgs(args []string) PreflightOption {
	return func(c *preflightConfig) {
		c.args = args
	}
}
