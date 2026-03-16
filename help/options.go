package help

// Option transforms help sections (composable post-processor).
type Option func([]Section) []Section

// Apply applies options to sections in order.
func Apply(sections []Section, opts ...Option) []Section {
	for _, o := range opts {
		sections = o(sections)
	}
	return sections
}

// WithHelpFlags replaces any combined help flag (Long=="help") with separate
// -h and --help entries. Appends as a new FlagGroup to the last section
// containing flag content. Removes empty FlagGroups/sections left behind.
func WithHelpFlags(shortDesc, longDesc string) Option {
	return func(sections []Section) []Section {
		return SplitHelpFlags(sections, shortDesc, longDesc)
	}
}

// WithFlagDefault appends a "[default: value]" suffix to the description of
// the flag with the given Long name. No-op if value is empty or the flag is
// not found.
func WithFlagDefault(flagLong, value string) Option {
	return func(sections []Section) []Section {
		if value == "" {
			return sections
		}
		patchFlag(sections, flagLong, func(f *Flag) {
			f.Desc += " [default: " + value + "]"
		})
		return sections
	}
}

// WithLongHelp appends sections only when args include --help (not -h).
func WithLongHelp(args []string, sections ...Section) Option {
	return func(s []Section) []Section {
		if IsLongHelp(args) {
			s = append(s, sections...)
		}
		return s
	}
}

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
// width are word-wrapped, with continuation lines indented to the description
// column. A value of 0 disables wrapping; by default the renderer auto-detects
// width from the output writer when possible.
func WithMaxWidth(n int) RendererOption {
	return func(r *Renderer) {
		r.maxWidth = n
	}
}
