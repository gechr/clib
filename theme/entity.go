package theme

import (
	"image/color"

	"charm.land/lipgloss/v2"
	xpalette "github.com/gechr/x/palette"
)

// EntityColor returns a stable theme color for text.
func (th *Theme) EntityColor(text string) color.Color {
	if th == nil {
		return nil
	}
	return xpalette.Palette(th.EntityColors).Color(text)
}

// EntityStyle returns a foreground style using [Theme.EntityColor].
func (th *Theme) EntityStyle(text string) lipgloss.Style {
	c := th.EntityColor(text)
	if c == nil {
		return lipgloss.NewStyle()
	}
	return lipgloss.NewStyle().Foreground(c)
}

// RenderEntity renders text with its stable theme entity color.
func (th *Theme) RenderEntity(text string) string {
	return th.EntityStyle(text).Render(text)
}
