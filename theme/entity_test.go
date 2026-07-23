package theme_test

import (
	"image/color"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestEntityColorStable(t *testing.T) {
	th := theme.Dark()

	first := th.EntityColor("production")
	second := th.EntityColor("production")

	require.NotNil(t, first)
	require.Equal(t, first, second)
}

func TestEntityColorAllInputsResolve(t *testing.T) {
	th := theme.Dark()
	palette := map[color.Color]bool{}
	for _, c := range th.EntityColors {
		palette[c] = true
	}

	// Hash values vary widely (including ones whose high bit is set); every
	// input must map to a color from the palette and never panic.
	for _, text := range []string{
		"production", "staging", "dev", "us-east-1", "prod",
		"qa", "canary", "test", "blue", "green", "alpha", "beta",
	} {
		c := th.EntityColor(text)
		require.NotNil(t, c, text)
		require.True(t, palette[c], text)
	}
}

func TestEntityColorUsesThemePalette(t *testing.T) {
	palette := []color.Color{lipgloss.Color("#112233")}
	th := theme.Dark().With(theme.WithEntityColors(palette...))

	require.Equal(t, palette[0], th.EntityColor("anything"))
}

func TestEntityColorEmptyPalette(t *testing.T) {
	th := theme.Dark().With(theme.WithEntityColors())

	require.Nil(t, th.EntityColor("anything"))
	require.Equal(t, "anything", th.RenderEntity("anything"))
}

func TestWithTrueColorGeneratesDistinctPalette(t *testing.T) {
	th := theme.Dark().With(theme.WithTrueColor())

	require.Len(t, th.EntityColors, 256)

	seen := map[color.Color]bool{}
	for _, c := range th.EntityColors {
		require.False(t, seen[c], c)
		seen[c] = true
	}
}

func TestWithTrueColorAdaptsToBackground(t *testing.T) {
	dark := theme.Dark().With(theme.WithTrueColor())
	light := theme.Light().With(theme.WithTrueColor())

	require.Len(t, light.EntityColors, len(dark.EntityColors))
	require.NotEqual(t, dark.EntityColors, light.EntityColors)
}

func TestMonochromeEntityColorEmpty(t *testing.T) {
	th := theme.Monochrome(theme.BackgroundDark)

	require.Nil(t, th.EntityColor("anything"))
	require.Equal(t, "anything", th.RenderEntity("anything"))
}

func TestLightEntityPaletteIsDarkerThanDarkPalette(t *testing.T) {
	light := theme.Light()
	dark := theme.Dark()

	require.Len(t, light.EntityColors, len(dark.EntityColors))
	require.Less(t, luminance(light.EntityColors[0]), luminance(dark.EntityColors[0]))
}

func luminance(c color.Color) uint32 {
	r, g, b, _ := c.RGBA()
	return r + g + b
}
