package urfave_test

import (
	"bytes"
	"context"
	"testing"

	urfavecli "github.com/gechr/clib/cli/urfave"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
	clilib "github.com/urfave/cli/v3"
)

func TestSections_Usage(t *testing.T) {
	cmd := &clilib.Command{
		Name:      "app",
		ArgsUsage: "<file>",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output format"},
		},
	}

	sections := urfavecli.Sections(cmd)

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
	cmd := &clilib.Command{
		Name:      "search",
		ArgsUsage: "[query]",
	}

	sections := urfavecli.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Len(t, usage.Args, 1)
	require.Equal(t, "query", usage.Args[0].Name)
	require.False(t, usage.Args[0].Required)
}

func TestSections_Usage_NoFlags(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}

	sections := urfavecli.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.False(t, usage.ShowOptions)
}

func TestSections_Aliases(t *testing.T) {
	cmd := &clilib.Command{
		Name:    "app",
		Aliases: []string{"a", "application"},
	}

	sections := urfavecli.Sections(cmd)

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

func TestSections_Commands(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Commands: []*clilib.Command{
			{Name: "run", Usage: "Run the app"},
			{Name: "build", Usage: "Build the app"},
		},
	}

	sections := urfavecli.Sections(cmd)

	var cmdSection *help.Section
	for i := range sections {
		if sections[i].Title == "Commands" {
			cmdSection = &sections[i]
			break
		}
	}
	require.NotNil(t, cmdSection)

	cmds, ok := cmdSection.Content[0].(help.CommandGroup)
	require.True(t, ok)
	require.Len(t, cmds, 2)
}

func TestSections_Flags(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output format"},
			&clilib.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output",
			},
		},
	}

	sections := urfavecli.Sections(cmd)

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
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "name", Usage: "Your name"},
			&clilib.BoolFlag{Name: "verbose", Usage: "Verbose output"},
		},
	}

	sections := urfavecli.Sections(cmd)

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

func TestSections_GroupedFlags(t *testing.T) {
	repoFlag := &clilib.StringFlag{Name: "repo", Aliases: []string{"R"}, Usage: "Filter by repo"}
	authorFlag := &clilib.StringFlag{
		Name:    "author",
		Aliases: []string{"a"},
		Usage:   "Filter by author",
	}
	outputFlag := &clilib.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output format"}
	limitFlag := &clilib.IntFlag{Name: "limit", Aliases: []string{"L"}, Usage: "Maximum results"}

	urfavecli.Extend(repoFlag, urfavecli.FlagExtra{Group: "Filters"})
	urfavecli.Extend(authorFlag, urfavecli.FlagExtra{Group: "Filters"})
	urfavecli.Extend(outputFlag, urfavecli.FlagExtra{Group: "Output"})
	urfavecli.Extend(limitFlag, urfavecli.FlagExtra{Group: "Output"})

	cmd := &clilib.Command{
		Name:  "app",
		Flags: []clilib.Flag{repoFlag, authorFlag, outputFlag, limitFlag},
	}

	sections := urfavecli.Sections(cmd)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Filters", "Output"}, titles)

	for _, s := range sections {
		if s.Title == "Filters" {
			flags, ok := s.Content[0].(help.FlagGroup)
			require.True(t, ok)
			require.Len(t, flags, 2)
		}
	}
}

func TestSections_Placeholder_Annotation(t *testing.T) {
	repoFlag := &clilib.StringFlag{Name: "repo", Aliases: []string{"R"}, Usage: "Filter by repo"}
	urfavecli.Extend(repoFlag, urfavecli.FlagExtra{Placeholder: "owner/repo"})

	cmd := &clilib.Command{
		Name:  "app",
		Flags: []clilib.Flag{repoFlag},
	}

	sections := urfavecli.Sections(cmd)

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

func TestSections_Negatable(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.BoolWithInverseFlag{Name: "draft", Usage: "Include drafts"},
		},
	}

	sections := urfavecli.Sections(cmd)

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
	require.Equal(t, "[no-]draft", flags[0].Long)
}

