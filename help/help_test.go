package help_test

import (
	"bytes"
	"errors"
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

func testTheme() *theme.Theme {
	return theme.Default()
}

type failingWriter struct{}

func (failingWriter) Write(_ []byte) (int, error) {
	return 0, errors.New("write failed")
}

func TestBracketArg(t *testing.T) {
	require.Equal(t, "[<query>]", help.BracketArg(help.Arg{Name: "query"}))
	require.Equal(t, "<file>", help.BracketArg(help.Arg{Name: "file", Required: true}))
	require.Equal(t, "[<query>…]", help.BracketArg(help.Arg{Name: "query", Repeatable: true}))
	require.Equal(
		t,
		"<file>…",
		help.BracketArg(help.Arg{Name: "file", Required: true, Repeatable: true}),
	)
}

func TestRender_Usage(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{
			help.Usage{Command: "prl", ShowOptions: true, Args: []help.Arg{
				{Name: "query"},
			}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Usage\n\n  prl [options] [<query>]\n", ansi.Strip(buf.String()))
}

func TestRender_Usage_SubcommandArg(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{
			help.Usage{Command: "app", ShowOptions: true, Args: []help.Arg{
				{Name: "command", Required: true, IsSubcommand: true},
			}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(
		t,
		th.HelpSection.Render(
			"Usage",
		)+"\n\n  "+th.HelpCommand.Render(
			"app",
		)+" "+th.HelpArg.Render(
			"<command>",
		)+" "+th.HelpFlag.Render(
			"[options]",
		)+"\n",
		buf.String(),
	)
}

func TestRender_Args(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{{Name: "query", Desc: "Search term"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Arguments\n\n  [<query>]  Search term\n", ansi.Strip(buf.String()))
}

func TestRender_Args_Required(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{{Name: "file", Desc: "Input file", Required: true}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Arguments\n\n  <file>  Input file\n", ansi.Strip(buf.String()))
}

func TestRender_Args_RepeatableOptional(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{{Name: "query", Desc: "Search terms", Repeatable: true}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Arguments\n\n  [<query>…]  Search terms\n", ansi.Strip(buf.String()))
}

func TestRender_Args_RequiredUsesHelpArg(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{{Name: "file", Desc: "Input file", Required: true}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Contains(t, buf.String(), th.HelpArg.Render("<file>"))
}

func TestRender_Args_OptionalOnlyUsesHelpArg(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{{Name: "query", Desc: "Search term"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// When all args are optional (no required args present), optional args
	// use HelpArg style since there is no need to distinguish.
	require.Contains(t, buf.String(), th.HelpArg.Render("[<query>]"))
}

func TestRender_Usage_RequiredAndOptionalArgStyles(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{
			help.Usage{Command: "rep", ShowOptions: true, Args: []help.Arg{
				{Name: "find", Required: true},
				{Name: "replace", Required: true},
				{Name: "path", Repeatable: true},
			}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	out := buf.String()
	require.Contains(t, out, th.HelpArg.Render("<find>"))
	require.Contains(t, out, th.HelpArg.Render("<replace>"))
	require.Contains(t, out, th.HelpArgOptional.Render("[<path>…]"))
}

func TestRender_PropagatesWriteErrors(t *testing.T) {
	r := help.NewRenderer(testTheme())
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{
			help.Usage{Command: "prl"},
		}},
	}

	err := r.Render(failingWriter{}, sections)
	require.EqualError(t, err, "write failed")
}

func TestRender_Flags_AutoIndent(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "organization", Placeholder: "org", Desc: "Limit to org"},
				{Short: "R", Long: "repo", Placeholder: "repo", Desc: "Limit to repo"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))
	lines := strings.Split(buf.String(), "\n")

	var flagLines []string
	for _, line := range lines {
		stripped := ansi.Strip(line)
		if strings.Contains(stripped, "--") {
			flagLines = append(flagLines, stripped)
		}
	}

	require.Len(t, flagLines, 2)
	orgLine := flagLines[0]
	repoLine := flagLines[1]
	orgIndent := len(orgLine) - len(strings.TrimLeft(orgLine, " "))
	repoIndent := len(repoLine) - len(strings.TrimLeft(repoLine, " "))
	require.Greater(t, orgIndent, repoIndent, "long-only flag should have extra indent")
}

func TestRender_Flags_NoExtraIndent_WhenAllLong(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "approve", Desc: "Approve PRs"},
				{Long: "close", Desc: "Close PRs"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(
		t,
		"Test\n\n  --approve  Approve PRs\n  --close    Close PRs\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_BlankLineBetweenFlagGroups(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Filters", Content: []help.Content{
			help.FlagGroup{{Short: "a", Long: "author", Desc: "Filter by author"}},
			help.FlagGroup{{Short: "c", Long: "created", Desc: "Filter by date"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Filters\n\n  -a, --author   Filter by author\n\n  -c, --created  Filter by date\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_MultipleSections(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "First", Content: []help.Content{
			help.FlagGroup{{Long: "alpha", Desc: "Alpha"}},
		}},
		{Title: "Second", Content: []help.Content{
			help.FlagGroup{{Long: "beta", Desc: "Beta"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"First\n\n  --alpha  Alpha\n\nSecond\n\n  --beta   Beta\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_Text(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Config", Content: []help.Content{
			help.Text("Some freeform text here"),
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Config\n\n  Some freeform text here\n", ansi.Strip(buf.String()))
}

func TestRender_Examples(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Examples", Content: []help.Content{
			help.Examples{
				{Comment: "List PRs", Command: "prl"},
				{Comment: "Search", Command: "prl foo"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Examples\n\n  # List PRs\n  $ prl\n\n  # Search\n  $ prl foo\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_NestedSection(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Configuration", Content: []help.Content{
			help.Text("Some intro text"),
			&help.Section{Title: "tf-github", Content: []help.Content{
				help.FlagGroup{{Long: "topic", Desc: "Resolve topics"}},
			}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Configuration\n\n  Some intro text\n\n    tf-github\n\n      --topic  Resolve topics\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_NestedSection_TitleUseCommandStyle(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Config", Content: []help.Content{
			&help.Section{Title: "sub", Content: []help.Content{
				help.FlagGroup{{Long: "flag", Desc: "desc"}},
			}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(
		t,
		th.HelpSection.Render(
			"Config",
		)+"\n\n    "+th.HelpCommand.Render(
			"sub",
		)+"\n\n      "+th.HelpFlag.Render(
			"--flag",
		)+"  desc\n",
		buf.String(),
	)
}

func TestRender_DescColAlignment(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Short: "a", Long: "author", Placeholder: "user", Desc: "Filter by author"},
				{Long: "organization", Placeholder: "org", Desc: "Limit to org"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))
	lines := strings.Split(buf.String(), "\n")

	var descPositions []int
	for _, line := range lines {
		stripped := ansi.Strip(line)
		for _, desc := range []string{"Filter by author", "Limit to org"} {
			if idx := strings.Index(stripped, desc); idx > 0 {
				descPositions = append(descPositions, idx)
			}
		}
	}

	require.Len(t, descPositions, 2)
	require.Equal(t, descPositions[0], descPositions[1],
		"descriptions should start at the same column")
}

func TestRender_FlagPlaceholderBrackets(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{{Long: "repo", Placeholder: "repo", Desc: "The repo"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Test\n\n  --repo <repo>  The repo\n", ansi.Strip(buf.String()))
}

func TestRender_FlagRepeatablePlaceholder(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "label", Placeholder: "tag", Repeatable: true, Desc: "Add labels"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Test\n\n  --label <tag>…  Add labels\n", ansi.Strip(buf.String()))
}

func TestRender_FlagRepeatableSuffix_UsesThemeStyle(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "label", Placeholder: "tag", Repeatable: true, Desc: "Add labels"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	sep := th.HelpKeyValueSeparatorStyle.Render(" ")
	require.Equal(t,
		th.HelpSection.Render("Test")+"\n\n  "+
			th.HelpFlag.Render(
				"--label",
			)+sep+th.HelpValuePlaceholder.Render("<tag>")+th.HelpRepeatEllipsis.Render(help.EllipsisShort)+
			"  Add labels\n",
		buf.String(),
	)
}

func TestRender_FlagRepeatableSuffix_Disabled(t *testing.T) {
	th := testTheme()
	th.HelpRepeatEllipsisEnabled = false
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "label", Placeholder: "tag", Repeatable: true, Desc: "Add labels"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// With repeat disabled, output should show <tag> without ,...
	require.Equal(t, "Test\n\n  --label <tag>  Add labels\n", ansi.Strip(buf.String()))
}

func TestRender_FlagKeyValueSeparator(t *testing.T) {
	th := testTheme()
	th.HelpKeyValueSeparator = '='
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "output", Placeholder: "fmt", Desc: "Output format"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Test\n\n  --output=<fmt>  Output format\n", ansi.Strip(buf.String()))
}

func TestRender_FlagKeyValueSeparatorStyle(t *testing.T) {
	th := testTheme()
	th.HelpKeyValueSeparator = '='
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("3"))
	th.HelpKeyValueSeparatorStyle = &s
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "output", Placeholder: "fmt", Desc: "Output format"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Test")+"\n\n  "+
			th.HelpFlag.Render("--output")+s.Render("=")+th.HelpValuePlaceholder.Render("<fmt>")+
			"  Output format\n",
		buf.String(),
	)
}

func TestRender_FlagDashesAdded(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{{Short: "R", Long: "repo", Desc: "The repo"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Test\n\n  -R, --repo  The repo\n", ansi.Strip(buf.String()))
}

func TestRender_CommandGroup(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Commands", Content: []help.Content{
			help.CommandGroup{
				{Name: "run", Desc: "Run the app"},
				{Name: "build", Desc: "Build the app"},
				{Name: "lint", Desc: "Lint code"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Commands")+"\n\n"+
			"  "+th.HelpSubcommand.Render("run")+"    Run the app\n"+
			"  "+th.HelpSubcommand.Render("build")+"  Build the app\n"+
			"  "+th.HelpSubcommand.Render("lint")+"   Lint code\n",
		buf.String(),
	)
}

func TestRender_DescNote(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:               "to",
					Placeholder:        "1.2.3",
					PlaceholderLiteral: true,
					Desc:               "Set version (implies --force)",
				},
				{Long: "debug", Desc: "Debug mode"},
				{Long: "tag", Placeholder: "tag", Desc: "Match tags (AND logic)"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	sep := th.HelpKeyValueSeparatorStyle.Render(" ")
	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n"+
			"  "+th.HelpFlag.Render("--to")+sep+th.HelpValuePlaceholder.Render("1.2.3")+"   Set version "+th.HelpFlagNote.Render("(implies --force)")+"\n"+
			"  "+th.HelpFlag.Render("--debug")+"      Debug mode\n"+
			"  "+th.HelpFlag.Render("--tag")+sep+th.HelpValuePlaceholder.Render("<tag>")+"  Match tags "+th.HelpFlagNote.Render("(AND logic)")+"\n",
		buf.String(),
	)
}

func TestRender_FlagEnum(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Short:       "s",
					Long:        "state",
					Placeholder: "state",
					Desc:        "Filter by state",
					Enum:        []string{"open", "closed", "merged", "all"},
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Flags\n\n  -s, --state <state>  Filter by state [open, closed, merged, all]\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_FlagEnum_NoDesc(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:        "state",
					Placeholder: "state",
					Enum:        []string{"open", "closed"},
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Flags\n\n  --state <state>  [open, closed]\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_FlagEnum_DoesNotAffectDescCol(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Short:       "s",
					Long:        "state",
					Placeholder: "state",
					Desc:        "Filter by state",
					Enum:        []string{"open", "closed", "merged", "all"},
				},
				{Short: "v", Long: "verbose", Desc: "Verbose output"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	lines := strings.Split(buf.String(), "\n")
	var descPositions []int
	for _, line := range lines {
		stripped := ansi.Strip(line)
		for _, desc := range []string{"Filter by state", "Verbose output"} {
			if idx := strings.Index(stripped, desc); idx > 0 {
				descPositions = append(descPositions, idx)
			}
		}
	}

	require.Len(t, descPositions, 2)
	require.Equal(t, descPositions[0], descPositions[1],
		"enum suffix should not affect description column alignment")
}

func TestRender_FlagEnum_Default_Plain(t *testing.T) {
	th := testTheme()
	th.EnumStyle = theme.EnumStylePlain
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Short:       "s",
					Long:        "state",
					Placeholder: "state",
					Desc:        "Filter by state",
					Enum:        []string{"open", "closed", "merged", "all"},
					EnumDefault: "open",
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(
		t,
		"Flags\n\n  -s, --state <state>  Filter by state [open, closed, merged, all] (default: open)\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_FlagEnum_Default_Plain_UsesThemeMethod(t *testing.T) {
	th := testTheme()
	th.EnumStyle = theme.EnumStylePlain
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:        "sort",
					Placeholder: "sort",
					Desc:        "Sort order",
					Enum:        []string{"asc", "desc"},
					EnumDefault: "asc",
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	sep := th.HelpKeyValueSeparatorStyle.Render(" ")
	enumPart := th.FmtEnumDefault("asc", []theme.EnumValue{{Name: "asc"}, {Name: "desc"}})
	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render("--sort")+sep+th.HelpValuePlaceholder.Render("<sort>")+
			"  Sort order "+enumPart+"\n",
		buf.String(),
	)
}

func TestRender_FlagEnum_HighlightPrefix(t *testing.T) {
	th := testTheme()
	th.EnumStyle = theme.EnumStyleHighlightPrefix
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:          "state",
					Placeholder:   "state",
					Desc:          "Filter by state",
					Enum:          []string{"open", "closed"},
					EnumHighlight: []string{"o", "c"},
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	sep := th.HelpKeyValueSeparatorStyle.Render(" ")
	enumPart := th.FmtEnum([]theme.EnumValue{
		{Name: "open", Bold: "o"},
		{Name: "closed", Bold: "c"},
	})
	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render("--state")+sep+th.HelpValuePlaceholder.Render("<state>")+
			"  Filter by state "+enumPart+"\n",
		buf.String(),
	)
}

func TestRender_FlagEnum_HighlightDefault(t *testing.T) {
	th := testTheme()
	th.EnumStyle = theme.EnumStyleHighlightDefault
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:        "color",
					Placeholder: "color",
					Desc:        "Color output mode",
					Enum:        []string{"auto", "always", "never"},
					EnumDefault: "auto",
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	sep := th.HelpKeyValueSeparatorStyle.Render(" ")
	enumPart := th.FmtEnum([]theme.EnumValue{
		{Name: "auto", IsDefault: true},
		{Name: "always"},
		{Name: "never"},
	})
	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render("--color")+sep+th.HelpValuePlaceholder.Render("<color>")+
			"  Color output mode "+enumPart+"\n",
		buf.String(),
	)
}

func TestRender_FlagEnum_HighlightDefault_NoDefault(t *testing.T) {
	th := testTheme()
	th.EnumStyle = theme.EnumStyleHighlightDefault
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:        "color",
					Placeholder: "color",
					Desc:        "Color output mode",
					Enum:        []string{"auto", "always", "never"},
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Flags\n\n  --color <color>  Color output mode [auto, always, never]\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_FlagEnum_MismatchedHighlightErrors(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:          "state",
					Desc:          "desc",
					Enum:          []string{"open", "closed"},
					EnumHighlight: []string{"o"},
				},
			},
		}},
	}
	err := r.Render(&buf, sections)
	require.EqualError(t, err, "help: EnumHighlight length must match Enum length")
}

func TestRender_BackticksStyled_ByDefault(t *testing.T) {
	th := testTheme()
	require.NotNil(t, th.HelpDescBacktick)

	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "debug", Desc: "Log HTTP requests to `stderr`"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(
		t,
		th.HelpSection.Render(
			"Flags",
		)+"\n\n  "+th.HelpFlag.Render(
			"--debug",
		)+"  Log HTTP requests to "+
			th.HelpDescBacktick.Render(
				"stderr",
			)+"\n",
		buf.String(),
	)
}

func TestRender_BackticksStyled_WhenConfigured(t *testing.T) {
	th := testTheme()
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th.HelpDescBacktick = &s

	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "debug", Desc: "Log HTTP requests to `stderr`"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render("--debug")+"  Log HTTP requests to "+s.Render("stderr")+"\n",
		buf.String(),
	)
}

func TestRender_DescNoteBackticksKeepsClosingParenStyled(t *testing.T) {
	th := testTheme()
	note := lipgloss.NewStyle().Faint(true)
	code := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th.HelpFlagNote = &note
	th.HelpDescBacktick = &code

	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "git", Desc: "Clone with `git` (alias for `--vcs=git`)"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render("--git")+"  Clone with "+code.Render("git")+" "+
			note.Render("(alias for ")+code.Inherit(note).Render("--vcs=git")+note.Render(")")+"\n",
		buf.String(),
	)
}

func TestRender_DescBracketDefault(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "output", Desc: "Output format [default: table]"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render(
				"--output",
			)+"  Output format "+th.HelpFlagDefault.Render("[default: table]")+"\n",
		buf.String(),
	)
}

func TestRender_DescBracketExample(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "format", Desc: "Format string [example: json]"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render(
				"--format",
			)+"  Format string "+th.HelpFlagExample.Render("[example: json]")+"\n",
		buf.String(),
	)
}

func TestRender_DescBracketGeneric(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "mode", Desc: "Mode [beta]"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Flags")+"\n\n  "+
			th.HelpFlag.Render("--mode")+"  Mode "+th.HelpDim.Render("[beta]")+"\n",
		buf.String(),
	)
}

func TestRender_DescBracketOnly(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "mode", Desc: "[value]"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// When desc is just brackets with no prefix, it should be left as-is.
	require.Equal(t,
		"Flags\n\n  --mode  [value]\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_DescNoteOnly(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "mode", Desc: "(note only)"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// When desc is entirely a note with no prefix, it should be left as-is.
	require.Equal(t,
		"Flags\n\n  --mode  (note only)\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_DescEmbeddedParens(t *testing.T) {
	th := testTheme()
	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "delete-context", Desc: "delete context(s)"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// Embedded parens without a preceding space must not be split.
	require.Equal(t,
		"Flags\n\n  --delete-context  delete context(s)\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_BackticksUnclosed(t *testing.T) {
	th := testTheme()
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th.HelpDescBacktick = &s

	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "debug", Desc: "Log to `stderr"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// Unclosed backtick should be left intact.
	require.Equal(t,
		"Flags\n\n  --debug  Log to `stderr\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_SingleQuotesStyled(t *testing.T) {
	th := testTheme()
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th.HelpDescBacktick = &s

	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Commands", Content: []help.Content{
			help.CommandGroup{
				{Name: "resolve", Desc: "Runtime resolution for 'prl' filters"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		th.HelpSection.Render("Commands")+"\n\n  "+
			th.HelpSubcommand.Render(
				"resolve",
			)+"  Runtime resolution for "+s.Render("prl")+" filters\n",
		buf.String(),
	)
}

func TestRender_SingleQuotesContraction(t *testing.T) {
	th := testTheme()
	s := lipgloss.NewStyle().Foreground(lipgloss.Color("5"))
	th.HelpDescBacktick = &s

	r := help.NewRenderer(th)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "x", Desc: "don't do that"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// Contractions must not be styled.
	require.Equal(t, "Flags\n\n  --x  don't do that\n", ansi.Strip(buf.String()))
}

func TestRender_FlagShortOnly(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Short: "v", Desc: "Verbose output"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Test\n\n  -v  Verbose output\n", ansi.Strip(buf.String()))
}

func TestRender_FlagShortOnlyWithPlaceholder(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Short: "n", Placeholder: "num", Desc: "Number of results"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Test\n\n  -n <num>  Number of results\n", ansi.Strip(buf.String()))
}

func TestRender_CommandNoDesc(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Commands", Content: []help.Content{
			help.CommandGroup{
				{Name: "help"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Commands\n\n  help\n", ansi.Strip(buf.String()))
}

func TestWithCommandAlign_Right(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithCommandAlign(help.AlignRight))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Commands", Content: []help.Content{
			help.CommandGroup{
				{Name: "run", Desc: "Run the app"},
				{Name: "build", Desc: "Build the app"},
				{Name: "lint", Desc: "Lint code"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	// Names should be right-aligned: "run" padded more, "build" less.
	require.Equal(t,
		"Commands\n\n    run  Run the app\n  build  Build the app\n   lint  Lint code\n",
		stripped,
	)
}

func TestRender_ArgNoDesc(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{{Name: "query"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Arguments\n\n  [<query>]\n", ansi.Strip(buf.String()))
}

func TestWithFlagPadding(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithFlagPadding(5))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "verbose", Desc: "Verbose output"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// 2 indent + 9 "--verbose" + 5 padding = 16, so desc starts at col 16.
	require.Equal(t, "Test\n\n  --verbose     Verbose output\n", ansi.Strip(buf.String()))
}

func TestWithArgumentPadding(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithArgumentPadding(4))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{{Name: "query", Desc: "Search term"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	// 2 indent + 9 "[<query>]" + 4 padding = 15, so desc starts at col 15.
	require.Equal(t, "Arguments\n\n  [<query>]    Search term\n", ansi.Strip(buf.String()))
}

func TestWithCommandPadding(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithCommandPadding(4))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Commands", Content: []help.Content{
			help.CommandGroup{
				{Name: "run", Desc: "Run the app"},
				{Name: "build", Desc: "Build the app"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Commands\n\n  run      Run the app\n  build    Build the app\n",
		ansi.Strip(buf.String()),
	)
}

func TestWithFlagAlign_Right(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithFlagAlign(help.AlignRight))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Short: "v", Long: "verbose", Desc: "Verbose output"},
				{Short: "o", Long: "output", Placeholder: "fmt", Desc: "Output format"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Flags\n\n       -v, --verbose  Verbose output\n  -o, --output <fmt>  Output format\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_FlagNoDesc(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Test", Content: []help.Content{
			help.FlagGroup{
				{Long: "verbose"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t, "Test\n\n  --verbose\n", ansi.Strip(buf.String()))
}

func TestWithMaxWidth_WrapsLongDescription(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(40))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "out", Desc: "A long description that should wrap to the next line"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	// Find the flag line and continuation line.
	var flagLines []string
	for _, line := range lines {
		if strings.Contains(line, "--out") ||
			(len(flagLines) > 0 && strings.TrimSpace(line) != "" && !strings.HasPrefix(line, "Flags")) {
			flagLines = append(flagLines, line)
		}
	}

	require.Greater(t, len(flagLines), 1, "description should wrap to multiple lines")

	// Continuation lines should be indented to the description column.
	firstDescCol := strings.Index(flagLines[0], "A long")
	require.Positive(t, firstDescCol)
	for _, line := range flagLines[1:] {
		leadingSpaces := len(line) - len(strings.TrimLeft(line, " "))
		require.Equal(t, firstDescCol, leadingSpaces,
			"continuation line should be indented to description column")
	}
}

func TestWithMaxWidth_Zero_NoWrapping(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(0))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long: "out",
					Desc: "A long description that should not wrap because max width is zero",
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Flags\n\n  --out  A long description that should not wrap because max width is zero\n",
		ansi.Strip(buf.String()),
	)
}

func TestWithMaxWidth_ShortDescNoWrap(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(80))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "out", Desc: "Short desc"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	require.Equal(t, "Flags\n\n  --out  Short desc\n", stripped)
}

func TestWithMaxWidth_WrapsCommandDescription(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(40))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Commands", Content: []help.Content{
			help.CommandGroup{
				{Name: "run", Desc: "A long description that should wrap to the next line"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(strings.TrimSuffix(stripped, "\n"), "\n")

	// Title + blank + at least 2 content lines (wrapped).
	require.Greater(t, len(lines), 3, "command description should wrap")
}

func TestWithMaxWidth_WrapsArgDescription(t *testing.T) {
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(40))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Arguments", Content: []help.Content{
			help.Args{
				{Name: "file", Desc: "A long description that should wrap to the next line"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(strings.TrimSuffix(stripped, "\n"), "\n")

	// Title + blank + at least 2 content lines (wrapped).
	require.Greater(t, len(lines), 3, "arg description should wrap")
}

func TestWrapStyle_BracketAlign_Default(t *testing.T) {
	// Default WrapBracketAlign: continuation lines align after the '['.
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(60))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:        "include",
					Placeholder: "include",
					Desc:        "Include relationships",
					Enum: []string{
						"balances",
						"details_blob",
						"access_token",
						"removal_notice",
						"migration",
						"audit_log",
					},
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	// Find content lines (skip title and blank).
	var contentLines []string
	for _, line := range lines {
		if strings.Contains(line, "--include") ||
			(len(contentLines) > 0 && strings.TrimSpace(line) != "") {
			contentLines = append(contentLines, line)
		}
	}

	require.Greater(t, len(contentLines), 1, "should wrap to multiple lines")

	// The '[' should be on the first line.
	bracketIdx := strings.Index(contentLines[0], "[")
	require.Positive(t, bracketIdx, "first line should contain '['")

	// Continuation lines should align to bracketIdx+1 (after '[').
	for _, line := range contentLines[1:] {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		require.Equal(t, bracketIdx+1, indent,
			"continuation should align after '['")
	}
}

func TestWrapStyle_Flush(t *testing.T) {
	// WrapFlush: continuation lines align to description column.
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(60), help.WithWrapStyle(help.WrapFlush))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:        "include",
					Placeholder: "include",
					Desc:        "Include relationships",
					Enum: []string{
						"balances",
						"details_blob",
						"access_token",
						"removal_notice",
						"migration",
						"audit_log",
					},
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	var contentLines []string
	for _, line := range lines {
		if strings.Contains(line, "--include") ||
			(len(contentLines) > 0 && strings.TrimSpace(line) != "") {
			contentLines = append(contentLines, line)
		}
	}

	require.Greater(t, len(contentLines), 1, "should wrap to multiple lines")

	// Description starts at same column as "Include".
	descCol := strings.Index(contentLines[0], "Include")
	require.Positive(t, descCol)

	// Continuation lines should align to descCol, NOT after '['.
	for _, line := range contentLines[1:] {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		require.Equal(t, descCol, indent,
			"continuation should align to description column")
	}
}

func TestWrapStyle_BracketBelow(t *testing.T) {
	// WrapBracketBelow: bracket drops to next line.
	r := help.NewRenderer(
		testTheme(),
		help.WithMaxWidth(80),
		help.WithWrapStyle(help.WrapBracketBelow),
	)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{
					Long:        "include",
					Placeholder: "include",
					Desc:        "Include relationships",
					Enum: []string{
						"balances",
						"details_blob",
						"access_token",
						"removal_notice",
						"migration",
						"manual_archive",
						"review_request",
						"summary_counts",
						"audit_log",
					},
				},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	var contentLines []string
	for _, line := range lines {
		if strings.Contains(line, "--include") ||
			(len(contentLines) > 0 && strings.TrimSpace(line) != "") {
			contentLines = append(contentLines, line)
		}
	}

	require.GreaterOrEqual(t, len(contentLines), 3, "should have desc + bracket + continuation")

	// First line should have description but NOT '['.
	require.NotContains(t, contentLines[0], "[",
		"first line should not contain bracket")
	require.Contains(t, contentLines[0], "Include relationships")

	// Second line should start with '[' at descCol.
	descCol := strings.Index(contentLines[0], "Include")
	require.Positive(t, descCol)
	bracketLine := contentLines[1]
	bracketIndent := len(bracketLine) - len(strings.TrimLeft(bracketLine, " "))
	require.Equal(t, descCol, bracketIndent,
		"bracket line should start at description column")
	require.Equal(t, byte('['), strings.TrimSpace(bracketLine)[0],
		"bracket line should start with '['")

	// Continuation lines within bracket should align at descCol+1.
	for _, line := range contentLines[2:] {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		require.Equal(t, descCol+1, indent,
			"bracket continuation should align after '['")
	}
}

func TestWrapStyle_BracketBelow_NoBracket(t *testing.T) {
	// Without brackets, BracketBelow falls back to flush.
	r := help.NewRenderer(
		testTheme(),
		help.WithMaxWidth(40),
		help.WithWrapStyle(help.WrapBracketBelow),
	)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "out", Desc: "A long description that should wrap to the next line"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	var contentLines []string
	for _, line := range lines {
		if strings.Contains(line, "--out") ||
			(len(contentLines) > 0 && strings.TrimSpace(line) != "") {
			contentLines = append(contentLines, line)
		}
	}

	require.Greater(t, len(contentLines), 1, "should wrap")
	descCol := strings.Index(contentLines[0], "A long")
	for _, line := range contentLines[1:] {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		require.Equal(t, descCol, indent,
			"without bracket, should fall back to flush")
	}
}

func TestWrapStyle_BracketAlign_NoBracket(t *testing.T) {
	// When no unclosed bracket exists, default BracketAlign falls back to flush.
	r := help.NewRenderer(testTheme(), help.WithMaxWidth(40))
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Flags", Content: []help.Content{
			help.FlagGroup{
				{Long: "out", Desc: "A long description that should wrap to the next line"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	var contentLines []string
	for _, line := range lines {
		if strings.Contains(line, "--out") ||
			(len(contentLines) > 0 && strings.TrimSpace(line) != "") {
			contentLines = append(contentLines, line)
		}
	}

	require.Greater(t, len(contentLines), 1, "should wrap")

	descCol := strings.Index(contentLines[0], "A long")
	require.Positive(t, descCol)
	for _, line := range contentLines[1:] {
		indent := len(line) - len(strings.TrimLeft(line, " "))
		require.Equal(t, descCol, indent,
			"without bracket, continuation should align to description column")
	}
}

func TestWithCommandAlignMode_Section(t *testing.T) {
	r := help.NewRenderer(testTheme(),
		help.WithCommandAlignMode(help.AlignModeSection),
		help.WithMaxWidth(0),
	)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Basic", Content: []help.Content{
			help.CommandGroup{
				{Name: "run", Desc: "Run"},
				{Name: "build", Desc: "Build"},
			},
		}},
		{Title: "Advanced", Content: []help.Content{
			help.CommandGroup{
				{Name: "configure", Desc: "Configure"},
				{Name: "deploy", Desc: "Deploy"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	// In section mode, "run" and "build" are aligned independently from
	// "configure" and "deploy". So the desc column differs between sections.
	runDescCol := strings.Index(findLine(lines, "run"), "Run")
	buildDescCol := strings.Index(findLine(lines, "build"), "Build")
	require.Equal(t, runDescCol, buildDescCol, "commands within section should be aligned")

	configureDescCol := strings.Index(findLine(lines, "configure"), "Configure")
	deployDescCol := strings.Index(findLine(lines, "deploy"), "Deploy")
	require.Equal(t, configureDescCol, deployDescCol, "commands within section should be aligned")

	// The two sections should have different desc columns because "build" (5)
	// is shorter than "configure" (9).
	require.NotEqual(t, buildDescCol, configureDescCol,
		"sections should have independent alignment in section mode")
}

func TestWithCommandAlignMode_Global(t *testing.T) {
	r := help.NewRenderer(testTheme(),
		help.WithCommandAlignMode(help.AlignModeGlobal),
		help.WithMaxWidth(0),
	)
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Basic", Content: []help.Content{
			help.CommandGroup{
				{Name: "run", Desc: "Run"},
				{Name: "build", Desc: "Build"},
			},
		}},
		{Title: "Advanced", Content: []help.Content{
			help.CommandGroup{
				{Name: "configure", Desc: "Configure"},
				{Name: "deploy", Desc: "Deploy"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	stripped := ansi.Strip(buf.String())
	lines := strings.Split(stripped, "\n")

	// In global mode, all desc columns should be the same.
	runDescCol := strings.Index(findLine(lines, "run"), "Run")
	buildDescCol := strings.Index(findLine(lines, "build"), "Build")
	configureDescCol := strings.Index(findLine(lines, "configure"), "Configure")
	deployDescCol := strings.Index(findLine(lines, "deploy"), "Deploy")

	require.Equal(t, runDescCol, buildDescCol)
	require.Equal(t, buildDescCol, configureDescCol)
	require.Equal(t, configureDescCol, deployDescCol,
		"all commands should share the same description column in global mode")
}

// findLine returns the first line containing substr.
func findLine(lines []string, substr string) string {
	for _, line := range lines {
		if strings.Contains(line, substr) {
			return line
		}
	}
	return ""
}

func TestParseArg(t *testing.T) {
	tests := []struct {
		input string
		want  help.Arg
	}{
		{
			input: "<context>",
			want:  help.Arg{Name: "context", Required: true, Repeatable: false},
		},
		{
			input: "[<namespace>]",
			want:  help.Arg{Name: "namespace", Required: false, Repeatable: false},
		},
		{
			input: "<file>...",
			want:  help.Arg{Name: "file", Required: true, Repeatable: true},
		},
		{
			input: "[<query>\u2026]",
			want:  help.Arg{Name: "query", Required: false, Repeatable: true},
		},
		{
			input: "plain",
			want:  help.Arg{Name: "plain", Required: true, Repeatable: false},
		},
		{
			// Ellipsis inside angle brackets: stripped as <files...> -> inner "files..."
			input: "<files...>",
			want:  help.Arg{Name: "files", Required: true, Repeatable: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := help.ParseArg(tt.input)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestRender_Flags_MixedIndentDoesNotLeakAcrossSections(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Mixed", Content: []help.Content{
			help.FlagGroup{{Short: "a", Long: "author", Desc: "Author"}},
		}},
		{Title: "LongOnly", Content: []help.Content{
			help.FlagGroup{{Long: "verbose", Desc: "Verbose"}},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Mixed\n\n  -a, --author  Author\n\nLongOnly\n\n  --verbose     Verbose\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_Flags_LongOnlyIndentedWhenSectionHasShort(t *testing.T) {
	// Long-only flags in a separate FlagGroup within the same section should
	// be indented when another group in the section has short flags.
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Long: "init", Desc: "Initialize config"},
			},
			help.FlagGroup{
				{Long: "color", Placeholder: "color", Desc: "When to use color"},
				{Short: "Q", Long: "quick", Desc: "Skip enrichment"},
				{Short: "v", Long: "verbose", Desc: "Enable verbose logging"},
			},
			help.FlagGroup{
				{Short: "h", Desc: "Print short help"},
				{Long: "help", Desc: "Print long help"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))
	got := ansi.Strip(buf.String())
	lines := strings.Split(got, "\n")

	// Find the key flag lines.
	var initLine, colorLine, quickLine, helpShortLine, helpLongLine string
	for _, l := range lines {
		stripped := strings.TrimLeft(l, " ")
		switch {
		case strings.HasPrefix(stripped, "--init"):
			initLine = l
		case strings.HasPrefix(stripped, "--color"):
			colorLine = l
		case strings.HasPrefix(stripped, "-Q"):
			quickLine = l
		case strings.HasPrefix(stripped, "-h"):
			helpShortLine = l
		case strings.HasPrefix(stripped, "--help"):
			helpLongLine = l
		}
	}

	indent := func(s string) int { return len(s) - len(strings.TrimLeft(s, " ")) }

	// --init (long-only) should have extra indent to align with long flags.
	require.Greater(t, indent(initLine), indent(quickLine),
		"long-only --init should be indented past short flags")

	// --color (long-only) should have same indent as --init.
	require.Equal(t, indent(initLine), indent(colorLine),
		"all long-only flags in section should share indent")

	// --help (long-only) should also be indented.
	require.Greater(t, indent(helpLongLine), indent(helpShortLine),
		"--help should be indented past -h in same group")
}

func TestRender_Flags_ShortOnlyGroupTriggersIndent(t *testing.T) {
	// A group with a short-only flag (no long) should still cause long-only
	// flags in the same section to be indented.
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Long: "verbose", Desc: "Verbose"},
			},
			help.FlagGroup{
				{Short: "h", Desc: "Print help"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))
	got := ansi.Strip(buf.String())

	// --verbose should be indented because -h exists in the section.
	require.Contains(t, got, "      --verbose")
}

func TestBuildFlagSections_Empty(t *testing.T) {
	result := help.BuildFlagSections(nil)
	require.Nil(t, result)
}

func TestBuildFlagSections_UngroupedLocalOnly(t *testing.T) {
	flags := []help.ClassifiedFlag{
		{Flag: help.Flag{Long: "verbose", Desc: "Verbose output"}, Group: "", Inherited: false},
		{Flag: help.Flag{Long: "debug", Desc: "Debug mode"}, Group: "", Inherited: false},
	}

	result := help.BuildFlagSections(flags)

	require.Len(t, result, 1)
	require.Equal(t, "Options", result[0].Title)
	require.Len(t, result[0].Content, 1)

	fg, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 2)
	require.Equal(t, "verbose", fg[0].Long)
	require.Equal(t, "debug", fg[1].Long)
}

func TestBuildFlagSections_UngroupedLocalAndInherited(t *testing.T) {
	flags := []help.ClassifiedFlag{
		{Flag: help.Flag{Long: "verbose", Desc: "Verbose"}, Group: "", Inherited: false},
		{Flag: help.Flag{Long: "config", Desc: "Config file"}, Group: "", Inherited: true},
	}

	result := help.BuildFlagSections(flags)

	require.Len(t, result, 2)
	require.Equal(t, "Options", result[0].Title)
	require.Equal(t, "Inherited Options", result[1].Title)

	localFg, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, localFg, 1)
	require.Equal(t, "verbose", localFg[0].Long)

	inheritedFg, ok := result[1].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, inheritedFg, 1)
	require.Equal(t, "config", inheritedFg[0].Long)
}

func TestBuildFlagSections_GroupedSortedAlphabetically(t *testing.T) {
	flags := []help.ClassifiedFlag{
		{Flag: help.Flag{Long: "format", Desc: "Output format"}, Group: "Output", Inherited: false},
		{Flag: help.Flag{Long: "author", Desc: "Filter author"}, Group: "Filter", Inherited: false},
		{Flag: help.Flag{Long: "color", Desc: "Color mode"}, Group: "Output", Inherited: false},
	}

	result := help.BuildFlagSections(flags)

	require.Len(t, result, 2)
	// Alphabetical order: Filter before Output.
	require.Equal(t, "Filter", result[0].Title)
	require.Equal(t, "Output", result[1].Title)

	filterFg, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, filterFg, 1)
	require.Equal(t, "author", filterFg[0].Long)

	outputFg, ok := result[1].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, outputFg, 2)
	require.Equal(t, "format", outputFg[0].Long)
	require.Equal(t, "color", outputFg[1].Long)
}

func TestBuildFlagSections_CompoundGroup(t *testing.T) {
	flags := []help.ClassifiedFlag{
		{
			Flag:      help.Flag{Long: "format", Desc: "Output format"},
			Group:     "Output/Format",
			Inherited: false,
		},
		{
			Flag:      help.Flag{Long: "color", Desc: "Color mode"},
			Group:     "Output/Color",
			Inherited: false,
		},
	}

	result := help.BuildFlagSections(flags)

	require.Len(t, result, 1)
	require.Equal(t, "Output", result[0].Title)
	// Two sub-groups should produce two FlagGroup content entries.
	require.Len(t, result[0].Content, 2)

	fg1, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg1, 1)
	require.Equal(t, "format", fg1[0].Long)

	fg2, ok := result[0].Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg2, 1)
	require.Equal(t, "color", fg2[0].Long)
}

func TestBuildFlagSections_KeepGroupOrder(t *testing.T) {
	flags := []help.ClassifiedFlag{
		{Flag: help.Flag{Long: "format", Desc: "Output format"}, Group: "Output", Inherited: false},
		{Flag: help.Flag{Long: "author", Desc: "Filter author"}, Group: "Filter", Inherited: false},
		{Flag: help.Flag{Long: "color", Desc: "Color mode"}, Group: "Output", Inherited: false},
	}

	result := help.BuildFlagSections(flags, help.KeepGroupOrder())

	require.Len(t, result, 2)
	// First-seen order: Output before Filter.
	require.Equal(t, "Output", result[0].Title)
	require.Equal(t, "Filter", result[1].Title)
}

func TestBuildFlagSections_GroupedWithUngroupedFlags(t *testing.T) {
	flags := []help.ClassifiedFlag{
		{Flag: help.Flag{Long: "format"}, Group: "Output", Inherited: false},
		{Flag: help.Flag{Long: "verbose"}, Group: "", Inherited: false},
		{Flag: help.Flag{Long: "config"}, Group: "", Inherited: true},
	}

	result := help.BuildFlagSections(flags)

	// "Output" group + "Options" (ungrouped local) + "Inherited Options" (ungrouped inherited).
	require.Len(t, result, 3)
	require.Equal(t, "Output", result[0].Title)
	require.Equal(t, "Options", result[1].Title)
	require.Equal(t, "Inherited Options", result[2].Title)

	localFg, ok := result[1].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "verbose", localFg[0].Long)

	inheritedFg, ok := result[2].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "config", inheritedFg[0].Long)
}

func TestWithHelpFlags(t *testing.T) {
	sections := []help.Section{
		{Title: "Options", Content: []help.Content{
			help.FlagGroup{
				{Short: "v", Long: "verbose", Desc: "Verbose"},
				{Short: "h", Long: "help", Desc: "Print help"},
			},
		}},
	}

	result := help.Apply(sections, help.WithHelpFlags("Show help", "Show detailed help"))

	require.Len(t, result, 1)
	require.Equal(t, "Options", result[0].Title)
	require.Len(t, result[0].Content, 2)

	// Original group should have help removed.
	orig, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, orig, 1)
	require.Equal(t, "verbose", orig[0].Long)

	// New help group with separate -h and --help entries.
	helpGroup, ok := result[0].Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup, 2)
	require.Equal(t, "h", helpGroup[0].Short)
	require.Empty(t, helpGroup[0].Long)
	require.Equal(t, "Show help", helpGroup[0].Desc)
	require.Equal(t, "help", helpGroup[1].Long)
	require.Empty(t, helpGroup[1].Short)
	require.Equal(t, "Show detailed help", helpGroup[1].Desc)
}

func TestWithHelpFlags_NoExistingFlagSections(t *testing.T) {
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{help.Text("app [options]")}},
	}

	result := help.Apply(sections, help.WithHelpFlags("Short", "Long"))

	// Should create a new "Options" section.
	require.Len(t, result, 2)
	require.Equal(t, "Usage", result[0].Title)
	require.Equal(t, "Options", result[1].Title)

	helpGroup, ok := result[1].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup, 2)
	require.Equal(t, "h", helpGroup[0].Short)
	require.Equal(t, "help", helpGroup[1].Long)
}

func TestWithHelpFlagsInSection(t *testing.T) {
	sections := []help.Section{
		{Title: "Output", Content: []help.Content{
			help.FlagGroup{
				{Long: "json", Desc: "JSON output"},
				{Short: "h", Long: "help", Desc: "Print help"},
			},
		}},
		{Title: "Miscellaneous", Content: []help.Content{
			help.FlagGroup{
				{Long: "debug", Desc: "Debug mode"},
			},
		}},
	}

	result := help.Apply(
		sections,
		help.WithHelpFlagsInSection("Miscellaneous", "Show help", "Show detailed help"),
	)

	require.Len(t, result, 2)
	require.Equal(t, "Output", result[0].Title)
	require.Equal(t, "Miscellaneous", result[1].Title)

	outputGroup, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, outputGroup, 1)
	require.Equal(t, "json", outputGroup[0].Long)

	require.Len(t, result[1].Content, 2)
	helpGroup, ok := result[1].Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup, 2)
	require.Equal(t, "h", helpGroup[0].Short)
	require.Equal(t, "help", helpGroup[1].Long)
}

func TestWithHelpFlagSection_PreservesCombinedHelpFlag(t *testing.T) {
	sections := []help.Section{
		{Title: "Output", Content: []help.Content{
			help.FlagGroup{
				{Long: "json", Desc: "JSON output"},
				{Short: "h", Long: "help", Desc: "Print help"},
			},
		}},
		{Title: "Miscellaneous", Content: []help.Content{
			help.FlagGroup{
				{Long: "debug", Desc: "Debug mode"},
			},
		}},
	}

	result := help.Apply(sections, help.WithHelpFlagSection("Miscellaneous"))

	require.Len(t, result, 2)
	require.Equal(t, "Output", result[0].Title)
	require.Equal(t, "Miscellaneous", result[1].Title)

	outputGroup, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, outputGroup, 1)
	require.Equal(t, "json", outputGroup[0].Long)

	require.Len(t, result[1].Content, 2)
	helpGroup, ok := result[1].Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup, 1)
	require.Equal(t, "help", helpGroup[0].Long)
	require.Equal(t, "h", helpGroup[0].Short)
}

func TestWithHelpFlagSection_MovesSplitHelpFlags(t *testing.T) {
	sections := []help.Section{
		{Title: "Output", Content: []help.Content{
			help.FlagGroup{
				{Long: "json", Desc: "JSON output"},
			},
			help.FlagGroup{
				{Short: "h", Desc: "Show help"},
				{Long: "help", Desc: "Show detailed help"},
			},
		}},
		{Title: "Miscellaneous", Content: []help.Content{
			help.FlagGroup{
				{Long: "debug", Desc: "Debug mode"},
			},
		}},
	}

	result := help.Apply(sections, help.WithHelpFlagSection("Miscellaneous"))

	require.Len(t, result, 2)
	require.Len(t, result[0].Content, 1)

	miscGroup, ok := result[1].Content[1].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, miscGroup, 2)
	require.Equal(t, "h", miscGroup[0].Short)
	require.Equal(t, "help", miscGroup[1].Long)
}

func TestWithHelpFlagsInSection_CreatesSection(t *testing.T) {
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{help.Text("app [options]")}},
	}

	result := help.Apply(
		sections,
		help.WithHelpFlagsInSection("Miscellaneous", "Short", "Long"),
	)

	require.Len(t, result, 2)
	require.Equal(t, "Usage", result[0].Title)
	require.Equal(t, "Miscellaneous", result[1].Title)

	helpGroup, ok := result[1].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, helpGroup, 2)
	require.Equal(t, "h", helpGroup[0].Short)
	require.Equal(t, "help", helpGroup[1].Long)
}

func TestWithRenamedSection(t *testing.T) {
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{help.Text("app [options]")}},
		{Title: "Inherited Options", Content: []help.Content{help.FlagGroup{{Long: "debug"}}}},
	}

	result := help.Apply(sections, help.WithRenamedSection("Inherited Options", "Global Options"))

	require.Len(t, result, 2)
	require.Equal(t, "Usage", result[0].Title)
	require.Equal(t, "Global Options", result[1].Title)
}

func TestWithoutSection(t *testing.T) {
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{help.Text("app [options]")}},
		{Title: "Global Options", Content: []help.Content{help.FlagGroup{{Long: "debug"}}}},
	}

	result := help.Apply(sections, help.WithoutSection("Global Options"))

	require.Len(t, result, 1)
	require.Equal(t, "Usage", result[0].Title)
}

func TestResolvePolicy_Default(t *testing.T) {
	b := help.ResolvePolicy(
		help.WithHelpFlags("Short", "Long"),
		help.WithRenamedSection("A", "B"),
	)
	require.False(t, b.AlwaysShowExamples)
}

func TestResolvePolicy_WithAlwaysShowExamples(t *testing.T) {
	b := help.ResolvePolicy(
		help.WithHelpFlags("Short", "Long"),
		help.WithAlwaysShowExamples(),
	)
	require.True(t, b.AlwaysShowExamples)
}

func TestWithAlwaysShowExamples_IsNoOpTransform(t *testing.T) {
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{help.Text("app")}},
		{Title: "Examples", Content: []help.Content{help.Text("$ app run")}},
	}
	result := help.Apply(sections, help.WithAlwaysShowExamples())
	require.Equal(t, sections, result)
}

func TestWithExamplesOnLongHelp_ShortHelp(t *testing.T) {
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{help.Text("app")}},
		{Title: "Examples", Content: []help.Content{help.Text("$ app run")}},
		{Title: "Options", Content: []help.Content{help.FlagGroup{{Long: "verbose"}}}},
	}

	result := help.Apply(sections, help.WithExamplesOnLongHelp([]string{"app", "-h"}))

	titles := make([]string, len(result))
	for i, s := range result {
		titles[i] = s.Title
	}
	require.Equal(t, []string{"Usage", "Options"}, titles)
}

