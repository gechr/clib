package help

import (
	"slices"
	"strings"
)

// ClassifiedCommand pairs a command with its native section and clib group.
type ClassifiedCommand struct {
	Command     Command
	NativeGroup string // native section title, or empty when ungrouped
	Group       string // clib section override or "/SubKey" within the native section
}

// BuildCommandSections builds sections in native/first-seen order.
// Group overrides NativeGroup; "/SubKey" inherits it. Numeric blocks sort ascending,
// non-negative above plain entries and negative below. Ungrouped commands appear last.
func BuildCommandSections(commands []ClassifiedCommand, opts ...CommandSectionsOption) []Section {
	if len(commands) == 0 {
		return nil
	}
	var cfg commandSectionsConfig
	for _, opt := range opts {
		opt(&cfg)
	}
	ungroupedTitle := "Commands"
	for _, command := range commands {
		if section, _ := commandGroup(command); section != "" {
			ungroupedTitle = "Additional Commands"
			break
		}
	}
	if len(cfg.groupOrder) > 0 {
		ungroupedTitle = "Additional Commands"
	}
	groups := make(map[string][]subGroup[Command])
	plain := make(map[string]CommandGroup)
	var order []string
	hasUngrouped := false
	for _, command := range commands {
		section, key := commandGroup(command)
		if section == "" {
			section = ungroupedTitle
			hasUngrouped = true
		}
		if _, seen := groups[section]; !seen {
			order = append(order, section)
			groups[section] = nil
		}
		if key == "" {
			plain[section] = append(plain[section], command.Command)
		} else {
			groups[section] = addSubGroup(groups[section], key, command.Command)
		}
	}
	sectionOrder := make([]string, 0, len(order))
	for _, title := range append(slices.Clone(cfg.groupOrder), order...) {
		if hasUngrouped && title == ungroupedTitle {
			continue
		}
		if _, exists := groups[title]; exists && !slices.Contains(sectionOrder, title) {
			sectionOrder = append(sectionOrder, title)
		}
	}
	if cfg.sorted {
		slices.Sort(sectionOrder)
	}
	if hasUngrouped {
		sectionOrder = append(sectionOrder, ungroupedTitle)
	}
	sections := make([]Section, 0, len(sectionOrder))
	for _, title := range sectionOrder {
		top, bottom := splitSubGroups(
			groups[title],
			func(entries []Command) Content { return CommandGroup(entries) },
		)
		content := top
		if len(plain[title]) > 0 {
			content = append(content, plain[title])
		}
		content = append(content, bottom...)
		sections = append(sections, Section{Title: title, Content: content})
	}
	return sections
}

func commandGroup(command ClassifiedCommand) (string, string) {
	section := command.NativeGroup
	if command.Group == "" {
		return section, ""
	}
	title, key, _ := strings.Cut(command.Group, "/")
	if title != "" {
		section = title
	}
	return section, key
}
