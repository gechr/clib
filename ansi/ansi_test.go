package ansi_test

import (
	"os"
	"testing"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/ansi"
	"github.com/stretchr/testify/require"
)

func TestNew_NoOptions(t *testing.T) {
	w := ansi.New()
	require.False(t, w.Terminal())
}

func TestNew_WithTerminalTrue(t *testing.T) {
	w := ansi.New(ansi.WithTerminal(true))
	require.True(t, w.Terminal())
}

func TestNew_WithTerminalFalse(t *testing.T) {
	w := ansi.New(ansi.WithTerminal(false))
	require.False(t, w.Terminal())
}

func TestNever(t *testing.T) {
	w := ansi.Never()
	require.False(t, w.Terminal())
}

func TestForce(t *testing.T) {
	w := ansi.Force()
	require.True(t, w.Terminal())
}

func TestAuto_DefaultNonTerminal(t *testing.T) {
	// In test environments, stdout is not a terminal.
	w := ansi.Auto()
	require.False(t, w.Terminal())
}

func TestAuto_ExplicitNonTerminal(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "non-terminal")
	require.NoError(t, err)
	defer f.Close()

	w := ansi.Auto(f)
	require.False(t, w.Terminal())
}

func TestAuto_MultipleFiles(t *testing.T) {
	f1, err := os.CreateTemp(t.TempDir(), "a")
	require.NoError(t, err)
	defer f1.Close()

	f2, err := os.CreateTemp(t.TempDir(), "b")
	require.NoError(t, err)
	defer f2.Close()

	w := ansi.Auto(f1, f2)
	require.False(t, w.Terminal())
}

func TestAuto_NilFile(t *testing.T) {
	w := ansi.Auto(nil)
	require.False(t, w.Terminal())
}

func TestAuto_NilAmongFiles(t *testing.T) {
	f, err := os.CreateTemp(t.TempDir(), "valid")
	require.NoError(t, err)
	defer f.Close()

	w := ansi.Auto(f, nil)
	require.False(t, w.Terminal())
}

func TestHyperlink_Never(t *testing.T) {
	w := ansi.Never()
	got := w.Hyperlink("https://example.com", "link")
	// Default fallback is parentheses.
	require.Equal(t, "link (https://example.com)", got)
}

func TestHyperlink_FallbackMarkdown(t *testing.T) {
	w := ansi.New(ansi.WithHyperlinkFallback(ansi.HyperlinkFallbackMarkdown))
	got := w.Hyperlink("https://example.com", "link")
	require.Equal(t, "[link](https://example.com)", got)
}

func TestHyperlink_FallbackExpanded(t *testing.T) {
	w := ansi.New(ansi.WithHyperlinkFallback(ansi.HyperlinkFallbackExpanded))
	got := w.Hyperlink("https://example.com", "link")
	require.Equal(t, "link (https://example.com)", got)
}

func TestHyperlink_FallbackText(t *testing.T) {
	w := ansi.New(ansi.WithHyperlinkFallback(ansi.HyperlinkFallbackText))
	got := w.Hyperlink("https://example.com", "link")
	require.Equal(t, "link", got)
}

func TestHyperlink_FallbackURL(t *testing.T) {
	w := ansi.New(ansi.WithHyperlinkFallback(ansi.HyperlinkFallbackURL))
	got := w.Hyperlink("https://example.com", "link")
	require.Equal(t, "https://example.com", got)
}

func TestHyperlink_FallbackUnknownDefaultsToExpanded(t *testing.T) {
	w := ansi.New(ansi.WithHyperlinkFallback(ansi.HyperlinkFallback(99)))
	got := w.Hyperlink("https://example.com", "link")
	require.Equal(t, "link (https://example.com)", got)
}

func TestHyperlink_Force(t *testing.T) {
	got := ansi.Force().Hyperlink("https://example.com", "link")

	// Should contain OSC 8 sequences.
	expected := xansi.SetHyperlink("https://example.com") + "link" + xansi.ResetHyperlink()
	require.Equal(t, expected, got)

	// Visible text should still be "link".
	require.Equal(t, "link", xansi.Strip(got))
}
