package theme

import (
	"image/color"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
	"github.com/gechr/clib/human"
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
	HelpSection                *lipgloss.Style
	HelpCommand                *lipgloss.Style
	HelpSubcommand             *lipgloss.Style
	HelpFlag                   *lipgloss.Style
	HelpArg                    *lipgloss.Style
	HelpArgOptional            *lipgloss.Style
	HelpValuePlaceholder       *lipgloss.Style
	HelpDim                    *lipgloss.Style
	HelpBoldDim                *lipgloss.Style
	HelpUsageExample           HelpUsageExampleStyle // Examples section prompt and command styling.
	HelpFlagExample            *lipgloss.Style       // [example: ...] annotations in flag descriptions.
	HelpFlagNote               *lipgloss.Style       // Trailing (...) notes in flag descriptions.
	HelpFlagDefault            *lipgloss.Style       // [default: ...] annotations in flag descriptions.
	HelpEnumDefault            *lipgloss.Style       // Default value in EnumStyleHighlightDefault lists (default: dim green).
	HelpDescBacktick           *lipgloss.Style       // Backtick-enclosed text in flag descriptions (nil = leave backticks intact).
	HelpFlagBacktick           *lipgloss.Style       // Override for backtick-enclosed flag-like text in descriptions (nil = fall back to HelpFlag).
	HelpKeyValueSeparator      rune                  // Separator between flag and placeholder (default: ' ').
	HelpKeyValueSeparatorStyle *lipgloss.Style       // Style applied to the separator (default: nil = unstyled).
	HelpRepeatEllipsis         *lipgloss.Style       // "…" suffix on repeatable flag placeholders (default: dim red).
	HelpRepeatEllipsisEnabled  bool                  // Whether to show "…" suffix on repeatable placeholders (default: true).
	EnumStyle                  EnumStyle             // How enum values are rendered in help output.

	// Markdown styles.
	MarkdownCode *lipgloss.Style
	MarkdownText *lipgloss.Style

	// Time-ago thresholds (ordered by MaxAge ascending).
	TimeAgoThresholds []TimeAgoThreshold

	// Entity color palette for unique entity colorization.
	EntityColors []color.Color
}

// Option configures a Theme.
type Option func(*Theme)

// Default returns the default theme matching prl's hardcoded styles.
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
		name:      "default",
		Bold:      new(lipgloss.NewStyle().Bold(true)),
		Dim:       new(lipgloss.NewStyle().Faint(true)),
		Red:       new(lipgloss.NewStyle().Foreground(lipgloss.Color("1"))),
		Green:     new(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))),
		Yellow:    new(lipgloss.NewStyle().Foreground(lipgloss.Color("3"))),
		Blue:      new(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))),
		Magenta:   new(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))),
		Orange:    new(lipgloss.NewStyle().Foreground(lipgloss.Color("208"))),
		BoldGreen: new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))),

		HelpSection:          new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("3"))),
		HelpCommand:          new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))),
		HelpSubcommand:       new(lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2"))),
		HelpFlag:             new(lipgloss.NewStyle().Foreground(lipgloss.Color("1"))),
		HelpArg:              new(lipgloss.NewStyle().Foreground(lipgloss.Color("4"))),
		HelpArgOptional:      new(lipgloss.NewStyle().Foreground(lipgloss.Color("5"))),
		HelpValuePlaceholder: new(lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("1"))),
		HelpDim:              new(lipgloss.NewStyle().Faint(true)),
		HelpBoldDim: new(lipgloss.NewStyle().
			Bold(true).
			Faint(true).
			Foreground(lipgloss.Color("4"))),
		HelpEnumDefault: new(
			lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("2")),
		),
		HelpUsageExample: HelpUsageExampleStyle{
			Prompt:      "$",
			PromptStyle: lipgloss.NewStyle().Foreground(lipgloss.Color("2")),
		},
		HelpFlagExample:       new(lipgloss.NewStyle().Foreground(lipgloss.Color("2"))),
		HelpFlagNote:          new(lipgloss.NewStyle().Faint(true)),
		HelpFlagDefault:       new(lipgloss.NewStyle().Faint(true)),
		HelpDescBacktick:      new(lipgloss.NewStyle().Foreground(lipgloss.Color("189"))),
		HelpKeyValueSeparator: ' ',
		HelpKeyValueSeparatorStyle: new(
			lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("1")),
		),
		HelpRepeatEllipsis: new(
			lipgloss.NewStyle().Faint(true).Foreground(lipgloss.Color("1")),
		),
		HelpRepeatEllipsisEnabled: true,
		EnumStyle:                 EnumStyleHighlightDefault,

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
	ensureStyle(&n.HelpSection)
	ensureStyle(&n.HelpCommand)
	ensureStyle(&n.HelpSubcommand)
	ensureStyle(&n.HelpFlag)
	ensureStyle(&n.HelpArg)
	ensureStyle(&n.HelpArgOptional)
	ensureStyle(&n.HelpValuePlaceholder)
	ensureStyle(&n.HelpDim)
	ensureStyle(&n.HelpBoldDim)
	ensureStyle(&n.HelpEnumDefault)
	ensureStyle(&n.HelpRepeatEllipsis)
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

// WithBlue sets the blue color style.
func WithBlue(s lipgloss.Style) Option {
	return func(t *Theme) { t.Blue = new(s) }
}

