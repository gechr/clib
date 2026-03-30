package help

import (
	"slices"
	"strings"
)

// Content is anything that can appear inside a help section.
type Content interface {
	helpContent()
}

// Flag describes a single flag entry.
// Short and Long should not include dashes - the renderer adds them.
// Placeholder is rendered as <placeholder> (angle brackets added by renderer).
// Set PlaceholderLiteral to true to render the placeholder as-is without <...>.
type Flag struct {
	Desc               string
	Enum               []string // values to render as [v1, v2, ...]
	EnumDefault        string   // default value annotation appended after enum list
	EnumHighlight      []string // highlight substrings (parallel to Enum, used with EnumStyleHighlightPrefix)
	Long               string   // "repo" -> rendered as --repo
	NoIndent           bool     // true -> suppress short-flag alignment indent for long-only flags
	Placeholder        string   // "repo" -> rendered as <repo>
	PlaceholderLiteral bool     // true -> renders placeholder without <...>
	Repeatable         bool     // true -> renders <placeholder>,…
	Short              string   // "R" -> rendered as -R
}

// Arg describes a positional argument.
type Arg struct {
	Name         string // "query"
	Desc         string
	Required     bool // true -> <query>, false -> [query]
	Repeatable   bool // true -> appends "…" suffix
	IsSubcommand bool // true -> this arg represents a subcommand placeholder
}

// FlagGroup is a group of flag entries (blank line separates adjacent groups).
type FlagGroup []Flag

func (FlagGroup) helpContent() {}

// Args is a group of positional argument entries.
type Args []Arg

func (Args) helpContent() {}

// Command describes a subcommand entry.
type Command struct {
	Name string
	Desc string
}

// CommandGroup is a group of subcommand entries.
type CommandGroup []Command

func (CommandGroup) helpContent() {}

// Usage is an auto-styled usage line.
type Usage struct {
	Command     string // "prl" -> styled as HelpCommand
	ShowOptions bool   // true -> renders [options] in HelpFlag style
	Args        []Arg  // positional args with bracket style
}

func (Usage) helpContent() {}

// Text is freeform pre-styled text.
type Text string

func (Text) helpContent() {}

// Example describes a help example with a comment and command.
type Example struct {
	Comment string
	Command string
}

// Examples is a group of example entries.
type Examples []Example

func (Examples) helpContent() {}

// Section is a named section containing content blocks.
type Section struct {
	Title   string
	Content []Content
}

func (*Section) helpContent() {}

// Alignment controls how names are aligned against the description column.
type Alignment int

const (
	AlignLeft  Alignment = iota // Left-align names (default).
	AlignRight                  // Right-align names against the description column.
)

// AlignMode controls whether alignment is computed per section or globally.
type AlignMode int

const (
	AlignModeSection AlignMode = iota // Align within each section independently (default).
	AlignModeGlobal                   // Align across all sections using a shared column.
)

// WrapStyle controls how wrapped description continuation lines are indented.
type WrapStyle int

const (
	// WrapBracketAlign indents continuation lines to the content after an
	// unclosed '[' on the first line, keeping bracketed lists (like enum
	// values) visually cohesive. Falls back to WrapFlush when no unclosed
	// bracket is present.
	WrapBracketAlign WrapStyle = iota

	// WrapBracketBelow breaks before a trailing '[...]', placing the bracket
	// content on a new line at the description column. Continuation lines
	// within the bracket are indented one column further to align with the
	// content after '['. Falls back to WrapFlush when no trailing bracket
	// is present.
	WrapBracketBelow

	// WrapFlush indents all continuation lines to the description column.
	WrapFlush
)

// Docopt-style argument syntax tokens.
const (
	EllipsisShort = "…"
	EllipsisLong  = "..."

	ArgOpen       = "<"
	ArgClose      = ">"
	OptOpen       = "["
	OptClose      = "]"
	ArgRepeatable = EllipsisShort
	NoteOpen      = "("
	NoteClose     = ")"
)

