package main

import (
	"fmt"
	"os"

	clib "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/examples"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	cobra "github.com/spf13/cobra"
)

func main() {
	th := theme.New(theme.WithHelpDescBacktick(*theme.Default().Magenta))
	r := help.NewRenderer(th)

	root := &cobra.Command{
		Use: "catalog [<query>...]",
		RunE: func(_ *cobra.Command, _ []string) error {
			fmt.Println(examples.DemoMessage())
			return nil
		},
	}

	// Define flags with clib extras for grouping and placeholders.
	f := root.Flags()

	// Filters
	f.StringP("query", "q", "", "Filter results by query")
	clib.Extend(f.Lookup("query"), clib.FlagExtra{
		Group: "Filters", Placeholder: "text", Terse: "Query", Complete: "predictor=query",
	})

	f.StringP("category", "g", "", "Filter by category")
	clib.Extend(f.Lookup("category"), clib.FlagExtra{
		Group: "Filters", Placeholder: "category",
	})

	f.BoolP("hidden", "H", false, "Include hidden items")
	clib.Extend(f.Lookup("hidden"), clib.FlagExtra{Group: "Filters"})

	f.StringP("created", "c", "", "Filter by creation date")
	clib.Extend(f.Lookup("created"), clib.FlagExtra{Group: "Filters", Placeholder: "duration"})

	f.StringSliceP("tag", "t", nil, "Filter by tag")
	clib.Extend(f.Lookup("tag"), clib.FlagExtra{Group: "Filters"})

	// Interactive
	f.Bool("select", false, "Select matching items interactively")
	clib.Extend(f.Lookup("select"), clib.FlagExtra{Group: "Interactive"})

	f.Bool("delete", false, "Delete matching items")
	clib.Extend(f.Lookup("delete"), clib.FlagExtra{Group: "Interactive"})

	f.BoolP("yes", "y", false, "Skip interactive confirmation prompt")
	clib.Extend(f.Lookup("yes"), clib.FlagExtra{Group: "Interactive"})

	// Actions
	f.BoolP("open", "O", false, "Open matching items in browser")
	clib.Extend(f.Lookup("open"), clib.FlagExtra{Group: "Actions"})

	f.BoolP("preview", "p", false, "Preview matching items")
	clib.Extend(f.Lookup("preview"), clib.FlagExtra{Group: "Actions"})

	// Output
	f.StringP("format", "f", "table", "Output format")
	clib.Extend(f.Lookup("format"), clib.FlagExtra{
		Group:       "Output",
		Placeholder: "format",
		Enum:        []string{"table", "json", "yaml"},
		EnumDefault: "table",
	})

	f.String("fields", "", "Fields to show (comma-separated)")
	clib.Extend(f.Lookup("fields"), clib.FlagExtra{
		Group: "Output", Placeholder: "field", Complete: "predictor=field,comma",
	})

	f.IntP("limit", "L", 30, "Maximum results") //nolint:mnd // this is fine
	clib.Extend(f.Lookup("limit"), clib.FlagExtra{Group: "Output", Placeholder: "n"})

	f.String("sort", "name", "Sort by")
	clib.Extend(f.Lookup("sort"), clib.FlagExtra{
		Group:       "Output",
		Placeholder: "field",
		Enum:        []string{"name", "created", "updated"},
		EnumDefault: "name",
	})

	// Miscellaneous
	f.Bool("debug", false, "Log HTTP requests to `stderr`")
	clib.Extend(f.Lookup("debug"), clib.FlagExtra{Group: "Miscellaneous"})

	// Hide the auto-generated help flag - WithHelpFlags adds split entries instead.
	f.BoolP("help", "h", false, "Print help")
	f.Lookup("help").Hidden = true

	// Completion flags (hidden).
	comp := clib.NewCompletion(root)

	// Auto-grouped help using extras, with split -h / --help and long-only examples.
	root.SetHelpFunc(clib.HelpFunc(r, clib.Sections,
		help.WithHelpFlags("Print short help", "Print long help with examples"),
		help.WithLongHelp(os.Args, help.Section{
			Title: "Examples",
			Content: []help.Content{
				help.Examples{
					{Comment: "List matching items", Command: "catalog"},
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
	))

	// Handle completions before running.
	root.PersistentPreRunE = func(_ *cobra.Command, _ []string) error {
		gen := complete.NewGenerator("catalog").FromFlags(clib.FlagMeta(root))
		gen.Specs = append(gen.Specs,
			complete.Spec{ShortFlag: "h", Terse: "Print short help"},
			complete.Spec{LongFlag: "help", Terse: "Print long help with examples"},
		)
		handled, err := comp.Handle(gen, nil, clib.WithQuiet(false))
		if err != nil {
			return err
		}
		if handled {
			os.Exit(0)
		}
		return nil
	}

	if err := root.Execute(); err != nil {
		os.Exit(1)
	}
}
