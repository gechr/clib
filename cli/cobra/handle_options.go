package cobra

type config struct {
	quiet bool
	args  []string
}

// Option configures Handle behavior.
type Option func(*config)

// WithQuiet suppresses output during install/uninstall.
func WithQuiet(quiet bool) Option {
	return func(c *config) {
		c.quiet = quiet
	}
}

// WithArgs passes preceding positional args to the completion handler.
func WithArgs(args []string) Option {
	return func(c *config) {
		c.args = args
	}
}
