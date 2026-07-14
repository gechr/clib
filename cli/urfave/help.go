package urfave

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/gechr/clib/help"
	placeholders "github.com/gechr/clib/internal/placeholder"
	xslices "github.com/gechr/x/slices"
	clilib "github.com/urfave/cli/v3"
)

// HelpPrinter returns a func(io.Writer, string, any) suitable for assigning
// to clilib.HelpPrinter (the global variable). The data parameter is the *Command.
// By default, the description blurb and examples are hidden on -h and shown
// on --help (examples last); pass [help.WithAlwaysShowDescription] and/or
// [help.WithAlwaysShowExamples] to disable this.
func HelpPrinter(
	r *help.Renderer,
	sections func(cmd *clilib.Command) []help.Section,
	opts ...help.Option,
) func(io.Writer, string, any) {
	behavior := help.ResolvePolicy(opts...)
	return func(w io.Writer, _ string, data any) {
		cmd, ok := data.(*clilib.Command)
		if !ok {
			// Non-*Command data is silently ignored; urfave calls HelpPrinter with various types.
			return
		}
		s := help.Apply(sections(cmd), opts...)
		if !behavior.AlwaysShowDescription {
			s = help.Apply(s, help.WithDescriptionOnLongHelp(os.Args))
		}
		if !behavior.AlwaysShowExamples {
			s = help.Apply(s, help.WithExamplesOnLongHelp(os.Args))
		}
		cw := colorprofile.NewWriter(w, os.Environ())
		_ = r.Render(cw, s)
	}
}

// Sections builds standard help sections from a urfave command.
// Extracts: Usage, Aliases, Commands, and grouped flag sections.
func Sections(cmd *clilib.Command) []help.Section {
	return buildSections(cmd)
}

// SectionsWithOptions builds standard help sections from a urfave command
// using configurable behavior.
func SectionsWithOptions(opts ...SectionsOption) func(*clilib.Command) []help.Section {
	return func(cmd *clilib.Command) []help.Section {
		return buildSections(cmd, opts...)
	}
}

