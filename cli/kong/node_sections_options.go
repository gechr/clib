package kong

// NodeSectionsOption configures NodeSections behavior.
type NodeSectionsOption func(*nodeSectionsConfig)

type nodeSectionsConfig struct {
	hideArguments  bool
	showAliases    bool
	separateGlobal bool
	globalTitle    string
	optionsTitle   string
	argsCLI        any // when set, use reflected args instead of kong's
}

// WithSeparateGlobalOptions splits inherited (ancestor) flags into their own
// "Global Options" section, below the selected command's local "Options".
// By default both share one "Options" section (local first, then a
// blank-line-separated inherited subgroup).
func WithSeparateGlobalOptions() NodeSectionsOption {
	return func(c *nodeSectionsConfig) { c.separateGlobal = true }
}

// WithSeparateGlobalOptionsName is WithSeparateGlobalOptions with a custom
// section title instead of the default "Global Options".
func WithSeparateGlobalOptionsName(title string) NodeSectionsOption {
	return func(c *nodeSectionsConfig) {
		c.separateGlobal = true
		c.globalTitle = title
	}
}

// WithOptionsTitle sets the section title for local and merged flags instead
// of the default "Options".
func WithOptionsTitle(title string) NodeSectionsOption {
	return func(c *nodeSectionsConfig) { c.optionsTitle = title }
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