func TestSections_NoAliases(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}

	sections := urfavecli.Sections(cmd)

	for _, s := range sections {
		require.NotEqual(t, "Aliases", s.Title)
	}
}

func TestHelpPrinter(t *testing.T) {
	r := help.NewRenderer(nil)
	printer := urfavecli.HelpPrinter(r, urfavecli.Sections)

	// Non-command data should be silently ignored.
	var buf bytes.Buffer
	printer(&buf, "", "not a command")
	require.Empty(t, buf.String())
}

func TestHelpNewRendererNilUsesDefaultTheme(t *testing.T) {
	t.Setenv("CLIB_THEME", "monochrome")
	theme.SetEnvPrefix("")
	t.Cleanup(func() {
		theme.SetEnvPrefix("")
	})

	r := help.NewRenderer(nil)
	require.Equal(t, theme.Monochrome().HelpSection.Render("x"), r.Theme.HelpSection.Render("x"))
}

func TestHelpPrinter_WithCommand(t *testing.T) {
	th := theme.Default()
	r := help.NewRenderer(th)
	printer := urfavecli.HelpPrinter(r, urfavecli.Sections)

	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Usage: "Output format"},
		},
	}

	var buf bytes.Buffer
	printer(&buf, "", cmd)
	require.NotEmpty(t, buf.String())
}

func TestSections_InheritedFlags(t *testing.T) {
	// The parent's verbose flag is added to the child's Flags to simulate
	// how a CLI app propagates inherited flags.
	verboseFlag := &clilib.BoolFlag{
		Name:    "verbose",
		Aliases: []string{"v"},
		Usage:   "Verbose output",
	}

	parent := &clilib.Command{
		Name:  "parent",
		Flags: []clilib.Flag{verboseFlag},
		Commands: []*clilib.Command{
			{
				Name: "child",
				Flags: []clilib.Flag{
					&clilib.StringFlag{Name: "output", Usage: "Output format"},
					verboseFlag, // inherited from parent
				},
			},
		},
	}

	// Run the parent command so Lineage is populated.
	var sections []help.Section
	parent.Commands[0].Action = func(_ context.Context, cmd *clilib.Command) error {
		sections = urfavecli.Sections(cmd)
		return nil
	}
	_ = parent.Run(context.Background(), []string{"parent", "child"})

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	// Inherited flags merge into "Options"; urfave's auto-injected --help
	// is pulled into its own trailing sub-group, so we expect three.
	require.Equal(t, []string{"Usage", "Options"}, titles)
	opts := sections[1]
	require.Len(t, opts.Content, 3, "local + inherited + help sub-groups")
}

func TestSections_GroupedWithInherited(t *testing.T) {
	filterFlag := &clilib.StringFlag{Name: "repo", Usage: "Filter by repo"}
	urfavecli.Extend(filterFlag, urfavecli.FlagExtra{Group: "Filters"})

	verboseFlag := &clilib.BoolFlag{Name: "verbose", Usage: "Verbose output"}

	parent := &clilib.Command{
		Name:  "parent",
		Flags: []clilib.Flag{verboseFlag},
		Commands: []*clilib.Command{
			{
				Name: "child",
				Flags: []clilib.Flag{
					filterFlag,
					verboseFlag, // inherited from parent
				},
			},
		},
	}

	var sections []help.Section
	parent.Commands[0].Action = func(_ context.Context, cmd *clilib.Command) error {
		sections = urfavecli.Sections(cmd)
		return nil
	}
	_ = parent.Run(context.Background(), []string{"parent", "child"})

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	// Inherited "verbose" merges into "Options" as a second sub-group by default.
	require.Equal(t, []string{"Usage", "Filters", "Options"}, titles)
}

