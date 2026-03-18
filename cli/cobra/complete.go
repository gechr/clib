package cobra

import (
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
	complete            string
	shell               string
	installCompletion   bool
	uninstallCompletion bool
	printCompletion     bool
}

// NewCompletion adds hidden persistent flags to cmd and returns a Completion.
// Flags added: --@complete, --@shell, --install-completion,
// --uninstall-completion, --print-completion.
func NewCompletion(cmd *cobralib.Command) *Completion {
	c := &Completion{}
	pf := cmd.PersistentFlags()
	pf.StringVar(&c.complete, complete.FlagComplete, "", "Dynamic completion type")
	pf.StringVar(&c.shell, complete.FlagShell, "", "Shell type for completions")
	pf.BoolVar(&c.installCompletion, "install-completion", false, "Install shell completions")
	pf.BoolVar(&c.uninstallCompletion, "uninstall-completion", false, "Uninstall shell completions")
	pf.BoolVar(&c.printCompletion, "print-completion", false, "Print completion script")

	_ = pf.MarkHidden(complete.FlagComplete)
	_ = pf.MarkHidden(complete.FlagShell)
	_ = pf.MarkHidden("install-completion")
	_ = pf.MarkHidden("uninstall-completion")
	_ = pf.MarkHidden("print-completion")

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
