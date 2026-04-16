package theme

import (
	"image/color"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/gechr/x/human"
)

// palette holds the core colors that define a theme preset.
// Each preset maps these onto the full [Theme] struct.
type palette struct {
	heading     color.Color // section headings
	command     color.Color // command name in usage line
	subcommand  color.Color // subcommands
	backtick    color.Color // backticks, markdown code
	flag        color.Color // flags
	arg         color.Color // args
	argOptional color.Color // optional args
	comment     color.Color // dim text, placeholders, defaults
}

// fromPalette builds a full [Theme] from a [palette].
func fromPalette(name string, p palette) *Theme {
	return &Theme{
		name:      name,
		Bold:      new(lipgloss.NewStyle().Bold(true)),
		Dim:       new(lipgloss.NewStyle().Faint(true)),
		Red:       new(lipgloss.NewStyle().Foreground(p.flag)),
		Green:     new(lipgloss.NewStyle().Foreground(p.arg)),
		Yellow:    new(lipgloss.NewStyle().Foreground(p.heading)),
		Blue:      new(lipgloss.NewStyle().Foreground(p.command)),
		Magenta:   new(lipgloss.NewStyle().Foreground(p.backtick)),
		Orange:    new(lipgloss.NewStyle().Foreground(p.backtick)),
		BoldGreen: new(lipgloss.NewStyle().Bold(true).Foreground(p.arg)),

		HelpSection:     new(lipgloss.NewStyle().Bold(true).Foreground(p.heading)),
		HelpCommand:     new(lipgloss.NewStyle().Bold(true).Foreground(p.command)),
		HelpSubcommand:  new(lipgloss.NewStyle().Bold(true).Foreground(p.subcommand)),
		HelpFlag:        new(lipgloss.NewStyle().Foreground(p.flag)),
		HelpArg:         new(lipgloss.NewStyle().Foreground(p.arg)),
		HelpArgOptional: new(lipgloss.NewStyle().Foreground(p.argOptional)),
		HelpValuePlaceholder: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.flag),
		),
		HelpDim: new(lipgloss.NewStyle().Faint(true)),
		HelpBoldDim: new(lipgloss.NewStyle().
			Bold(true).
			Faint(true).
			Foreground(p.command)),
		HelpEnumDefault: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.command),
		),
		HelpUsageExample: HelpUsageExampleStyle{
			Prompt:      "$",
			PromptStyle: lipgloss.NewStyle().Foreground(p.arg),
		},
		HelpFlagExample:       new(lipgloss.NewStyle().Foreground(p.arg)),
		HelpFlagNote:          new(lipgloss.NewStyle().Faint(true)),
		HelpFlagDefault:       new(lipgloss.NewStyle().Faint(true)),
		HelpDescBacktick:      new(lipgloss.NewStyle().Foreground(p.backtick)),
		HelpKeyValueSeparator: ' ',
		HelpKeyValueSeparatorStyle: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.flag),
		),
		HelpRepeatEllipsis: new(
			lipgloss.NewStyle().Faint(true).Foreground(p.flag),
		),
		HelpRepeatEllipsisEnabled: true,
		EnumStyle:                 EnumStyleHighlightDefault,

		MarkdownCode: new(lipgloss.NewStyle().Foreground(p.backtick)),
		MarkdownText: new(lipgloss.NewStyle().Foreground(p.comment)),

		TimeAgoThresholds: []TimeAgoThreshold{
			{MaxAge: time.Minute, Style: lipgloss.NewStyle().Bold(true).Foreground(p.arg)},
			{MaxAge: time.Hour, Style: lipgloss.NewStyle().Bold(true).Foreground(p.arg)},
			{MaxAge: human.HoursPerDay * time.Hour, Style: lipgloss.NewStyle().Foreground(p.arg)},
			{
				MaxAge: 14 * human.HoursPerDay * time.Hour,
				Style:  lipgloss.NewStyle().Foreground(p.heading),
			},
			{
				MaxAge: 30 * human.HoursPerDay * time.Hour,
				Style:  lipgloss.NewStyle().Foreground(p.backtick),
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

// Plain returns a theme with no styling at all.
func Plain() *Theme {
	plain := lipgloss.NewStyle()

	return &Theme{
		name:      "plain",
		Bold:      new(plain),
		Dim:       new(plain),
		Red:       new(plain),
		Green:     new(plain),
		Yellow:    new(plain),
		Blue:      new(plain),
		Magenta:   new(plain),
		Orange:    new(plain),
		BoldGreen: new(plain),

		HelpSection:                new(plain),
		HelpCommand:                new(plain),
		HelpSubcommand:             new(plain),
		HelpFlag:                   new(plain),
		HelpArg:                    new(plain),
		HelpArgOptional:            new(plain),
		HelpValuePlaceholder:       new(plain),
		HelpDim:                    new(plain),
		HelpBoldDim:                new(plain),
		HelpEnumDefault:            new(plain),
		HelpUsageExample:           HelpUsageExampleStyle{Prompt: "$"},
		HelpFlagExample:            new(plain),
		HelpFlagNote:               new(plain),
		HelpFlagDefault:            new(plain),
		HelpKeyValueSeparator:      ' ',
		HelpKeyValueSeparatorStyle: new(plain),
		HelpRepeatEllipsis:         new(plain),

		MarkdownCode: new(plain),
		MarkdownText: new(plain),

		TimeAgoThresholds: []TimeAgoThreshold{
			{MaxAge: time.Minute, Style: plain},
			{MaxAge: time.Hour, Style: plain},
			{MaxAge: human.HoursPerDay * time.Hour, Style: plain},
			{MaxAge: 14 * human.HoursPerDay * time.Hour, Style: plain},
			{MaxAge: 30 * human.HoursPerDay * time.Hour, Style: plain},
		},

		EntityColors: defaultEntityColors(),
	}
}

// Monochrome returns a theme with no colors - only bold and dim.
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
		HelpArgOptional:            new(plain),
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
		heading:     lipgloss.Color("#ffd866"), // yellow
		command:     lipgloss.Color("#a9dc76"), // green
		subcommand:  lipgloss.Color("#78dce8"), // cyan
		backtick:    lipgloss.Color("#fc9867"), // orange
		flag:        lipgloss.Color("#ff6188"), // pink
		arg:         lipgloss.Color("#78dce8"), // cyan
		argOptional: lipgloss.Color("#ab9df2"), // purple
		comment:     lipgloss.Color("#939293"), // comment
	})
}

