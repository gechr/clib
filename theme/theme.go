package theme

import (
	"image/color"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/gechr/x/human"
)

// HelpUsageExampleStyle controls the rendering of examples in the
// "Examples" help section (the "$ command" lines).
type HelpUsageExampleStyle struct {
	Prompt       string         // Prefix string shown before commands (default "$").
	PromptStyle  lipgloss.Style // Style applied to the prompt.
	CommandStyle lipgloss.Style // Style applied to the command text.
}

// TimeAgoThreshold defines a coloring threshold for time-ago rendering.
// Thresholds are evaluated in order; the first one where the duration is
// less than MaxAge wins. Entries beyond the last threshold use the
// theme's Red style as a fallback.
type TimeAgoThreshold struct {
	MaxAge time.Duration
	Style  lipgloss.Style
}

// Theme holds all style definitions for CLI output.
// All lipgloss.Style fields are pointers so that nil means "not configured".
// Use a preset to construct a Theme, or [Init] to fill nil
// fields on an existing theme.
type Theme struct {
	name string

	// Background declares the terminal background this theme is designed for.
	Background Background

	// Base styles.
	Bold *lipgloss.Style
	Dim  *lipgloss.Style

	// Semantic color styles.
	Red       *lipgloss.Style
	Green     *lipgloss.Style
	Yellow    *lipgloss.Style
	Blue      *lipgloss.Style
	Magenta   *lipgloss.Style
	Orange    *lipgloss.Style
	BoldGreen *lipgloss.Style

	// Help styles.
	HelpAlias                  *lipgloss.Style
	HelpArg                    *lipgloss.Style
	HelpArgOptional            *lipgloss.Style
	HelpBoldDim                *lipgloss.Style
	HelpCommand                *lipgloss.Style
	HelpDescBacktick           *lipgloss.Style // Backtick-enclosed text in flag descriptions (nil = leave backticks intact).
	HelpDim                    *lipgloss.Style
	HelpEnumDefault            *lipgloss.Style // Default value in EnumStyleHighlightDefault lists (default: dim green).
	HelpFlag                   *lipgloss.Style
	HelpFlagBacktick           *lipgloss.Style // Override for backtick-enclosed flag-like text in descriptions (nil = fall back to HelpFlag).
	HelpDefaultOpen            string          // Opening bracket for default-value annotations (default "(", e.g. "(default: X)").
	HelpDefaultClose           string          // Closing bracket for default-value annotations (default ")").
	HelpExampleOpen            string          // Opening bracket for example annotations (default "(", e.g. "(example: X)").
	HelpExampleClose           string          // Closing bracket for example annotations (default ")").
	HelpFlagDefault            *lipgloss.Style // [default: ...] annotations in flag descriptions.
	HelpFlagExample            *lipgloss.Style // [example: ...] annotations in flag descriptions.
	HelpFlagNote               *lipgloss.Style // Trailing (...) notes in flag descriptions.
	HelpKeyValueSeparator      rune            // Separator between flag and placeholder (default: ' ').
	HelpKeyValueSeparatorStyle *lipgloss.Style // Style applied to the separator (default: nil = unstyled).
	HelpRepeatEllipsis         *lipgloss.Style // "…" suffix on repeatable flag placeholders (default: dim red).
	HelpRepeatEllipsisEnabled  bool            // Whether to show "…" suffix on repeatable placeholders (default: true).
	HelpSection                *lipgloss.Style
	HelpSubcommand             *lipgloss.Style
	HelpUsageExample           HelpUsageExampleStyle // Examples section prompt and command styling.
	HelpValuePlaceholder       *lipgloss.Style
	EnumStyle                  EnumStyle // How enum values are rendered in help output.

	// Markdown styles.
	MarkdownCode *lipgloss.Style
	MarkdownText *lipgloss.Style

	// Time-ago thresholds (ordered by MaxAge ascending).
	TimeAgoThresholds []TimeAgoThreshold

	// Entity color palette for unique entity colorization.
	EntityColors []color.Color
}

