package kong

import (
	"os"
	"reflect"
	"slices"
	"strings"

	konglib "github.com/alecthomas/kong"
	"github.com/charmbracelet/colorprofile"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/internal/tag"
)

// HelpPrinter returns a kong.HelpPrinter that renders the sections
// returned by the provided callback.
func HelpPrinter(
	r *help.Renderer,
	sections func() []help.Section,
	opts ...help.Option,
) konglib.HelpPrinter {
	return func(_ konglib.HelpOptions, ctx *konglib.Context) error {
		w := colorprofile.NewWriter(ctx.Stdout, os.Environ())
		return r.Render(w, help.Apply(sections(), opts...))
	}
}

// HelpPrinterFunc returns a context-aware kong.HelpPrinter.
// The sections callback receives the kong context, allowing help output
// to vary by subcommand.
func HelpPrinterFunc(
	r *help.Renderer,
	sections func(*konglib.Context) []help.Section,
	opts ...help.Option,
) konglib.HelpPrinter {
	return func(_ konglib.HelpOptions, ctx *konglib.Context) error {
		w := colorprofile.NewWriter(ctx.Stdout, os.Environ())
		return r.Render(w, help.Apply(sections(ctx), opts...))
	}
}

// NodeSectionsOption configures NodeSections behavior.
type NodeSectionsOption func(*nodeSectionsConfig)

