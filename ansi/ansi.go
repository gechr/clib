package ansi

import (
	"os"

	xansi "github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/terminal"
)

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
