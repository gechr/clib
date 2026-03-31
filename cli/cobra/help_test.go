package cobra_test

import (
	"bytes"
	"testing"

	"github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestSections_Usage(t *testing.T) {
	cmd := &cobralib.Command{Use: "app <file>"}
	cmd.Flags().StringP("output", "o", "", "Output format")

	sections := cobra.Sections(cmd)

	require.NotEmpty(t, sections)
	require.Equal(t, "Usage", sections[0].Title)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Equal(t, "app", usage.Command)
	require.True(t, usage.ShowOptions)
	require.Len(t, usage.Args, 1)
	require.Equal(t, "file", usage.Args[0].Name)
	require.True(t, usage.Args[0].Required)
}

func TestSections_Usage_OptionalArg(t *testing.T) {
	cmd := &cobralib.Command{Use: "search [query]"}

	sections := cobra.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Len(t, usage.Args, 1)
	require.Equal(t, "query", usage.Args[0].Name)
	require.False(t, usage.Args[0].Required)
}

func TestSections_Usage_RepeatableArg(t *testing.T) {
	cmd := &cobralib.Command{Use: "cp <src>... <dst>"}

	sections := cobra.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Len(t, usage.Args, 2)
	require.Equal(t, "src", usage.Args[0].Name)
	require.True(t, usage.Args[0].Required)
	require.True(t, usage.Args[0].Repeatable)
	require.Equal(t, "dst", usage.Args[1].Name)
	require.True(t, usage.Args[1].Required)
	require.False(t, usage.Args[1].Repeatable)
}

func TestSections_Usage_NoFlags(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}

	sections := cobra.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.False(t, usage.ShowOptions)
}

func TestSections_Aliases(t *testing.T) {
	cmd := &cobralib.Command{
		Use:     "app",
		Aliases: []string{"a", "application"},
	}

	sections := cobra.Sections(cmd)

	var aliasSection *help.Section
	for i := range sections {
		if sections[i].Title == "Aliases" {
			aliasSection = &sections[i]
			break
		}
	}
	require.NotNil(t, aliasSection)

	text, ok := aliasSection.Content[0].(help.Text)
	require.True(t, ok)
	require.Equal(t, "a, application", string(text))
}

func TestSections_Examples(t *testing.T) {
	cmd := &cobralib.Command{
		Use:     "app",
		Example: "# Run the app\n$ app run\n\n# Build the app\n$ app build",
	}

	sections := cobra.Sections(cmd)

	var exSection *help.Section
	for i := range sections {
		if sections[i].Title == "Examples" {
			exSection = &sections[i]
			break
		}
	}
	require.NotNil(t, exSection)

	examples, ok := exSection.Content[0].(help.Examples)
	require.True(t, ok)
	require.Len(t, examples, 2)
	require.Equal(t, "Run the app", examples[0].Comment)
	require.Equal(t, "app run", examples[0].Command)
	require.Equal(t, "Build the app", examples[1].Comment)
	require.Equal(t, "app build", examples[1].Command)
}

func TestSections_Flags(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringP("output", "o", "", "Output format")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, flags, 2)
}

func TestSections_Flags_Placeholder(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("name", "", "Your name")
	cmd.Flags().Bool("verbose", false, "Verbose output")

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)

	var nameFlag, verboseFlag *help.Flag
	for i := range flags {
		switch flags[i].Long {
		case "name":
			nameFlag = &flags[i]
		case "verbose":
			verboseFlag = &flags[i]
		}
	}
	require.NotNil(t, nameFlag)
	require.Equal(t, "name", nameFlag.Placeholder)
	require.NotNil(t, verboseFlag)
	require.Empty(t, verboseFlag.Placeholder)
}

func TestSections_Subcommands(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.AddCommand(
		&cobralib.Command{Use: "run", Short: "Run the app", RunE: noop},
		&cobralib.Command{Use: "build", Short: "Build the app", RunE: noop},
	)

	sections := cobra.Sections(root)

	var cmdSection *help.Section
	for i := range sections {
		if sections[i].Title == "Commands" {
			cmdSection = &sections[i]
			break
		}
	}
	require.NotNil(t, cmdSection)
	require.Len(t, cmdSection.Content, 1)

	cmds, ok := cmdSection.Content[0].(help.CommandGroup)
	require.True(t, ok)
	require.Len(t, cmds, 2)
	require.Equal(t, "build", cmds[0].Name)
	require.Equal(t, "run", cmds[1].Name)
}

func TestSections_SubcommandGroups(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.AddGroup(&cobralib.Group{ID: "core", Title: "Core Commands"})
	root.AddCommand(
		&cobralib.Command{Use: "run", Short: "Run the app", GroupID: "core", RunE: noop},
		&cobralib.Command{Use: "extra", Short: "Extra stuff", RunE: noop},
	)

	sections := cobra.Sections(root)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Core Commands", "Additional Commands"}, titles)
}

