package kong

import "github.com/gechr/clib/complete"

// Option configures Handle behavior.
type Option = complete.PreflightOption

// WithQuiet suppresses output during install/uninstall.
func WithQuiet(quiet bool) Option {
	return complete.WithQuiet(quiet)
}

// WithArgs passes preceding positional args to the completion handler.
func WithArgs(args []string) Option {
	return complete.WithArgs(args)
}