// ParseArg parses a docopt-style argument token into an Arg.
// It handles optional brackets ([...]), angle brackets (<...>), and
// ellipsis (...) for repeated arguments.
func ParseArg(s string) Arg {
	// Strip outer brackets to determine optional vs required.
	required := true
	inner := s
	if after, ok := strings.CutPrefix(inner, OptOpen); ok {
		if trimmed, ok := strings.CutSuffix(after, OptClose); ok {
			required = false
			inner = trimmed
		}
	}

	// Strip repeated suffix (may be inside or outside angle brackets).
	repeated := false
	if after, ok := strings.CutSuffix(inner, EllipsisLong); ok {
		repeated = true
		inner = after
	} else if after, ok := strings.CutSuffix(inner, EllipsisShort); ok {
		repeated = true
		inner = after
	}

	// Strip angle brackets.
	if after, ok := strings.CutPrefix(inner, ArgOpen); ok {
		if trimmed, ok := strings.CutSuffix(after, ArgClose); ok {
			inner = trimmed
		}
	}
	if !repeated {
		if after, ok := strings.CutSuffix(inner, EllipsisLong); ok {
			repeated = true
			inner = after
		} else if after, ok := strings.CutSuffix(inner, EllipsisShort); ok {
			repeated = true
			inner = after
		}
	}

	return Arg{
		Name:       inner,
		Required:   required,
		Repeatable: repeated,
	}
}

// BracketArg returns the arg formatted in docopt style:
//
//	required:          <name>
//	required+repeated: <name>…
//	optional:          [<name>]
//	optional+repeated: [<name>…]
func BracketArg(a Arg) string {
	inner := ArgOpen + a.Name + ArgClose
	if a.Repeatable {
		inner += ArgRepeatable
	}
	if a.Required {
		return inner
	}
	return OptOpen + inner + OptClose
}

// IsLongHelp reports whether --help appears in args (before any "--"
// separator). args is expected to include the program name at index 0.
func IsLongHelp(args []string) bool {
	for i, arg := range args {
		if i == 0 {
			continue
		}
		if arg == "--help" {
			return true
		}
		if arg == "--" {
			break
		}
	}
	return false
}

// SplitHelpFlags removes any Flag with Long=="help" from all sections,
// then appends separate -h and --help entries as a new FlagGroup to the
// last section containing flag content. Empty FlagGroups and sections
// are cleaned up.
func SplitHelpFlags(sections []Section, shortDesc, longDesc string) []Section {
	return SplitHelpFlagsInSection(sections, "", shortDesc, longDesc)
}

// SplitHelpFlagsInSection removes any Flag with Long=="help" from all
// sections, then appends separate -h and --help entries as a new FlagGroup to
// sectionTitle. When sectionTitle is empty, the help group is appended to the
// last section containing flag content and falls back to "Options" if no flag
// sections exist. Empty FlagGroups and sections are cleaned up.
func SplitHelpFlagsInSection(
	sections []Section,
	sectionTitle, shortDesc, longDesc string,
) []Section {
	sections = removeHelpFlags(sections)
	helpGroup := newHelpFlagGroup(shortDesc, longDesc)
	return appendFlagGroupToSection(sections, sectionTitle, helpGroup)
}

// MoveHelpFlagsToSection moves existing help flags into sectionTitle. It
// preserves whether the help flags are combined or already split. When
// sectionTitle is empty, help flags are appended to the last section
// containing flag content and fall back to "Options" if no flag sections
// exist. Empty FlagGroups and sections are cleaned up.
func MoveHelpFlagsToSection(sections []Section, sectionTitle string) []Section {
	remainingSections := make([]Section, 0, len(sections))
	var movedHelpFlags FlagGroup

	for _, section := range sections {
		sectionContent := section.Content
		filteredContent := make([]Content, 0, len(sectionContent))
		for _, content := range sectionContent {
			flagGroup, ok := content.(FlagGroup)
			if !ok {
				filteredContent = append(filteredContent, content)
				continue
			}

			remainingFlags := make(FlagGroup, 0, len(flagGroup))
			for _, flag := range flagGroup {
				if isHelpFlag(flag) {
					movedHelpFlags = append(movedHelpFlags, flag)
					continue
				}
				remainingFlags = append(remainingFlags, flag)
			}
			if len(remainingFlags) > 0 {
				filteredContent = append(filteredContent, remainingFlags)
			}
		}

		if len(filteredContent) == 0 {
			continue
		}
		section.Content = filteredContent
		remainingSections = append(remainingSections, section)
	}

	if len(movedHelpFlags) == 0 {
		return cleanEmpty(remainingSections)
	}

	return appendFlagGroupToSection(cleanEmpty(remainingSections), sectionTitle, movedHelpFlags)
}

