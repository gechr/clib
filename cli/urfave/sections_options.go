package urfave

type sectionsConfig struct {
	keepGroupOrder        bool
	sortGroupOrder        bool
	hideCommandAliases    bool
	inlineCommandAliases  bool
	lowercasePlaceholders bool
	rawUsage              bool
	optionsTitle          string
	globalOptionsTitle    string
}

// SectionsOption configures urfave help-section generation.
type SectionsOption func(*sectionsConfig)

// WithHideCommandAliases omits command aliases from help output.
func WithHideCommandAliases() SectionsOption {
	return func(c *sectionsConfig) { c.hideCommandAliases = true }
}

// WithInlineCommandAliases keeps alias commands in the Commands section instead of
// placing them in a separate Aliases section.
func WithInlineCommandAliases() SectionsOption {
	return func(c *sectionsConfig) { c.inlineCommandAliases = true }
}

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

// WithOptionsTitle sets the section title for local and merged flags instead
// of the default "Options".
func WithOptionsTitle(title string) SectionsOption {
	return func(c *sectionsConfig) { c.optionsTitle = title }
}

// WithGlobalOptionsTitle separates inherited flags under the given section
// title instead of the default merged-options layout.
func WithGlobalOptionsTitle(title string) SectionsOption {
	return func(c *sectionsConfig) { c.globalOptionsTitle = title }
}

// WithKeepGroupOrder preserves first-seen command and flag group order.
// By default, commands keep first-seen order and flags sort alphabetically.
func WithKeepGroupOrder() SectionsOption {
	return func(c *sectionsConfig) { c.keepGroupOrder = true; c.sortGroupOrder = false }
}

// WithSortedGroupOrder sorts command and flag sections alphabetically.
func WithSortedGroupOrder() SectionsOption {
	return func(c *sectionsConfig) { c.keepGroupOrder = false; c.sortGroupOrder = true }
}
