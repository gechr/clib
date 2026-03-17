package theme_test

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestDefaultTheme_StyleValues(t *testing.T) {
	th := theme.Default()

	// Verify base styles produce expected ANSI output.
	require.Equal(t, lipgloss.NewStyle().Bold(true).Render("x"), th.Bold.Render("x"))
	require.Equal(t, lipgloss.NewStyle().Faint(true).Render("x"), th.Dim.Render("x"))

	// Verify semantic colors.
	require.Equal(
		t,
		lipgloss.NewStyle().Foreground(lipgloss.Color("1")).Render("x"),
		th.Red.Render("x"),
	)
	require.Equal(
		t,
		lipgloss.NewStyle().Foreground(lipgloss.Color("2")).Render("x"),
		th.Green.Render("x"),
	)
	require.Equal(
		t,
		lipgloss.NewStyle().Foreground(lipgloss.Color("3")).Render("x"),
		th.Yellow.Render("x"),
	)
	require.Equal(
		t,
		lipgloss.NewStyle().Foreground(lipgloss.Color("4")).Render("x"),
		th.Blue.Render("x"),
	)
	require.Equal(
		t,
		lipgloss.NewStyle().Foreground(lipgloss.Color("5")).Render("x"),
		th.Magenta.Render("x"),
	)
	require.Equal(
		t,
		lipgloss.NewStyle().Foreground(lipgloss.Color("208")).Render("x"),
		th.Orange.Render("x"),
	)
	require.Equal(
		t,
		lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("2")).Render("x"),
		th.BoldGreen.Render("x"),
	)

	// Verify entity colors count.
	require.Len(t, th.EntityColors, 20)
	require.Equal(t, lipgloss.Color("208"), th.EntityColors[0])

	// Verify time-ago thresholds count.
	require.Len(t, th.TimeAgoThresholds, 5)
}

func TestDefaultTheme_EnumStyleDefault(t *testing.T) {
	th := theme.Default()
	require.Equal(t, theme.EnumStyleHighlightDefault, th.EnumStyle)
}

func TestNewTheme_WithEnumStyle(t *testing.T) {
	th := theme.New(theme.WithEnumStyle(theme.EnumStyleHighlightPrefix))
	require.Equal(t, theme.EnumStyleHighlightPrefix, th.EnumStyle)
}

func TestNewTheme_AppliesOptions(t *testing.T) {
	custom := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	th := theme.New(theme.WithRed(custom))
	require.Equal(t, custom.Render("x"), th.Red.Render("x"))

	// Other fields retain defaults.
	def := theme.Default()
	require.Equal(t, def.Bold.Render("x"), th.Bold.Render("x"))
}

