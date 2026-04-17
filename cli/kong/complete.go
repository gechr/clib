package kong

import (
	"maps"
	"os"
	"slices"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/bash" // register shell generators
	_ "github.com/gechr/clib/complete/fish" // register shell generators
	_ "github.com/gechr/clib/complete/zsh"  // register shell generators
)

// CompletionFlags provides embeddable Kong flags for shell completion management.
// Embed this in your CLI struct to get --@complete, --@shell, --install-completion,
// --uninstall-completion, and --print-completion flags automatically.
//
// For pre-parse usage, see [Preflight].
type CompletionFlags struct {
	Complete            string `name:"@complete"            help:"Dynamic completion type"     hidden:""`
	Shell               string `name:"@shell"               help:"Shell type for completions"  hidden:""`
	InstallCompletion   bool   `name:"install-completion"   help:"Install shell completions"   hidden:""`
	UninstallCompletion bool   `name:"uninstall-completion" help:"Uninstall shell completions" hidden:""`
	PrintCompletion     bool   `name:"print-completion"     help:"Print completion script"     hidden:""`
}

// Handle checks whether a completion action was requested and executes it.
// Returns true if a completion action was handled (caller should exit).
// The handler callback is invoked for --@complete=<type> requests;
// it receives the completion type and resolved shell name.
func (f *CompletionFlags) Handle(
	gen *complete.Generator,
	handler complete.Handler,
	opts ...complete.PreflightOption,
) (bool, error) {
	cf := complete.CompletionFlags{
		Complete:            f.Complete,
		Shell:               f.Shell,
		InstallCompletion:   f.InstallCompletion,
		UninstallCompletion: f.UninstallCompletion,
		PrintCompletion:     f.PrintCompletion,
	}
	return cf.Handle(gen, handler, opts...)
}

// Preflight scans os.Args for completion flags, allowing completion to be
// handled before the CLI parser. This is useful for subcommand-based CLIs
// where the parser requires a subcommand but completion flags are standalone.
//
// Returns a populated CompletionFlags, any positional args found after "--",
// and true if a completion flag was found. When ok is false, the caller should
// proceed with normal CLI parsing.
//
// Usage:
//
//	if f, args, ok := clib.Preflight(); ok {
//	    gen := complete.NewGenerator("myapp").FromFlags(flags)
//	    gen.Subs = clib.Subcommands(parser)
//	    f.Handle(gen, handler, complete.WithArgs(args))
//	    return
//	}
func Preflight() (CompletionFlags, []string, bool) {
	cf, positional, ok := complete.Preflight()
	if !ok {
		return CompletionFlags{}, nil, false
	}
	return CompletionFlags{
		Complete:            cf.Complete,
		Shell:               cf.Shell,
		InstallCompletion:   cf.InstallCompletion,
		UninstallCompletion: cf.UninstallCompletion,
		PrintCompletion:     cf.PrintCompletion,
	}, positional, true
}

// completionCmd is the struct registered via [CompletionCommand] as a kong
// dynamic command. It accepts a shell name as a positional arg.
type completionCmd struct {
	Shell   string `help:"Shell type" arg:"" enum:"bash,zsh,fish"`
	genFunc func() *complete.Generator
}

func (c *completionCmd) Run() error {
	return c.genFunc().Print(os.Stdout, c.Shell)
}

// CompletionCommand returns a [konglib.Option] that registers a hidden
// "completion" subcommand powered by clib. Pass this to [konglib.New].
//
// The genFunc is called at run time to build the generator, so the full
// parser model is available.
func CompletionCommand(genFunc func() *complete.Generator) konglib.Option {
	cmd := &completionCmd{genFunc: genFunc}
	return konglib.DynamicCommand("completion", "Print completion script", "", cmd, `hidden:""`)
}

// Subcommands extracts subcommand completion specs from a kong parser's model.
// Each visible subcommand produces a SubSpec with its flags (excluding kong's
// built-in --help flag).
func Subcommands(parser *konglib.Kong) []complete.SubSpec {
	if parser == nil || parser.Model == nil {
		return nil
	}
	return nodeSubSpecs(parser.Model.Node)
}

func nodeSubSpecs(node *konglib.Node) []complete.SubSpec {
	var subs []complete.SubSpec
	for _, child := range node.Children {
		if child == nil || child.Hidden {
			continue
		}
		sub := complete.SubSpec{
			Name:    child.Name,
			Aliases: child.Aliases,
			Terse:   child.Help,
		}
		for _, flag := range child.Flags {
			if flag == nil || flag.Hidden {
				continue
			}
			// Skip kong's built-in help flag.
			if flag.Name == "help" {
				continue
			}
			meta := flagMeta(flag)
			sub.Specs = append(sub.Specs, complete.SpecsFromFlagMeta(meta)...)
		}
		// Read clib annotations on the cmd field for terse and path completion.
		if child.Tag != nil && child.Tag.Has(tagClib) {
			var meta complete.FlagMeta
			if err := meta.ParseClibTag(child.Tag.Get(tagClib)); err != nil {
				panic(err)
			}
			if meta.Terse != "" {
				sub.Terse = meta.Terse
			}
			if meta.Complete == predictorPath {
				sub.PathArgs = true
			}
		}
		for _, arg := range child.Positional {
			if arg.Tag == nil || !arg.Tag.Has(tagPredictor) {
				break
			}
			predictor := arg.Tag.Get(tagPredictor)
			if predictor == predictorPath {
				sub.PathArgs = true
				break
			}
			sub.DynamicArgs = append(sub.DynamicArgs, predictor)
		}
		sub.MaxPositionalArgs, sub.HasMaxPositionalArgs = positionalLimit(child)
		// Recurse into nested subcommands.
		sub.Subs = nodeSubSpecs(child)
		// Subcommand-only groupers: flags at this level cannot take effect
		// without picking a subcommand, so drop them from completions to
		// match help output.
		if len(sub.Subs) > 0 {
			sub.Specs = nil
		}
		subs = append(subs, sub)
	}
	return subs
}

func positionalLimit(node *konglib.Node) (int, bool) {
	if node == nil {
		return 0, false
	}

	count := 0
	for _, arg := range node.Positional {
		if arg == nil {
			continue
		}
		if arg.IsSlice() {
			return 0, false
		}
		count++
	}

	return count, true
}

func flagMeta(flag *konglib.Flag) complete.FlagMeta {
	meta := complete.FlagMeta{
		Name:       flag.Name,
		Help:       flag.Help,
		HasArg:     !flag.IsBool(),
		Persistent: true,
	}
	if flag.Short != 0 {
		meta.Short = string(flag.Short)
	}
	if flag.Enum != "" {
		meta.Enum = slices.Sorted(maps.Keys(flag.EnumMap()))
	}
	if flag.Tag != nil && flag.Tag.Has(tagPredictor) {
		if p := flag.Tag.Get(tagPredictor); p != "" && p != predictorPath {
			meta.Complete = "predictor=" + p
		}
	}
	if flag.Tag != nil && flag.Tag.Has(tagClib) {
		if err := meta.ParseClibTag(flag.Tag.Get(tagClib)); err != nil {
			panic(err)
		}
	}
	if flag.Tag != nil && flag.Tag.Has(tagNegatable) {
		meta.Negatable = true
	}
	return meta
}
