package complete

import (
	"os"

	"github.com/gechr/clib/shell"
)

// CompletionFlags provides standalone completion flags for pre-parse handling.
// Use [Preflight] to populate this struct from os.Args before the CLI parser runs.
type CompletionFlags struct {
	Complete            string
	Shell               string
	InstallCompletion   bool
	UninstallCompletion bool
	PrintCompletion     bool
}

// Preflight scans os.Args for completion flags, allowing completion to be
// handled before the CLI parser. This is useful for subcommand-based CLIs
// where the parser requires a subcommand but completion flags are standalone.
//
// Returns a populated CompletionFlags, any positional args found after "--",
// and true if a completion flag was found. When ok is false, the caller should
// proceed with normal CLI parsing.
func Preflight() (CompletionFlags, []string, bool) {
	args := os.Args[1:]

	var action Action
	ApplyActionArgs(&action, args)

	if !action.InstallCompletion && !action.UninstallCompletion &&
		!action.PrintCompletion && action.Complete == "" {
		return CompletionFlags{}, nil, false
	}

	// Extract positional args after "--" for dynamic completion handlers.
	var positional []string
	for i, arg := range args {
		if arg == "--" {
			positional = args[i+1:]
			break
		}
	}

	f := CompletionFlags{
		Complete:            action.Complete,
		Shell:               action.Shell,
		InstallCompletion:   action.InstallCompletion,
		UninstallCompletion: action.UninstallCompletion,
		PrintCompletion:     action.PrintCompletion,
	}
	return f, positional, true
}

// Handle checks whether a completion action was requested and executes it.
// Returns true if a completion action was handled (caller should exit).
// The handler callback is invoked for --@complete=<type> requests;
// it receives the completion type and resolved shell name.
func (f *CompletionFlags) Handle(
	gen *Generator,
	handler Handler,
	opts ...PreflightOption,
) (bool, error) {
	var cfg preflightConfig
	for _, o := range opts {
		o(&cfg)
	}

	sh := f.Shell
	if sh == "" {
		sh = shell.Detect()
	}

	return HandleAction(Action{
		Shell:               sh,
		Complete:            f.Complete,
		Args:                cfg.args,
		InstallCompletion:   f.InstallCompletion,
		UninstallCompletion: f.UninstallCompletion,
		PrintCompletion:     f.PrintCompletion,
	}, gen, handler, cfg.quiet)
}

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
