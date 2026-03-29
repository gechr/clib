package cobra

import (
	"os"
	"strings"

	"github.com/charmbracelet/colorprofile"
	"github.com/gechr/clib/help"
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// HelpFunc returns a cobra-compatible help function that renders themed help.
// The sections callback receives the command and returns sections to render.
func HelpFunc(
	r *help.Renderer,
	sections func(cmd *cobralib.Command) []help.Section,
	opts ...help.Option,
) func(*cobralib.Command, []string) {
	return func(cmd *cobralib.Command, _ []string) {
		w := colorprofile.NewWriter(cmd.OutOrStdout(), os.Environ())
		renderSections := help.Apply(sections(cmd), opts...)
		setUsageShowOptions(renderSections)
		_ = r.Render(w, renderSections)
	}
}

type sectionsConfig struct {
	keepGroupOrder                  bool
	sortGroupOrder                  bool
	hideInheritedFlags              bool
	hideInheritedFlagsOnSubcommands bool
	showInheritedFlagsOnSubcommands bool
}

// SectionsOption configures cobra help-section generation.
type SectionsOption func(*sectionsConfig)

// WithKeepGroupOrder preserves first-seen order of grouped flag sections
// instead of sorting them alphabetically. This is the default.
func WithKeepGroupOrder() SectionsOption {
	return func(c *sectionsConfig) {
		c.keepGroupOrder = true
		c.sortGroupOrder = false
	}
}

// WithSortedGroupOrder sorts grouped flag sections alphabetically.
func WithSortedGroupOrder() SectionsOption {
	return func(c *sectionsConfig) {
		c.keepGroupOrder = false
		c.sortGroupOrder = true
	}
}

// WithHideInheritedFlags omits inherited/global flags from help output.
func WithHideInheritedFlags() SectionsOption {
	return func(c *sectionsConfig) {
		c.hideInheritedFlags = true
	}
}

// WithHideInheritedFlagsOnSubcommands omits inherited/global flags from
// subcommand help output while leaving root-command help unchanged. This is
// the default.
func WithHideInheritedFlagsOnSubcommands() SectionsOption {
	return func(c *sectionsConfig) {
		c.hideInheritedFlagsOnSubcommands = true
		c.showInheritedFlagsOnSubcommands = false
	}
}

// WithShowInheritedFlagsOnSubcommands keeps inherited/global flags visible in
// subcommand help output.
func WithShowInheritedFlagsOnSubcommands() SectionsOption {
	return func(c *sectionsConfig) {
		c.hideInheritedFlagsOnSubcommands = false
		c.showInheritedFlagsOnSubcommands = true
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
	}
	for _, opt := range opts {
		opt(&cfg)
	}

	var sections []help.Section

	flagSections, hasFlags := buildFlagSections(cmd, cfg)

	sections = append(sections, usageSection(cmd, hasFlags))

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

func usageSection(cmd *cobralib.Command, hasFlags bool) help.Section {
	u := help.Usage{
		Command:     cmd.CommandPath(),
		ShowOptions: hasFlags,
		Args:        parseUseArgs(cmd.Use),
	}
	if len(availableCommands(cmd)) > 0 {
		u.Args = append(
			u.Args,
			help.Arg{Name: "command", Required: !cmd.Runnable(), IsSubcommand: true},
		)
	}
	return help.Section{
		Title:   "Usage",
		Content: []help.Content{u},
	}
}

func aliasSection(cmd *cobralib.Command) help.Section {
	return help.Section{
		Title:   "Aliases",
		Content: []help.Content{help.Text(strings.Join(cmd.Aliases, ", "))},
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
// through to "Options" (local) or "Inherited Options" (inherited).
// If no flag has a group, the flat "Options" / "Inherited Options" layout is used.
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

	classifyFlags := func(flags *pflag.FlagSet, inherited bool) {
		if inherited && hideInherited {
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
				Flag:      pflagToHelpFlag(f),
				Group:     group,
				Inherited: inherited,
			})
		})
	}
	classifyFlags(cmd.LocalFlags(), false)
	classifyFlags(cmd.InheritedFlags(), true)

	var opts []help.FlagSectionsOption
	if cfg.keepGroupOrder && !cfg.sortGroupOrder {
		opts = append(opts, help.KeepGroupOrder())
	}

	sections := help.BuildFlagSections(classified, opts...)
	return sections, len(sections) > 0
}

func pflagToHelpFlag(f *pflag.Flag) help.Flag {
	hf := help.Flag{
		Short: f.Shorthand,
		Long:  f.Name,
		Desc:  f.Usage,
	}
	typeName := f.Value.Type()
	isRepeatable := strings.Contains(typeName, "Slice") || strings.Contains(typeName, "Array")
	extra := getExtra(f)
	if extra != nil && extra.Placeholder != "" {
		hf.Placeholder = extra.Placeholder
		hf.Repeatable = isRepeatable
	} else if typeName != pflagTypeBool {
		hf.Placeholder = f.Name
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
	lines := strings.Split(s, "\n")
	var examples help.Examples
	var currentComment string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
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
