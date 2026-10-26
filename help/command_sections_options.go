package help

// CommandSectionsOption configures BuildCommandSections.
type CommandSectionsOption func(*commandSectionsConfig)

type commandSectionsConfig struct {
	sorted     bool
	groupOrder []string
}

// WithCommandGroupOrder puts populated native sections first in the supplied order.
func WithCommandGroupOrder(titles ...string) CommandSectionsOption {
	return func(c *commandSectionsConfig) { c.groupOrder = titles }
}

// WithKeepCommandGroupOrder preserves native/first-seen section order (the default).
func WithKeepCommandGroupOrder() CommandSectionsOption {
	return func(c *commandSectionsConfig) { c.sorted = false }
}

// WithSortedCommandGroupOrder sorts named sections alphabetically; ungrouped commands stay last.
func WithSortedCommandGroupOrder() CommandSectionsOption {
	return func(c *commandSectionsConfig) { c.sorted = true }
}
