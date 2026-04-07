package theme_test

import (
	"testing"

	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestPresets(t *testing.T) {
	presets := map[string]func() *theme.Theme{
		"Monochrome":          theme.Monochrome,
		"Monokai":             theme.Monokai,
		"CatppuccinLatte":     theme.CatppuccinLatte,
		"CatppuccinFrappe":    theme.CatppuccinFrappe,
		"CatppuccinMacchiato": theme.CatppuccinMacchiato,
		"CatppuccinMocha":     theme.CatppuccinMocha,
		"Dracula":             theme.Dracula,
	}

	for name, fn := range presets {
		t.Run(name, func(t *testing.T) {
			th := fn()
			require.NotNil(t, th)

			// All style pointers must be non-nil.
			require.NotNil(t, th.Bold, "Bold")
			require.NotNil(t, th.Dim, "Dim")
			require.NotNil(t, th.Red, "Red")
			require.NotNil(t, th.Green, "Green")
			require.NotNil(t, th.Yellow, "Yellow")
			require.NotNil(t, th.Blue, "Blue")
			require.NotNil(t, th.Magenta, "Magenta")
			require.NotNil(t, th.Orange, "Orange")
			require.NotNil(t, th.BoldGreen, "BoldGreen")
			require.NotNil(t, th.HelpSection, "HelpSection")
			require.NotNil(t, th.HelpCommand, "HelpCommand")
			require.NotNil(t, th.HelpSubcommand, "HelpSubcommand")
			require.NotNil(t, th.HelpFlag, "HelpFlag")
			require.NotNil(t, th.HelpArg, "HelpArg")
			require.NotNil(t, th.HelpValuePlaceholder, "HelpValuePlaceholder")
			require.NotNil(t, th.HelpDim, "HelpDim")
			require.NotNil(t, th.HelpBoldDim, "HelpBoldDim")
			require.NotNil(t, th.HelpEnumDefault, "HelpEnumDefault")
			require.NotNil(t, th.HelpFlagExample, "HelpFlagExample")
			require.NotNil(t, th.HelpFlagNote, "HelpFlagNote")
			require.NotNil(t, th.HelpFlagDefault, "HelpFlagDefault")
			require.NotNil(t, th.HelpDescBacktick, "HelpDescBacktick")
			require.NotNil(t, th.HelpKeyValueSeparatorStyle, "HelpKeyValueSeparatorStyle")
			require.NotNil(t, th.HelpRepeatEllipsis, "HelpRepeatEllipsis")
			require.NotNil(t, th.MarkdownCode, "MarkdownCode")
			require.NotNil(t, th.MarkdownText, "MarkdownText")

			require.Equal(t, "$", th.HelpUsageExample.Prompt)
			require.Equal(t, ' ', th.HelpKeyValueSeparator)
			require.True(t, th.HelpRepeatEllipsisEnabled)
			require.Equal(t, theme.EnumStyleHighlightDefault, th.EnumStyle)
			require.Len(t, th.EntityColors, 20)
			require.Len(t, th.TimeAgoThresholds, 5)
		})
	}
}

func TestMonochrome_NoColors(t *testing.T) {
	th := theme.Monochrome()

	// Semantic color styles should render without any ANSI color codes —
	// they use only bold/dim/plain formatting.
	require.Equal(t, "x", th.Red.Render("x"))
	require.Equal(t, "x", th.Green.Render("x"))
	require.Equal(t, "x", th.Yellow.Render("x"))
	require.Equal(t, "x", th.Blue.Render("x"))
	require.Equal(t, "x", th.Magenta.Render("x"))
	require.Equal(t, "x", th.Orange.Render("x"))
}

func TestPresets_WithApplies(t *testing.T) {
	th := theme.Monokai().With(theme.WithHelpKeyValueSeparator('='))
	require.Equal(t, '=', th.HelpKeyValueSeparator)
}
