package help

// RendererOption configures a Renderer.
type RendererOption func(*Renderer)

// WithFlagPadding sets the padding (in spaces) between a flag and its
// description. Default is 2.
func WithFlagPadding(n int) RendererOption {
	return func(r *Renderer) {
		r.flagPad = n
	}
}

// WithArgumentPadding sets the padding (in spaces) between an argument and its
// description. Default is 2.
func WithArgumentPadding(n int) RendererOption {
	return func(r *Renderer) {
		r.argPad = n
	}
}

// WithCommandPadding sets the padding (in spaces) between a command and its
// description. Default is 1.
func WithCommandPadding(n int) RendererOption {
	return func(r *Renderer) {
		r.cmdPad = n
	}
}

// WithFlagAlign sets the alignment of flag names in flag sections.
func WithFlagAlign(a Alignment) RendererOption {
	return func(r *Renderer) {
		r.flagAlign = a
	}
}

// WithCommandAlign sets the alignment of command names in the Commands section.
func WithCommandAlign(a Alignment) RendererOption {
	return func(r *Renderer) {
		r.cmdAlign = a
	}
}

// WithCommandAlignMode sets whether command names are aligned per section
// (default) or globally across all command sections.
func WithCommandAlignMode(m AlignMode) RendererOption {
	return func(r *Renderer) {
		r.cmdAlignMode = m
	}
}

// WithMaxWidth sets the maximum output width. Descriptions that exceed this
// width are word-wrapped, with continuation lines indented according to the
// configured [WrapStyle]. A value of 0 disables wrapping; by default the
// renderer auto-detects width from the output writer when possible.
func WithMaxWidth(n int) RendererOption {
	return func(r *Renderer) {
		r.maxWidth = n
	}
}

// WithDescriptionIndent sets the extra indent (in columns) applied to
// [Description] content beyond the section's normal content indent. The
// default is 2, which nests a description visually under the preceding
// content (e.g. a Usage line) rather than aligning flush with it. Pass 0
// to align descriptions with regular section content.
func WithDescriptionIndent(n int) RendererOption {
	return func(r *Renderer) {
		r.descriptionIndent = n
	}
}

// WithDescriptionWidth sets a fixed wrap width for [Description] content (e.g.
// the long-form help surfaced by kong's HelpProvider interface). It overrides
// the default flexible [WithDescriptionWidthRange], pinning descriptions to
// exactly n columns. Pass 0 to disable wrapping for descriptions specifically
// while keeping flag/arg wrapping intact.
func WithDescriptionWidth(n int) RendererOption {
	return func(r *Renderer) {
		r.descriptionWidth = n
		// A fixed width overrides the default range; clear it so the fixed
		// path in descriptionWrapAvail takes effect.
		r.descriptionWidthMin = 0
		r.descriptionWidthMax = 0
	}
}

// WithDescriptionWidthRange sets a flexible wrap width for [Description]
// content. Instead of wrapping strictly at one column, the renderer tries
// every width from minWidth to maxWidth and keeps the one whose wrapped
// lines form the most even right edge - so a short word never pokes out
// past an otherwise clean margin just to satisfy an exact width. One width
// is chosen per Description block, so all its paragraphs share the same
// right edge. The upper bound is capped at [WithMaxWidth] when that is set.
//
// A range of 70-100 is the default; call this to widen or narrow it, or call
// [WithDescriptionWidth] to pin descriptions to one fixed column instead.
func WithDescriptionWidthRange(minWidth, maxWidth int) RendererOption {
	return func(r *Renderer) {
		if minWidth > maxWidth {
			minWidth, maxWidth = maxWidth, minWidth
		}
		r.descriptionWidthMin = minWidth
		r.descriptionWidthMax = maxWidth
		// A range overrides any fixed width; reset to auto so the range path
		// in descriptionWrapAvail takes effect.
		r.descriptionWidth = autoDescriptionWidth
	}
}

// WithListIndent sets the leading indent (in columns) applied to list items
// auto-detected in [Description] content, relative to the description's base
// indent. Both unordered ("-", "*", "+") and ordered ("1.", "2)") markers are
// re-indented to this width regardless of how many spaces the author wrote, so
// list indentation stays uniform. The default is 2.
func WithListIndent(n int) RendererOption {
	return func(r *Renderer) {
		r.listIndent = n
	}
}

// WithBacktickStyle sets how backticked tokens in descriptions are styled.
// See [BacktickStyle] for the supported modes. The default is
// [BacktickStyleSmart].
func WithBacktickStyle(s BacktickStyle) RendererOption {
	return func(r *Renderer) {
		r.backtickStyle = s
	}
}

// WithHideDefaults suppresses the " (default: X)" annotation that the
// renderer would otherwise append to non-enum flag descriptions for any flag
// whose [Flag.Default] is set. Per-flag [Flag.HideDefault] is unaffected and
// always wins. Useful when the caller would rather surface defaults in their
// own description text, or in a separate footer.
func WithHideDefaults() RendererOption {
	return func(r *Renderer) {
		r.hideDefaults = true
	}
}

// WithWrapStyle sets how wrapped description continuation lines are indented.
// The default is [WrapBracketAlign], which aligns continuation lines to the
// content after an unclosed '[' on the first line (e.g. for enum value lists).
// Use [WrapBracketBelow] to break before the bracket, or [WrapFlush] for
// uniform indentation to the description column.
func WithWrapStyle(s WrapStyle) RendererOption {
	return func(r *Renderer) {
		r.wrapStyle = s
	}
}
