package cobra

import (
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/gechr/clib/help"
	xstrings "github.com/gechr/x/strings"
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// HelpFunc returns a cobra-compatible help function that renders themed help.
// The sections callback receives the command and returns sections to render.
// By default, the description blurb and examples are hidden on -h and shown
// on --help (examples last); pass [help.WithAlwaysShowDescription] and/or
// [help.WithAlwaysShowExamples] to disable this.
func HelpFunc(
	r *help.Renderer,
	sections func(cmd *cobralib.Command) []help.Section,
	opts ...help.Option,
) func(*cobralib.Command, []string) {
	return func(cmd *cobralib.Command, _ []string) {
		w := colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ())
		behavior := help.ResolvePolicy(opts...)
		renderSections := help.Apply(sections(cmd), opts...)
		if !behavior.AlwaysShowDescription {
			renderSections = help.Apply(renderSections, help.WithDescriptionOnLongHelp(os.Args))
		}
		if !behavior.AlwaysShowExamples {
			renderSections = help.Apply(renderSections, help.WithExamplesOnLongHelp(os.Args))
		}
		setUsageShowOptions(renderSections)
		_ = r.Render(w, renderSections)
	}
}

// SectionsWithOptions builds standard help sections from a cobra command using
// configurable flag-section behavior.
func SectionsWithOptions(opts ...SectionsOption) func(*cobralib.Command) []help.Section {
	return func(cmd *cobralib.Command) []help.Section {
		return buildSections(cmd, opts...)
	}
}

// Sections builds standard help sections from a cobra command.
// Extracts: Usage, Aliases, Examples, grouped Subcommands, Flags, Inherited Flags.
//
// When any flag carries a clib "group" extra, flags are organized into
// one section per group (alphabetical), with ungrouped local flags under "Flags"
// and ungrouped inherited flags under "Inherited Flags".
func Sections(cmd *cobralib.Command) []help.Section {
	return buildSections(cmd)
}

