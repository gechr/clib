package theme

import (
	"image/color"
	"strings"
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
// Use [Default] or [New] to construct a Theme, or [Init] to fill nil
// fields on an existing theme.
type Theme struct {
	name string

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

// Default returns the default theme.
func Default() *Theme {
	if name := strings.TrimSpace(getEnv(envTheme)); name != "" {
		var th Theme
		if err := th.UnmarshalText([]byte(name)); err == nil {
			return &th
		}
	}
	return defaultTheme()
}

func defaultTheme() *Theme {
	return &Theme{
		name:      themeNameDefault,
		Bold:      new(lipgloss.NewStyle().Bold(true)),
		Dim:       new(lipgloss.NewStyle().Faint(true)),
		Red:       new(lipgloss.NewStyle().Foreground(lipgloss.Color("1"))),
		Green:     new(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))),
		Yellow:    new(lipgloss.NewStyle().Foreground(lipgloss.Color("3"))),
		Blue:      new(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))),
		Magenta:   new(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))),
		Orange:    new(lipgloss.NewStyle().Foreground(lipgloss.Color("208"))),
		BoldGreen: new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))),

		HelpArg:         new(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))),
		HelpArgOptional: new(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))),
		HelpBoldDim: new(lipgloss.NewStyle().
			Bold(true).
			Faint(true).
			Foreground(lipgloss.Color("4"))),
		HelpCommand:      new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))),
		HelpDescBacktick: new(lipgloss.NewStyle().Foreground(lipgloss.Color("189"))),
		HelpDim:          new(lipgloss.NewStyle().Faint(true)),
		HelpEnumDefault: new(
			lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("2")),
		),
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

		EntityColors: []color.Color{
			lipgloss.Color("208"),
			lipgloss.Color("51"),
			lipgloss.Color("226"),
			lipgloss.Color("207"),
			lipgloss.Color("82"),
			lipgloss.Color("75"),
			lipgloss.Color("214"),
			lipgloss.Color("177"),
			lipgloss.Color("48"),
			lipgloss.Color("87"),
			lipgloss.Color("220"),
			lipgloss.Color("141"),
			lipgloss.Color("118"),
			lipgloss.Color("50"),
			lipgloss.Color("213"),
			lipgloss.Color("111"),
			lipgloss.Color("156"),
			lipgloss.Color("183"),
			lipgloss.Color("229"),
			lipgloss.Color("123"),
		},
	}
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

	return &n
}
