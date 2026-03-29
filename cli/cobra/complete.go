package cobra

import (
	"os"

	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/bash" // register shell generators
	_ "github.com/gechr/clib/complete/fish" // register shell generators
	_ "github.com/gechr/clib/complete/zsh"  // register shell generators
	"github.com/gechr/clib/internal/tag"
	shellpkg "github.com/gechr/clib/shell"
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// pflagTypeBool is the pflag type name for boolean flags.
const pflagTypeBool = "bool"

// Completion manages hidden completion flags on a cobra command.
type Completion struct {
	cmd                 *cobralib.Command
	complete            string
	shell               string
	installCompletion   bool
	uninstallCompletion bool
	printCompletion     bool
}

// NewCompletion adds hidden persistent flags to cmd and returns a Completion.
// Flags added: --@complete, --@shell, --install-completion,
// --uninstall-completion, --print-completion.
// It also hides cobra's built-in "completion" subcommand.
func NewCompletion(cmd *cobralib.Command) *Completion {
	cmd.CompletionOptions.HiddenDefaultCmd = true

	c := &Completion{cmd: cmd}
	pf := cmd.PersistentFlags()
	pf.StringVar(&c.complete, complete.FlagComplete, "", "Dynamic completion type")
	pf.StringVar(&c.shell, complete.FlagShell, "", "Shell type for completions")
	pf.BoolVar(
		&c.installCompletion,
		complete.FlagInstallCompletion,
		false,
		"Install shell completions",
	)
	pf.BoolVar(
		&c.uninstallCompletion,
		complete.FlagUninstallCompletion,
		false,
		"Uninstall shell completions",
	)
	pf.BoolVar(&c.printCompletion, complete.FlagPrintCompletion, false, "Print completion script")

	_ = pf.MarkHidden(complete.FlagComplete)
	_ = pf.MarkHidden(complete.FlagShell)
	_ = pf.MarkHidden(complete.FlagInstallCompletion)
	_ = pf.MarkHidden(complete.FlagUninstallCompletion)
	_ = pf.MarkHidden(complete.FlagPrintCompletion)

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

	action := c.action(cfg.args)
	return complete.HandleAction(action, gen, handler, cfg.quiet)
}

func (c *Completion) action(args []string) complete.Action {
	action := complete.Action{
		Shell:               c.shell,
		Complete:            c.complete,
		Args:                args,
		InstallCompletion:   c.installCompletion,
		UninstallCompletion: c.uninstallCompletion,
		PrintCompletion:     c.printCompletion,
	}

	if c.cmd != nil && !completionFlagsChanged(c.cmd.PersistentFlags()) {
		complete.ApplyActionArgs(&action, os.Args[1:])
	}
	if action.Shell == "" {
		action.Shell = shellpkg.Detect()
	}

	return action
}

func completionFlagsChanged(fs *pflag.FlagSet) bool {
	if fs == nil {
		return false
	}
	for _, name := range []string{
		complete.FlagComplete,
		complete.FlagShell,
		complete.FlagInstallCompletion,
		complete.FlagUninstallCompletion,
		complete.FlagPrintCompletion,
	} {
		if flag := fs.Lookup(name); flag != nil && flag.Changed {
			return true
		}
	}
	return false
}

// Subcommands extracts subcommand completion specs from a cobra command tree.
// Each visible subcommand produces a SubSpec with its local flags (excluding
// the built-in --help flag).
func Subcommands(cmd *cobralib.Command) []complete.SubSpec {
	if cmd == nil {
		return nil
	}
	return commandSubSpecs(cmd)
}

func commandSubSpecs(cmd *cobralib.Command) []complete.SubSpec {
	var subs []complete.SubSpec
	for _, child := range cmd.Commands() {
		if !child.IsAvailableCommand() || child.Deprecated != "" {
			continue
		}
		sub := complete.SubSpec{
			Name:    child.Name(),
			Aliases: child.Aliases,
			Terse:   child.Short,
		}

		appendFlags := func(fs *pflag.FlagSet, persistent bool) {
			fs.VisitAll(func(f *pflag.Flag) {
				if f.Hidden || f.Deprecated != "" || f.Name == "help" {
					return
				}
				meta := pflagToMeta(f)
				meta.Persistent = persistent
				sub.Specs = append(sub.Specs, complete.SpecsFromFlagMeta(meta)...)
			})
		}

		appendFlags(child.LocalNonPersistentFlags(), false)
		appendFlags(child.PersistentFlags(), true)

		// PathArgs: check cmd.Annotations["clib"] for complete='path'.
		if clib, ok := child.Annotations["clib"]; ok {
			if val, found, err := tag.Parse(
				clib,
				tag.Complete,
			); err != nil {
				panic(err)
			} else if found && val == "path" {
				sub.PathArgs = true
			}
		}
		sub.Subs = commandSubSpecs(child)
		subs = append(subs, sub)
	}
	return subs
}
