package theme

import (
	"image/color"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/gechr/clib/human"
)

// palette holds the core colors that define a theme preset.
// Each preset maps these onto the full [Theme] struct.
type palette struct {
	accent    color.Color // section headers, accents
	key       color.Color // commands
	secondary color.Color // subcommands, backticks
	flag      color.Color // flags, repeat ellipsis
	arg       color.Color // args, examples, enum defaults
	comment   color.Color // dim text, placeholders, notes, defaults
}

// fromPalette builds a full [Theme] from a [palette].
func fromPalette(name string, p palette) *Theme {
	return &Theme{
		name:      name,
		Bold:      new(lipgloss.NewStyle().Bold(true)),
		Dim:       new(lipgloss.NewStyle().Faint(true)),
		Red:       new(lipgloss.NewStyle().Foreground(p.flag)),
		Green:     new(lipgloss.NewStyle().Foreground(p.arg)),
		Yellow:    new(lipgloss.NewStyle().Foreground(p.accent)),
		Blue:      new(lipgloss.NewStyle().Foreground(p.key)),
		Magenta:   new(lipgloss.NewStyle().Foreground(p.secondary)),
		Orange:    new(lipgloss.NewStyle().Foreground(p.secondary)),
		BoldGreen: new(lipgloss.NewStyle().Bold(true).Foreground(p.arg)),

		HelpSection:    new(lipgloss.NewStyle().Bold(true).Foreground(p.accent)),
		HelpCommand:    new(lipgloss.NewStyle().Bold(true).Foreground(p.key)),
		HelpSubcommand: new(lipgloss.NewStyle().Bold(true).Foreground(p.secondary)),
		HelpFlag:       new(lipgloss.NewStyle().Foreground(p.flag)),
		HelpArg:        new(lipgloss.NewStyle().Foreground(p.arg)),
		HelpValuePlaceholder: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.flag),
		),
		HelpDim: new(lipgloss.NewStyle().Faint(true)),
		HelpBoldDim: new(lipgloss.NewStyle().
			Bold(true).
			Faint(true).
			Foreground(p.key)),
		HelpEnumDefault: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.arg),
		),
		HelpUsageExample: HelpUsageExampleStyle{
			Prompt:      "$",
			PromptStyle: lipgloss.NewStyle().Foreground(p.arg),
		},
		HelpFlagExample:       new(lipgloss.NewStyle().Foreground(p.arg)),
		HelpFlagNote:          new(lipgloss.NewStyle().Faint(true)),
		HelpFlagDefault:       new(lipgloss.NewStyle().Faint(true)),
		HelpDescBacktick:      new(lipgloss.NewStyle().Foreground(p.secondary)),
		HelpKeyValueSeparator: ' ',
		HelpKeyValueSeparatorStyle: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.flag),
		),
		HelpRepeatEllipsis: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.flag),
		),
		HelpRepeatEllipsisEnabled: true,
		EnumStyle:                 EnumStyleHighlightDefault,

		MarkdownCode: new(lipgloss.NewStyle().Foreground(p.secondary)),
		MarkdownText: new(lipgloss.NewStyle().Foreground(p.comment)),

		TimeAgoThresholds: []TimeAgoThreshold{
			{MaxAge: time.Minute, Style: lipgloss.NewStyle().Bold(true).Foreground(p.arg)},
			{MaxAge: time.Hour, Style: lipgloss.NewStyle().Bold(true).Foreground(p.arg)},
			{MaxAge: human.HoursPerDay * time.Hour, Style: lipgloss.NewStyle().Foreground(p.arg)},
			{
				MaxAge: 14 * human.HoursPerDay * time.Hour,
				Style:  lipgloss.NewStyle().Foreground(p.accent),
			},
			{
				MaxAge: 30 * human.HoursPerDay * time.Hour,
				Style:  lipgloss.NewStyle().Foreground(p.secondary),
			},
		},

		EntityColors: defaultEntityColors(),
	}
}

func defaultEntityColors() []color.Color {
	return []color.Color{
		lipgloss.Color(
			"208",
		),
		lipgloss.Color("51"),
		lipgloss.Color("226"),
		lipgloss.Color("207"),
		lipgloss.Color("82"),
		lipgloss.Color(
			"75",
		),
		lipgloss.Color("214"),
		lipgloss.Color("177"),
		lipgloss.Color("48"),
		lipgloss.Color("87"),
		lipgloss.Color(
			"220",
		),
		lipgloss.Color("141"),
		lipgloss.Color("118"),
		lipgloss.Color("50"),
		lipgloss.Color("213"),
		lipgloss.Color(
			"111",
		),
		lipgloss.Color("156"),
		lipgloss.Color("183"),
		lipgloss.Color("229"),
		lipgloss.Color("123"),
	}
}