// WithBold sets the bold style.
func WithBold(s lipgloss.Style) Option {
	return func(t *Theme) { t.Bold = new(s) }
}

// WithBoldGreen sets the bold green color style.
func WithBoldGreen(s lipgloss.Style) Option {
	return func(t *Theme) { t.BoldGreen = new(s) }
}

// WithDim sets the dim (faint) style.
func WithDim(s lipgloss.Style) Option {
	return func(t *Theme) { t.Dim = new(s) }
}

// WithEntityColors sets the color palette for entity colorization.
func WithEntityColors(c []color.Color) Option {
	return func(t *Theme) { t.EntityColors = c }
}

// WithEnumStyle sets how enum values are rendered in help output.
func WithEnumStyle(s EnumStyle) Option {
	return func(t *Theme) { t.EnumStyle = s }
}

// WithGreen sets the green color style.
func WithGreen(s lipgloss.Style) Option {
	return func(t *Theme) { t.Green = new(s) }
}

// WithHelpArg sets the style for argument names in help output.
func WithHelpArg(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpArg = new(s) }
}

// WithHelpArgOptional sets the style for optional argument names in help output
// when they coexist with required arguments.
func WithHelpArgOptional(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpArgOptional = new(s) }
}

// WithHelpBoldDim sets the bold-dim style used in help output.
func WithHelpBoldDim(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpBoldDim = new(s) }
}

// WithHelpCommand sets the style for command names in help output.
func WithHelpCommand(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpCommand = new(s) }
}

// WithHelpDescBacktick sets the style for backtick-enclosed text in flag descriptions.
func WithHelpDescBacktick(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpDescBacktick = new(s) }
}

// WithHelpDim sets the dim style used in help output.
func WithHelpDim(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpDim = new(s) }
}

// WithHelpEnumDefault sets the style for the default value in enum lists.
func WithHelpEnumDefault(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpEnumDefault = new(s) }
}

// WithHelpFlag sets the style for flag names in help output.
func WithHelpFlag(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpFlag = new(s) }
}

// WithHelpFlagBacktick sets the style for backtick-enclosed flag-like text in flag descriptions.
func WithHelpFlagBacktick(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpFlagBacktick = new(s) }
}

// WithHelpFlagDefault sets the style for flag default annotations in help output.
func WithHelpFlagDefault(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpFlagDefault = new(s) }
}

// WithHelpFlagExample sets the style for flag example annotations in help output.
func WithHelpFlagExample(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpFlagExample = new(s) }
}

// WithHelpFlagNote sets the style for trailing flag notes in help output.
func WithHelpFlagNote(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpFlagNote = new(s) }
}

// WithHelpKeyValueSeparator sets the separator rune between flag and placeholder.
func WithHelpKeyValueSeparator(sep rune) Option {
	return func(t *Theme) { t.HelpKeyValueSeparator = sep }
}

// WithHelpKeyValueSeparatorStyle sets the style for the flag-placeholder separator.
func WithHelpKeyValueSeparatorStyle(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpKeyValueSeparatorStyle = new(s) }
}

// WithHelpPlaceholder sets the style for value placeholders in help output.
func WithHelpPlaceholder(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpValuePlaceholder = new(s) }
}

// WithHelpRepeatEllipsis sets the style for the repeat ellipsis on repeatable flags.
func WithHelpRepeatEllipsis(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpRepeatEllipsis = new(s) }
}

// WithHelpRepeatEllipsisEnabled sets whether repeatable flag placeholders show an ellipsis.
func WithHelpRepeatEllipsisEnabled(enabled bool) Option {
	return func(t *Theme) { t.HelpRepeatEllipsisEnabled = enabled }
}

// WithHelpSection sets the style for help section headers.
func WithHelpSection(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpSection = new(s) }
}

// WithHelpSubcommand sets the style for subcommand names in help output.
func WithHelpSubcommand(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpSubcommand = new(s) }
}

// WithHelpUsageExample sets the style for usage examples in help output.
func WithHelpUsageExample(s HelpUsageExampleStyle) Option {
	return func(t *Theme) { t.HelpUsageExample = s }
}

// WithMagenta sets the magenta color style.
func WithMagenta(s lipgloss.Style) Option {
	return func(t *Theme) { t.Magenta = new(s) }
}

// WithMarkdownCode sets the style for inline code in markdown rendering.
func WithMarkdownCode(s lipgloss.Style) Option {
	return func(t *Theme) { t.MarkdownCode = new(s) }
}

// WithMarkdownText sets the style for plain text in markdown rendering.
func WithMarkdownText(s lipgloss.Style) Option {
	return func(t *Theme) { t.MarkdownText = new(s) }
}

// WithOrange sets the orange color style.
func WithOrange(s lipgloss.Style) Option {
	return func(t *Theme) { t.Orange = new(s) }
}

// WithRed sets the red color style.
func WithRed(s lipgloss.Style) Option {
	return func(t *Theme) { t.Red = new(s) }
}

// WithTimeAgoThresholds sets the time-ago color thresholds.
func WithTimeAgoThresholds(th []TimeAgoThreshold) Option {
	return func(t *Theme) { t.TimeAgoThresholds = th }
}

// WithYellow sets the yellow color style.
func WithYellow(s lipgloss.Style) Option {
	return func(t *Theme) { t.Yellow = new(s) }
}
