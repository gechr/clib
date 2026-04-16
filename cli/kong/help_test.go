package kong_test

import (
	"bytes"
	"io"
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/cli/kong"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func TestHelpPrinter(t *testing.T) {
	var buf bytes.Buffer
	renderer := help.NewRenderer(theme.Default())

	var called bool
	printer := kong.HelpPrinter(renderer, func() ([]help.Section, error) {
		called = true
		return []help.Section{
			{
				Title:   "Usage",
				Content: []help.Content{help.Text("myapp [options]")},
			},
			{
				Title: "Flags",
				Content: []help.Content{
					help.FlagGroup{
						{Short: "v", Long: "verbose", Desc: "Verbose output"},
					},
				},
			},
		}, nil
	})

	type CLI struct{}
	k, err := konglib.New(&CLI{},
		konglib.Writers(&buf, io.Discard),
		konglib.Help(printer),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	_, _ = k.Parse([]string{"--help"})

	require.True(t, called)
	require.Equal(
		t,
		"Usage\n\n  myapp [options]\n\nFlags\n\n  -v, --verbose  Verbose output\n",
		buf.String(),
	)
}

func TestHelpPrinter_MultipleSections(t *testing.T) {
	var buf bytes.Buffer
	renderer := help.NewRenderer(theme.Default())

	printer := kong.HelpPrinter(renderer, func() ([]help.Section, error) {
		return []help.Section{
			{
				Title:   "Usage",
				Content: []help.Content{help.Text("app [options]")},
			},
			{
				Title: "Filters",
				Content: []help.Content{
					help.FlagGroup{
						{Short: "a", Long: "author", Placeholder: "user", Desc: "Filter by author"},
						{Short: "s", Long: "state", Placeholder: "state", Desc: "PR state"},
					},
				},
			},
			{
				Title: "Output",
				Content: []help.Content{
					help.FlagGroup{
						{Short: "o", Long: "output", Placeholder: "fmt", Desc: "Output format"},
					},
				},
			},
		}, nil
	})

	type CLI struct{}
	k, err := konglib.New(&CLI{},
		konglib.Writers(&buf, io.Discard),
		konglib.Help(printer),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	_, _ = k.Parse([]string{"--help"})

	expected := `Usage

  app [options]

Filters

  -a, --author <user>  Filter by author
  -s, --state <state>  PR state

Output

  -o, --output <fmt>   Output format
`
	require.Equal(t, expected, buf.String())
}

func TestSections_GroupsByGroup(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "repo", Short: "R", Help: "Limit to repo", Group: "Filters", Placeholder: "repo"},
		{
			Name:        "author",
			Short:       "a",
			Help:        "Filter by author",
			Group:       "Filters",
			Placeholder: "user",
		},
		{Name: "output", Short: "o", Help: "Output format", Group: "Output", Placeholder: "fmt"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 2)
	require.Equal(t, "Filters", sections[0].Title)
	require.Equal(t, "Output", sections[1].Title)

	filters, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, filters, 2)
	require.Equal(t, "repo", filters[0].Long)
	require.Equal(t, "author", filters[1].Long)
}

func TestSections_FirstSeenOrdering(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "output", Help: "Output format", Group: "Output"},
		{Name: "repo", Help: "Limit to repo", Group: "Filters"},
		{Name: "open", Help: "Open in browser", Group: "Actions"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 3)
	require.Equal(t, "Output", sections[0].Title)
	require.Equal(t, "Filters", sections[1].Title)
	require.Equal(t, "Actions", sections[2].Title)
}

func TestSections_UngroupedGoToFlags(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "repo", Help: "Limit to repo", Group: "Filters"},
		{Name: "debug", Help: "Debug mode"},
		{Name: "verbose", Help: "Verbose output"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 2)
	require.Equal(t, "Filters", sections[0].Title)
	require.Equal(t, "Options", sections[1].Title)

	ungrouped, ok := sections[1].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, ungrouped, 2)
	require.Equal(t, "debug", ungrouped[0].Long)
	require.Equal(t, "verbose", ungrouped[1].Long)
}