// Monochrome returns a theme with no colors — only bold and dim.
func Monochrome() *Theme {
	bold := lipgloss.NewStyle().Bold(true)
	dim := lipgloss.NewStyle().Faint(true)
	boldDim := lipgloss.NewStyle().Bold(true).Faint(true)
	plain := lipgloss.NewStyle()

	return &Theme{
		name:      "monochrome",
		Bold:      new(bold),
		Dim:       new(dim),
		Red:       new(plain),
		Green:     new(plain),
		Yellow:    new(plain),
		Blue:      new(plain),
		Magenta:   new(plain),
		Orange:    new(plain),
		BoldGreen: new(bold),

		HelpSection:                new(bold),
		HelpCommand:                new(bold),
		HelpSubcommand:             new(bold),
		HelpFlag:                   new(plain),
		HelpArg:                    new(plain),
		HelpValuePlaceholder:       new(dim),
		HelpDim:                    new(dim),
		HelpBoldDim:                new(boldDim),
		HelpEnumDefault:            new(dim),
		HelpUsageExample:           HelpUsageExampleStyle{Prompt: "$"},
		HelpFlagExample:            new(dim),
		HelpFlagNote:               new(dim),
		HelpFlagDefault:            new(dim),
		HelpDescBacktick:           new(bold),
		HelpKeyValueSeparator:      ' ',
		HelpKeyValueSeparatorStyle: new(dim),
		HelpRepeatEllipsis:         new(dim),
		HelpRepeatEllipsisEnabled:  true,
		EnumStyle:                  EnumStyleHighlightDefault,

		MarkdownCode: new(bold),
		MarkdownText: new(plain),

		TimeAgoThresholds: []TimeAgoThreshold{
			{MaxAge: time.Minute, Style: bold},
			{MaxAge: time.Hour, Style: bold},
			{MaxAge: human.HoursPerDay * time.Hour, Style: lipgloss.NewStyle()},
			{MaxAge: 14 * human.HoursPerDay * time.Hour, Style: dim},
			{MaxAge: 30 * human.HoursPerDay * time.Hour, Style: dim},
		},

		EntityColors: defaultEntityColors(),
	}
}

// Monokai returns a theme inspired by the Monokai color scheme.
func Monokai() *Theme {
	return fromPalette("monokai", palette{
		accent:    lipgloss.Color("#66d9ef"), // cyan
		key:       lipgloss.Color("#ae81ff"), // purple
		secondary: lipgloss.Color("#fd971f"), // orange
		flag:      lipgloss.Color("#f92672"), // pink
		arg:       lipgloss.Color("#a6e22e"), // green
		comment:   lipgloss.Color("#88846f"), // comment
	})
}

// CatppuccinLatte returns a theme based on the Catppuccin Latte (light) palette.
func CatppuccinLatte() *Theme {
	return fromPalette("catppuccin-latte", palette{
		accent:    lipgloss.Color("#179299"), // teal
		key:       lipgloss.Color("#1e66f5"), // blue
		secondary: lipgloss.Color("#dc8a78"), // rosewater
		flag:      lipgloss.Color("#d20f39"), // red
		arg:       lipgloss.Color("#40a02b"), // green
		comment:   lipgloss.Color("#7c7f93"), // overlay2
	})
}

// CatppuccinFrappe returns a theme based on the Catppuccin Frappé (dark) palette.
func CatppuccinFrappe() *Theme {
	return fromPalette("catppuccin-frappe", palette{
		accent:    lipgloss.Color("#81c8be"), // teal
		key:       lipgloss.Color("#8caaee"), // blue
		secondary: lipgloss.Color("#f2d5cf"), // rosewater
		flag:      lipgloss.Color("#e78284"), // red
		arg:       lipgloss.Color("#a6d189"), // green
		comment:   lipgloss.Color("#949cbb"), // overlay2
	})
}

// CatppuccinMacchiato returns a theme based on the Catppuccin Macchiato (dark) palette.
func CatppuccinMacchiato() *Theme {
	return fromPalette("catppuccin-macchiato", palette{
		accent:    lipgloss.Color("#8bd5ca"), // teal
		key:       lipgloss.Color("#8aadf4"), // blue
		secondary: lipgloss.Color("#f4dbd6"), // rosewater
		flag:      lipgloss.Color("#ed8796"), // red
		arg:       lipgloss.Color("#a6da95"), // green
		comment:   lipgloss.Color("#939ab7"), // overlay2
	})
}

// CatppuccinMocha returns a theme based on the Catppuccin Mocha (dark) palette.
func CatppuccinMocha() *Theme {
	return fromPalette("catppuccin-mocha", palette{
		accent:    lipgloss.Color("#94e2d5"), // teal
		key:       lipgloss.Color("#89b4fa"), // blue
		secondary: lipgloss.Color("#f5e0dc"), // rosewater
		flag:      lipgloss.Color("#f38ba8"), // red
		arg:       lipgloss.Color("#a6e3a1"), // green
		comment:   lipgloss.Color("#9399b2"), // overlay2
	})
}

// Dracula returns a theme based on the Dracula color scheme.
func Dracula() *Theme {
	return fromPalette("dracula", palette{
		accent:    lipgloss.Color("#8be9fd"), // cyan
		key:       lipgloss.Color("#bd93f9"), // purple
		secondary: lipgloss.Color("#ffb86c"), // orange
		flag:      lipgloss.Color("#ff5555"), // red
		arg:       lipgloss.Color("#50fa7b"), // green
		comment:   lipgloss.Color("#6272a4"), // comment
	})
}
