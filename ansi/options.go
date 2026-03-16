package ansi

// Option configures an ANSI.
type Option func(*ANSI)

// WithHyperlinkFallback sets how hyperlinks render when the output is not a terminal.
func WithHyperlinkFallback(fallback HyperlinkFallback) Option {
	return func(w *ANSI) {
		w.hyperlinkFallback = fallback
	}
}

// WithTerminal sets whether the output target is a terminal.
func WithTerminal(v bool) Option {
	return func(w *ANSI) {
		w.terminal = v
	}
}
