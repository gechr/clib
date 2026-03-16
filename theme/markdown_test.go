package theme_test

import (
	"testing"

	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestRenderMarkdownInline_NoCode(t *testing.T) {
	th := theme.Default()
	got := th.RenderMarkdown("plain text")
	plain := ansi.Strip(got)
	require.Equal(t, "plain text", plain)
}

func TestRenderMarkdownInline_WithCode(t *testing.T) {
	th := theme.Default()
	got := th.RenderMarkdown("run `go test` now")
	plain := ansi.Strip(got)
	require.Equal(t, "run `go test` now", plain)

	// Should contain styled segments (non-plain).
	require.NotEqual(t, "run `go test` now", got)
}

func TestRenderMarkdownInline_OnlyCode(t *testing.T) {
	th := theme.Default()
	got := th.RenderMarkdown("`code`")
	plain := ansi.Strip(got)
	require.Equal(t, "`code`", plain)
}

func TestRenderMarkdownInline_MultipleCode(t *testing.T) {
	th := theme.Default()
	got := th.RenderMarkdown("`a` and `b`")
	plain := ansi.Strip(got)
	require.Equal(t, "`a` and `b`", plain)
}