func TestSections_InheritedFlags(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.PersistentFlags().Bool("debug", false, "Debug mode")

	sub := &cobralib.Command{Use: "run", RunE: noop}
	sub.Flags().StringP("output", "o", "", "Output format")
	root.AddCommand(sub)

	sections := cobra.Sections(sub)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Options"}, titles)
}

func TestSections_NoAliases(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}

	sections := cobra.Sections(cmd)

	for _, s := range sections {
		require.NotEqual(t, "Aliases", s.Title)
	}
}

func TestSections_NoExamples(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}

	sections := cobra.Sections(cmd)

	for _, s := range sections {
		require.NotEqual(t, "Examples", s.Title)
	}
}

func TestSections_GroupedFlags(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringP("output", "o", "", "Output format")
	cmd.Flags().IntP("limit", "L", 30, "Maximum results")
	cmd.Flags().StringP("repo", "R", "", "Filter by repo")
	cmd.Flags().StringP("author", "a", "", "Filter by author")

	cobra.Extend(cmd.Flags().Lookup("repo"), cobra.FlagExtra{Group: "Filters"})
	cobra.Extend(cmd.Flags().Lookup("author"), cobra.FlagExtra{Group: "Filters"})
	cobra.Extend(cmd.Flags().Lookup("output"), cobra.FlagExtra{Group: "Output"})
	cobra.Extend(cmd.Flags().Lookup("limit"), cobra.FlagExtra{Group: "Output"})

	sections := cobra.Sections(cmd)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Output", "Filters"}, titles)

	// Check Filters section has 2 flags.
	for _, s := range sections {
		if s.Title == "Filters" {
			flags, ok := s.Content[0].(help.FlagGroup)
			require.True(t, ok)
			require.Len(t, flags, 2)
		}
	}
}

func TestSectionsWithOptions_KeepGroupOrder(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("output", "", "Output format")
	cmd.Flags().String("repo", "", "Filter by repo")
	cmd.Flags().String("debug", "", "Debug mode")

	cobra.Extend(cmd.Flags().Lookup("output"), cobra.FlagExtra{Group: "Output"})
	cobra.Extend(cmd.Flags().Lookup("repo"), cobra.FlagExtra{Group: "Filters"})
	cobra.Extend(cmd.Flags().Lookup("debug"), cobra.FlagExtra{Group: "Miscellaneous"})

	sections := cobra.SectionsWithOptions(cobra.WithKeepGroupOrder())(cmd)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Output", "Filters", "Miscellaneous"}, titles)
}

func TestSectionsWithOptions_SortedGroupOrder(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("output", "", "Output format")
	cmd.Flags().String("repo", "", "Filter by repo")
	cmd.Flags().String("debug", "", "Debug mode")

	cobra.Extend(cmd.Flags().Lookup("output"), cobra.FlagExtra{Group: "Output"})
	cobra.Extend(cmd.Flags().Lookup("repo"), cobra.FlagExtra{Group: "Filters"})
	cobra.Extend(cmd.Flags().Lookup("debug"), cobra.FlagExtra{Group: "Miscellaneous"})

	sections := cobra.SectionsWithOptions(cobra.WithSortedGroupOrder())(cmd)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Filters", "Miscellaneous", "Output"}, titles)
}

func TestSections_Placeholder_Annotation(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringP("repo", "R", "", "Filter by repo")
	cobra.Extend(cmd.Flags().Lookup("repo"), cobra.FlagExtra{Placeholder: "owner/repo"})

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, flags, 1)
	require.Equal(t, "owner/repo", flags[0].Placeholder)
}

func TestSections_Flags_SlicePlaceholder(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringSlice("label", nil, "Add labels")
	cmd.Flags().StringSlice("assignee", nil, "Add assignees")
	cobra.Extend(cmd.Flags().Lookup("assignee"), cobra.FlagExtra{Placeholder: "user"})

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)

	var labelFlag, assigneeFlag *help.Flag
	for i := range flags {
		switch flags[i].Long {
		case "label":
			labelFlag = &flags[i]
		case "assignee":
			assigneeFlag = &flags[i]
		}
	}

	// Slice without explicit placeholder: repeatable with flag name.
	require.NotNil(t, labelFlag)
	require.Equal(t, "label", labelFlag.Placeholder)
	require.True(t, labelFlag.Repeatable)

	// Slice with explicit clib placeholder: still repeatable.
	require.NotNil(t, assigneeFlag)
	require.Equal(t, "user", assigneeFlag.Placeholder)
	require.True(t, assigneeFlag.Repeatable)
}

