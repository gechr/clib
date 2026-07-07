package theme_test

import (
	"fmt"
	"strings"
	"testing"

	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestPresets(t *testing.T) {
	presets := map[string]struct {
		fn           func() *theme.Theme
		entityColors bool
	}{
		"Dark":  {fn: theme.Dark, entityColors: true},
		"Light": {fn: theme.Light, entityColors: true},
		"PlainDark": {
			fn: func() *theme.Theme { return theme.Plain(theme.BackgroundDark) },
		},
		"PlainLight": {
			fn: func() *theme.Theme { return theme.Plain(theme.BackgroundLight) },
		},
		"MonochromeDark": {
			fn: func() *theme.Theme { return theme.Monochrome(theme.BackgroundDark) },
		},
		"MonochromeLight": {
			fn: func() *theme.Theme { return theme.Monochrome(theme.BackgroundLight) },
		},
		"Monokai":             {fn: theme.Monokai, entityColors: true},
		"CatppuccinLatte":     {fn: theme.CatppuccinLatte, entityColors: true},
		"CatppuccinFrappe":    {fn: theme.CatppuccinFrappe, entityColors: true},
		"CatppuccinMacchiato": {fn: theme.CatppuccinMacchiato, entityColors: true},
		"CatppuccinMocha":     {fn: theme.CatppuccinMocha, entityColors: true},
		"Dracula":             {fn: theme.Dracula, entityColors: true},
		"GruvboxDark":         {fn: theme.GruvboxDark, entityColors: true},
		"GruvboxLight":        {fn: theme.GruvboxLight, entityColors: true},
		"Nord":                {fn: theme.Nord, entityColors: true},
		"OneDark":             {fn: theme.OneDark, entityColors: true},
		"Synthwave":           {fn: theme.Synthwave, entityColors: true},
		"SolarizedDark":       {fn: theme.SolarizedDark, entityColors: true},
		"SolarizedLight":      {fn: theme.SolarizedLight, entityColors: true},
		"TokyoNight":          {fn: theme.TokyoNight, entityColors: true},
	}

	for name, preset := range presets {
		t.Run(name, func(t *testing.T) {
			th := preset.fn()
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
			require.NotNil(t, th.HelpArgOptional, "HelpArgOptional")
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

			require.Nil(t, th.HelpAlias)
			require.Equal(t, "$", th.HelpUsageExample.Prompt)
			require.Equal(t, ' ', th.HelpKeyValueSeparator)
			require.True(t, th.HelpRepeatEllipsisEnabled)
			require.Equal(t, theme.EnumStyleHighlightDefault, th.EnumStyle)
			require.Nil(t, th.HelpFlagBacktick)
			if preset.entityColors {
				require.Len(t, th.EntityColors, 20)
			} else {
				require.Empty(t, th.EntityColors)
			}
			require.Len(t, th.TimeAgoThresholds, 5)
		})
	}
}

func TestMonochrome_NoColors(t *testing.T) {
	th := theme.Monochrome(theme.BackgroundDark)

	// Semantic color styles should render without any ANSI color codes -
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

func TestThemeUnmarshalText(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  func() *theme.Theme
	}{
		{name: "dark", input: "dark", want: theme.Dark},
		{name: "light", input: "LIGHT", want: theme.Light},
		{
			name:  "plain dark",
			input: "plain dark",
			want:  func() *theme.Theme { return theme.Plain(theme.BackgroundDark) },
		},
		{
			name:  "monochrome light",
			input: "monochrome_light",
			want:  func() *theme.Theme { return theme.Monochrome(theme.BackgroundLight) },
		},
		{name: "monokai", input: "MONOKAI", want: theme.Monokai},
		{name: "catppuccin hyphen", input: "catppuccin-mocha", want: theme.CatppuccinMocha},
		{
			name:  "catppuccin underscore",
			input: "catppuccin_macchiato",
			want:  theme.CatppuccinMacchiato,
		},
		{name: "catppuccin compact", input: "catppuccinfrappe", want: theme.CatppuccinFrappe},
		{name: "dracula", input: "dracula", want: theme.Dracula},
		{name: "gruvbox-dark", input: "gruvbox-dark", want: theme.GruvboxDark},
		{name: "gruvbox-light", input: "gruvbox-light", want: theme.GruvboxLight},
		{name: "nord", input: "nord", want: theme.Nord},
		{name: "one-dark", input: "one-dark", want: theme.OneDark},
		{name: "synthwave", input: "synthwave", want: theme.Synthwave},
		{name: "solarized dark", input: "solarized-dark", want: theme.SolarizedDark},
		{name: "solarized light", input: "solarized-light", want: theme.SolarizedLight},
		{name: "tokyo-night", input: "tokyo-night", want: theme.TokyoNight},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var got theme.Theme
			require.NoError(t, got.UnmarshalText([]byte(tt.input)))
			require.Equal(t, tt.want().String(), got.String())
			require.Equal(t, tt.want().HelpSection.Render("x"), got.HelpSection.Render("x"))
			require.Equal(t, tt.want().HelpFlag.Render("x"), got.HelpFlag.Render("x"))
			require.Equal(t, tt.want().MarkdownCode.Render("x"), got.MarkdownCode.Render("x"))
		})
	}
}

func TestThemeMarshalText(t *testing.T) {
	got, err := theme.Dracula().MarshalText()
	require.NoError(t, err)
	require.Equal(t, "dracula", string(got))
}

func TestThemeMarshalTextCustom(t *testing.T) {
	_, err := theme.Dark().With(theme.WithHelpKeyValueSeparator('=')).MarshalText()
	require.EqualError(t, err, "cannot marshal custom theme")
}

func TestNames(t *testing.T) {
	want := []string{
		"dark",
		"light",
		"catppuccin-frappe",
		"catppuccin-latte",
		"catppuccin-macchiato",
		"catppuccin-mocha",
		"dracula",
		"gruvbox-dark",
		"gruvbox-light",
		"monokai",
		"monochrome-dark",
		"monochrome-light",
		"nord",
		"one-dark",
		"plain-dark",
		"plain-light",
		"synthwave",
		"solarized-dark",
		"solarized-light",
		"tokyo-night",
	}

	got := theme.Names()
	require.Equal(t, want, got)

	got[0] = "mutated"
	require.Equal(t, want, theme.Names())

	for _, name := range want {
		t.Run(name, func(t *testing.T) {
			var got theme.Theme
			require.NoError(t, got.UnmarshalText([]byte(name)))
		})
	}
}

func TestThemeUnmarshalTextInvalid(t *testing.T) {
	var got theme.Theme
	err := got.UnmarshalText([]byte("bogus"))
	require.EqualError(t, err, unknownThemeError("bogus"))
}

func TestThemeUnmarshalTextInvalidNames(t *testing.T) {
	for _, input := range []string{
		"default",
		"default-dark",
		"default-light",
		"plain",
		"monochrome",
		"solarized",
	} {
		t.Run(input, func(t *testing.T) {
			var got theme.Theme
			err := got.UnmarshalText([]byte(input))
			require.EqualError(t, err, unknownThemeError(input))
		})
	}
}

func unknownThemeError(input string) string {
	return fmt.Sprintf("unknown theme %q (valid: %s)", input, strings.Join(theme.Names(), ", "))
}