// Dark returns clib's default dark-background theme.
func Dark() *Theme {
	return &Theme{
		name:       themeNameDark,
		Background: BackgroundDark,
		Bold:       new(lipgloss.NewStyle().Bold(true)),
		Dim:        new(lipgloss.NewStyle().Faint(true)),
		Red:        new(lipgloss.NewStyle().Foreground(lipgloss.Color("1"))),
		Green:      new(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))),
		Yellow:     new(lipgloss.NewStyle().Foreground(lipgloss.Color("3"))),
		Blue:       new(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))),
		Magenta:    new(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))),
		Orange:     new(lipgloss.NewStyle().Foreground(lipgloss.Color("208"))),
		BoldGreen:  new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))),

		HelpArg:         new(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))),
		HelpArgOptional: new(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))),
		HelpBoldDim: new(lipgloss.NewStyle().
			Bold(true).
			Faint(true).
			Foreground(lipgloss.Color("4"))),
		HelpCommand:      new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))),
		HelpDescBacktick: new(lipgloss.NewStyle().Foreground(lipgloss.Color("183"))),
		HelpDim:          new(lipgloss.NewStyle().Faint(true)),
		HelpEnumDefault: new(
			lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("2")),
		),
		HelpDefaultOpen:       "(",
		HelpDefaultClose:      ")",
		HelpExampleOpen:       "(",
		HelpExampleClose:      ")",
		HelpFlag:              new(lipgloss.NewStyle().Foreground(lipgloss.Color("1"))),
		HelpFlagDefault:       new(lipgloss.NewStyle().Faint(true)),
		HelpFlagExample:       new(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))),
		HelpFlagNote:          new(lipgloss.NewStyle().Faint(true)),
		HelpKeyValueSeparator: ' ',
		HelpKeyValueSeparatorStyle: new(
			lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("1")),
		),
		HelpRepeatEllipsis: new(
			lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("1")),
		),
		HelpRepeatEllipsisEnabled: true,
		HelpSection: new(
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3")),
		),
		HelpSubcommand: new(
			lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("4")),
		),
		HelpUsageExample: HelpUsageExampleStyle{
			Prompt:      "$",
			PromptStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		},
		HelpValuePlaceholder: new(lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("1"))),
		EnumStyle:            EnumStyleHighlightDefault,

		MarkdownCode: new(lipgloss.NewStyle().Foreground(lipgloss.Color("#98d5d3"))),
		MarkdownText: new(lipgloss.NewStyle().Foreground(lipgloss.Color("#D8DEE9"))),

		TimeAgoThresholds: []TimeAgoThreshold{
			{
				MaxAge: time.Minute,
				Style:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")),
			},
			{
				MaxAge: time.Hour,
				Style:  lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")),
			},
			{
				MaxAge: human.HoursPerDay * time.Hour,
				Style:  lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
			},
			{
				MaxAge: 14 * human.HoursPerDay * time.Hour,
				Style:  lipgloss.NewStyle().Foreground(lipgloss.Color("3")),
			},
			{
				MaxAge: 30 * human.HoursPerDay * time.Hour,
				Style:  lipgloss.NewStyle().Foreground(lipgloss.Color("208")),
			},
		},

		EntityColors: defaultEntityColors(BackgroundDark),
	}
}