func TestSections_MixedGroupedUngrouped(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.PersistentFlags().Bool("debug", false, "Debug mode")

	sub := &cobralib.Command{Use: "run", RunE: noop}
	sub.Flags().StringP("repo", "R", "", "Filter by repo")
	sub.Flags().BoolP("verbose", "v", false, "Verbose output")
	cobra.Extend(sub.Flags().Lookup("repo"), cobra.FlagExtra{Group: "Filters"})
	root.AddCommand(sub)

	sections := cobra.Sections(sub)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}

	require.Equal(t, []string{"Usage", "Filters", "Options"}, titles)
}

func TestSectionsWithOptions_HideInheritedFlags(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.PersistentFlags().Bool("debug", false, "Debug mode")

	sub := &cobralib.Command{Use: "run", RunE: noop}
	sub.Flags().String("output", "", "Output format")
	root.AddCommand(sub)

	sections := cobra.SectionsWithOptions(cobra.WithHideInheritedFlags())(sub)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Options"}, titles)
}

func TestSectionsWithOptions_HideInheritedFlagsOnSubcommands(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.PersistentFlags().Bool("debug", false, "Debug mode")
	root.Flags().String("output", "", "Output format")

	sub := &cobralib.Command{Use: "run", RunE: noop}
	sub.Flags().String("repo", "", "Filter by repo")
	root.AddCommand(sub)

	rootSections := cobra.SectionsWithOptions(cobra.WithHideInheritedFlagsOnSubcommands())(root)
	subSections := cobra.SectionsWithOptions(cobra.WithHideInheritedFlagsOnSubcommands())(sub)

	var rootTitles []string
	for _, s := range rootSections {
		rootTitles = append(rootTitles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Commands", "Options"}, rootTitles)

	var subTitles []string
	for _, s := range subSections {
		subTitles = append(subTitles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Options"}, subTitles)
}

func TestSectionsWithOptions_ShowInheritedFlagsOnSubcommands(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.PersistentFlags().Bool("debug", false, "Debug mode")

	sub := &cobralib.Command{Use: "run", RunE: noop}
	sub.Flags().String("output", "", "Output format")
	root.AddCommand(sub)

	sections := cobra.SectionsWithOptions(cobra.WithShowInheritedFlagsOnSubcommands())(sub)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Options", "Inherited Options"}, titles)
}

func TestHelpFunc_UsageReflectsPostProcessedFlags(t *testing.T) {
	r := help.NewRenderer(theme.Default())
	cmd := &cobralib.Command{Use: "app"}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	helpFn := cobra.HelpFunc(
		r,
		func(_ *cobralib.Command) []help.Section {
			return []help.Section{
				{
					Title: "Usage",
					Content: []help.Content{
						help.Usage{Command: "app"},
					},
				},
			}
		},
		help.WithHelpFlagsInSection("Miscellaneous", "Show help", "Show detailed help"),
	)
	helpFn(cmd, nil)

	require.Contains(t, buf.String(), "app [options]")
	require.Contains(t, buf.String(), "Miscellaneous")
}

func TestHelpFunc_RendersOutput(t *testing.T) {
	r := help.NewRenderer(theme.Default())
	cmd := &cobralib.Command{Use: "app"}
	var buf bytes.Buffer
	cmd.SetOut(&buf)

	helpFn := cobra.HelpFunc(r, func(_ *cobralib.Command) []help.Section {
		return []help.Section{
			{Title: "Test", Content: []help.Content{help.Text("hello")}},
		}
	})
	helpFn(cmd, nil)
	require.Equal(t, "Test\n\n  hello\n", buf.String())
}

func TestSections_Usage_FiltersFlagsAndCommandTokens(t *testing.T) {
	// Use string contains [flags] and [command] which should be filtered out.
	cmd := &cobralib.Command{Use: "app [flags] <file> [command]"}
	cmd.Flags().Bool("verbose", false, "Verbose")

	sections := cobra.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	// Only <file> should remain; [flags] and [command] are filtered.
	require.Len(t, usage.Args, 1)
	require.Equal(t, "file", usage.Args[0].Name)
}

func TestSections_Usage_CommandOnly(t *testing.T) {
	// Use string with no args at all.
	cmd := &cobralib.Command{Use: "app"}

	sections := cobra.Sections(cmd)
	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Empty(t, usage.Args)
}

func TestSections_Flags_AllHidden(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("secret", "", "Secret")
	_ = cmd.Flags().MarkHidden("secret")

	sections := cobra.Sections(cmd)
	// No Flags section should be present.
	for _, s := range sections {
		require.NotEqual(t, "Flags", s.Title)
	}
}

func TestSections_Flags_EnumAnnotation(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("state", "open", "PR state")
	cobra.Extend(
		cmd.Flags().Lookup("state"),
		cobra.FlagExtra{Enum: []string{"open", "closed", "merged", "all"}},
	)

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)

	var stateFlag *help.Flag
	for i := range flags {
		if flags[i].Long == "state" {
			stateFlag = &flags[i]
			break
		}
	}
	require.NotNil(t, stateFlag)
	require.Equal(t, []string{"open", "closed", "merged", "all"}, stateFlag.Enum)
}

func TestSections_Flags_EnumHighlightAnnotation(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("state", "open", "PR state")
	cobra.Extend(cmd.Flags().Lookup("state"), cobra.FlagExtra{
		Enum:          []string{"open", "closed", "merged", "all"},
		EnumHighlight: []string{"o", "c", "m", "a"},
	})

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)

	var stateFlag *help.Flag
	for i := range flags {
		if flags[i].Long == "state" {
			stateFlag = &flags[i]
			break
		}
	}
	require.NotNil(t, stateFlag)
	require.Equal(t, []string{"open", "closed", "merged", "all"}, stateFlag.Enum)
	require.Equal(t, []string{"o", "c", "m", "a"}, stateFlag.EnumHighlight)
}

func TestSections_SubcommandGroups_EmptyGroup(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.AddGroup(
		&cobralib.Group{ID: "core", Title: "Core Commands"},
		&cobralib.Group{ID: "empty", Title: "Empty Group"},
	)
	root.AddCommand(
		&cobralib.Command{Use: "run", Short: "Run", GroupID: "core", RunE: noop},
	)

	sections := cobra.Sections(root)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Core Commands"}, titles)
}

func TestSections_GroupedFlags_WithHiddenFlag(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringP("repo", "R", "", "Filter by repo")
	cmd.Flags().String("secret", "", "Hidden flag")
	cobra.Extend(cmd.Flags().Lookup("repo"), cobra.FlagExtra{Group: "Filters"})
	_ = cmd.Flags().MarkHidden("secret")

	sections := cobra.Sections(cmd)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Filters"}, titles)
}

func TestSections_Flags_EnumDefaultFromDefValue(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("state", "open", "PR state")
	cobra.Extend(cmd.Flags().Lookup("state"), cobra.FlagExtra{
		Enum: []string{"open", "closed", "merged", "all"},
	})

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)

	var stateFlag *help.Flag
	for i := range flags {
		if flags[i].Long == "state" {
			stateFlag = &flags[i]
			break
		}
	}
	require.NotNil(t, stateFlag)
	require.Equal(t, "open", stateFlag.EnumDefault, "should fall back to pflag DefValue")
}

func TestSections_Flags_EnumDefaultExtraOverridesDefValue(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("state", "closed", "PR state")
	cobra.Extend(cmd.Flags().Lookup("state"), cobra.FlagExtra{
		Enum:        []string{"open", "closed", "merged", "all"},
		EnumDefault: "open",
	})

	sections := cobra.Sections(cmd)

	var flagSection *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			flagSection = &sections[i]
			break
		}
	}
	require.NotNil(t, flagSection)

	flags, ok := flagSection.Content[0].(help.FlagGroup)
	require.True(t, ok)

	var stateFlag *help.Flag
	for i := range flags {
		if flags[i].Long == "state" {
			stateFlag = &flags[i]
			break
		}
	}
	require.NotNil(t, stateFlag)
	require.Equal(t, "open", stateFlag.EnumDefault, "FlagExtra.EnumDefault should take precedence")
}

