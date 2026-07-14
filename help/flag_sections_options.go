package help

// FlagSectionsOption configures BuildFlagSections behavior.
type FlagSectionsOption func(*flagSectionsConfig)

type flagSectionsConfig struct {
	keepGroupOrder        bool
	separateGlobalOptions bool
	globalOptionsTitle    string
	optionsTitle          string
}

const (
	defaultGlobalOptionsTitle = "Global Options"
	defaultOptionsTitle       = "Options"
)

// globalTitle returns the configured global-options section title, falling
// back to the default.
func (c flagSectionsConfig) globalTitle() string {
	if c.globalOptionsTitle != "" {
		return c.globalOptionsTitle
	}
	return defaultGlobalOptionsTitle
}

// localTitle returns the configured local-options section title, falling
// back to the default.
func (c flagSectionsConfig) localTitle() string {
	if c.optionsTitle != "" {
		return c.optionsTitle
	}
	return defaultOptionsTitle
}

// WithKeepGroupOrder preserves first-seen order of groups instead of sorting
// them alphabetically. Use this when the caller controls insertion order
// and wants it preserved in the output.
func WithKeepGroupOrder() FlagSectionsOption {
	return func(c *flagSectionsConfig) { c.keepGroupOrder = true }
}

// WithSeparateGlobalOptions emits inherited flags under a dedicated
// "Global Options" section instead of the default behavior, which merges
// them into the "Options" section as a blank-line-separated sub-group.
func WithSeparateGlobalOptions() FlagSectionsOption {
	return func(c *flagSectionsConfig) { c.separateGlobalOptions = true }
}

// WithGlobalOptionsTitle separates inherited flags into their own section
// (like WithSeparateGlobalOptions) under a custom title instead of the
// default "Global Options".
func WithGlobalOptionsTitle(title string) FlagSectionsOption {
	return func(c *flagSectionsConfig) {
		c.separateGlobalOptions = true
		c.globalOptionsTitle = title
	}
}

// WithOptionsTitle sets the section title for local and merged flags instead
// of the default "Options".
func WithOptionsTitle(title string) FlagSectionsOption {
	return func(c *flagSectionsConfig) { c.optionsTitle = title }
}