func TestSections_SkipsHidden(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "repo", Help: "Limit to repo", Group: "Filters"},
		{Name: complete.FlagComplete, Help: "Completion type", Hidden: true},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)
	require.Equal(t, "Filters", sections[0].Title)

	filters, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, filters, 1)
}

func TestSections_SkipsArgs(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "repo", Help: "Limit to repo", Group: "Filters"},
		{Origin: "Query", Help: "Search query", IsArg: true},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	filters, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, filters, 1)
}

func TestSections_NegatableGetsNoPrefix(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "draft", Short: "D", Help: "Filter by draft", Group: "Filters", Negatable: true},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	filters, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, filters, 1)
	require.Equal(t, "[no-]draft", filters[0].Long)
	require.Equal(t, "D", filters[0].Short)
}

func TestSections_PlaceholderPassedThrough(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "output", Short: "o", Help: "Output format", Placeholder: "fmt", Group: "Output"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	output, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, output, 1)
	require.Equal(t, "fmt", output[0].Placeholder)
}

func TestSections_RepeatableSlice(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "label", Help: "Add labels", Placeholder: "tag", IsSlice: true, Group: "Flags"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Equal(t, "tag", fg[0].Placeholder)
	require.True(t, fg[0].Repeatable)
}

func TestSections_RepeatableSlice_PlaceholderOverrideOverride(t *testing.T) {
	flags := []complete.FlagMeta{
		{
			Name:                "label",
			Help:                "Add labels",
			Placeholder:         "custom",
			PlaceholderOverride: true,
			IsSlice:             true,
			Group:               "Flags",
		},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Equal(t, "custom", fg[0].Placeholder)
	require.False(t, fg[0].Repeatable)
}

func TestSections_CSVFlagRepeatableWithPlaceholderOverride(t *testing.T) {
	flags := []complete.FlagMeta{
		{
			Name:                "tags",
			Help:                "Filter by tags",
			HasArg:              true,
			IsCSV:               true,
			Placeholder:         "tag",
			PlaceholderOverride: true,
			Group:               "Flags",
		},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.True(t, fg[0].Repeatable)
}

func TestSections_PrefersHelpOverDescription(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "debug", Help: "Debug mode", Terse: "Debug"},
		{Name: "verbose", Help: "Enable verbose output"},
		{Name: "quiet", Terse: "Suppress output"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	ungrouped, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "Debug mode", ungrouped[0].Desc)
	require.Equal(t, "Enable verbose output", ungrouped[1].Desc)
	require.Equal(t, "Suppress output", ungrouped[2].Desc)
}

func TestSections_EnumPassedThrough(t *testing.T) {
	flags := []complete.FlagMeta{
		{
			Name:          "state",
			Short:         "s",
			Help:          "Filter by state",
			Enum:          []string{"open", "closed", "merged", "all"},
			EnumHighlight: []string{"o", "c", "m", "a"},
			Group:         "Filters",
		},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Equal(t, []string{"open", "closed", "merged", "all"}, fg[0].Enum)
	require.Equal(t, []string{"o", "c", "m", "a"}, fg[0].EnumHighlight)
}

func TestSections_SubGroupsSplitIntoFlagGroups(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "org", Help: "Organization", Group: "Filters/1"},
		{Name: "repo", Short: "R", Help: "Repository", Group: "Filters/1"},
		{Name: "author", Short: "a", Help: "Author", Group: "Filters/2"},
		{Name: "output", Short: "o", Help: "Output format", Group: "Output"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 2)
	require.Equal(t, "Filters", sections[0].Title)
	require.Equal(t, "Output", sections[1].Title)

	// Filters should have 2 content entries (one per sub-group).
	require.Len(t, sections[0].Content, 2)

	fg1, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg1, 2)
	require.Equal(t, "org", fg1[0].Long)
	require.Equal(t, "repo", fg1[1].Long)

	fg2, ok := sections[0].Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg2, 1)
	require.Equal(t, "author", fg2[0].Long)
}