// CatppuccinLatte returns a theme based on the Catppuccin Latte (light) palette.
func CatppuccinLatte() *Theme {
	return fromPalette("catppuccin-latte", palette{
		heading:     lipgloss.Color("#df8e1d"), // yellow
		command:     lipgloss.Color("#40a02b"), // green
		subcommand:  lipgloss.Color("#179299"), // teal
		backtick:    lipgloss.Color("#dc8a78"), // rosewater
		flag:        lipgloss.Color("#d20f39"), // red
		arg:         lipgloss.Color("#8839ef"), // mauve
		argOptional: lipgloss.Color("#1e66f5"), // blue
		comment:     lipgloss.Color("#7c7f93"), // overlay2
	})
}

// CatppuccinFrappe returns a theme based on the Catppuccin Frappé (dark) palette.
func CatppuccinFrappe() *Theme {
	return fromPalette("catppuccin-frappe", palette{
		heading:     lipgloss.Color("#e5c890"), // yellow
		command:     lipgloss.Color("#a6d189"), // green
		subcommand:  lipgloss.Color("#81c8be"), // teal
		backtick:    lipgloss.Color("#f2d5cf"), // rosewater
		flag:        lipgloss.Color("#e78284"), // red
		arg:         lipgloss.Color("#ca9ee6"), // mauve
		argOptional: lipgloss.Color("#8caaee"), // blue
		comment:     lipgloss.Color("#949cbb"), // overlay2
	})
}

// CatppuccinMacchiato returns a theme based on the Catppuccin Macchiato (dark) palette.
func CatppuccinMacchiato() *Theme {
	return fromPalette("catppuccin-macchiato", palette{
		heading:     lipgloss.Color("#eed49f"), // yellow
		command:     lipgloss.Color("#a6da95"), // green
		subcommand:  lipgloss.Color("#8bd5ca"), // teal
		backtick:    lipgloss.Color("#f4dbd6"), // rosewater
		flag:        lipgloss.Color("#ed8796"), // red
		arg:         lipgloss.Color("#c6a0f6"), // mauve
		argOptional: lipgloss.Color("#8aadf4"), // blue
		comment:     lipgloss.Color("#939ab7"), // overlay2
	})
}

// CatppuccinMocha returns a theme based on the Catppuccin Mocha (dark) palette.
func CatppuccinMocha() *Theme {
	return fromPalette("catppuccin-mocha", palette{
		heading:     lipgloss.Color("#f9e2af"), // yellow
		command:     lipgloss.Color("#a6e3a1"), // green
		subcommand:  lipgloss.Color("#94e2d5"), // teal
		backtick:    lipgloss.Color("#f5e0dc"), // rosewater
		flag:        lipgloss.Color("#f38ba8"), // red
		arg:         lipgloss.Color("#cba6f7"), // mauve
		argOptional: lipgloss.Color("#89b4fa"), // blue
		comment:     lipgloss.Color("#9399b2"), // overlay2
	})
}

// Dracula returns a theme based on the Dracula color scheme.
func Dracula() *Theme {
	return fromPalette("dracula", palette{
		heading:     lipgloss.Color("#f1fa8c"), // yellow
		command:     lipgloss.Color("#50fa7b"), // green
		subcommand:  lipgloss.Color("#8be9fd"), // cyan
		backtick:    lipgloss.Color("#ffb86c"), // orange
		flag:        lipgloss.Color("#ff5555"), // red
		arg:         lipgloss.Color("#ff79c6"), // pink
		argOptional: lipgloss.Color("#bd93f9"), // purple
		comment:     lipgloss.Color("#6272a4"), // comment
	})
}