func TestSections_PerLevelAncestorDepth(t *testing.T) {
	// root (quiet) -> parent (filter) -> leaf (limit).
	rootFlag := &clilib.BoolFlag{Name: "quiet", Usage: "Quiet mode"}
	parentFlag := &clilib.StringFlag{Name: "filter", Usage: "Filter string"}
	leafFlag := &clilib.IntFlag{Name: "limit", Usage: "Max results"}

	// Each level declares only its own flag; the leaf re-declares the ancestors
	// (urfave inheritance model).
	root := &clilib.Command{
		Name:  "root",
		Flags: []clilib.Flag{rootFlag},
		Commands: []*clilib.Command{
			{
				Name:  "parent",
				Flags: []clilib.Flag{parentFlag},
				Commands: []*clilib.Command{
					{
						Name:  "leaf",
						Flags: []clilib.Flag{leafFlag, parentFlag, rootFlag},
					},
				},
			},
		},
	}

	var sections []help.Section
	root.Commands[0].Commands[0].Action = func(_ context.Context, cmd *clilib.Command) error {
		sections = urfavecli.Sections(cmd)
		return nil
	}
	_ = root.Run(context.Background(), []string{"root", "parent", "leaf"})

	var opts *help.Section
	for i := range sections {
		if sections[i].Title == "Options" {
			opts = &sections[i]
			break
		}
	}
	require.NotNil(t, opts)
	// Four sub-groups: local + parent + root, plus urfave's --help as the
	// mandatory trailing sub-group.
	require.Len(t, opts.Content, 4)

	fg0, ok := opts.Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "limit", fg0[0].Long)

	fg1, ok := opts.Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "filter", fg1[0].Long)

	fg2, ok := opts.Content[2].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "quiet", fg2[0].Long)

	fg3, ok := opts.Content[3].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "help", fg3[0].Long)
}

func TestSections_HiddenFlagsFiltered(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Usage: "Output format"},
			&clilib.StringFlag{Name: "secret", Usage: "Secret flag", Hidden: true},
		},
	}

	sections := urfavecli.Sections(cmd)

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
	// Only the visible flag should appear.
	require.Len(t, flags, 1)
	require.Equal(t, "output", flags[0].Long)
}

func TestSections_AllHiddenFlags(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "secret", Usage: "Secret", Hidden: true},
		},
	}

	sections := urfavecli.Sections(cmd)

	// No flag sections should appear.
	for _, s := range sections {
		require.NotEqual(t, "Options", s.Title)
	}

	// Usage should show ShowOptions=false.
	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.False(t, usage.ShowOptions)
}

func TestSections_RepeatedArg(t *testing.T) {
	cmd := &clilib.Command{
		Name:      "app",
		ArgsUsage: "<file>...",
	}

	sections := urfavecli.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Len(t, usage.Args, 1)
	require.Equal(t, "file", usage.Args[0].Name)
	require.True(t, usage.Args[0].Repeatable)
	require.True(t, usage.Args[0].Required)
}

func TestSections_OptionalRepeatedArg(t *testing.T) {
	cmd := &clilib.Command{
		Name:      "app",
		ArgsUsage: "[<file>...]",
	}

	sections := urfavecli.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Len(t, usage.Args, 1)
	require.Equal(t, "file", usage.Args[0].Name)
	require.True(t, usage.Args[0].Repeatable)
	require.False(t, usage.Args[0].Required)
}

func TestSections_FlagEnum(t *testing.T) {
	flag := &clilib.StringFlag{Name: "format", Usage: "Output format"}
	urfavecli.Extend(flag, urfavecli.FlagExtra{
		Enum:          []string{"json", "yaml", "table"},
		EnumHighlight: []string{"j", "y", "t"},
	})

	cmd := &clilib.Command{
		Name:  "app",
		Flags: []clilib.Flag{flag},
	}

	sections := urfavecli.Sections(cmd)

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
	require.Equal(t, []string{"json", "yaml", "table"}, flags[0].Enum)
	require.Equal(t, []string{"j", "y", "t"}, flags[0].EnumHighlight)
}

