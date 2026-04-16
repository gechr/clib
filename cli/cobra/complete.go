package cobra

import (
	"fmt"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/bash" // register shell generators
	_ "github.com/gechr/clib/complete/fish" // register shell generators
	_ "github.com/gechr/clib/complete/zsh"  // register shell generators
	"github.com/gechr/clib/internal/tag"
	"github.com/gechr/x/shell"
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// pflagTypeBool is the pflag type name for boolean flags.
const pflagTypeBool = "bool"

const commandDynamicArgsKey = "dynamic-args"

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
	cmd.CompletionOptions.DisableDefaultCmd = true

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
		action.Shell = shell.Detect()
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

		applyCommandAnnotations(&sub, child)
		sub.MaxPositionalArgs, sub.HasMaxPositionalArgs = commandPositionalLimit(child)
		sub.Subs = commandSubSpecs(child)
		subs = append(subs, sub)
	}
	return subs
}

var (
	cobraExactArgsRe   = regexp.MustCompile(`accepts (\d+) arg\(s\), received \d+`)
	cobraMaximumArgsRe = regexp.MustCompile(`accepts at most (\d+) arg\(s\), received \d+`)
	cobraRangeArgsRe   = regexp.MustCompile(`accepts between \d+ and (\d+) arg\(s\), received \d+`)
)

func commandPositionalLimit(cmd *cobralib.Command) (int, bool) {
	if cmd == nil {
		return 0, false
	}
	if limit, ok := limitFromArgsValidator(cmd); ok {
		return limit, true
	}
	return limitFromUse(cmd.Use)
}

func limitFromArgsValidator(cmd *cobralib.Command) (int, bool) {
	if cmd == nil || cmd.Args == nil {
		return 0, false
	}

	sample := "x"
	if len(cmd.ValidArgs) > 0 {
		sample, _, _ = strings.Cut(cmd.ValidArgs[0], "\t")
	}

	for n := range 33 {
		args := make([]string, n)
		for i := range args {
			args[i] = sample
		}

		err := cmd.Args(cmd, args)
		if err == nil {
			continue
		}

		msg := err.Error()
		for _, re := range []*regexp.Regexp{cobraMaximumArgsRe, cobraExactArgsRe, cobraRangeArgsRe} {
			matches := re.FindStringSubmatch(msg)
			if len(matches) != 1+1 {
				continue
			}
			limit, convErr := strconv.Atoi(matches[1])
			if convErr == nil {
				return limit, true
			}
		}
	}

	return 0, false
}

func limitFromUse(use string) (int, bool) {
	fields := strings.Fields(use)
	if len(fields) <= 1 {
		return 0, false
	}

	count := 0
	sawPositional := false
	for _, field := range fields[1:] {
		switch {
		case field == "[flags]", field == "[flag]":
			continue
		case strings.ContainsAny(field, "|{}()"):
			return 0, false
		}

		token := strings.Trim(field, ",")
		repeatable := strings.Contains(token, "...") || strings.Contains(token, "…")
		token = strings.TrimSuffix(token, "...")
		token = strings.TrimSuffix(token, "…")
		token = strings.TrimPrefix(token, "[")
		token = strings.TrimSuffix(token, "]")
		token = strings.TrimPrefix(token, "<")
		token = strings.TrimSuffix(token, ">")
		if token == "" || strings.HasPrefix(token, "-") {
			continue
		}
		sawPositional = true
		if repeatable {
			return 0, false
		}
		count++
	}

	return count, sawPositional
}

// CompletionCommand returns a cobra subcommand that replaces cobra's built-in
// "completion" command with one powered by clib. It disables cobra's default
// completion subcommand on parent and generates scripts via [complete.Generator].
//
// The genFunc is called at run time to build the generator, so the full command
// tree is available.
//
// Usage:
//
//	cmd.AddCommand(clib.CompletionCommand(cmd, func() *complete.Generator {
//	    gen := complete.NewGenerator("myapp").FromFlags(clib.FlagMeta(cmd))
//	    gen.Subs = clib.Subcommands(cmd)
//	    return gen
//	}))
func CompletionCommand(
	parent *cobralib.Command,
	genFunc func() *complete.Generator,
) *cobralib.Command {
	parent.CompletionOptions.DisableDefaultCmd = true

	shellCmd := func(sh string) *cobralib.Command {
		return &cobralib.Command{
			Use:   sh,
			Short: fmt.Sprintf("Print %s completion script", sh),
			Args:  cobralib.NoArgs,
			RunE: func(_ *cobralib.Command, _ []string) error {
				return genFunc().Print(parent.OutOrStdout(), sh)
			},
		}
	}

	cmd := &cobralib.Command{
		Use:    "completion",
		Short:  "Print completion script",
		Hidden: true,
	}
	cmd.AddCommand(
		shellCmd(shell.Bash),
		shellCmd(shell.Fish),
		shellCmd(shell.Zsh),
	)
	return cmd
}

func applyCommandAnnotations(sub *complete.SubSpec, cmd *cobralib.Command) {
	if sub == nil || cmd == nil {
		return
	}

	clib, ok := cmd.Annotations["clib"]
	if !ok {
		return
	}

	if val, found, err := tag.Parse(clib, tag.Complete); err != nil {
		panic(err)
	} else if found && val == "path" {
		sub.PathArgs = true
	}

	if val, found, err := tag.Parse(clib, commandDynamicArgsKey); err != nil {
		panic(err)
	} else if found {
		sub.DynamicArgs = tag.SplitCSV(val)
	}
}
