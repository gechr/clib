package help

import "strings"

// Option transforms help sections (composable post-processor).
// Use [OptionFunc] to create an Option from a plain function.
type Option interface {
	apply([]Section) []Section
}

// OptionFunc is a function that implements [Option].
type OptionFunc func([]Section) []Section

func (f OptionFunc) apply(s []Section) []Section { return f(s) }

// Apply applies options to sections in order.
func Apply(sections []Section, opts ...Option) []Section {
	for _, o := range opts {
		sections = o.apply(sections)
	}
	return sections
}

// behaviorOption is a secondary interface that some options implement to
// configure framework-level behavior (e.g. disabling default examples handling).
// It is unexported; framework packages use [ResolvePolicy] instead.
type behaviorOption interface {
	applyPolicy(*Policy)
}

// Policy holds framework-level help configuration resolved from options.
type Policy struct {
	// AlwaysShowExamples disables the default [WithExamplesOnLongHelp] behavior,
	// making examples visible on both -h and --help.
	AlwaysShowExamples bool
}

// ResolvePolicy walks opts and builds a [Policy] from any options
// that implement the internal behaviorOption interface.
func ResolvePolicy(opts ...Option) Policy {
	var b Policy
	for _, o := range opts {
		if bh, ok := o.(behaviorOption); ok {
			bh.applyPolicy(&b)
		}
	}
	return b
}

// alwaysShowExamplesOption implements both [Option] (no-op transform) and
// behaviorOption (sets AlwaysShowExamples).
type alwaysShowExamplesOption struct{}

func (alwaysShowExamplesOption) apply(s []Section) []Section { return s }
func (alwaysShowExamplesOption) applyPolicy(b *Policy) {
	b.AlwaysShowExamples = true
}

// WithAlwaysShowExamples disables the default [WithExamplesOnLongHelp]
// behavior, making examples visible on both -h and --help.
func WithAlwaysShowExamples() Option { return alwaysShowExamplesOption{} }

// WithHelpFlags replaces any combined help flag (Long=="help") with separate
// -h and --help entries. Appends as a new FlagGroup to the last section
// containing flag content. Removes empty FlagGroups/sections left behind.
func WithHelpFlags(shortDesc, longDesc string) Option {
	return OptionFunc(func(sections []Section) []Section {
		return SplitHelpFlags(sections, shortDesc, longDesc)
	})
}

// WithHelpFlagSection moves existing help flags into the named section.
// It preserves their current rendering shape, whether combined or already
// split into separate -h and --help entries. If the section does not exist,
// it is created.
func WithHelpFlagSection(sectionTitle string) Option {
	return OptionFunc(func(sections []Section) []Section {
		return MoveHelpFlagsToSection(sections, sectionTitle)
	})
}

// WithHelpFlagsInSection replaces any combined help flag (Long=="help") with
// separate -h and --help entries, then appends them to the named section.
// When sectionTitle is empty, it uses the last section containing flag content
// and falls back to "Options" if no flag sections exist.
func WithHelpFlagsInSection(sectionTitle, shortDesc, longDesc string) Option {
	return OptionFunc(func(sections []Section) []Section {
		sections = SplitHelpFlags(sections, shortDesc, longDesc)
		return MoveHelpFlagsToSection(sections, sectionTitle)
	})
}

// WithRenamedSection renames any section whose title exactly matches from.
func WithRenamedSection(from, to string) Option {
	return OptionFunc(func(sections []Section) []Section {
		for i := range sections {
			if sections[i].Title == from {
				sections[i].Title = to
			}
		}
		return sections
	})
}

// WithoutSection removes any section whose title exactly matches title.
func WithoutSection(title string) Option {
	return OptionFunc(func(sections []Section) []Section {
		out := make([]Section, 0, len(sections))
		for _, section := range sections {
			if section.Title == title {
				continue
			}
			out = append(out, section)
		}
		return out
	})
}

// WithFlagDefault appends a "[default: value]" suffix to the description of
// the flag with the given Long name. No-op if value is empty or the flag is
// not found.
func WithFlagDefault(flagLong, value string) Option {
	return OptionFunc(func(sections []Section) []Section {
		if value == "" {
			return sections
		}
		patchFlag(sections, flagLong, func(f *Flag) {
			f.Desc += " [default: " + value + "]"
		})
		return sections
	})
}

// WithExamplesOnLongHelp hides the Examples section on short help (-h) and
// moves it to the end on long help (--help), ensuring it is always last
// regardless of option ordering.
func WithExamplesOnLongHelp(args []string) Option {
	return OptionFunc(func(sections []Section) []Section {
		var examples []Section
		out := make([]Section, 0, len(sections))
		for _, s := range sections {
			if strings.EqualFold(s.Title, "examples") {
				examples = append(examples, s)
				continue
			}
			out = append(out, s)
		}
		if IsLongHelp(args) {
			out = append(out, examples...)
		}
		return out
	})
}

// WithLongHelp appends sections only when args include --help (not -h).
func WithLongHelp(args []string, sections ...Section) Option {
	return OptionFunc(func(s []Section) []Section {
		if IsLongHelp(args) {
			s = append(s, sections...)
		}
		return s
	})
}
