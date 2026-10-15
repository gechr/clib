package cobra

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// CompletionOption configures the native completion bridge.
type CompletionOption func(*completionConfig)

type completionConfig struct {
	handler complete.ContextHandler
}

// WithCompletionHandler supplies predictors for flags and dynamic-args.
func WithCompletionHandler(handler complete.ContextHandler) CompletionOption {
	return func(cfg *completionConfig) { cfg.handler = handler }
}

// RegisterCompletions bridges clib metadata to Cobra's native completion API.
// Call it after constructing the command tree and registering native callbacks.
// Existing callbacks and ValidArgs are preserved. The returned generator uses
// native Bash, Fish and Zsh clients; other shells retain their legacy clients.
// Keep handling --@complete requests for previously installed scripts.
//
// The bridge captures completion words in the root's persistent pre-run hook.
// Existing hooks still run; use complete.IsRequest to bypass application setup
// during completion. Register again if commands are added later.
func RegisterCompletions(
	root *cobralib.Command,
	opts ...CompletionOption,
) (*complete.Generator, error) {
	if root == nil {
		return nil, fmt.Errorf("completion root must not be nil")
	}
	var cfg completionConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	if err := registerNativeCommand(root, cfg.handler); err != nil {
		return nil, err
	}
	root.CompletionOptions.SetDefaultShellCompDirective(cobralib.ShellCompDirectiveNoFileComp)
	installCompletionHook(root)
	gen := complete.NewGenerator(root.Name(),
		complete.WithShellGenerator("bash", nativeBash),
		complete.WithShellGenerator("fish", nativeFish),
		complete.WithShellGenerator("zsh", nativeZsh),
	).FromFlags(FlagMeta(root))
	gen.Subs = Subcommands(root)
	gen.DynamicArgs = CommandMeta(root).DynamicArgs
	return gen, nil
}

func registerNativeCommand(cmd *cobralib.Command, handler complete.ContextHandler) error {
	for _, meta := range FlagMeta(cmd) {
		if _, exists := cmd.GetFlagCompletionFunc(meta.Name); exists {
			continue
		}
		// Flags with complete='-' also need a callback to suppress file fallback.
		spec := complete.Spec{}
		complete.ApplyMeta(&spec, &meta)
		if meta.Complete == "-" {
			spec = complete.Spec{}
		}
		flag := cmd.Flags().Lookup(meta.Name)
		if flag == nil {
			flag = cmd.PersistentFlags().Lookup(meta.Name)
		}
		if err := cmd.RegisterFlagCompletionFunc(
			meta.Name,
			nativeFlag(flag, spec, handler),
		); err != nil {
			return fmt.Errorf(
				"register completion for %s --%s: %w",
				cmd.CommandPath(),
				meta.Name,
				err,
			)
		}
	}
	if cmd.ValidArgsFunction == nil && len(cmd.ValidArgs) == 0 {
		cmd.ValidArgsFunction = nativePositionals(cmd, handler)
	}
	for _, child := range cmd.Commands() {
		if err := registerNativeCommand(child, handler); err != nil {
			return err
		}
	}
	return nil
}

func nativeFlag(
	flag *pflag.Flag,
	spec complete.Spec,
	handler complete.ContextHandler,
) cobralib.CompletionFunc {
	return func(cmd *cobralib.Command, args []string, word string) ([]string, cobralib.ShellCompDirective) {
		directive := cobralib.ShellCompDirectiveNoFileComp
		var values []string
		switch {
		case spec.Dynamic != "":
			if handler != nil {
				values, directive = handler(complete.Context{
					Command: cmd, Args: args, Flag: flag, Kind: spec.Dynamic,
					ToComplete: word, AfterDash: AfterDash(cmd),
				})
			}
		case spec.ValueHint == complete.HintDir:
			return nil, cobralib.ShellCompDirectiveFilterDirs
		case spec.Extension != "":
			exts := strings.Split(spec.Extension, ",")
			for i := range exts {
				exts[i] = strings.TrimPrefix(strings.TrimSpace(exts[i]), ".")
			}
			return exts, cobralib.ShellCompDirectiveFilterFileExt
		case spec.ValueHint == complete.HintFile:
			return nil, cobralib.ShellCompDirectiveDefault
		case flag.Value.Type() == pflagTypeBool:
			values = []string{"true", "false"}
		case len(spec.ValueDescs) > 0:
			for _, value := range spec.ValueDescs {
				values = append(values, cobralib.CompletionWithDesc(value.Value, value.Desc))
			}
		default:
			values = spec.Values
		}
		if spec.Order == complete.OrderKeep {
			directive |= cobralib.ShellCompDirectiveKeepOrder
		}
		if directive&(cobralib.ShellCompDirectiveFilterDirs|cobralib.ShellCompDirectiveFilterFileExt) == 0 &&
			(spec.CommaList || flag.Value.Type() == "stringSlice") {
			values = nativeCommaList(values, word)
			directive |= cobralib.ShellCompDirectiveNoSpace
		}
		return values, directive
	}
}

func nativeCommaList(values []string, word string) []string {
	index := strings.LastIndex(word, ",") + 1
	if index == 0 {
		return values
	}
	used := strings.Split(word[:index-1], ",")
	var result []string
	for _, value := range values {
		name, _, _ := strings.Cut(value, "\t")
		if !slices.Contains(used, name) {
			result = append(result, word[:index]+value)
		}
	}
	return result
}

