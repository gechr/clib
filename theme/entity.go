package theme

import (
	"hash/fnv"
	"image/color"
	"math"

	"charm.land/lipgloss/v2"
)

// EntityColor returns a stable theme color for text.
func (th *Theme) EntityColor(text string) color.Color {
	if th == nil || len(th.EntityColors) == 0 {
		return nil
	}

	hasher := fnv.New32a()
	_, _ = hasher.Write([]byte(text))
	return th.EntityColors[int(hasher.Sum32()&math.MaxInt32)%len(th.EntityColors)]
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
