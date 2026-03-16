package theme_test

import (
	"testing"
	"time"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestRenderTimeAgo_Now_TTY(t *testing.T) {
	th := theme.Default()
	got := th.RenderTimeAgo(time.Now().UTC().Add(-10*time.Second), true)
	plain := ansi.Strip(got)
	require.Equal(t, "now", plain)
}

func TestRenderTimeAgo_NoTTY(t *testing.T) {
	th := theme.Default()

	got := th.RenderTimeAgo(time.Now().UTC().Add(-5*time.Minute), false)
	require.Equal(t, "5 minutes ago", got)

	got = th.RenderTimeAgo(time.Now().UTC().Add(-10*time.Second), false)
	require.Equal(t, "now", got)
}

func TestRenderTimeAgo_BeyondAllThresholds_FallsBackToRed(t *testing.T) {
	th := theme.Default()
	// 365 days ago exceeds all default thresholds (max is 30 days).
	old := time.Now().UTC().Add(-365 * 24 * time.Hour)
	got := th.RenderTimeAgo(old, true)
	plain := ansi.Strip(got)
	require.NotEmpty(t, plain)
	// Should use the Red style (styled output, not plain).
	require.NotEqual(t, plain, got)
	// Verify it uses Red specifically.
	expected := th.Red.Render(plain)
	require.Equal(t, expected, got)
}

func TestRenderTimeAgo_FutureTime(t *testing.T) {
	th := theme.Default()
	// Future time - the absolute duration should still match a threshold.
	future := time.Now().UTC().Add(30 * time.Second)
	got := th.RenderTimeAgo(future, true)
	plain := ansi.Strip(got)
	require.NotEmpty(t, plain)
	// Should be styled (TTY on).
	require.NotEqual(t, plain, got)
}

func TestRenderTimeAgo_EachThreshold(t *testing.T) {
	th := theme.Default()

	tests := []struct {
		name   string
		offset time.Duration
	}{
		{"within_minute", -30 * time.Second},
		{"within_hour", -30 * time.Minute},
		{"within_day", -12 * time.Hour},
		{"within_2_weeks", -7 * 24 * time.Hour},
		{"within_30_days", -20 * 24 * time.Hour},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ts := time.Now().UTC().Add(tt.offset)
			got := th.RenderTimeAgo(ts, true)
			plain := ansi.Strip(got)
			require.NotEmpty(t, plain)
			// All should be styled in TTY mode.
			require.NotEqual(t, plain, got)
		})
	}
}

func TestRenderTimeAgo_NilRed_DoesNotPanic(t *testing.T) {
	th := &theme.Theme{} // bare literal: nil Red, no thresholds
	old := time.Now().UTC().Add(-365 * 24 * time.Hour)
	got := th.RenderTimeAgo(old, true)
	// Should return plain text when Red is nil.
	require.NotEmpty(t, got)
	require.Equal(t, got, ansi.Strip(got))
}

func TestRenderTimeAgo_NoThresholds_AlwaysRed(t *testing.T) {
	th := theme.New(theme.WithTimeAgoThresholds(nil))
	got := th.RenderTimeAgo(time.Now().UTC().Add(-5*time.Second), true)
	plain := ansi.Strip(got)
	require.NotEmpty(t, plain)
	// With no thresholds, everything falls through to Red.
	expected := th.Red.Render(plain)
	require.Equal(t, expected, got)
}

func TestRenderTimeAgoCompact_NoTTY(t *testing.T) {
	th := theme.Default()

	got := th.RenderTimeAgoCompact(time.Now().UTC().Add(-5*time.Minute), false)
	require.Equal(t, "5m ago", got)

	got = th.RenderTimeAgoCompact(time.Now().UTC().Add(-10*time.Second), false)
	require.Equal(t, "now", got)
}

func TestRenderTimeAgoCompact_TTY(t *testing.T) {
	th := theme.Default()
	got := th.RenderTimeAgoCompact(time.Now().UTC().Add(-30*time.Second), true)
	require.NotEqual(t, ansi.Strip(got), got)
}

func TestRenderTimeAgoCompact_NilRed(t *testing.T) {
	th := &theme.Theme{}
	old := time.Now().UTC().Add(-365 * 24 * time.Hour)
	got := th.RenderTimeAgoCompact(old, true)
	require.NotEmpty(t, got)
	require.Equal(t, got, ansi.Strip(got))
}
