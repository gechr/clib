package urfave

import (
	"context"
	"fmt"
	"os"
	"reflect"

	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/bash" // register shell generators
	_ "github.com/gechr/clib/complete/fish" // register shell generators
	_ "github.com/gechr/clib/complete/zsh"  // register shell generators
	"github.com/gechr/clib/shell"
	clilib "github.com/urfave/cli/v3"
)

// CompletionFlags is an alias for [complete.CompletionFlags].
// See [Preflight] for pre-parse usage.
type CompletionFlags = complete.CompletionFlags

// Preflight scans os.Args for completion flags, allowing completion to be
// handled before the CLI parser. This is useful for subcommand-based CLIs
// where the parser requires a subcommand but completion flags are standalone.
//
// Returns a populated CompletionFlags, any positional args found after "--",
// and true if a completion flag was found. When ok is false, the caller should
// proceed with normal CLI parsing.
var Preflight = complete.Preflight

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
			Name:        complete.FlagInstallCompletion,
			Usage:       "Install shell completions",
			Hidden:      true,
			Destination: &c.installCompletion,
		},
		&clilib.BoolFlag{
			Name:        complete.FlagUninstallCompletion,
			Usage:       "Uninstall shell completions",
			Hidden:      true,
			Destination: &c.uninstallCompletion,
		},
		&clilib.BoolFlag{
			Name:        complete.FlagPrintCompletion,
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

	action := complete.Action{
		Shell:               c.shell,
		Complete:            c.complete,
		Args:                cfg.args,
		InstallCompletion:   c.installCompletion,
		UninstallCompletion: c.uninstallCompletion,
		PrintCompletion:     c.printCompletion,
	}
	complete.ApplyActionArgs(&action, os.Args[1:])
	if action.Shell == "" {
		action.Shell = shell.Detect()
	}

	return complete.HandleAction(action, gen, handler, cfg.quiet)
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
		sub.MaxPositionalArgs, sub.HasMaxPositionalArgs = positionalLimit(child)
		sub.Subs = commandSubSpecs(child)
		subs = append(subs, sub)
	}
	return subs
}

// CompletionCommand returns a hidden urfave subcommand that prints clib-powered
// completion scripts. It replaces urfave's default shell completion behavior.
//
// The genFunc is called at run time to build the generator, so the full command
// tree is available.
func CompletionCommand(genFunc func() *complete.Generator) *clilib.Command {
	shellCmd := func(sh string) *clilib.Command {
		return &clilib.Command{
			Name:  sh,
			Usage: fmt.Sprintf("Print %s completion script", sh),
			Action: func(_ context.Context, _ *clilib.Command) error {
				return genFunc().Print(os.Stdout, sh)
			},
		}
	}

	return &clilib.Command{
		Name:   "completion",
		Usage:  "Print completion script",
		Hidden: true,
		Commands: []*clilib.Command{
			shellCmd(shell.Bash),
			shellCmd(shell.Fish),
			shellCmd(shell.Zsh),
		},
	}
}

func positionalLimit(cmd *clilib.Command) (int, bool) {
	if cmd == nil {
		return 0, false
	}

	if len(cmd.Arguments) == 0 {
		return 0, true
	}

	total := 0
	for _, arg := range cmd.Arguments {
		if arg == nil {
			continue
		}
		v := reflect.ValueOf(arg)
		if !v.IsValid() {
			continue
		}
		if v.Kind() == reflect.Pointer {
			if v.IsNil() {
				continue
			}
			v = v.Elem()
		}
		if v.Kind() != reflect.Struct {
			return 0, false
		}

		maxField := v.FieldByName("Max")
		if maxField.IsValid() && maxField.Kind() == reflect.Int {
			limit := int(maxField.Int())
			if limit < 0 {
				return 0, false
			}
			total += limit
			continue
		}

		total++
	}

	return total, true
}
