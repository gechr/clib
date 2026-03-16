package urfave

import (
	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/bash" // register shell generators
	_ "github.com/gechr/clib/complete/fish" // register shell generators
	_ "github.com/gechr/clib/complete/zsh"  // register shell generators
	shellpkg "github.com/gechr/clib/shell"
	clilib "github.com/urfave/cli/v3"
)

// Completion manages hidden completion flags on a urfave command.
type Completion struct {
	complete            string
	shell               string
	installCompletion   bool
	uninstallCompletion bool
	printCompletion     bool
}

// NewCompletion adds hidden completion flags to cmd and returns a Completion.
// Flags added: --@complete, --@shell, --install-completion,
// --uninstall-completion, --print-completion.
func NewCompletion(cmd *clilib.Command) *Completion {
	c := &Completion{}
	cmd.Flags = append(cmd.Flags,
		&clilib.StringFlag{
			Name:        complete.FlagComplete,
			Usage:       "Dynamic completion type",
			Hidden:      true,
			Destination: &c.complete,
		},
		&clilib.StringFlag{
			Name:        complete.FlagShell,
			Usage:       "Shell type for completions",
			Hidden:      true,
			Destination: &c.shell,
		},
		&clilib.BoolFlag{
			Name:        "install-completion",
			Usage:       "Install shell completions",
			Hidden:      true,
			Destination: &c.installCompletion,
		},
		&clilib.BoolFlag{
			Name:        "uninstall-completion",
			Usage:       "Uninstall shell completions",
			Hidden:      true,
			Destination: &c.uninstallCompletion,
		},
		&clilib.BoolFlag{
			Name:        "print-completion",
			Usage:       "Print completion script",
			Hidden:      true,
			Destination: &c.printCompletion,
		},
	)
	return c
}

// Handle checks whether a completion action was requested and executes it.
// Returns true if a completion action was handled (caller should exit).
// The handler callback is invoked for --@complete=<type> requests;
// it receives the completion type and resolved shell name.
func (c *Completion) Handle(
	gen *complete.Generator,
	handler complete.Handler,
	opts ...Option,
) (bool, error) {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}

	shell := c.shell
	if shell == "" {
		shell = shellpkg.Detect()
	}

	return complete.HandleAction(complete.Action{
		Shell:               shell,
		Complete:            c.complete,
		Args:                cfg.args,
		InstallCompletion:   c.installCompletion,
		UninstallCompletion: c.uninstallCompletion,
		PrintCompletion:     c.printCompletion,
	}, gen, handler, cfg.quiet)
}

// Subcommands extracts subcommand completion specs from a urfave command tree.
// Each visible subcommand produces a SubSpec with its flags (excluding hidden
// flags and the built-in --help flag).
func Subcommands(cmd *clilib.Command) []complete.SubSpec {
	if cmd == nil {
		return nil
	}
	prepareFlagExtras(cmd)
	return commandSubSpecs(cmd)
}

func commandSubSpecs(cmd *clilib.Command) []complete.SubSpec {
	var subs []complete.SubSpec
	for _, child := range cmd.Commands {
		if child.Hidden || child.Name == "help" {
			continue
		}
		sub := complete.SubSpec{
			Name:    child.Name,
			Aliases: child.Aliases,
			Terse:   child.Usage,
		}
		for _, f := range child.Flags {
			// Skip hidden flags.
			if vf, ok := f.(clilib.VisibleFlag); ok && !vf.IsVisible() {
				continue
			}
			meta := flagToMeta(child, f)
			if meta.Name == "help" {
				continue
			}
			sub.Specs = append(sub.Specs, complete.SpecsFromFlagMeta(meta)...)
		}
		// PathArgs: check command extra.
		if extra := getCommandExtra(child); extra != nil && extra.PathArgs {
			sub.PathArgs = true
		}
		sub.Subs = commandSubSpecs(child)
		subs = append(subs, sub)
	}
	return subs
}