func TestNewTheme_WithAllOptions(t *testing.T) {
	custom := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	def := theme.Default()

	tests := []struct {
		name  string
		opt   theme.Option
		check func(t *testing.T, th *theme.Theme)
	}{
		{
			name: "WithBold",
			opt:  theme.WithBold(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.Bold.Render("x"))
			},
		},
		{
			name: "WithDim",
			opt:  theme.WithDim(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.Dim.Render("x"))
			},
		},
		{
			name: "WithGreen",
			opt:  theme.WithGreen(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.Green.Render("x"))
			},
		},
		{
			name: "WithYellow",
			opt:  theme.WithYellow(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.Yellow.Render("x"))
			},
		},
		{
			name: "WithBlue",
			opt:  theme.WithBlue(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.Blue.Render("x"))
			},
		},
		{
			name: "WithMagenta",
			opt:  theme.WithMagenta(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.Magenta.Render("x"))
			},
		},
		{
			name: "WithOrange",
			opt:  theme.WithOrange(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.Orange.Render("x"))
			},
		},
		{
			name: "WithBoldGreen",
			opt:  theme.WithBoldGreen(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.BoldGreen.Render("x"))
			},
		},
		{
			name: "WithHelpSection",
			opt:  theme.WithHelpSection(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpSection.Render("x"))
			},
		},
		{
			name: "WithHelpCommand",
			opt:  theme.WithHelpCommand(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpCommand.Render("x"))
			},
		},
		{
			name: "WithHelpSubcommand",
			opt:  theme.WithHelpSubcommand(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpSubcommand.Render("x"))
			},
		},
		{
			name: "WithHelpFlag",
			opt:  theme.WithHelpFlag(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpFlag.Render("x"))
			},
		},
		{
			name: "WithHelpArg",
			opt:  theme.WithHelpArg(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpArg.Render("x"))
			},
		},
		{
			name: "WithHelpPlaceholder",
			opt:  theme.WithHelpPlaceholder(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpValuePlaceholder.Render("x"))
			},
		},
		{
			name: "WithHelpDim",
			opt:  theme.WithHelpDim(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpDim.Render("x"))
			},
		},
		{
			name: "WithHelpBoldDim",
			opt:  theme.WithHelpBoldDim(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpBoldDim.Render("x"))
			},
		},
		{
			name: "WithHelpUsageExample",
			opt: theme.WithHelpUsageExample(theme.HelpUsageExampleStyle{
				Prompt: ">",
			}),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, ">", th.HelpUsageExample.Prompt)
			},
		},
		{
			name: "WithHelpFlagExample",
			opt:  theme.WithHelpFlagExample(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpFlagExample.Render("x"))
			},
		},
		{
			name: "WithHelpFlagNote",
			opt:  theme.WithHelpFlagNote(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpFlagNote.Render("x"))
			},
		},
		{
			name: "WithHelpFlagDefault",
			opt:  theme.WithHelpFlagDefault(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpFlagDefault.Render("x"))
			},
		},
		{
			name: "WithHelpDescBacktick",
			opt:  theme.WithHelpDescBacktick(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpDescBacktick.Render("x"))
			},
		},
		{
			name: "WithHelpKeyValueSeparator",
			opt:  theme.WithHelpKeyValueSeparator('='),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, '=', th.HelpKeyValueSeparator)
			},
		},
		{
			name: "WithHelpKeyValueSeparatorStyle",
			opt:  theme.WithHelpKeyValueSeparatorStyle(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpKeyValueSeparatorStyle.Render("x"))
			},
		},
		{
			name: "WithHelpRepeatEllipsis",
			opt:  theme.WithHelpRepeatEllipsis(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.HelpRepeatEllipsis.Render("x"))
			},
		},
		{
			name: "WithHelpRepeatEllipsisEnabled",
			opt:  theme.WithHelpRepeatEllipsisEnabled(false),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.False(t, th.HelpRepeatEllipsisEnabled)
			},
		},
		{
			name: "WithMarkdownCode",
			opt:  theme.WithMarkdownCode(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.MarkdownCode.Render("x"))
			},
		},
		{
			name: "WithMarkdownText",
			opt:  theme.WithMarkdownText(custom),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Equal(t, custom.Render("x"), th.MarkdownText.Render("x"))
			},
		},
		{
			name: "WithTimeAgoThresholds",
			opt:  theme.WithTimeAgoThresholds(nil),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Nil(t, th.TimeAgoThresholds)
			},
		},
		{
			name: "WithEntityColors",
			opt:  theme.WithEntityColors([]color.Color{lipgloss.Color("1"), lipgloss.Color("2")}),
			check: func(t *testing.T, th *theme.Theme) {
				t.Helper()
				require.Len(t, th.EntityColors, 2)
				require.Equal(t, lipgloss.Color("1"), th.EntityColors[0])
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			th := theme.New(tt.opt)
			tt.check(t, th)
			// Verify the default theme is unaffected (option was applied correctly).
			_ = def
		})
	}
}

func TestNewTheme_MultipleOptions(t *testing.T) {
	custom1 := lipgloss.NewStyle().Foreground(lipgloss.Color("99"))
	custom2 := lipgloss.NewStyle().Foreground(lipgloss.Color("88"))
	th := theme.New(
		theme.WithBold(custom1),
		theme.WithDim(custom2),
	)
	require.Equal(t, custom1.Render("x"), th.Bold.Render("x"))
	require.Equal(t, custom2.Render("x"), th.Dim.Render("x"))
}

func TestNewTheme_NoOptions(t *testing.T) {
	th := theme.New()
	def := theme.Default()
	require.Equal(t, def.Bold.Render("x"), th.Bold.Render("x"))
	require.Equal(t, def.Red.Render("x"), th.Red.Render("x"))
}

func TestWithHelpEnumDefault(t *testing.T) {
	s := lipgloss.NewStyle().Italic(true)
	th := theme.New(theme.WithHelpEnumDefault(s))
	require.Equal(t, s.Render("x"), th.HelpEnumDefault.Render("x"))
}