type nodeSectionsConfig struct {
	hideArguments bool
	argsCLI       any // when set, use reflected args instead of kong's
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

// NodeSectionsFunc returns a sections callback for use with HelpPrinterFunc,
// with the given options applied.
func NodeSectionsFunc(opts ...NodeSectionsOption) func(*konglib.Context) []help.Section {
	return func(ctx *konglib.Context) []help.Section {
		return NodeSections(ctx, opts...)
	}
}

// NodeSections builds help sections from the kong parse context.
// It determines the active node via ctx.Selected() (falls back to the
// application root) and produces Usage, Arguments, Aliases, Commands,
// and flag sections.
func NodeSections(ctx *konglib.Context, opts ...NodeSectionsOption) []help.Section {
	var cfg nodeSectionsConfig
	for _, o := range opts {
		o(&cfg)
	}

	node := ctx.Selected()
	if node == nil {
		node = ctx.Model.Node
	}

	var sections []help.Section

	// Usage section.
	sections = append(sections, nodeUsageSection(node))

	// Arguments section.
	switch {
	case cfg.argsCLI != nil:
		if args := Args(cfg.argsCLI); len(args) > 0 {
			sections = append(sections, help.Section{
				Title:   "Arguments",
				Content: []help.Content{help.Args(args)},
			})
		}
	case !cfg.hideArguments:
		if args := nodeArgs(node); len(args) > 0 {
			sections = append(sections, help.Section{
				Title:   "Arguments",
				Content: []help.Content{args},
			})
		}
	}

	// Aliases section.
	if len(node.Aliases) > 0 {
		sections = append(sections, help.Section{
			Title:   "Aliases",
			Content: []help.Content{help.Text(strings.Join(node.Aliases, ", "))},
		})
	}

	// Commands section.
	if cmds := visibleChildren(node); len(cmds) > 0 {
		var group help.CommandGroup
		for _, child := range cmds {
			group = append(group, help.Command{Name: child.Name, Desc: child.Help})
		}
		sections = append(sections, help.Section{
			Title:   "Commands",
			Content: []help.Content{group},
		})
	}

	// Flag sections.
	sections = append(sections, buildNodeFlagSections(node)...)

	return sections
}

// nodeUsageSection builds the Usage section for a node.
func nodeUsageSection(node *konglib.Node) help.Section {
	u := help.Usage{
		Command: nodePath(node),
	}

	// Add positional args.
	for _, pos := range node.Positional {
		u.Args = append(u.Args, help.Arg{
			Name:       pos.Name,
			Required:   pos.Required,
			Repeatable: pos.IsCumulative(),
		})
	}

	// Add <command> if the node has visible children.
	if hasVisibleChildren(node) {
		u.Args = append(u.Args, help.Arg{Name: "command", Required: true, IsSubcommand: true})
	}

	// Show [options] if there are any visible flags.
	for _, f := range node.Flags {
		if !f.Hidden {
			u.ShowOptions = true
			break
		}
	}
	// Also check ancestor flags.
	if !u.ShowOptions && node.Parent != nil {
		for _, f := range ancestorFlags(node) {
			if !f.Hidden {
				u.ShowOptions = true
				break
			}
		}
	}

	return help.Section{
		Title:   "Usage",
		Content: []help.Content{u},
	}
}

// nodePath walks parents to build the command path (e.g. "app run"),
// without including aliases (unlike kong's FullPath).
func nodePath(node *konglib.Node) string {
	var parts []string
	for n := node; n != nil; n = n.Parent {
		if n.Type == konglib.ApplicationNode || n.Type == konglib.CommandNode {
			parts = append(parts, n.Name)
		}
	}
	slices.Reverse(parts)
	return strings.Join(parts, " ")
}

// visibleChildren returns non-hidden children of the node.
func visibleChildren(node *konglib.Node) []*konglib.Node {
	var children []*konglib.Node
	for _, child := range node.Children {
		if !child.Hidden {
			children = append(children, child)
		}
	}
	return children
}

// nodeArgs builds help.Args from positional arguments that have help text.
// Returns nil if no positional arg has a description.
func nodeArgs(node *konglib.Node) help.Args {
	hasHelp := false
	for _, pos := range node.Positional {
		if pos.Help != "" {
			hasHelp = true
			break
		}
	}
	if !hasHelp {
		return nil
	}
	var args help.Args
	for _, pos := range node.Positional {
		args = append(args, help.Arg{
			Name:       pos.Name,
			Desc:       pos.Help,
			Required:   pos.Required,
			Repeatable: pos.IsCumulative(),
		})
	}
	return args
}

// hasVisibleChildren returns true if the node has any non-hidden children.
func hasVisibleChildren(node *konglib.Node) bool {
	for _, child := range node.Children {
		if !child.Hidden {
			return true
		}
	}
	return false
}

// kongFlagToHelp converts a kong flag to a help.Flag.
func kongFlagToHelp(f *konglib.Flag) help.Flag {
	meta := complete.FlagMeta{
		Name:                f.Name,
		Help:                f.Help,
		HasArg:              !f.IsBool() && !f.IsCounter(),
		IsCSV:               isCSVFlag(f),
		IsSlice:             f.IsSlice() || f.IsCumulative(),
		PlaceholderOverride: f.PlaceHolder != "",
	}
	if f.Tag != nil && f.Tag.Negatable != "" {
		meta.Negatable = true
	}

	if f.Short != 0 {
		meta.Short = string(f.Short)
	}

	placeholder, placeholderLiteral := flagPlaceholder(f)
	meta.Placeholder = placeholder

	// Thread enum values from tags.
	// Prefer clib enum over kong enum (display-only, no kong validation).
	var clibTag string
	if f.Tag != nil {
		clibTag = f.Tag.Get(tagClib)
	}
	if clibTag != "" {
		if clibEnum, ok := tag.Parse(clibTag, tag.Enum); ok && clibEnum != "" {
			meta.Enum = tag.SplitCSV(clibEnum)
		}
		if hl, ok := tag.Parse(clibTag, tag.Highlight); ok && hl != "" {
			meta.EnumHighlight = tag.SplitCSV(hl)
		}
		if def, ok := tag.Parse(clibTag, tag.Default); ok && def != "" {
			meta.EnumDefault = def
		}
		if inv, ok := tag.Parse(clibTag, tag.Inverse); ok && inv != "" {
			meta.InversePrefix = inv
		}
	}
	if len(meta.Enum) == 0 && f.Tag != nil {
		if enum := f.Tag.Get(tagEnum); enum != "" {
			meta.Enum = tag.SplitCSV(enum)
		}
	}

	// Fall back to kong's native default for enum highlighting.
	if meta.EnumDefault == "" && len(meta.Enum) > 0 && f.HasDefault {
		meta.EnumDefault = f.Default
	}

	hf := helpFlagFromMeta(meta)
	hf.PlaceholderLiteral = placeholderLiteral
	return hf
}

// flagPlaceholder returns the placeholder string and whether it's a literal
// (i.e. should not be wrapped in <...> by the renderer).
func flagPlaceholder(f *konglib.Flag) (string, bool) {
	if f.IsBool() || f.IsCounter() {
		return "", false
	}
	if f.PlaceHolder == "" {
		return f.Name, false
	}
	// Explicit placeholder from the struct tag. If it has angle brackets,
	// strip them - the renderer adds its own. Otherwise it's a literal
	// example value (e.g. "1w2d3h4m") rendered as-is without <...>.
	ph := strings.TrimPrefix(f.PlaceHolder, help.ArgOpen)
	ph = strings.TrimSuffix(ph, help.ArgClose)
	return ph, ph == f.PlaceHolder
}

// isCSVFlag reports whether the kong flag's target type is CSVFlag or *CSVFlag.
func isCSVFlag(f *konglib.Flag) bool {
	csvType := reflect.TypeFor[CSVFlag]()
	t := f.Target.Type()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t == csvType
}

// clibGroup reads the clib tag from a kong flag and returns the group value.
func clibGroup(f *konglib.Flag) string {
	if f.Tag == nil {
		return ""
	}
	clibTag := f.Tag.Get(tagClib)
	if clibTag == "" {
		return ""
	}
	group, _ := tag.Parse(clibTag, tag.Group)
	return group
}

// ancestorFlags collects flags from all ancestor nodes.
func ancestorFlags(node *konglib.Node) []*konglib.Flag {
	var flags []*konglib.Flag
	for p := node.Parent; p != nil; p = p.Parent {
		flags = append(flags, p.Flags...)
	}
	return flags
}

// buildNodeFlagSections builds flag sections from a kong node:
//   - If any flag has a clib group, split into group sections (sorted alphabetically),
//     plus "Options" (ungrouped local) and "Inherited Options" (ungrouped inherited).
//   - Compound group names ("Section/SubGroup") split flags within the same
//     section into separate FlagGroup entries (blank line separator).
//   - Otherwise: flat "Options" (local) + "Inherited Options" (ancestor).
func buildNodeFlagSections(node *konglib.Node) []help.Section {
	inherited := ancestorFlags(node)

	var classified []help.ClassifiedFlag
	classifyKongFlags := func(flags []*konglib.Flag, isInherited bool) {
		for _, f := range flags {
			if f.Hidden {
				continue
			}
			classified = append(classified, help.ClassifiedFlag{
				Flag:      kongFlagToHelp(f),
				Group:     clibGroup(f),
				Inherited: isInherited,
			})
		}
	}
	classifyKongFlags(node.Flags, false)
	classifyKongFlags(inherited, true)

	return help.BuildFlagSections(classified, help.KeepGroupOrder())
}

// Args extracts positional argument entries from a CLI struct's reflected
// metadata. It returns entries suitable for use in help.Args content.
func Args(cli any) []help.Arg {
	var args []help.Arg
	for _, f := range Reflect(cli) {
		if f.IsArg {
			name := f.Name
			if name == "" {
				name = strings.ToLower(f.Origin)
			}
			args = append(args, help.Arg{
				Name:       name,
				Desc:       f.Desc(),
				Required:   !f.Optional,
				Repeatable: f.IsSlice,
			})
		}
	}
	return args
}

// FlagSections builds flag help sections from reflected FlagMeta.
// Flags are grouped by their Group field (from clib:"group='...'").
// Compound group names ("Section/SubGroup") split flags within the same
// section into separate FlagGroup entries (blank line separator).
// Sections appear in first-seen order, with ungrouped flags in an "Options" section.
// Hidden flags and positional args are skipped.
func FlagSections(flags []complete.FlagMeta) []help.Section {
	var classified []help.ClassifiedFlag
	for _, f := range flags {
		if f.Hidden || f.IsArg {
			continue
		}
		classified = append(classified, help.ClassifiedFlag{
			Flag:  helpFlagFromMeta(f),
			Group: f.Group,
			// All flags from FlagMeta are local (no inherited concept).
		})
	}
	return help.BuildFlagSections(classified, help.KeepGroupOrder())
}

func helpFlagFromMeta(f complete.FlagMeta) help.Flag {
	long := f.Name
	if f.Negatable {
		prefix := f.InversePrefix
		if prefix == "" {
			prefix = "no-"
		}
		long = "[" + prefix + "]" + long
	}
	placeholder := f.Placeholder
	if placeholder == "" && f.HasArg {
		placeholder = f.Name
	}
	// Help is for usage output; Terse is for completions.
	desc := f.Help
	if desc == "" {
		desc = f.Terse
	}
	repeatable := placeholder != "" && (f.IsCSV || (f.IsSlice && !f.PlaceholderOverride))
	return help.Flag{
		Short:         f.Short,
		Long:          long,
		Desc:          desc,
		Enum:          f.Enum,
		EnumDefault:   f.EnumDefault,
		EnumHighlight: f.EnumHighlight,
		Placeholder:   placeholder,
		Repeatable:    repeatable,
	}
}
