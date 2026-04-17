package help

// FlagSectionsOption configures BuildFlagSections behavior.
type FlagSectionsOption func(*flagSectionsConfig)

type flagSectionsConfig struct {
	keepGroupOrder        bool
	separateGlobalOptions bool
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
