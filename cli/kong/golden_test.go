package kong_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/cli/kong"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func sectionsBasic() []help.Section {
	type CLI struct {
		Channel string `name:"channel"        help:"Channel ID, name, or alias" short:"c" placeholder:"CHANNEL"`
		Max     int    `name:"max-items"      help:"Maximum messages to return" short:"M" placeholder:"N"`
		Verbose bool   `help:"Verbose output" short:"v"`
	}

	ctx := parseGoldenForHelp(&CLI{}, []string{"--help"}, konglib.Name("catalog"))
	sections, err := kong.NodeSections(ctx)
	if err != nil {
		panic(err)
	}
	return sections
}

func sectionsGrouped() []help.Section {
	type CLI struct {
		Query  string       `name:"query"  help:"Filter results by query" short:"q"              clib:"group='Filters'" placeholder:"text"`
		Tags   kong.CSVFlag `name:"tags"   help:"Filter by tags"          clib:"group='Filters'" placeholder:"<tag>"`
		Format string       `name:"format" help:"Output format"           short:"f"              clib:"group='Output'"  default:"table"    enum:"table,json" placeholder:"format"`
	}

	ctx := parseGoldenForHelp(&CLI{}, []string{"--help"}, konglib.Name("catalog"))
	sections, err := kong.NodeSections(ctx)
	if err != nil {
		panic(err)
	}
	return sections
}

// inspectCmd implements kong's HelpProvider so the Description section is
// populated from Help() and we can exercise the smart backtick styling
// against the command's own args.
type inspectCmd struct {
	Name    string `help:"Widget name to inspect" arg:""    optional:""`
	Verbose bool   `help:"Show all fields"        short:"v"`
}

func (inspectCmd) Help() string {
	return `Without ` + "`<name>`" + `, lists every known widget.

With ` + "`<name>`" + `, prints that widget's details. Pass ` + "`--verbose`" + ` to
include all fields, or use the ` + "`remove`" + ` command to delete it.`
}

func sectionsDescription() []help.Section {
	type CLI struct {
		Inspect inspectCmd `help:"Inspect a widget" cmd:""`
		Remove  struct {
			Name string `help:"Widget name" arg:""`
		} `help:"Remove a widget"  cmd:""`
	}

	ctx := parseGoldenForHelp(
		&CLI{},
		[]string{"inspect", "--help"},
		konglib.Name("widgets"),
	)
	sections, err := kong.NodeSections(ctx)
	if err != nil {
		panic(err)
	}
	return sections
}

func parseGoldenForHelp(cli any, args []string, opts ...konglib.Option) *konglib.Context {
	var captured *konglib.Context
	printer := func(_ konglib.HelpOptions, ctx *konglib.Context) error {
		captured = ctx
		return nil
	}

	defaults := []konglib.Option{
		konglib.Writers(os.Stdout, os.Stderr),
		konglib.Help(printer),
		konglib.Exit(func(int) {}),
	}
	defaults = append(defaults, opts...)

	k, err := konglib.New(cli, defaults...)
	if err != nil {
		panic(err)
	}
	_, _ = k.Parse(args)
	if captured == nil {
		panic("help printer was not invoked")
	}
	return captured
}

func TestGolden(t *testing.T) {
	r := help.NewRenderer(theme.Dark())

	scenarios := map[string][]help.Section{
		"basic":       sectionsBasic(),
		"grouped":     sectionsGrouped(),
		"description": sectionsDescription(),
	}

	for name, sections := range scenarios {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, r.Render(&buf, sections))

			got := buf.String()
			goldenFile := filepath.Join("testdata", name+".golden")

			if *update {
				require.NoError(t, os.WriteFile(goldenFile, []byte(got), 0o644))
				return
			}

			want, err := os.ReadFile(goldenFile)
			require.NoError(t, err, "golden file missing; run with -update to create")
			require.Equal(t, string(want), got)
		})
	}
}
