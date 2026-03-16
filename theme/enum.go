package theme

import "strings"

// EnumStyle controls how enum values are rendered in help output.
type EnumStyle int

const (
	EnumStylePlain            EnumStyle = iota // [open, closed, merged, all] - all dim
	EnumStyleHighlightDefault                  // [auto, always, never] - default value highlighted
	EnumStyleHighlightPrefix                   // [open, closed, merged, all] - bold hints highlighted
	EnumStyleHighlightBoth                     // bold hints + default value highlighted
)

// EnumValue represents a single enum option for help display.
// Bold is the substring to highlight within Name (first occurrence).
// If Bold is empty, the entire Name renders as dim with no bold prefix.
// IsDefault marks this value as the default (rendered with HelpEnumDefault style).
type EnumValue struct {
	Name      string
	Bold      string
	IsDefault bool
}

// FmtEnum formats an enum list with bold shortcut substrings inside dim brackets.
// e.g. FmtEnum([]EnumValue{{Name: "open", Bold: "o"}, {Name: "closed", Bold: "c"}})
// renders as dim("[") + boldDim("o") + dim("pen, ") + boldDim("c") + dim("losed]").
func (th *Theme) FmtEnum(values []EnumValue) string {
	th = th.Init()

	var items []string
	for _, v := range values {
		items = append(items, th.fmtEnumValue(v))
	}

	return th.HelpDim.Render("[") +
		strings.Join(items, th.HelpDim.Render(", ")) +
		th.HelpDim.Render("]")
}

// fmtEnumValue renders a single enum value with its bold substring highlighted.
func (th *Theme) fmtEnumValue(v EnumValue) string {
	// Default with no bold prefix: render entire value in default style.
	if v.IsDefault && v.Bold == "" {
		return th.HelpEnumDefault.Render(v.Name)
	}
	if v.Bold == "" {
		return th.HelpDim.Render(v.Name)
	}
	idx := strings.Index(v.Name, v.Bold)
	if idx < 0 {
		if v.IsDefault {
			return th.HelpEnumDefault.Render(v.Name)
		}
		return th.HelpDim.Render(v.Name)
	}
	// Pick base and bold styles depending on default status.
	base := th.HelpDim
	bold := th.HelpBoldDim
	if v.IsDefault {
		base = th.HelpEnumDefault
		boldStyle := th.HelpEnumDefault.Bold(true)
		bold = &boldStyle
	}
	before := v.Name[:idx]
	after := v.Name[idx+len(v.Bold):]
	var s string
	if before != "" {
		s += base.Render(before)
	}
	s += bold.Render(v.Bold)
	if after != "" {
		s += base.Render(after)
	}
	return s
}

// FmtEnumDefault formats an enum list followed by a dim default annotation.
func (th *Theme) FmtEnumDefault(defaultVal string, values []EnumValue) string {
	th = th.Init()
	return th.FmtEnum(values) + th.HelpDim.Render(" (default: "+defaultVal+")")
}

// DimDefault formats a default value annotation in dim.
// Returns empty string if value is empty.
func (th *Theme) DimDefault(value string) string {
	th = th.Init()
	if value == "" {
		return ""
	}
	return " " + th.HelpDim.Render("(default: "+value+")")
}

// DimNote formats a parenthetical note in dim.
func (th *Theme) DimNote(text string) string {
	th = th.Init()
	return th.HelpDim.Render("(" + text + ")")
}
