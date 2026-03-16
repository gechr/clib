package main

import (
	"fmt"
	"os"

	kong "github.com/alecthomas/kong"
	clib "github.com/gechr/clib/cli/kong"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/examples"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
)

type CLI struct {
	clib.CompletionFlags

	Query []string `help:"Search query" arg:"" optional:""`

	// Filters
	Filter   string   `name:"query"    help:"Filter results by query" short:"q" clib:"group='Filters',terse='Query',placeholder='text',complete='predictor=query'"`
	Category string   `name:"category" help:"Filter by category"      short:"g" clib:"group='Filters',terse='Category',placeholder='category'"`
	Hidden   *bool    `name:"hidden"   help:"Include hidden items"    short:"H" clib:"group='Filters',terse='Hidden filter'"                                       negatable:""`
	Created  string   `name:"created"  help:"Filter by creation date" short:"c" clib:"group='Filters',terse='Creation date',placeholder='duration'"`
	Tag      []string `name:"tag"      help:"Filter by tag"           short:"t" clib:"group='Filters',terse='Tag'"`

	// Interactive
	Select bool `name:"select" help:"Select matching items interactively" clib:"group='Interactive',terse='Select'"`
	Delete bool `name:"delete" help:"Delete matching items"               clib:"group='Interactive',terse='Delete'"`
	Yes    bool `name:"yes"    help:"Skip confirmation"                   short:"y"                                 clib:"group='Interactive',terse='Skip confirmation'"`

	// Actions
	Open    bool `name:"open"    help:"Open matching items in browser" short:"O" clib:"group='Actions',terse='Open in browser'"`
	Preview bool `name:"preview" help:"Preview matching items"         short:"p" clib:"group='Actions',terse='Preview'"`

	// Output
	Format string       `name:"format" help:"Output format"     short:"f"                                                                                 clib:"group='Output',terse='Output format',placeholder='format'" default:"table"             enum:"table,json,yaml"`
	Fields clib.CSVFlag `name:"fields" help:"Fields to show"    clib:"group='Output',terse='Fields',placeholder='field',complete='predictor=field,comma'"`
	Color  string       `name:"color"  help:"Color output mode" clib:"group='Output',terse='Color mode',placeholder='color'"                              default:"auto"                                                   enum:"auto,always,never"`
	Limit  int          `name:"limit"  help:"Maximum results"   short:"L"                                                                                 clib:"group='Output',terse='Max results',placeholder='n'"`
	Sort   string       `name:"sort"   help:"Sort by"           clib:"group='Output',terse='Sort field',placeholder='field'"                              default:"name"                                                   enum:"name,created,updated"`

	// Miscellaneous
	Debug bool `name:"debug" help:"Log HTTP requests to stderr" clib:"group='Miscellaneous',terse='Debug mode'"`
}

func main() {
	th := theme.New(
		theme.WithHelpDescBacktick(*theme.Default().Magenta),
	)
	r := help.NewRenderer(th)

	var cli CLI
	flags := clib.Reflect(&cli)

	k := kong.Must(&cli,
		kong.Name("catalog"),
		kong.Help(clib.HelpPrinter(r, func() []help.Section {
			args := clib.Args(&cli)

			sections := []help.Section{
				{
					Title: "Usage",
					Content: []help.Content{
						help.Usage{
							Command:     "catalog",
							ShowOptions: true,
							Args:        args,
						},
					},
				},
			}

			sections = append(sections, clib.FlagSections(flags)...)
			return sections
		},
			help.WithHelpFlags("Print short help", "Print long help with examples"),
			help.WithLongHelp(os.Args, help.Section{
				Title: "Examples",
				Content: []help.Content{
					help.Examples{
						{
							Comment: "List matching items",
							Command: "catalog",
						},
						{
							Comment: "List items in JSON format",
							Command: "catalog --format json",
						},
						{
							Comment: "Show only selected fields",
							Command: "catalog --fields name,status,updated",
						},
						{
							Comment: "Search for items matching 'status page'",
							Command: "catalog status page",
						},
					},
				},
			}),
		)),
	)

	_, parseErr := k.Parse(os.Args[1:])

	// Handle completions before checking parse errors,
	// since tab completion may produce incomplete args.
	gen := complete.NewGenerator("catalog").FromFlags(flags)
	gen.Specs = append(gen.Specs,
		complete.Spec{ShortFlag: "h", Terse: "Print short help"},
		complete.Spec{LongFlag: "help", Terse: "Print long help with examples"},
	)
	handled, err := cli.Handle(gen, nil)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	if handled {
		os.Exit(0)
	}

	if parseErr != nil {
		k.FatalIfErrorf(parseErr)
	}

	fmt.Println(examples.DemoMessage())
}