func nativePositionals(
	cmd *cobralib.Command,
	handler complete.ContextHandler,
) cobralib.CompletionFunc {
	meta := CommandMeta(cmd)
	legacyArgs := cmd.Args == nil
	limit, limited := commandPositionalLimit(cmd)
	return func(cmd *cobralib.Command, args []string, word string) ([]string, cobralib.ShellCompDirective) {
		if legacyArgs && !cmd.HasParent() && cmd.HasSubCommands() && len(args) > 0 {
			return nil, cobralib.ShellCompDirectiveNoFileComp
		}
		if len(meta.DynamicArgs) > 0 && handler != nil {
			// Let the predictor decide how to handle passthrough arguments after --.
			if !AfterDash(cmd) && limited && len(args) >= limit {
				return nil, cobralib.ShellCompDirectiveNoFileComp
			}
			kind := meta.DynamicArgs[min(len(args), len(meta.DynamicArgs)-1)]
			return handler(complete.Context{
				Command: cmd, Args: args, Kind: kind, ToComplete: word, AfterDash: AfterDash(cmd),
			})
		}
		if meta.PathArgs && (!limited || len(args) < limit) {
			return nil, cobralib.ShellCompDirectiveDefault
		}
		return nil, cobralib.ShellCompDirectiveNoFileComp
	}
}

type completionWordsKey struct{}

// AfterDash reports whether the native completion cursor follows a -- boundary.
// It is also available to callbacks registered directly with Cobra. Unlike
// ArgsLenAtDash, this is not affected by Cobra's probe parse of an extra --.
// This is a lexical check: a standalone -- used as a flag value also counts.
func AfterDash(cmd *cobralib.Command) bool {
	if cmd == nil || cmd.Context() == nil {
		return false
	}
	words, _ := cmd.Context().Value(completionWordsKey{}).([]string)
	return len(words) > 0 && slices.Contains(words[:len(words)-1], "--")
}

const nativeHookKey = "clib.native-completion"

func installCompletionHook(root *cobralib.Command) {
	if root.Annotations[nativeHookKey] == "registered" {
		return
	}
	if root.Annotations == nil {
		root.Annotations = make(map[string]string)
	}
	root.Annotations[nativeHookKey] = "registered"
	preRunE, preRun := root.PersistentPreRunE, root.PersistentPreRun
	root.PersistentPreRunE = func(cmd *cobralib.Command, args []string) error {
		var restore func()
		if cmd.Name() == cobralib.ShellCompRequestCmd {
			cmd.SetContext(
				context.WithValue(cmd.Context(), completionWordsKey{}, slices.Clone(args)),
			)
			restore = tolerateCompletionErrors(root)
			run := cmd.Run
			cmd.Run = func(cmd *cobralib.Command, args []string) {
				defer restore()
				defer func() { cmd.Run = run }()
				if unknownCompletionFlag(root, args) {
					fmt.Fprintln(cmd.OutOrStdout(), ":4")
					return
				}
				run(cmd, args)
			}
		}
		if preRunE != nil {
			if err := preRunE(cmd, args); err != nil {
				if restore != nil {
					restore()
				}
				return err
			}
		} else if preRun != nil {
			preRun(cmd, args)
		}
		return nil
	}
}

func tolerateCompletionErrors(root *cobralib.Command) func() {
	args := root.Args
	if root.Args == nil {
		root.Args = cobralib.ArbitraryArgs
	}
	var restores []func()
	var walk func(*cobralib.Command)
	walk = func(cmd *cobralib.Command) {
		unknown := cmd.FParseErrWhitelist.UnknownFlags
		cmd.FParseErrWhitelist.UnknownFlags = true
		restores = append(restores, func() { cmd.FParseErrWhitelist.UnknownFlags = unknown })
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)
	return func() {
		root.Args = args
		for _, restore := range restores {
			restore()
		}
	}
}

// Cobra's UnknownFlags whitelist does not cover a flag whose value is currently
// being completed. Guard that one case before Cobra emits a file directive.
func unknownCompletionFlag(root *cobralib.Command, words []string) bool {
	if len(words) == 0 || slices.Contains(words[:len(words)-1], "--") {
		return false
	}
	cmd, args, err := root.Find(words[:len(words)-1])
	if err != nil || cmd.DisableFlagParsing {
		return false
	}
	word := words[len(words)-1]
	if strings.HasPrefix(word, "-") {
		var hasEqual bool
		word, _, hasEqual = strings.Cut(word, "=")
		if !hasEqual {
			return false
		}
	} else if len(args) > 0 {
		word = args[len(args)-1]
		if strings.Contains(word, "=") {
			return false
		}
	}
	if strings.HasPrefix(word, "--") {
		return cmd.Flags().Lookup(word[2:]) == nil &&
			cmd.InheritedFlags().Lookup(word[2:]) == nil &&
			cmd.PersistentFlags().Lookup(word[2:]) == nil &&
			word != "--help" &&
			word != "--version"
	}
	if strings.HasPrefix(word, "-") && len(word) > 1 {
		name := word[len(word)-1:]
		return cmd.Flags().ShorthandLookup(name) == nil &&
			cmd.InheritedFlags().ShorthandLookup(name) == nil &&
			cmd.PersistentFlags().ShorthandLookup(name) == nil &&
			name != "h"
	}
	return false
}
