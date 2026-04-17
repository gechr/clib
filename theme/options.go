package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
)

// Option configures a Theme.
type Option func(*Theme)

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

// WithHelpAlias sets the style for alias names in the Aliases section.
// When unset, aliases fall back to the HelpCommand style.
func WithHelpAlias(s lipgloss.Style) Option {
	return func(t *Theme) { t.HelpAlias = new(s) }
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