func appendFlagGroupToSection(
	sections []Section,
	sectionTitle string,
	flagGroup FlagGroup,
) []Section {
	if sectionTitle != "" {
		for i := range sections {
			section := &sections[i]
			if section.Title != sectionTitle {
				continue
			}
			section.Content = append(section.Content, flagGroup)
			return sections
		}
		return append(sections, Section{
			Title:   sectionTitle,
			Content: []Content{flagGroup},
		})
	}

	appended := false
	for i := len(sections) - 1; i >= 0; i-- {
		section := &sections[i]
		if hasFlagContent(section.Content) {
			section.Content = append(section.Content, flagGroup)
			appended = true
			break
		}
	}
	if !appended {
		sections = append(sections, Section{
			Title:   "Options",
			Content: []Content{flagGroup},
		})
	}

	return sections
}

func newHelpFlagGroup(shortDesc, longDesc string) FlagGroup {
	return FlagGroup{
		{Short: "h", Desc: shortDesc},
		{Long: "help", Desc: longDesc},
	}
}

func isHelpFlag(flag Flag) bool {
	return flag.Long == "help" || (flag.Long == "" && flag.Short == "h")
}

// patchFlag walks sections looking for a flag by Long name and applies fn.
func patchFlag(sections []Section, flagLong string, fn func(*Flag)) {
	for i := range sections {
		for j := range sections[i].Content {
			fg, ok := sections[i].Content[j].(FlagGroup)
			if !ok {
				continue
			}
			for k := range fg {
				if fg[k].Long == flagLong {
					fn(&fg[k])
					return
				}
			}
		}
	}
}

// removeFlagLong removes any Flag with Long==name from FlagGroups in content.
func removeFlagLong(content []Content, name string) []Content {
	out := make([]Content, 0, len(content))
	for _, c := range content {
		fg, ok := c.(FlagGroup)
		if !ok {
			out = append(out, c)
			continue
		}
		var filtered FlagGroup
		for _, f := range fg {
			if f.Long != name {
				filtered = append(filtered, f)
			}
		}
		out = append(out, filtered)
	}
	return out
}

func removeHelpFlags(sections []Section) []Section {
	for i := range sections {
		section := &sections[i]
		section.Content = removeFlagLong(section.Content, "help")
	}
	return cleanEmpty(sections)
}

// cleanEmpty removes empty FlagGroups from sections and drops sections
// that have no content left.
func cleanEmpty(sections []Section) []Section {
	var out []Section
	for _, s := range sections {
		var content []Content
		for _, c := range s.Content {
			if fg, ok := c.(FlagGroup); ok && len(fg) == 0 {
				continue
			}
			content = append(content, c)
		}
		if len(content) > 0 {
			s.Content = content
			out = append(out, s)
		}
	}
	return out
}

// hasFlagContent reports whether any content item is a FlagGroup.
func hasFlagContent(content []Content) bool {
	for _, c := range content {
		if _, ok := c.(FlagGroup); ok {
			return true
		}
	}
	return false
}

// ClassifiedFlag pairs a help.Flag with its group name and locality.
type ClassifiedFlag struct {
	Flag      Flag
	Group     string // group name ("" = ungrouped); may contain "/" for sub-groups
	Inherited bool   // true = inherited/ancestor flag, false = local
}

// FlagSectionsOption configures BuildFlagSections behavior.
type FlagSectionsOption func(*flagSectionsConfig)

type flagSectionsConfig struct {
	keepGroupOrder bool
}

// KeepGroupOrder preserves first-seen order of groups instead of sorting
// them alphabetically. Use this when the caller controls insertion order
// and wants it preserved in the output.
func KeepGroupOrder() FlagSectionsOption {
	return func(c *flagSectionsConfig) { c.keepGroupOrder = true }
}