// Light returns clib's default light-background theme.
//
// The help styles come from the light palette, but the semantic color slots
// (Red/Green/Yellow/Blue/Magenta/Orange/BoldGreen) and the time-ago gradient
// are set explicitly so they faithfully mirror [Dark]'s meaning with
// light-background contrast. Deriving them from the help palette would scramble
// semantics (e.g. Green would render as the palette's blue "arg" color).
func Light() *Theme {
	t := fromPalette(themeNameLight, BackgroundLight, palette{
		heading:     lipgloss.Color("#8a6500"),
		command:     lipgloss.Color("#256d1b"),
		subcommand:  lipgloss.Color("#006d75"),
		backtick:    lipgloss.Color("#0e7490"),
		flag:        lipgloss.Color("#a11d33"),
		arg:         lipgloss.Color("#2459b3"),
		argOptional: lipgloss.Color("#7047b5"),
		comment:     lipgloss.Color("#5f6368"),
	})

	// Faithful light-contrast semantic colors (mirroring Dark's meaning).
	red := lipgloss.Color("#cf222e")
	green := lipgloss.Color("#1a7f37")
	yellow := lipgloss.Color("#9a6700")
	blue := lipgloss.Color("#0969da")
	magenta := lipgloss.Color("#8250df")
	orange := lipgloss.Color("#bc4c00")

	t.Red = new(lipgloss.NewStyle().Foreground(red))
	t.Green = new(lipgloss.NewStyle().Foreground(green))
	t.Yellow = new(lipgloss.NewStyle().Foreground(yellow))
	t.Blue = new(lipgloss.NewStyle().Foreground(blue))
	t.Magenta = new(lipgloss.NewStyle().Foreground(magenta))
	t.Orange = new(lipgloss.NewStyle().Foreground(orange))
	t.BoldGreen = new(lipgloss.NewStyle().Bold(true).Foreground(green))

	// Time-ago gradient: recent = green, aging through yellow/orange, old = red.
	t.TimeAgoThresholds = []TimeAgoThreshold{
		{MaxAge: time.Minute, Style: lipgloss.NewStyle().Bold(true).Foreground(green)},
		{MaxAge: time.Hour, Style: lipgloss.NewStyle().Bold(true).Foreground(green)},
		{MaxAge: human.HoursPerDay * time.Hour, Style: lipgloss.NewStyle().Foreground(green)},
		{
			MaxAge: 14 * human.HoursPerDay * time.Hour,
			Style:  lipgloss.NewStyle().Foreground(yellow),
		},
		{
			MaxAge: 30 * human.HoursPerDay * time.Hour,
			Style:  lipgloss.NewStyle().Foreground(orange),
		},
	}

	return t
}

// With returns a copy of t with the given options applied.
func (t *Theme) With(opts ...Option) *Theme {
	n := *t
	n.name = ""
	for _, opt := range opts {
		opt(&n)
	}
	return &n
}

// Init returns a copy of t with nil styles replaced by zero-value styles
// and structural defaults filled in so zero-value themes remain usable.
func (t *Theme) Init() *Theme {
	if t == nil {
		t = &Theme{}
	}

	n := *t

	ensureStyle := func(s **lipgloss.Style) {
		if *s == nil {
			*s = new(lipgloss.Style)
		}
	}

	ensureStyle(&n.Bold)
	ensureStyle(&n.Dim)
	ensureStyle(&n.Red)
	ensureStyle(&n.Green)
	ensureStyle(&n.Yellow)
	ensureStyle(&n.Blue)
	ensureStyle(&n.Magenta)
	ensureStyle(&n.Orange)
	ensureStyle(&n.BoldGreen)
	ensureStyle(&n.HelpArg)
	ensureStyle(&n.HelpArgOptional)
	ensureStyle(&n.HelpBoldDim)
	ensureStyle(&n.HelpCommand)
	ensureStyle(&n.HelpDim)
	ensureStyle(&n.HelpEnumDefault)
	ensureStyle(&n.HelpFlag)
	ensureStyle(&n.HelpRepeatEllipsis)
	ensureStyle(&n.HelpSection)
	ensureStyle(&n.HelpSubcommand)
	ensureStyle(&n.HelpValuePlaceholder)
	ensureStyle(&n.MarkdownCode)
	ensureStyle(&n.MarkdownText)

	if n.HelpUsageExample.Prompt == "" {
		n.HelpUsageExample.Prompt = "$"
	}
	if n.HelpKeyValueSeparator == 0 {
		n.HelpKeyValueSeparator = ' '
	}
	if n.HelpDefaultOpen == "" {
		n.HelpDefaultOpen = "("
	}
	if n.HelpDefaultClose == "" {
		n.HelpDefaultClose = ")"
	}
	if n.HelpExampleOpen == "" {
		n.HelpExampleOpen = "("
	}
	if n.HelpExampleClose == "" {
		n.HelpExampleClose = ")"
	}

	return &n
}
