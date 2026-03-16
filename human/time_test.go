package human

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestFormatTimeAgoFrom(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		// Past times.
		{"just now", now.Add(-10 * time.Second), "now"},
		{"5 minutes", now.Add(-5 * time.Minute), "5 minutes ago"},
		{"90 seconds", now.Add(-90 * time.Second), "1 minute ago"},
		{"3 hours", now.Add(-3 * time.Hour), "3 hours ago"},
		{"90 minutes", now.Add(-90 * time.Minute), "1 hour ago"},
		{"3 days", now.Add(-3 * 24 * time.Hour), "3 days ago"},
		{"36 hours", now.Add(-36 * time.Hour), "1 day ago"},
		{"14 days", now.Add(-14 * 24 * time.Hour), "2 weeks ago"},
		{"10 days", now.Add(-10 * 24 * time.Hour), "1 week ago"},
		{"60 days", now.Add(-60 * 24 * time.Hour), "2 months ago"},
		{"35 days", now.Add(-35 * 24 * time.Hour), "1 month ago"},
		{"400 days", now.Add(-400 * 24 * time.Hour), "1 year ago"},
		{"800 days", now.Add(-800 * 24 * time.Hour), "2 years ago"},

		// Future times.
		{"future small", now.Add(30 * time.Second), "now"},
		{"future 5 minutes", now.Add(5 * time.Minute), "in 5 minutes"},
		{"future 90 seconds", now.Add(90 * time.Second), "in 1 minute"},
		{"future 3 hours", now.Add(3 * time.Hour), "in 3 hours"},
		{"future 90 minutes", now.Add(90 * time.Minute), "in 1 hour"},
		{"future 3 days", now.Add(3 * 24 * time.Hour), "in 3 days"},
		{"future 36 hours", now.Add(36 * time.Hour), "in 1 day"},
		{"future 14 days", now.Add(14 * 24 * time.Hour), "in 2 weeks"},
		{"future 10 days", now.Add(10 * 24 * time.Hour), "in 1 week"},
		{"future 60 days", now.Add(60 * 24 * time.Hour), "in 2 months"},
		{"future 35 days", now.Add(35 * 24 * time.Hour), "in 1 month"},
		{"future 400 days", now.Add(400 * 24 * time.Hour), "in 1 year"},
		{"future 800 days", now.Add(800 * 24 * time.Hour), "in 2 years"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimeAgoFrom(tt.t, now)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFormatTimeAgo(t *testing.T) {
	// Smoke test: FormatTimeAgo delegates to FormatTimeAgoFrom with time.Now().
	got := FormatTimeAgo(time.Now().UTC().Add(-10 * time.Second))
	require.Equal(t, "now", got)
}

func TestFormatTimeAgoCompactFrom(t *testing.T) {
	now := time.Date(2025, 6, 15, 12, 0, 0, 0, time.UTC)

	tests := []struct {
		name string
		t    time.Time
		want string
	}{
		{"just now", now.Add(-10 * time.Second), "now"},
		{"5 minutes", now.Add(-5 * time.Minute), "5m ago"},
		{"90 seconds", now.Add(-90 * time.Second), "1m ago"},
		{"3 hours", now.Add(-3 * time.Hour), "3h ago"},
		{"90 minutes", now.Add(-90 * time.Minute), "1h ago"},
		{"3 days", now.Add(-3 * 24 * time.Hour), "3d ago"},
		{"36 hours", now.Add(-36 * time.Hour), "1d ago"},
		{"14 days", now.Add(-14 * 24 * time.Hour), "2w ago"},
		{"10 days", now.Add(-10 * 24 * time.Hour), "1w ago"},
		{"60 days", now.Add(-60 * 24 * time.Hour), "2mo ago"},
		{"35 days", now.Add(-35 * 24 * time.Hour), "1mo ago"},
		{"400 days", now.Add(-400 * 24 * time.Hour), "1y ago"},
		{"800 days", now.Add(-800 * 24 * time.Hour), "2y ago"},
		// Future.
		{"future 5 minutes", now.Add(5 * time.Minute), "in 5m"},
		{"future 3 hours", now.Add(3 * time.Hour), "in 3h"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatTimeAgoCompactFrom(tt.t, now)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestFormatTimeAgoCompact(t *testing.T) {
	got := FormatTimeAgoCompact(time.Now().UTC().Add(-10 * time.Second))
	require.Equal(t, "now", got)
}