func TestSections_Subcommand_RequiredByDefault(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app", RunE: noop}
	root.AddCommand(&cobralib.Command{Use: "sub", RunE: noop})

	sections := cobra.Sections(root)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)

	var subArg *help.Arg
	for i := range usage.Args {
		if usage.Args[i].IsSubcommand {
			subArg = &usage.Args[i]
			break
		}
	}
	require.NotNil(t, subArg)
	require.True(
		t,
		subArg.Required,
		"subcommand arg should be required by default even when root is runnable",
	)
}

func TestSectionsWithOptions_OptionalSubcommand(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app", RunE: noop}
	root.AddCommand(&cobralib.Command{Use: "sub", RunE: noop})

	sections := cobra.SectionsWithOptions(cobra.WithSubcommandOptional())(root)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)

	var subArg *help.Arg
	for i := range usage.Args {
		if usage.Args[i].IsSubcommand {
			subArg = &usage.Args[i]
			break
		}
	}
	require.NotNil(t, subArg)
	require.False(
		t,
		subArg.Required,
		"subcommand arg should be optional with WithSubcommandOptional()",
	)
}

func TestSections_SubcommandGroups_AllGrouped(t *testing.T) {
	noop := func(*cobralib.Command, []string) error { return nil }

	root := &cobralib.Command{Use: "app"}
	root.AddGroup(&cobralib.Group{ID: "core", Title: "Core Commands"})
	root.AddCommand(
		&cobralib.Command{Use: "run", Short: "Run", GroupID: "core", RunE: noop},
	)

	sections := cobra.Sections(root)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Core Commands"}, titles)
}
