package urfave

type sectionsConfig struct {
	lowercasePlaceholders bool
	rawUsage              bool
}

// SectionsOption configures urfave help-section generation.
type SectionsOption func(*sectionsConfig)

// WithRawUsage passes cmd.ArgsUsage through to the usage line verbatim instead
// of parsing it into structured Args. Use this for urfave commands whose
// ArgsUsage contains shell metacharacters (pipes, parens, ellipses) that
// clib's arg grammar would otherwise mangle.
func WithRawUsage() SectionsOption {
	return func(c *sectionsConfig) { c.rawUsage = true }
}

// WithPreservePlaceholders keeps placeholders exactly as provided by clib
// metadata. By default, explicit urfave flag placeholders are lowercased for
// consistency with clib's help style.
func WithPreservePlaceholders() SectionsOption {
	return func(c *sectionsConfig) { c.lowercasePlaceholders = false }
}