func TestSections_SubGroupsOrdering(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "author", Help: "Author", Group: "Filters/2"},
		{Name: "org", Help: "Organization", Group: "Filters/1"},
		{Name: "state", Help: "State", Group: "Filters/3"},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)
	require.Equal(t, "Filters", sections[0].Title)

	// Sub-groups ordered by first appearance: 2, 1, 3.
	require.Len(t, sections[0].Content, 3)

	fg1, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "author", fg1[0].Long)

	fg2, ok := sections[0].Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "org", fg2[0].Long)

	fg3, ok := sections[0].Content[2].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "state", fg3[0].Long)
}

func TestSections_EnumDefaultPassedThrough(t *testing.T) {
	flags := []complete.FlagMeta{
		{
			Name:        "state",
			Short:       "s",
			Help:        "Filter by state",
			Enum:        []string{"open", "closed", "merged", "all"},
			EnumDefault: "open",
			Group:       "Filters",
		},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Equal(t, "open", fg[0].EnumDefault)
}

func TestSections_Empty(t *testing.T) {
	sections := kong.FlagSections(nil)
	require.Empty(t, sections)
}

func TestSections_AllHidden(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: complete.FlagComplete, Hidden: true},
		{Name: complete.FlagShell, Hidden: true},
	}

	sections := kong.FlagSections(flags)
	require.Empty(t, sections)
}

// parseForHelp creates a kong parser and parses args, capturing the context
// passed to the help printer. Returns the context and the rendered output.
func parseForHelp(t *testing.T, cli any, args []string, opts ...konglib.Option) *konglib.Context {
	t.Helper()

	var captured *konglib.Context
	printer := func(_ konglib.HelpOptions, ctx *konglib.Context) error {
		captured = ctx
		return nil
	}

	defaults := []konglib.Option{
		konglib.Writers(io.Discard, io.Discard),
		konglib.Help(printer),
		konglib.Exit(func(int) {}),
	}
	defaults = append(defaults, opts...)

	k, err := konglib.New(cli, defaults...)
	require.NoError(t, err)

	_, _ = k.Parse(args)
	require.NotNil(t, captured, "help printer was not invoked")
	return captured
}

