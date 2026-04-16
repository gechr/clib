package urfave

import (
	"io"
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/gechr/clib/help"
	clilib "github.com/urfave/cli/v3"
)

// HelpPrinter returns a func(io.Writer, string, any) suitable for assigning
// to clilib.HelpPrinter (the global variable). The data parameter is the *Command.
// By default, examples are hidden on -h and shown last on --help;
// pass [help.WithAlwaysShowExamples] to disable this.
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
	prepareFlagExtras(cmd)

	var sections []help.Section

	flagSections, hasFlags := buildFlagSections(cmd)

	sections = append(sections, usageSection(cmd, hasFlags))

	if len(cmd.Aliases) > 0 {
		sections = append(sections, aliasSection(cmd))
	}

	if cmds := cmd.VisibleCommands(); len(cmds) > 0 {
		sections = append(sections, commandSection(cmds))
	}

	sections = append(sections, flagSections...)

	return sections
}

func usageSection(cmd *clilib.Command, hasFlags bool) help.Section {
	u := help.Usage{
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

	return help.Section{
		Title:   "Usage",
		Content: []help.Content{u},
	}
}

func aliasSection(cmd *clilib.Command) help.Section {
	return help.Section{
		Title:   "Aliases",
		Content: []help.Content{help.Text(strings.Join(cmd.Aliases, ", "))},
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
func buildFlagSections(cmd *clilib.Command) ([]help.Section, bool) {
	localFlags, inheritedFlags := splitFlags(cmd)

	var classified []help.ClassifiedFlag
	classifyFlags := func(flags []clilib.Flag, inherited bool) {
		for _, f := range flags {
			if !isVisible(f) {
				continue
			}
			classified = append(classified, help.ClassifiedFlag{
				Flag:      flagToHelp(cmd, f),
				Group:     flagGroup(cmd, f),
				Inherited: inherited,
			})
		}
	}
	classifyFlags(localFlags, false)
	classifyFlags(inheritedFlags, true)

	sections := help.BuildFlagSections(classified)
	return sections, len(sections) > 0
}

// splitFlags separates a command's flags into local and inherited.
// Inherited flags come from ancestor commands (Lineage()[1:]).
func splitFlags(cmd *clilib.Command) ([]clilib.Flag, []clilib.Flag) {
	// Collect inherited flag names from ancestors.
	inheritedNames := make(map[string]bool)
	lineage := cmd.Lineage()
	for _, ancestor := range lineage[1:] {
		for _, f := range ancestor.Flags {
			for _, n := range f.Names() {
				inheritedNames[n] = true
			}
		}
	}

	var local, inherited []clilib.Flag
	for _, f := range cmd.Flags {
		names := f.Names()
		if len(names) > 0 && inheritedNames[names[0]] {
			// Non-local flag that came from an ancestor.
			if lf, ok := f.(clilib.LocalFlag); ok && lf.IsLocal() {
				local = append(local, f)
			} else {
				inherited = append(inherited, f)
			}
		} else {
			local = append(local, f)
		}
	}
	return local, inherited
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

func flagToHelp(cmd *clilib.Command, f clilib.Flag) help.Flag {
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
		hf.Placeholder = extra.Placeholder
		hf.Repeatable = isMultiValue
	} else if hasArg {
		// Default placeholder is the flag name.
		hf.Placeholder = long
		hf.Repeatable = isMultiValue
	}

	if extra != nil && len(extra.Enum) > 0 {
		hf.Enum = extra.Enum
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

// parseArgsUsage parses an ArgsUsage string into help.Arg entries.
func parseArgsUsage(argsUsage string) []help.Arg {
	parts := strings.Fields(argsUsage)
	var args []help.Arg
	for _, p := range parts {
		args = append(args, help.ParseArg(p))
	}
	return args
}