// BuildFlagSections assembles flag help sections from pre-classified flags.
//
// When no flag carries a group name, two flat sections are produced:
// "Options" (local) and "Inherited Options" (inherited), each omitted if empty.
//
// When any flag has a group, flags are organized into one section per group
// (sorted alphabetically by default), with ungrouped local flags under "Options"
// and ungrouped inherited flags under "Inherited Options".
// Pass KeepGroupOrder() to preserve first-seen order instead of sorting.
//
// Compound group names ("Section/SubGroup") split flags within the same
// section into separate FlagGroup content entries (rendered with a blank-line
// separator). Sub-groups appear in first-seen order within each section.
func BuildFlagSections(flags []ClassifiedFlag, opts ...FlagSectionsOption) []Section {
	if len(flags) == 0 {
		return nil
	}

	var cfg flagSectionsConfig
	for _, o := range opts {
		o(&cfg)
	}

	hasAnyGroup := false
	for i := range flags {
		if flags[i].Group != "" {
			hasAnyGroup = true
			break
		}
	}

	if !hasAnyGroup {
		return buildFlatFlagSections(flags)
	}

	return buildGroupedFlagSections(flags, cfg.keepGroupOrder)
}

// buildFlatFlagSections builds simple "Options" / "Inherited Options" sections
// when no flag carries a group name.
func buildFlatFlagSections(flags []ClassifiedFlag) []Section {
	var local, inherited FlagGroup
	for i := range flags {
		if flags[i].Inherited {
			inherited = append(inherited, flags[i].Flag)
		} else {
			local = append(local, flags[i].Flag)
		}
	}
	var sections []Section
	if len(local) > 0 {
		sections = append(sections, Section{
			Title:   "Options",
			Content: []Content{local},
		})
	}
	if len(inherited) > 0 {
		sections = append(sections, Section{
			Title:   "Inherited Options",
			Content: []Content{inherited},
		})
	}
	return sections
}

// buildGroupedFlagSections builds sections when at least one flag has a group.
// Groups are sorted alphabetically unless keepOrder is true (first-seen order).
// Compound names ("Section/SubGroup") split into sub-groups within the same section.
func buildGroupedFlagSections(flags []ClassifiedFlag, keepOrder bool) []Section {
	type subGroup struct {
		key   string
		flags FlagGroup
	}
	sectionGroups := make(map[string][]subGroup)
	var sectionOrder []string
	var ungroupedLocal, ungroupedInherited FlagGroup

	for i := range flags {
		f := &flags[i]
		switch {
		case f.Group != "":
			section, subKey, _ := strings.Cut(f.Group, "/")
			if _, exists := sectionGroups[section]; !exists {
				sectionOrder = append(sectionOrder, section)
			}
			subs := sectionGroups[section]
			found := false
			for j := range subs {
				if subs[j].key == subKey {
					subs[j].flags = append(subs[j].flags, f.Flag)
					found = true
					sectionGroups[section] = subs
					break
				}
			}
			if !found {
				sectionGroups[section] = append(
					subs,
					subGroup{key: subKey, flags: FlagGroup{f.Flag}},
				)
			}
		case f.Inherited:
			ungroupedInherited = append(ungroupedInherited, f.Flag)
		default:
			ungroupedLocal = append(ungroupedLocal, f.Flag)
		}
	}

	if !keepOrder {
		slices.Sort(sectionOrder)
	}

	var sections []Section
	for _, section := range sectionOrder {
		var content []Content
		for _, sg := range sectionGroups[section] {
			content = append(content, sg.flags)
		}
		sections = append(sections, Section{
			Title:   section,
			Content: content,
		})
	}
	if len(ungroupedLocal) > 0 {
		sections = append(sections, Section{
			Title:   "Options",
			Content: []Content{ungroupedLocal},
		})
	}
	if len(ungroupedInherited) > 0 {
		sections = append(sections, Section{
			Title:   "Inherited Options",
			Content: []Content{ungroupedInherited},
		})
	}
	return sections
}
