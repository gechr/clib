package complete

// Order controls how shell completion candidates are ordered.
type Order string

const (
	// OrderKeep preserves the candidate order for shells that support it.
	OrderKeep Order = "keep"
	// OrderShell uses the shell's normal ordering behavior.
	OrderShell Order = "shell"
)

// Option configures a Generator.
type Option func(*Generator)

// WithShellGenerator overrides a shell client for this generator only.
// Other generators continue to use the registered shell implementations.
func WithShellGenerator(name string, fn ShellFunc) Option {
	return func(g *Generator) {
		if g.shells == nil {
			g.shells = make(map[string]ShellFunc)
		}
		g.shells[name] = fn
	}
}

// WithOrder sets the default completion ordering for flags that do not specify
// an explicit order.
func WithOrder(order Order) Option {
	return func(g *Generator) {
		g.Order = order
	}
}

// WithIncludeHidden offers hidden flags as completions instead of omitting
// them. By default a flag hidden from --help is also withheld from completions;
// enabling this surfaces every hidden flag across the command tree. For a
// single flag, prefer the per-flag complete-hidden opt-in instead.
func WithIncludeHidden() Option {
	return func(g *Generator) {
		g.IncludeHidden = true
	}
}
