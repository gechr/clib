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

// WithDescriptionWidth sets the wrap width for [Description] content (e.g.
// the long-form help surfaced by kong's HelpProvider interface). When left
// unset, descriptions inherit the [WithMaxWidth] value so they wrap to the
// same column as flag descriptions. Pass 0 to disable wrapping for
// descriptions specifically while keeping flag/arg wrapping intact.
func WithDescriptionWidth(n int) RendererOption {
	return func(r *Renderer) {
		r.descriptionWidth = n
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