// Nord returns a theme based on the Nord Arctic color scheme.
func Nord() *Theme {
	return fromPalette("nord", palette{
		heading:     lipgloss.Color("#ebcb8b"), // nord13, yellow
		command:     lipgloss.Color("#a3be8c"), // nord14, green
		subcommand:  lipgloss.Color("#88c0d0"), // nord8, cyan
		backtick:    lipgloss.Color("#d08770"), // nord12, orange
		flag:        lipgloss.Color("#bf616a"), // nord11, red
		arg:         lipgloss.Color("#81a1c1"), // nord9, blue
		argOptional: lipgloss.Color("#b48ead"), // nord15, purple
		comment:     lipgloss.Color("#4c566a"), // nord3, dark gray
	})
}

// Solarized returns a theme based on the Solarized color scheme.
func Solarized() *Theme {
	return fromPalette("solarized", palette{
		heading:     lipgloss.Color("#b58900"), // yellow
		command:     lipgloss.Color("#859900"), // green
		subcommand:  lipgloss.Color("#2aa198"), // cyan
		backtick:    lipgloss.Color("#cb4b16"), // orange
		flag:        lipgloss.Color("#dc322f"), // red
		arg:         lipgloss.Color("#268bd2"), // blue
		argOptional: lipgloss.Color("#6c71c4"), // violet
		comment:     lipgloss.Color("#586e75"), // base01
	})
}

// GruvboxDark returns a theme based on the Gruvbox Dark color scheme.
func GruvboxDark() *Theme {
	return fromPalette("gruvbox-dark", palette{
		heading:     lipgloss.Color("#fabd2f"), // yellow
		command:     lipgloss.Color("#b8bb26"), // green
		subcommand:  lipgloss.Color("#8ec07c"), // aqua
		backtick:    lipgloss.Color("#fe8019"), // orange
		flag:        lipgloss.Color("#fb4934"), // red
		arg:         lipgloss.Color("#83a598"), // blue
		argOptional: lipgloss.Color("#d3869b"), // purple
		comment:     lipgloss.Color("#928374"), // gray
	})
}

// GruvboxLight returns a theme based on the Gruvbox Light color scheme.
func GruvboxLight() *Theme {
	return fromPalette("gruvbox-light", palette{
		heading:     lipgloss.Color("#b57614"), // yellow
		command:     lipgloss.Color("#79740e"), // green
		subcommand:  lipgloss.Color("#427b58"), // aqua
		backtick:    lipgloss.Color("#af3a03"), // orange
		flag:        lipgloss.Color("#9d0006"), // red
		arg:         lipgloss.Color("#076678"), // blue
		argOptional: lipgloss.Color("#8f3f71"), // purple
		comment:     lipgloss.Color("#928374"), // gray
	})
}

// TokyoNight returns a theme based on the Tokyo Night color scheme.
func TokyoNight() *Theme {
	return fromPalette("tokyo-night", palette{
		heading:     lipgloss.Color("#e0af68"), // yellow
		command:     lipgloss.Color("#9ece6a"), // green
		subcommand:  lipgloss.Color("#7dcfff"), // cyan
		backtick:    lipgloss.Color("#ff9e64"), // orange
		flag:        lipgloss.Color("#f7768e"), // red
		arg:         lipgloss.Color("#7aa2f7"), // blue
		argOptional: lipgloss.Color("#bb9af7"), // magenta
		comment:     lipgloss.Color("#565f89"), // comment
	})
}

// Synthwave returns a theme based on the Synthwave '84 color scheme.
func Synthwave() *Theme {
	return fromPalette("synthwave", palette{
		heading:     lipgloss.Color("#fede5d"), // yellow
		command:     lipgloss.Color("#72f1b8"), // green
		subcommand:  lipgloss.Color("#36f9f6"), // cyan
		backtick:    lipgloss.Color("#ff8b39"), // orange
		flag:        lipgloss.Color("#fe4450"), // red
		arg:         lipgloss.Color("#03edf9"), // blue
		argOptional: lipgloss.Color("#bb9af7"), // purple
		comment:     lipgloss.Color("#848bbd"), // comment
	})
}

// OneDark returns a theme based on the Atom One Dark color scheme.
func OneDark() *Theme {
	return fromPalette("one-dark", palette{
		heading:     lipgloss.Color("#e5c07b"), // chalky/yellow
		command:     lipgloss.Color("#98c379"), // sage/green
		subcommand:  lipgloss.Color("#56b6c2"), // cyan
		backtick:    lipgloss.Color("#d19a66"), // whiskey/orange
		flag:        lipgloss.Color("#e06c75"), // coral/red
		arg:         lipgloss.Color("#61afef"), // malibu/blue
		argOptional: lipgloss.Color("#c678dd"), // violet/purple
		comment:     lipgloss.Color("#5c6370"), // stone/gray
	})
}
