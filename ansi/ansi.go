package ansi

import (
	"os"
	"strings"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/terminal"
)

const sgrReset = "\x1b[0m"

// HyperlinkFallback controls how hyperlinks render when the output is not a terminal.
type HyperlinkFallback int

const (
	// HyperlinkFallbackExpanded renders "text (url)".
	HyperlinkFallbackExpanded HyperlinkFallback = iota
	// HyperlinkFallbackMarkdown renders "[text](url)".
	HyperlinkFallbackMarkdown
	// HyperlinkFallbackText renders only the display text, discarding the URL.
	HyperlinkFallbackText
	// HyperlinkFallbackURL renders only the URL, discarding the display text.
	HyperlinkFallbackURL
)

// ANSI produces ANSI-aware output, falling back to plain text
// when the output is not a terminal.
type ANSI struct {
	terminal          bool
	hyperlinkFallback HyperlinkFallback
}

// New creates an ANSI with the given options.
func New(opts ...Option) *ANSI {
	w := &ANSI{}
	for _, o := range opts {
		o(w)
	}
	return w
}

// Never creates an ANSI with ANSI output unconditionally disabled.
func Never() *ANSI {
	return &ANSI{}
}

// Force creates an ANSI with ANSI output unconditionally enabled.
func Force() *ANSI {
	return &ANSI{terminal: true}
}

// Auto creates an ANSI that auto-detects whether the output is a terminal.
// All provided files must be terminals for ANSI output to be enabled.
// Defaults to os.Stdout if no files are provided.
func Auto(files ...*os.File) *ANSI {
	if len(files) == 0 {
		files = []*os.File{os.Stdout}
	}
	for _, f := range files {
		if f == nil || !terminal.Is(f) {
			return Never()
		}
	}
	return Force()
}

// Terminal reports whether the output target is a terminal.
func (w *ANSI) Terminal() bool { return w.terminal }

// Hyperlink creates an OSC 8 terminal hyperlink.
// When the output is not a terminal, the HyperlinkFallback mode controls
// how the link is rendered in plain text.
func (w *ANSI) Hyperlink(url, text string) string {
	if !w.terminal {
		switch w.hyperlinkFallback {
		case HyperlinkFallbackExpanded:
			return text + " (" + url + ")"
		case HyperlinkFallbackMarkdown:
			return "[" + text + "](" + url + ")"
		case HyperlinkFallbackText:
			return text
		case HyperlinkFallbackURL:
			return url
		default:
			return text + " (" + url + ")"
		}
	}
	return xansi.SetHyperlink(url) + text + xansi.ResetHyperlink()
}

// PreserveBackground wraps a line with a background escape and re-applies it
// after every embedded SGR sequence so inner ANSI styling does not clear the
// row background.
func PreserveBackground(line, bg string) string {
	var b strings.Builder
	b.WriteString(bg)

	i := 0
	for i < len(line) {
		if line[i] == '\x1b' && i+1 < len(line) && line[i+1] == '[' {
			j := i + 2 //nolint:mnd // skip ESC[
			for j < len(line) && ((line[j] >= '0' && line[j] <= '9') || line[j] == ';') {
				j++
			}
			if j < len(line) && line[j] == 'm' {
				j++
				b.WriteString(line[i:j])
				b.WriteString(bg)
				i = j
				continue
			}
		}
		b.WriteByte(line[i])
		i++
	}

	b.WriteString(sgrReset)
	return b.String()
}

// PreserveBackgroundWidth behaves like PreserveBackground and pads the visible
// line to the requested terminal width.
func PreserveBackgroundWidth(line, bg string, width int) string {
	preserved := PreserveBackground(line, bg)
	if pad := width - xansi.WcWidth.StringWidth(line); pad > 0 {
		preserved = strings.TrimSuffix(preserved, sgrReset) + strings.Repeat(" ", pad) + sgrReset
	}
	return preserved
}
