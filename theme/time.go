package theme

import (
	"time"

	"github.com/gechr/clib/human"
)

// RenderTimeAgo formats a time as a colored relative string using the theme's
// TimeAgoThresholds. When tty is false, returns plain text.
func (th *Theme) RenderTimeAgo(t time.Time, tty bool) string {
	return th.renderTimeAgo(t, tty, false)
}

// RenderTimeAgoCompact formats a time as a compact colored relative string
// (e.g. "15m ago" instead of "15 minutes ago").
func (th *Theme) RenderTimeAgoCompact(t time.Time, tty bool) string {
	return th.renderTimeAgo(t, tty, true)
}

func (th *Theme) renderTimeAgo(t time.Time, tty, compact bool) string {
	now := time.Now().UTC()
	var display string
	if compact {
		display = human.FormatTimeAgoCompactFrom(t, now)
	} else {
		display = human.FormatTimeAgoFrom(t, now)
	}
	if !tty {
		return display
	}

	d := now.Sub(t)
	if d < 0 {
		d = -d
	}

	for _, threshold := range th.TimeAgoThresholds {
		if d < threshold.MaxAge {
			return threshold.Style.Render(display)
		}
	}
	if th.Red == nil {
		return display
	}
	return th.Red.Render(display)
}
