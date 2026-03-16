package theme

import (
	"regexp"
	"strings"
)

var reInlineCode = regexp.MustCompile("`([^`]+)`")

// RenderMarkdown renders a short markdown string for inline display.
// Handles inline code spans with themed colors.
func (th *Theme) RenderMarkdown(text string) string {
	var sb strings.Builder
	last := 0

	for _, loc := range reInlineCode.FindAllStringIndex(text, -1) {
		if loc[0] > last {
			sb.WriteString(th.MarkdownText.Render(text[last:loc[0]]))
		}
		inner := text[loc[0]+1 : loc[1]-1]
		sb.WriteString(th.MarkdownCode.Render("`" + inner + "`"))
		last = loc[1]
	}

	if last == 0 {
		return th.MarkdownText.Render(text)
	}

	if last < len(text) {
		sb.WriteString(th.MarkdownText.Render(text[last:]))
	}

	return sb.String()
}