func TestHelpPrinterFunc(t *testing.T) {
	var buf bytes.Buffer
	renderer := help.NewRenderer(theme.Default())

	var receivedCtx *konglib.Context
	printer := kong.HelpPrinterFunc(renderer, func(ctx *konglib.Context) ([]help.Section, error) {
		receivedCtx = ctx
		return []help.Section{
			{Title: "Usage", Content: []help.Content{help.Text("app [options]")}},
		}, nil
	})

	type CLI struct{}
	k, err := konglib.New(&CLI{},
		konglib.Writers(&buf, io.Discard),
		konglib.Help(printer),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	_, _ = k.Parse([]string{"--help"})
	require.NotNil(t, receivedCtx)
	require.Equal(t, "Usage\n\n  app [options]\n", buf.String())
}

func TestNodeSections_Root(t *testing.T) {
	type CLI struct {
		Run   struct{} `help:"Run the app"       cmd:""`
		Build struct{} `help:"Build the app"     cmd:""`
		Debug bool     `help:"Enable debug mode"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	require.Equal(t, []string{"Usage", "Commands", "Options"}, sectionTitles(sections))

	// Usage should contain the app name.
	usage := findSection(sections, "Usage")
	require.NotNil(t, usage)
	u, ok := usage.Content[0].(help.Usage)
	require.True(t, ok)
	require.Equal(t, "myapp", u.Command)
	require.True(t, u.ShowOptions)

	// Commands should list visible children.
	cmds := findSection(sections, "Commands")
	require.NotNil(t, cmds)
	cg, ok := cmds.Content[0].(help.CommandGroup)
	require.True(t, ok)
	require.Len(t, cg, 2)

	// Flags should include debug.
	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found := false
	for _, f := range fg {
		if f.Long == "debug" {
			found = true
		}
	}
	require.True(t, found, "expected debug flag")
}

func TestNodeSections_Subcommand(t *testing.T) {
	type CLI struct {
		Verbose bool `help:"Verbose output"`
		Run     struct {
			Output string `name:"output" help:"Output format" short:"o"`
		} `help:"Run the app"    cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"run", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	require.Equal(t, []string{"Usage", "Options", "Inherited Options"}, sectionTitles(sections))

	// Usage path should be "myapp run".
	usage := findSection(sections, "Usage")
	u, ok := usage.Content[0].(help.Usage)
	require.True(t, ok)
	require.Equal(t, "myapp run", u.Command)

	// Flags should have output.
	flags := findSection(sections, "Options")
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found := false
	for _, f := range fg {
		if f.Long == "output" && f.Short == "o" {
			found = true
		}
	}
	require.True(t, found, "expected output flag")

	// Inherited Flags should have verbose.
	inherited := findSection(sections, "Inherited Options")
	ifg, ok := inherited.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found = false
	for _, f := range ifg {
		if f.Long == "verbose" {
			found = true
		}
	}
	require.True(t, found, "expected inherited verbose flag")
}

func TestNodeSections_AliasesHiddenByDefault(t *testing.T) {
	type CLI struct {
		Format struct{} `help:"Format code" aliases:"fmt" cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"format", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	require.Nil(t, findSection(sections, "Aliases"))
}

func TestNodeSections_ShowAliasesOption(t *testing.T) {
	type CLI struct {
		Format struct{} `help:"Format code" aliases:"fmt" cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"format", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx, kong.WithShowAliases())
	require.NoError(t, err)

	aliases := findSection(sections, "Aliases")
	require.NotNil(t, aliases)
	text, ok := aliases.Content[0].(help.Text)
	require.True(t, ok)
	require.Equal(t, help.Text("fmt"), text)
}

func TestNodeSections_ShowAliasesTag(t *testing.T) {
	type CLI struct {
		Format struct{} `help:"Format code" aliases:"fmt" cmd:"" show-aliases:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"format", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	aliases := findSection(sections, "Aliases")
	require.NotNil(t, aliases)
	text, ok := aliases.Content[0].(help.Text)
	require.True(t, ok)
	require.Equal(t, help.Text("fmt"), text)
}

func TestNodeSections_PositionalArgs(t *testing.T) {
	type CLI struct {
		Bump struct {
			File string `help:"File to bump" arg:""`
		} `help:"Bump versions" cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"bump", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	usage := findSection(sections, "Usage")
	u, ok := usage.Content[0].(help.Usage)
	require.True(t, ok)
	require.Len(t, u.Args, 1)
	require.Equal(t, "file", u.Args[0].Name)
	require.True(t, u.Args[0].Required)
}

func TestNodeSections_HiddenChildrenSkipped(t *testing.T) {
	type CLI struct {
		Run        struct{} `help:"Run the app"       cmd:""`
		Completion struct{} `help:"Shell completions" hidden:"" cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	cmds := findSection(sections, "Commands")
	require.NotNil(t, cmds)
	cg, ok := cmds.Content[0].(help.CommandGroup)
	require.True(t, ok)
	require.Len(t, cg, 1)
	require.Equal(t, "run", cg[0].Name)
}

func TestNodeSections_HiddenFlagsSkipped(t *testing.T) {
	type CLI struct {
		Debug    bool `help:"Debug mode"`
		Internal bool `help:"Internal flag" hidden:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	for _, f := range fg {
		require.NotEqual(t, "internal", f.Long, "hidden flag should be skipped")
	}
}

func TestNodeSections_Negatable(t *testing.T) {
	type CLI struct {
		Draft *bool `help:"Filter drafts" negatable:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found := false
	for _, f := range fg {
		if f.Long == "[no-]draft" {
			found = true
		}
	}
	require.True(t, found, "expected [no-]draft flag")
}

func TestNodeSections_Short(t *testing.T) {
	type CLI struct {
		Verbose bool `help:"Verbose output" short:"v"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found := false
	for _, f := range fg {
		if f.Long == "verbose" && f.Short == "v" {
			found = true
		}
	}
	require.True(t, found, "expected -v/--verbose flag")
}

func TestNodeSections_CSVFlagRepeatable(t *testing.T) {
	type CLI struct {
		Tags kong.CSVFlag `name:"tags" help:"Filter by tags" placeholder:"<tag>"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found := false
	for _, f := range fg {
		if f.Long == "tags" {
			found = true
			require.True(t, f.Repeatable, "CSVFlag should be repeatable")
		}
	}
	require.True(t, found, "expected tags flag")
}

func TestNodeSections_CSVFlagPtrRepeatable(t *testing.T) {
	type CLI struct {
		Author *kong.CSVFlag `name:"author" help:"Filter by author" placeholder:"<user>"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found := false
	for _, f := range fg {
		if f.Long == "author" {
			found = true
			require.True(t, f.Repeatable, "*CSVFlag should be repeatable")
		}
	}
	require.True(t, found, "expected author flag")
}

func sectionTitles(sections []help.Section) []string {
	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	return titles
}

func findSection(sections []help.Section, title string) *help.Section {
	for i := range sections {
		if sections[i].Title == title {
			return &sections[i]
		}
	}
	return nil
}

func TestNodeSectionsFunc_Basic(t *testing.T) {
	type CLI struct {
		Debug bool `help:"Debug mode"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))

	fn := kong.NodeSectionsFunc()
	sections, err := fn(ctx)
	require.NoError(t, err)

	require.Equal(t, []string{"Usage", "Options"}, sectionTitles(sections))
}

func TestNodeSectionsFunc_WithHideArguments(t *testing.T) {
	type CLI struct {
		Bump struct {
			File string `help:"File to bump" arg:""`
		} `help:"Bump versions" cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"bump", "--help"}, konglib.Name("myapp"))

	fn := kong.NodeSectionsFunc(kong.WithHideArguments())
	sections, err := fn(ctx)
	require.NoError(t, err)

	require.Nil(t, findSection(sections, "Arguments"))
}

func TestNodeSections_WithHideArguments(t *testing.T) {
	type CLI struct {
		Bump struct {
			File string `help:"File to bump" arg:""`
		} `help:"Bump versions" cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"bump", "--help"}, konglib.Name("myapp"))

	// Without hide arguments - should have Arguments section.
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)
	require.NotNil(t, findSection(sections, "Arguments"))

	// With hide arguments - should not have Arguments section.
	sections, err = kong.NodeSections(ctx, kong.WithHideArguments())
	require.NoError(t, err)
	require.Nil(t, findSection(sections, "Arguments"))
}

func TestNodeSections_GroupedFlags(t *testing.T) {
	type CLI struct {
		Author string `name:"author" help:"Filter by author" short:"a" clib:"group='Filters'"`
		State  string `name:"state"  help:"PR state"         short:"s" clib:"group='Filters'"`
		Output string `name:"output" help:"Output format"    short:"o" clib:"group='Output'"`
		Debug  bool   `name:"debug"  help:"Debug mode"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	require.Equal(t, []string{"Usage", "Filters", "Output", "Options"}, sectionTitles(sections))

	// Filters should have 2 flags.
	filters := findSection(sections, "Filters")
	require.NotNil(t, filters)
	fg, ok := filters.Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 2)
	require.Equal(t, "author", fg[0].Long)
	require.Equal(t, "state", fg[1].Long)
}

func TestNodeSections_GroupedSubgroups(t *testing.T) {
	type CLI struct {
		Org    string `name:"org"    help:"Organization"  clib:"group='Filters/1'"`
		Repo   string `name:"repo"   help:"Repository"    short:"R"                clib:"group='Filters/1'"`
		Author string `name:"author" help:"Author"        short:"a"                clib:"group='Filters/2'"`
		Output string `name:"output" help:"Output format" short:"o"                clib:"group='Output'"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	filters := findSection(sections, "Filters")
	require.NotNil(t, filters)
	// 2 sub-groups within Filters.
	require.Len(t, filters.Content, 2)

	fg1, ok := filters.Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg1, 2)
	require.Equal(t, "org", fg1[0].Long)
	require.Equal(t, "repo", fg1[1].Long)

	fg2, ok := filters.Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg2, 1)
	require.Equal(t, "author", fg2[0].Long)
}

func TestNodeSections_GroupedWithUngroupedLocalAndInherited(t *testing.T) {
	type CLI struct {
		Verbose bool `name:"verbose" help:"Verbose"` // inherited, ungrouped
		Run     struct {
			Author string `name:"author" help:"Author" short:"a" clib:"group='Filters'"`
			Debug  bool   `name:"debug"  help:"Debug"` // local, ungrouped
		} `               help:"Run"     cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"run", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	require.Equal(
		t,
		[]string{"Usage", "Filters", "Options", "Inherited Options"},
		sectionTitles(sections),
	)
}

func TestNodeSections_GroupedInheritedTriggersGroupedMode(t *testing.T) {
	type CLI struct {
		Verbose bool `name:"verbose" help:"Verbose" clib:"group='Output'"`
		Run     struct {
			Limit int `name:"limit" help:"Max results" short:"L"`
		} `               help:"Run"                           cmd:""`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"run", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	require.Equal(
		t,
		[]string{"Usage", "Output", "Options", "Inherited Options"},
		sectionTitles(sections),
	)
}

func TestNodeSections_ClibEnumHighlightDefault(t *testing.T) {
	type CLI struct {
		State string `name:"state" help:"PR state" clib:"enum='open,closed,merged',highlight='o,c,m',default='open'"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)

	found := false
	for _, f := range fg {
		if f.Long == "state" {
			found = true
			require.Equal(t, []string{"open", "closed", "merged"}, f.Enum)
			require.Equal(t, []string{"o", "c", "m"}, f.EnumHighlight)
			require.Equal(t, "open", f.EnumDefault)
		}
	}
	require.True(t, found, "expected state flag with clib enum")
}

func TestNodeSections_ClibEnumOverridesKongEnum(t *testing.T) {
	type CLI struct {
		State string `name:"state" help:"PR state" clib:"enum='open,closed,merged,all'" default:"open" enum:"open,closed"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)

	found := false
	for _, f := range fg {
		if f.Long == "state" {
			found = true
			// Clib enum should take precedence over kong enum.
			require.Equal(t, []string{"open", "closed", "merged", "all"}, f.Enum)
		}
	}
	require.True(t, found)
}

func TestNodeSections_KongNativeDefaultFallback(t *testing.T) {
	type CLI struct {
		Color string `name:"color" help:"Color output mode" default:"auto" enum:"auto,always,never"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)

	found := false
	for _, f := range fg {
		if f.Long == "color" {
			found = true
			require.Equal(t, []string{"auto", "always", "never"}, f.Enum)
			require.Equal(t, "auto", f.EnumDefault, "should fall back to kong's native default")
		}
	}
	require.True(t, found, "expected color flag")
}

func TestNodeSections_ClibDefaultOverridesKongDefault(t *testing.T) {
	type CLI struct {
		State string `name:"state" help:"PR state" clib:"default='open'" default:"closed" enum:"open,closed,merged"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)

	found := false
	for _, f := range fg {
		if f.Long == "state" {
			found = true
			require.Equal(t, "open", f.EnumDefault, "clib default should take precedence")
		}
	}
	require.True(t, found, "expected state flag")
}

func TestNodeSections_WithArguments(t *testing.T) {
	type RunCmd struct {
		File string `name:"file" help:"File to process" arg:"" clib:"terse='Target file'"`
	}
	type CLI struct {
		Run RunCmd `help:"Run command" cmd:""`
	}
	cli := &CLI{}
	ctx := parseForHelp(t, cli, []string{"run", "--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx, kong.WithArguments(&cli.Run))
	require.NoError(t, err)
	args := findSection(sections, "Arguments")
	require.NotNil(t, args)
	require.Len(t, args.Content, 1)
	a, ok := args.Content[0].(help.Args)
	require.True(t, ok)
	require.Len(t, a, 1)
	require.Equal(t, "file", a[0].Name)
	require.Equal(t, "Target file", a[0].Desc)
}

func TestArgs_Basic(t *testing.T) {
	type CLI struct {
		File string `name:"file" help:"File to process" arg:"" clib:"terse='Target file'"`
	}
	args, err := kong.Args(&CLI{})
	require.NoError(t, err)
	require.Len(t, args, 1)
	require.Equal(t, "file", args[0].Name)
	require.Equal(t, "Target file", args[0].Desc)
	require.True(t, args[0].Required)
	require.False(t, args[0].Repeatable)
}

func TestArgs_Optional(t *testing.T) {
	type CLI struct {
		Query string `name:"query" help:"Search query" arg:"" optional:""`
	}
	args, err := kong.Args(&CLI{})
	require.NoError(t, err)
	require.Len(t, args, 1)
	require.False(t, args[0].Required)
}

func TestArgs_Slice(t *testing.T) {
	type CLI struct {
		Files []string `name:"files" help:"Files to process" arg:""`
	}
	args, err := kong.Args(&CLI{})
	require.NoError(t, err)
	require.Len(t, args, 1)
	require.True(t, args[0].Repeatable)
}

func TestArgs_NameFallback(t *testing.T) {
	type CLI struct {
		QueryTerm string `help:"Search query" arg:""`
	}
	args, err := kong.Args(&CLI{})
	require.NoError(t, err)
	require.Len(t, args, 1)
	require.Equal(t, "queryterm", args[0].Name)
}

func TestArgs_NoArgs(t *testing.T) {
	type CLI struct {
		Verbose bool `help:"Verbose"`
	}
	args, err := kong.Args(&CLI{})
	require.NoError(t, err)
	require.Empty(t, args)
}

func TestSections_HideLong(t *testing.T) {
	flags := []complete.FlagMeta{
		{
			Name:        "include-pattern",
			Short:       "i",
			Help:        "Filter by regex",
			Placeholder: "regex",
			HideLong:    true,
			Group:       "Filters",
		},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Empty(t, fg[0].Long)
	require.Equal(t, "i", fg[0].Short)
}

func TestSections_HideShort(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "verbose", Short: "v", Help: "Verbose output", HideShort: true},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Equal(t, "verbose", fg[0].Long)
	require.Empty(t, fg[0].Short)
}

func TestNodeSections_HideLong(t *testing.T) {
	type CLI struct {
		Pattern string `name:"include-pattern" help:"Filter by regex" short:"i" clib:"hide-long,group='Filters'"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	filters := findSection(sections, "Filters")
	require.NotNil(t, filters)
	fg, ok := filters.Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Empty(t, fg[0].Long)
	require.Equal(t, "i", fg[0].Short)
}

func TestNodeSections_HideShort(t *testing.T) {
	type CLI struct {
		Verbose bool `name:"verbose" help:"Verbose output" short:"v" clib:"hide-short"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	flags := findSection(sections, "Options")
	require.NotNil(t, flags)
	fg, ok := flags.Content[0].(help.FlagGroup)
	require.True(t, ok)
	found := false
	for _, f := range fg {
		if f.Long == "verbose" {
			found = true
			require.Empty(t, f.Short)
		}
	}
	require.True(t, found)
}

func TestNodeSections_NoIndent(t *testing.T) {
	type CLI struct {
		Pattern string `name:"include-pattern" help:"Filter by regex" short:"i"                        clib:"hide-long,group='Filters'"`
		Include string `name:"include"         help:"Include by name" clib:"no-indent,group='Filters'"`
	}

	ctx := parseForHelp(t, &CLI{}, []string{"--help"}, konglib.Name("myapp"))
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)

	filters := findSection(sections, "Filters")
	require.NotNil(t, filters)
	fg, ok := filters.Content[0].(help.FlagGroup)
	require.True(t, ok)

	for _, f := range fg {
		if f.Long == "include" {
			require.True(t, f.NoIndent, "expected NoIndent on --include")
			return
		}
	}
	t.Fatal("expected include flag")
}

func TestSections_HasArgNoPlaceholder(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "output", Help: "Output format", HasArg: true},
	}

	sections := kong.FlagSections(flags)
	require.Len(t, sections, 1)

	fg, ok := sections[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	// When HasArg is true and no explicit placeholder, defaults to flag name.
	require.Equal(t, "output", fg[0].Placeholder)
}