func TestWithExamplesOnLongHelp_LongHelp(t *testing.T) {
	sections := []help.Section{
		{Title: "Usage", Content: []help.Content{help.Text("app")}},
		{Title: "Examples", Content: []help.Content{help.Text("$ app run")}},
		{Title: "Options", Content: []help.Content{help.FlagGroup{{Long: "verbose"}}}},
	}

	result := help.Apply(sections, help.WithExamplesOnLongHelp([]string{"app", "--help"}))

	titles := make([]string, len(result))
	for i, s := range result {
		titles[i] = s.Title
	}
	// Examples is moved to the end.
	require.Equal(t, []string{"Usage", "Options", "Examples"}, titles)
}

func TestBuildFlagSections_InheritedOnly(t *testing.T) {
	flags := []help.ClassifiedFlag{
		{Flag: help.Flag{Long: "config", Desc: "Config file"}, Group: "", Inherited: true},
	}

	result := help.BuildFlagSections(flags)

	require.Len(t, result, 1)
	require.Equal(t, "Inherited Options", result[0].Title)

	fg, ok := result[0].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Len(t, fg, 1)
	require.Equal(t, "config", fg[0].Long)
}

func TestRender_Examples_SingleExample(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Examples", Content: []help.Content{
			help.Examples{
				{Comment: "Run tests", Command: "myapp test"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(t,
		"Examples\n\n  # Run tests\n  $ myapp test\n",
		ansi.Strip(buf.String()),
	)
}

func TestRender_CommandGroup_LongName(t *testing.T) {
	r := help.NewRenderer(testTheme())
	var buf bytes.Buffer
	sections := []help.Section{
		{Title: "Commands", Content: []help.Content{
			help.CommandGroup{
				{Name: "very-long-command-name", Desc: "A command with a long name"},
				{Name: "ls", Desc: "List things"},
			},
		}},
	}
	require.NoError(t, r.Render(&buf, sections))

	require.Equal(
		t,
		"Commands\n\n  very-long-command-name  A command with a long name\n  ls                      List things\n",
		ansi.Strip(buf.String()),
	)
}