func TestSections_CategoryGroup(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Usage: "Output format", Category: "Output"},
			&clilib.StringFlag{Name: "name", Usage: "Your name"},
		},
	}

	sections := urfavecli.Sections(cmd)

	var titles []string
	for _, s := range sections {
		titles = append(titles, s.Title)
	}
	require.Equal(t, []string{"Usage", "Output", "Options"}, titles)
}

func TestSections_ShortOnlyFlag(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.BoolFlag{Name: "v", Usage: "Verbose output"},
		},
	}

	sections := urfavecli.Sections(cmd)

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
	// Single-char name should end up as "long" (fallback).
	require.Equal(t, "v", flags[0].Long)
}

func TestSections_MultiValuePlaceholder(t *testing.T) {
	flag := &clilib.StringSliceFlag{Name: "tags", Usage: "Tags to add"}
	urfavecli.Extend(flag, urfavecli.FlagExtra{Placeholder: "tag"})

	cmd := &clilib.Command{
		Name:  "app",
		Flags: []clilib.Flag{flag},
	}

	sections := urfavecli.Sections(cmd)

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
	require.Equal(t, "tag", flags[0].Placeholder)
	require.True(t, flags[0].Repeatable)
}

func TestSections_NegatableWithShort(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Flags: []clilib.Flag{
			&clilib.BoolWithInverseFlag{
				Name:    "draft",
				Aliases: []string{"d"},
				Usage:   "Include drafts",
			},
		},
	}

	sections := urfavecli.Sections(cmd)

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
	require.Equal(t, "[no-]draft", flags[0].Long)
	require.Equal(t, "d", flags[0].Short)
}

func TestSections_MultipleArgs(t *testing.T) {
	cmd := &clilib.Command{
		Name:      "app",
		ArgsUsage: "<src> <dst>",
	}

	sections := urfavecli.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	require.Len(t, usage.Args, 2)
	require.Equal(t, "src", usage.Args[0].Name)
	require.Equal(t, "dst", usage.Args[1].Name)
}

func TestSections_Flags_EnumDefaultFromDefaultText(t *testing.T) {
	flag := &clilib.StringFlag{
		Name:        "color",
		Usage:       "Color mode",
		Value:       "auto",
		DefaultText: "auto",
	}
	urfavecli.Extend(flag, urfavecli.FlagExtra{
		Enum: []string{"auto", "always", "never"},
	})

	cmd := &clilib.Command{
		Name:  "app",
		Flags: []clilib.Flag{flag},
	}

	sections := urfavecli.Sections(cmd)

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

	var colorFlag *help.Flag
	for i := range flags {
		if flags[i].Long == "color" {
			colorFlag = &flags[i]
			break
		}
	}
	require.NotNil(t, colorFlag)
	require.Equal(t, "auto", colorFlag.EnumDefault, "should fall back to urfave default text")
}

func TestSections_Flags_EnumDefaultExtraOverridesDefault(t *testing.T) {
	flag := &clilib.StringFlag{
		Name:        "state",
		Usage:       "PR state",
		Value:       "closed",
		DefaultText: "closed",
	}
	urfavecli.Extend(flag, urfavecli.FlagExtra{
		Enum:        []string{"open", "closed", "merged", "all"},
		EnumDefault: "open",
	})

	cmd := &clilib.Command{
		Name:  "app",
		Flags: []clilib.Flag{flag},
	}

	sections := urfavecli.Sections(cmd)

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

func TestSections_CommandsAddSubcommandArg(t *testing.T) {
	cmd := &clilib.Command{
		Name: "app",
		Commands: []*clilib.Command{
			{Name: "run", Usage: "Run the app"},
		},
	}

	sections := urfavecli.Sections(cmd)

	usage, ok := sections[0].Content[0].(help.Usage)
	require.True(t, ok)
	// Should have a subcommand arg added.
	require.NotEmpty(t, usage.Args)
	last := usage.Args[len(usage.Args)-1]
	require.True(t, last.IsSubcommand)
	require.Equal(t, "command", last.Name)
}