func buildSections(cmd *cobralib.Command, opts ...SectionsOption) []help.Section {
	cfg := sectionsConfig{
		keepGroupOrder:                  true,
		hideInheritedFlagsOnSubcommands: true,
		lowercasePlaceholders:           true,
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var sections []help.Section

	// Commands that only dispatch to subcommands have their flags suppressed -
	// they cannot take effect without picking a subcommand.
	subcommandOnlyGrouper := len(availableCommands(cmd)) > 0

	flagSections, hasFlags := buildFlagSections(cmd, cfg)
	if subcommandOnlyGrouper {
		flagSections = nil
		hasFlags = false
	}

	sections = append(sections, usageSection(cmd, hasFlags, cfg))

	if len(cmd.Aliases) > 0 {
		sections = append(sections, aliasSection(cmd))
	}

	if cmd.Example != "" {
		sections = append(sections, examplesSection(cmd))
	}

	sections = append(sections, subcommandSections(cmd)...)
	sections = append(sections, flagSections...)

	return sections
}

func usageSection(cmd *cobralib.Command, hasFlags bool, cfg sectionsConfig) help.Section {
	var u help.Usage
	if cfg.rawUsage {
		u = help.Usage{Command: cmd.CommandPath(), Raw: rawUseSuffix(cmd)}
	} else {
		u = help.Usage{
			Command:     cmd.CommandPath(),
			ShowOptions: hasFlags,
			Args:        withValidArgsEnum(parseUseArgs(cmd.Use), cmd.ValidArgs),
		}
		if len(availableCommands(cmd)) > 0 {
			u.Args = append(
				u.Args,
				help.Arg{Name: "command", Required: !cfg.subcommandOptional, IsSubcommand: true},
			)
		}
	}

	content := []help.Content{u}
	// Surface cmd.Long as a Description blurb below the Usage line, mirroring
	// the kong adapter's handling of node.Detail. cmd.Short is intentionally
	// not used: it already appears in the parent's command list.
	if cmd.Long != "" {
		content = append(content, help.Description(cmd.Long))
	}
	return help.Section{Title: "Usage", Content: content}
}

// rawUseSuffix returns cmd.Use with the leading command name stripped, so the
// caller can prepend cmd.CommandPath() without duplicating the name.
func rawUseSuffix(cmd *cobralib.Command) string {
	return strings.TrimSpace(strings.TrimPrefix(cmd.Use, cmd.Name()))
}

func aliasSection(cmd *cobralib.Command) help.Section {
	return help.Section{
		Title:   "Aliases",
		Content: []help.Content{help.Aliases(cmd.Aliases)},
	}
}

func examplesSection(cmd *cobralib.Command) help.Section {
	return help.Section{
		Title:   "Examples",
		Content: []help.Content{parseExamples(cmd.Example)},
	}
}

func subcommandSections(cmd *cobralib.Command) []help.Section {
	available := availableCommands(cmd)
	if len(available) == 0 {
		return nil
	}

	groups := cmd.Groups()
	if len(groups) == 0 {
		return []help.Section{{
			Title:   "Commands",
			Content: []help.Content{formatCommandList(available)},
		}}
	}

	var sections []help.Section
	grouped := make(map[string][]*cobralib.Command)
	var ungrouped []*cobralib.Command

	for _, c := range available {
		if c.GroupID != "" {
			grouped[c.GroupID] = append(grouped[c.GroupID], c)
		} else {
			ungrouped = append(ungrouped, c)
		}
	}

	for _, g := range groups {
		cmds := grouped[g.ID]
		if len(cmds) == 0 {
			continue
		}
		sections = append(sections, help.Section{
			Title:   g.Title,
			Content: []help.Content{formatCommandList(cmds)},
		})
	}

	if len(ungrouped) > 0 {
		sections = append(sections, help.Section{
			Title:   "Additional Commands",
			Content: []help.Content{formatCommandList(ungrouped)},
		})
	}

	return sections
}

func availableCommands(cmd *cobralib.Command) []*cobralib.Command {
	var cmds []*cobralib.Command
	for _, c := range cmd.Commands() {
		if c.IsAvailableCommand() {
			cmds = append(cmds, c)
		}
	}
	return cmds
}

func formatCommandList(cmds []*cobralib.Command) help.CommandGroup {
	var group help.CommandGroup
	for _, c := range cmds {
		group = append(group, help.Command{Name: c.Name(), Desc: c.Short})
	}
	return group
}

// buildFlagSections builds flag help sections from a cobra command.
// If any visible flag has a clib "group" extra, flags are split into
// one section per group (alphabetical order), with ungrouped flags falling
// through to "Options" (local) or "Global Options" (inherited).
// If no flag has a group, the flat "Options" / "Global Options" layout is used.
func buildFlagSections(cmd *cobralib.Command, cfg sectionsConfig) ([]help.Section, bool) {
	var classified []help.ClassifiedFlag
	hideInherited := cfg.hideInheritedFlags ||
		(cfg.hideInheritedFlagsOnSubcommands && cmd.HasParent())

	sortLocal := cmd.Flags().SortFlags
	sortPersistent := cmd.PersistentFlags().SortFlags
	if cfg.keepGroupOrder {
		cmd.Flags().SortFlags = false
		cmd.PersistentFlags().SortFlags = false
		defer func() {
			cmd.Flags().SortFlags = sortLocal
			cmd.PersistentFlags().SortFlags = sortPersistent
		}()
	}

	classifyFlags := func(flags *pflag.FlagSet, depth int) {
		if depth > 0 && hideInherited {
			return
		}
		flags.VisitAll(func(f *pflag.Flag) {
			if f.Hidden {
				return
			}
			var group string
			if extra := getExtra(f); extra != nil {
				group = extra.Group
			}
			classified = append(classified, help.ClassifiedFlag{
				Flag:          pflagToHelpFlag(cfg, f),
				Group:         group,
				AncestorDepth: depth,
			})
		})
	}
	classifyFlags(cmd.LocalFlags(), 0)
	depth := 1
	for p := cmd.Parent(); p != nil; p = p.Parent() {
		classifyFlags(p.PersistentFlags(), depth)
		depth++
	}

	var opts []help.FlagSectionsOption
	if cfg.keepGroupOrder && !cfg.sortGroupOrder {
		opts = append(opts, help.WithKeepGroupOrder())
	}

	sections := help.BuildFlagSections(classified, opts...)
	return sections, len(sections) > 0
}

func pflagToHelpFlag(cfg sectionsConfig, f *pflag.Flag) help.Flag {
	hf := help.Flag{
		Short: f.Shorthand,
		Long:  f.Name,
		Desc:  f.Usage,
	}
	typeName := f.Value.Type()
	isRepeatable := xstrings.ContainsAny(typeName, "Slice", "Array")
	extra := getExtra(f)
	// Only extract a backquoted placeholder for non-bool flags: bool flags
	// take no value, so a placeholder is nonsensical and backticks in their
	// usage are inline-code markers that must survive to the renderer.
	switch {
	case extra != nil && extra.Placeholder != "":
		hf.Placeholder = normalizePlaceholder(extra.Placeholder, cfg)
		hf.Repeatable = isRepeatable
	case typeName != pflagTypeBool:
		usagePlaceholder, usage, hasUsagePlaceholder := splitPflagUsagePlaceholder(f.Usage)
		if hasUsagePlaceholder {
			hf.Placeholder = normalizePlaceholder(usagePlaceholder, cfg)
			hf.Desc = usage
		} else if placeholder := completionPlaceholder(f); placeholder != "" {
			hf.Placeholder = placeholder
		} else if isIntegerPflagType(typeName) {
			hf.Placeholder = "n"
		} else {
			hf.Placeholder = f.Name
		}
		hf.Repeatable = isRepeatable
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

	// Fall back to pflag's default value for enum highlighting.
	if hf.EnumDefault == "" && len(hf.Enum) > 0 && f.DefValue != "" {
		hf.EnumDefault = f.DefValue
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

// completionPlaceholder derives a placeholder from Cobra's native file and
// directory completion annotations.
func completionPlaceholder(f *pflag.Flag) string {
	if _, ok := f.Annotations[cobralib.BashCompSubdirsInDir]; ok {
		return "dir"
	}
	if _, ok := f.Annotations[cobralib.BashCompFilenameExt]; ok {
		return "file"
	}
	return ""
}

func isIntegerPflagType(typeName string) bool {
	typeName = strings.TrimSuffix(typeName, "Slice")
	typeName = strings.TrimSuffix(typeName, "Array")
	switch typeName {
	case "int", "int8", "int16", "int32", "int64",
		"uint", "uint8", "uint16", "uint32", "uint64":
		return true
	default:
		return false
	}
}

func normalizePlaceholder(placeholder string, cfg sectionsConfig) string {
	if !cfg.lowercasePlaceholders {
		return placeholder
	}
	return strings.ToLower(placeholder)
}

// splitPflagUsagePlaceholder extracts the first backquoted token from a
// pflag-style usage string and returns it as a value placeholder (mirroring
// the convention used by `flag.UnquoteUsage`).
//
// Backquoted content that looks like a flag reference (e.g. `--verbose`) is
// NOT treated as a placeholder - flag references in descriptions are common
// (e.g. "Alias for `--quiet=0`") and should be rendered as inline code, not
// pulled out as the placeholder. In that case the original usage string is
// returned unchanged so downstream renderers can preserve the backticks as
// inline code markers. When a backquoted token IS treated as a placeholder,
// the surrounding backticks are stripped from the returned usage string
// (pflag convention).
func splitPflagUsagePlaceholder(usage string) (string, string, bool) {
	for i := range len(usage) {
		if usage[i] != '`' {
			continue
		}
		for j := i + 1; j < len(usage); j++ {
			if usage[j] != '`' {
				continue
			}
			content := usage[i+1 : j]
			// Backticked flag references are inline code, not placeholders.
			// Preserve the original usage (with backticks) for the renderer.
			if strings.HasPrefix(content, "-") {
				return "", usage, false
			}
			unquotedUsage := usage[:i] + content + usage[j+1:]
			return content, unquotedUsage, true
		}
		break
	}
	return "", usage, false
}

// withValidArgsEnum attaches cobra's ValidArgs to each parsed positional arg
// as its Enum set, so backtick references to those values pick up the arg
// color in help output. cobra validates every positional against the same
// ValidArgs list, so the set is shared across args. Entries may carry a
// "value\tdescription" completion annotation; only the value is kept.
func withValidArgsEnum(args []help.Arg, validArgs []string) []help.Arg {
	if len(args) == 0 || len(validArgs) == 0 {
		return args
	}
	enum := make([]string, 0, len(validArgs))
	for _, v := range validArgs {
		value, _, _ := strings.Cut(v, "\t")
		if value != "" {
			enum = append(enum, value)
		}
	}
	if len(enum) == 0 {
		return args
	}
	for i := range args {
		args[i].Enum = enum
	}
	return args
}

func parseUseArgs(use string) []help.Arg {
	parts := strings.Fields(use)
	if len(parts) <= 1 {
		return nil
	}
	var args []help.Arg
	for _, p := range parts[1:] {
		if p == help.OptOpen+"flags"+help.OptClose || p == help.OptOpen+"command"+help.OptClose {
			continue
		}
		args = append(args, help.ParseArg(p))
	}
	return args
}

func setUsageShowOptions(sections []help.Section) {
	hasFlags := false
	for _, section := range sections {
		if sectionHasFlagContent(section.Content) {
			hasFlags = true
			break
		}
	}
	if len(sections) == 0 || len(sections[0].Content) == 0 {
		return
	}
	firstSection := &sections[0]
	firstContent := firstSection.Content
	firstItem := firstContent[0]

	usage, ok := firstItem.(help.Usage)
	if !ok {
		return
	}
	usage.ShowOptions = hasFlags
	firstContent[0] = usage
	firstSection.Content = firstContent
}

func sectionHasFlagContent(content []help.Content) bool {
	for _, item := range content {
		if _, ok := item.(help.FlagGroup); ok {
			return true
		}
	}
	return false
}

func parseExamples(s string) help.Examples {
	var examples help.Examples
	var currentComment string

	for _, line := range xstrings.SplitLines(s) {
		if after, ok := strings.CutPrefix(line, "#"); ok {
			currentComment = strings.TrimSpace(after)
		} else if after, ok := strings.CutPrefix(line, "$"); ok {
			cmd := strings.TrimSpace(after)
			examples = append(examples, help.Example{
				Comment: currentComment,
				Command: cmd,
			})
			currentComment = ""
		}
	}
	return examples
}
