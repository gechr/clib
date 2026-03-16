package kong

import (
	"maps"
	"slices"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/bash" // register shell generators
	_ "github.com/gechr/clib/complete/fish" // register shell generators
	_ "github.com/gechr/clib/complete/zsh"  // register shell generators
	shellpkg "github.com/gechr/clib/shell"
)

type config struct {
	quiet bool
	args  []string
}

// CompletionFlags provides embeddable Kong flags for shell completion management.
// Embed this in your CLI struct to get --@complete, --@shell, --install-completion,
// --uninstall-completion, and --print-completion flags automatically.
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
	opts ...Option,
) (bool, error) {
	var cfg config
	for _, o := range opts {
		o(&cfg)
	}

	shell := f.Shell
	if shell == "" {
		shell = shellpkg.Detect()
	}

	return complete.HandleAction(complete.Action{
		Shell:               shell,
		Complete:            f.Complete,
		Args:                cfg.args,
		InstallCompletion:   f.InstallCompletion,
		UninstallCompletion: f.UninstallCompletion,
		PrintCompletion:     f.PrintCompletion,
	}, gen, handler, cfg.quiet)
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
			meta.ParseClibTag(child.Tag.Get(tagClib))
			if meta.Terse != "" {
				sub.Terse = meta.Terse
			}
			if meta.Complete == predictorPath {
				sub.PathArgs = true
			}
		}
		if !sub.PathArgs {
			for _, arg := range child.Positional {
				if arg.Tag != nil && arg.Tag.Has(tagPredictor) &&
					arg.Tag.Get(tagPredictor) == predictorPath {
					sub.PathArgs = true
					break
				}
			}
		}
		// Recurse into nested subcommands.
		sub.Subs = nodeSubSpecs(child)
		subs = append(subs, sub)
	}
	return subs
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
		meta.ParseClibTag(flag.Tag.Get(tagClib))
	}
	if flag.Tag != nil && flag.Tag.Has(tagNegatable) {
		meta.Negatable = true
	}
	return meta
}
