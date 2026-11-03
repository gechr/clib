package kong_test

import (
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/cli/kong"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/help"
	"github.com/stretchr/testify/require"
)

type lastWinsFlags struct {
	Quiet    bool       `help:"Suppress output"       short:"q"         lastwins:"verbosity"`
	Verbose  customBool `help:"Enable verbose output" short:"v"         lastwins:"verbosity" negatable:""`
	Output   string     `help:"Output file"           short:"o"         lastwins:"output"`
	Template string     `help:"Output template"       lastwins:"output"`
	Plain    bool       `help:"Ordinary flag"         short:"p"`
}

func checkLastWinsSpecs(t *testing.T, specs []complete.Spec) {
	t.Helper()
	require.Len(t, specs, 6)
	for _, spec := range specs {
		require.Equal(t, spec.LongFlag != "plain", spec.Repeatable, spec.LongFlag)
	}
}

func TestLastWins_ReflectedCompletions(t *testing.T) {
	var cli lastWinsFlags
	flags, err := kong.Reflect(&cli)
	require.NoError(t, err)
	gen := complete.NewGenerator("app").FromFlags(flags)
	checkLastWinsSpecs(t, gen.Specs)

	parser, err := konglib.New(&cli)
	require.NoError(t, err)
	_, err = parser.Parse([]string{
		"--verbose", "--quiet", "--verbose",
		"--output=old", "--template=other", "--output=final",
	})
	require.NoError(t, err)
	require.Equal(t, customBool(1), cli.Verbose)
	require.False(t, cli.Quiet)
	require.Equal(t, "final", cli.Output)
	require.Empty(t, cli.Template)
}

func TestLastWins_ModelCompletions(t *testing.T) {
	var cli struct {
		Run lastWinsFlags `cmd:""`
	}
	parser, err := konglib.New(&cli)
	require.NoError(t, err)
	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	checkLastWinsSpecs(t, subs[0].Specs)

	ctx := parseGoldenForHelp(&cli, []string{"run", "--help"})
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)
	var flags help.FlagGroup
	for _, section := range sections {
		for _, content := range section.Content {
			if group, ok := content.(help.FlagGroup); ok {
				flags = append(flags, group...)
			}
		}
	}
	require.Len(t, flags, 6)
	require.Equal(t, "quiet", flags[0].Long)
	require.Equal(t, "[no-]verbose", flags[1].Long)
	require.Empty(t, flags[0].Placeholder)
	require.Empty(t, flags[1].Placeholder)
}

func TestLastWins_XorPrefix(t *testing.T) {
	type outputFlags struct {
		JSON bool `help:"JSON output" lastwins:"format"`
		Text bool `help:"Text output" lastwins:"format"`
	}
	var cli struct {
		Run struct {
			First  outputFlags `embed:"" prefix:"first-"  xorprefix:"first-"`
			Second outputFlags `embed:"" prefix:"second-" xorprefix:"second-"`
		} `cmd:""`
	}
	parser, err := konglib.New(&cli)
	require.NoError(t, err)
	_, err = parser.Parse([]string{"run", "--first-json", "--second-json", "--first-text"})
	require.NoError(t, err)
	require.False(t, cli.Run.First.JSON)
	require.True(t, cli.Run.First.Text)
	require.True(t, cli.Run.Second.JSON)
	require.False(t, cli.Run.Second.Text)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.Len(t, subs[0].Specs, 4)
	for _, spec := range subs[0].Specs {
		require.True(t, spec.Repeatable, spec.LongFlag)
	}
}

func TestNodeSections_EmptyInterpolationFallback(t *testing.T) {
	var cli struct {
		Output string `help:"Write output${suffix=}" default:"${output=}"`
	}
	ctx := parseGoldenForHelp(&cli, []string{"--help"})
	sections, err := kong.NodeSections(ctx)
	require.NoError(t, err)
	flags, ok := sections[1].Content[0].(help.FlagGroup)
	require.True(t, ok)
	require.Equal(t, "Write output", flags[0].Desc)
	require.Empty(t, flags[0].Default)
}

func TestSubcommands_EnvPrefixPropagation(t *testing.T) {
	var cli struct {
		Run struct {
			Output string `help:"Write output" env:"OUTPUT"`
		} `cmd:"" envprefix:"CLIB_TEST_"`
	}
	t.Setenv("CLIB_TEST_OUTPUT", "example")
	parser, err := konglib.New(&cli)
	require.NoError(t, err)
	_, err = parser.Parse([]string{"run"})
	require.NoError(t, err)
	require.Equal(t, "example", cli.Run.Output)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.Equal(t, []complete.Spec{{
		LongFlag: "output", HasArg: true, Terse: "Write output", Persistent: true,
	}}, subs[0].Specs)
}
