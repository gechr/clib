package urfave

type sectionsConfig struct {
	rawUsage bool
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
