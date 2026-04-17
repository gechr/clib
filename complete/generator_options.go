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

// WithOrder sets the default completion ordering for flags that do not specify
// an explicit order.
func WithOrder(order Order) Option {
	return func(g *Generator) {
		g.Order = order
	}
}
