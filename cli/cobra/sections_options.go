package cobra

type sectionsConfig struct {
	keepGroupOrder                  bool
	sortGroupOrder                  bool
	hideInheritedFlags              bool
	hideInheritedFlagsOnSubcommands bool
	showInheritedFlagsOnSubcommands bool
	subcommandOptional              bool
	lowercasePlaceholders           bool
	rawUsage                        bool
	optionsTitle                    string
	globalOptionsTitle              string
}

// SectionsOption configures cobra help-section generation.
type SectionsOption func(*sectionsConfig)

// WithKeepGroupOrder preserves first-seen order of grouped flag sections
// instead of sorting them alphabetically. This is the default.
func WithKeepGroupOrder() SectionsOption {
	return func(c *sectionsConfig) {
		c.keepGroupOrder = true
		c.sortGroupOrder = false
	}
}

// WithSortedGroupOrder sorts grouped flag sections alphabetically.
func WithSortedGroupOrder() SectionsOption {
	return func(c *sectionsConfig) {
		c.keepGroupOrder = false
		c.sortGroupOrder = true
	}
}

// WithHideInheritedFlags omits inherited/global flags from help output.
func WithHideInheritedFlags() SectionsOption {
	return func(c *sectionsConfig) {
		c.hideInheritedFlags = true
	}
}

// WithHideInheritedFlagsOnSubcommands omits inherited/global flags from
// subcommand help output while leaving root-command help unchanged. This is
// the default.
func WithHideInheritedFlagsOnSubcommands() SectionsOption {
	return func(c *sectionsConfig) {
		c.hideInheritedFlagsOnSubcommands = true
		c.showInheritedFlagsOnSubcommands = false
	}
}

// WithSubcommandOptional marks the auto-appended subcommand placeholder as
// optional ([<command>] instead of <command>). Use this when the root command
// is genuinely runnable without a subcommand.
func WithSubcommandOptional() SectionsOption {
	return func(c *sectionsConfig) { c.subcommandOptional = true }
}

// WithPreservePlaceholders keeps placeholders exactly as provided by clib
// metadata or pflag usage annotations. By default, explicit Cobra flag
// placeholders are lowercased for consistency with clib's help style.
func WithPreservePlaceholders() SectionsOption {
	return func(c *sectionsConfig) { c.lowercasePlaceholders = false }
}

// WithShowInheritedFlagsOnSubcommands keeps inherited/global flags visible in
// subcommand help output.
func WithShowInheritedFlagsOnSubcommands() SectionsOption {
	return func(c *sectionsConfig) {
		c.hideInheritedFlagsOnSubcommands = false
		c.showInheritedFlagsOnSubcommands = true
	}
}

// WithRawUsage passes cmd.Use through to the usage line verbatim instead of
// parsing it into structured Args. Use this for cobra commands whose Use:
// strings contain shell metacharacters (pipes, parens, ellipses) that clib's
// arg grammar would otherwise mangle.
func WithRawUsage() SectionsOption {
	return func(c *sectionsConfig) { c.rawUsage = true }
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
