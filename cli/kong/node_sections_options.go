package kong

// NodeSectionsOption configures NodeSections behavior.
type NodeSectionsOption func(*nodeSectionsConfig)

type nodeSectionsConfig struct {
	hideArguments bool
	showAliases   bool
	argsCLI       any // when set, use reflected args instead of kong's
}

// WithShowAliases opts into rendering the "Aliases" section. By default
// aliases are hidden - they exist to make commands callable by alternate
// names but are not advertised in help output unless explicitly enabled
// globally with this option, or per-command via the `show-aliases:""`
// struct tag.
func WithShowAliases() NodeSectionsOption {
	return func(c *nodeSectionsConfig) { c.showAliases = true }
}

// WithHideArguments suppresses the "Arguments" section from the output.
func WithHideArguments() NodeSectionsOption {
	return func(c *nodeSectionsConfig) { c.hideArguments = true }
}

// WithArguments uses reflected struct tag metadata for the Arguments section
// instead of kong's parse context. This provides richer descriptions from
// clib tags (e.g. terse, help).
func WithArguments(cli any) NodeSectionsOption {
	return func(c *nodeSectionsConfig) { c.argsCLI = cli }
}