func buildSections(cmd *clilib.Command, opts ...SectionsOption) []help.Section {
	cfg := sectionsConfig{
		lowercasePlaceholders: true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	prepareFlagExtras(cmd)

	var sections []help.Section

	flagSections, hasFlags := buildFlagSections(cmd, cfg)

	// Commands that only dispatch to subcommands have their flags suppressed -
	// they cannot take effect without picking a subcommand.
	if len(cmd.VisibleCommands()) > 0 {
		flagSections = nil
		hasFlags = false
	}

	sections = append(sections, usageSection(cmd, hasFlags, cfg))

	if len(cmd.Aliases) > 0 {
		sections = append(sections, aliasSection(cmd))
	}

	if cmds := cmd.VisibleCommands(); len(cmds) > 0 {
		sections = append(sections, commandSection(cmds))
	}

	sections = append(sections, flagSections...)

	return sections
}

func usageSection(cmd *clilib.Command, hasFlags bool, cfg sectionsConfig) help.Section {
	var u help.Usage
	if cfg.rawUsage {
		u = help.Usage{Command: cmd.FullName(), Raw: cmd.ArgsUsage}
	} else {
		u = help.Usage{
			Command:     cmd.FullName(),
			ShowOptions: hasFlags,
		}
		if cmd.ArgsUsage != "" {
			u.Args = parseArgsUsage(cmd.ArgsUsage)
		}
		if cmds := cmd.VisibleCommands(); len(cmds) > 0 {
			u.Args = append(
				u.Args,
				help.Arg{Name: "command", Required: cmd.Action == nil, IsSubcommand: true},
			)
		}
	}

	content := []help.Content{u}
	// Surface cmd.Description as a Description blurb below the Usage line,
	// mirroring the kong adapter's handling of node.Detail. cmd.Usage (the
	// short text) is intentionally not used: it already appears in the
	// parent's command list.
	if cmd.Description != "" {
		content = append(content, help.Description(cmd.Description))
	}
	return help.Section{Title: "Usage", Content: content}
}

func aliasSection(cmd *clilib.Command) help.Section {
	return help.Section{
		Title:   "Aliases",
		Content: []help.Content{help.Aliases(cmd.Aliases)},
	}
}

func commandSection(cmds []*clilib.Command) help.Section {
	var group help.CommandGroup
	for _, c := range cmds {
		group = append(group, help.Command{Name: c.Name, Desc: c.Usage})
	}
	return help.Section{
		Title:   "Commands",
		Content: []help.Content{group},
	}
}

// buildFlagSections builds flag help sections from a urfave command.
// If any visible flag has a group (via urfave Category or clib extra),
// flags are split into group sections (alphabetical), with ungrouped local
// flags under "Options" and ungrouped inherited flags under "Global Options".
func buildFlagSections(cmd *clilib.Command, cfg sectionsConfig) ([]help.Section, bool) {
	flagDepths := flagAncestorDepths(cmd)

	var classified []help.ClassifiedFlag
	for _, f := range cmd.Flags {
		if !isVisible(f) {
			continue
		}
		depth := 0
		names := f.Names()
		if len(names) > 0 {
			if d, ok := flagDepths[names[0]]; ok {
				if lf, isLocal := f.(clilib.LocalFlag); !isLocal || !lf.IsLocal() {
					depth = d
				}
			}
		}
		classified = append(classified, help.ClassifiedFlag{
			Flag:          flagToHelp(cfg, cmd, f),
			Group:         flagGroup(cmd, f),
			AncestorDepth: depth,
		})
	}

	var opts []help.FlagSectionsOption
	if cfg.optionsTitle != "" {
		opts = append(opts, help.WithOptionsTitle(cfg.optionsTitle))
	}
	if cfg.globalOptionsTitle != "" {
		opts = append(opts, help.WithGlobalOptionsTitle(cfg.globalOptionsTitle))
	}
	sections := help.BuildFlagSections(classified, opts...)
	return sections, len(sections) > 0
}

// flagAncestorDepths maps a flag name to the depth of the nearest ancestor
// that defines it (1 = immediate parent, 2 = grandparent, ...). Flags defined
// only on the current command are absent from the map.
func flagAncestorDepths(cmd *clilib.Command) map[string]int {
	depths := make(map[string]int)
	lineage := cmd.Lineage()
	for d := 1; d < len(lineage); d++ {
		for _, f := range lineage[d].Flags {
			for _, n := range f.Names() {
				if _, seen := depths[n]; !seen {
					depths[n] = d
				}
			}
		}
	}
	return depths
}

func isVisible(f clilib.Flag) bool {
	if vf, ok := f.(clilib.VisibleFlag); ok {
		return vf.IsVisible()
	}
	return true
}

func flagGroup(cmd *clilib.Command, f clilib.Flag) string {
	if extra := getFlagExtra(cmd, f); extra != nil && extra.Group != "" {
		return extra.Group
	}
	if cf, ok := f.(clilib.CategorizableFlag); ok {
		return cf.GetCategory()
	}
	return ""
}

func flagToHelp(cfg sectionsConfig, cmd *clilib.Command, f clilib.Flag) help.Flag {
	names := f.Names()
	var short, long string
	var isNegatable bool

	if bif, ok := f.(*clilib.BoolWithInverseFlag); ok {
		isNegatable = true
		prefix := bif.InversePrefix
		if prefix == "" {
			prefix = clilib.DefaultInverseBoolPrefix
		}
		long = "[" + prefix + "]" + bif.Name
		for _, n := range bif.Aliases {
			if len(n) == 1 && short == "" {
				short = n
			}
		}
	} else {
		for _, n := range names {
			switch {
			case len(n) > 1 && long == "":
				long = n
			case len(n) == 1 && short == "":
				short = n
			}
		}
		if long == "" && len(names) > 0 {
			long = names[0]
		}
	}

	var usage string
	hasArg := false
	if df, ok := f.(clilib.DocGenerationFlag); ok {
		usage = df.GetUsage()
		hasArg = df.TakesValue()
	}

	isMultiValue := false
	if mv, ok := f.(clilib.DocGenerationMultiValueFlag); ok {
		isMultiValue = mv.IsMultiValueFlag()
	}

	hf := help.Flag{
		Short: short,
		Long:  long,
		Desc:  usage,
	}

	if isNegatable {
		// Placeholder already handled by [no-] prefix in long name.
		return hf
	}

	extra := getFlagExtra(cmd, f)
	if extra != nil && extra.Placeholder != "" {
		hf.Placeholder = normalizePlaceholder(extra.Placeholder, cfg)
		hf.Repeatable = isMultiValue
	} else if hasArg {
		switch {
		case takesFile(f):
			hf.Placeholder = "file"
		case takesInteger(f):
			hf.Placeholder = "n"
		default:
			// Default placeholder is the flag name.
			hf.Placeholder = long
		}
		hf.Repeatable = isMultiValue
	}

	if extra != nil && len(extra.Enum) > 0 {
		hf.Enum = xslices.Unique(extra.Enum)
	}
	if extra != nil && extra.Placeholder == "" {
		if inferred := placeholders.ForEnum(hf.Enum); inferred != "" {
			hf.Placeholder = inferred
		}
	}
	if extra != nil && len(extra.EnumHighlight) > 0 {
		hf.EnumHighlight = extra.EnumHighlight
	}
	if extra != nil && extra.EnumDefault != "" {
		hf.EnumDefault = extra.EnumDefault
	}

	// Fall back to urfave's default value for enum highlighting.
	if hf.EnumDefault == "" && len(hf.Enum) > 0 {
		if df, ok := f.(clilib.DocGenerationFlag); ok {
			if def := df.GetDefaultText(); def != "" {
				hf.EnumDefault = def
			}
		}
	}

	if extra != nil {
		if extra.HideLong {
			hf.Long = ""
		}
		if extra.HideShort {
			hf.Short = ""
		}
		if extra.NoIndent {
			hf.NoIndent = true
		}
	}

	return hf
}

// takesFile reports whether urfave has been told that a string flag accepts a
// file. TakesFile is also used by urfave's own shell-completion generators.
func takesFile(f clilib.Flag) bool {
	switch f := f.(type) {
	case *clilib.StringFlag:
		return f.TakesFile
	case *clilib.StringSliceFlag:
		return f.TakesFile
	default:
		return false
	}
}

// takesInteger reports whether urfave identifies a flag's value as an int or
// uint. Duration flags have their own type name and are not included.
func takesInteger(f clilib.Flag) bool {
	docFlag, ok := f.(clilib.DocGenerationFlag)
	if !ok {
		return false
	}
	return docFlag.TypeName() == "int" || docFlag.TypeName() == "uint"
}

func normalizePlaceholder(placeholder string, cfg sectionsConfig) string {
	if !cfg.lowercasePlaceholders {
		return placeholder
	}
	return strings.ToLower(placeholder)
}

// parseArgsUsage parses an ArgsUsage string into help.Arg entries.
func parseArgsUsage(argsUsage string) []help.Arg {
	parts := strings.Fields(argsUsage)
	var args []help.Arg
	for _, p := range parts {
		args = append(args, help.ParseArg(p))
	}
	return args
}
