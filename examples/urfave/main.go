package main

import (
	"context"
	"fmt"
	"os"

	clib "github.com/gechr/clib/cli/urfave"
	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/examples"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	cli "github.com/urfave/cli/v3"
)

func main() {
	th := theme.New(theme.WithHelpDescBacktick(*theme.Default().Magenta))
	r := help.NewRenderer(th)

	// Filters
	queryFlag := &cli.StringFlag{
		Name:    "query",
		Aliases: []string{"q"},
		Usage:   "Filter results by query",
	}
	clib.Extend(queryFlag, clib.FlagExtra{
		Group: "Filters", Placeholder: "text", Terse: "Query", Complete: "predictor=query",
	})

	categoryFlag := &cli.StringFlag{
		Name:    "category",
		Aliases: []string{"g"},
		Usage:   "Filter by category",
	}
	clib.Extend(categoryFlag, clib.FlagExtra{
		Group: "Filters", Placeholder: "category",
	})

	hiddenFlag := &cli.BoolWithInverseFlag{
		Name:    "hidden",
		Aliases: []string{"H"},
		Usage:   "Include hidden items",
	}
	clib.Extend(hiddenFlag, clib.FlagExtra{Group: "Filters"})

	createdFlag := &cli.StringFlag{
		Name:    "created",
		Aliases: []string{"c"},
		Usage:   "Filter by creation date",
	}
	clib.Extend(createdFlag, clib.FlagExtra{Group: "Filters", Placeholder: "duration"})

	tagFlag := &cli.StringSliceFlag{
		Name:    "tag",
		Aliases: []string{"t"},
		Usage:   "Filter by tag",
	}
	clib.Extend(tagFlag, clib.FlagExtra{Group: "Filters"})

	// Interactive
	selectFlag := &cli.BoolFlag{Name: "select", Usage: "Select matching items interactively"}
	clib.Extend(selectFlag, clib.FlagExtra{Group: "Interactive"})

	deleteFlag := &cli.BoolFlag{Name: "delete", Usage: "Delete matching items"}
	clib.Extend(deleteFlag, clib.FlagExtra{Group: "Interactive"})

	yesFlag := &cli.BoolFlag{
		Name:    "yes",
		Aliases: []string{"y"},
		Usage:   "Skip interactive confirmation prompt",
	}
	clib.Extend(yesFlag, clib.FlagExtra{Group: "Interactive"})

	// Actions
	openFlag := &cli.BoolFlag{
		Name:    "open",
		Aliases: []string{"O"},
		Usage:   "Open matching items in browser",
	}
	clib.Extend(openFlag, clib.FlagExtra{Group: "Actions"})

	previewFlag := &cli.BoolFlag{
		Name:    "preview",
		Aliases: []string{"p"},
		Usage:   "Preview matching items",
	}
	clib.Extend(previewFlag, clib.FlagExtra{Group: "Actions"})

	// Output
	formatFlag := &cli.StringFlag{
		Name:    "format",
		Aliases: []string{"f"},
		Usage:   "Output format",
		Value:   "table",
	}
	clib.Extend(formatFlag, clib.FlagExtra{
		Group:       "Output",
		Placeholder: "format",
		Enum:        []string{"table", "json", "yaml"},
		EnumDefault: "table",
	})

	fieldsFlag := &cli.GenericFlag{
		Name:  "fields",
		Usage: "Fields to show",
		Value: &clib.CSVFlag{},
	}
	clib.Extend(fieldsFlag, clib.FlagExtra{
		Group: "Output", Placeholder: "field", Complete: "predictor=field,comma",
	})

	limitFlag := &cli.IntFlag{
		Name:    "limit",
		Aliases: []string{"L"},
		Usage:   "Maximum results",
		Value:   30, //nolint:mnd // this is fine
	}
	clib.Extend(limitFlag, clib.FlagExtra{Group: "Output", Placeholder: "n"})

	sortFlag := &cli.StringFlag{
		Name:  "sort",
		Usage: "Sort by",
		Value: "name",
	}
	clib.Extend(sortFlag, clib.FlagExtra{
		Group:       "Output",
		Placeholder: "field",
		Enum:        []string{"name", "created", "updated"},
		EnumDefault: "name",
	})

	// Miscellaneous
	debugFlag := &cli.BoolFlag{Name: "debug", Usage: "Log HTTP requests to `stderr`"}
	clib.Extend(debugFlag, clib.FlagExtra{Group: "Miscellaneous"})

	root := &cli.Command{
		Name:      "catalog",
		ArgsUsage: "[<query>...]",
		Flags: []cli.Flag{
			queryFlag,
			categoryFlag,
			hiddenFlag,
			createdFlag,
			tagFlag,
			selectFlag,
			deleteFlag,
			yesFlag,
			openFlag,
			previewFlag,
			formatFlag,
			fieldsFlag,
			limitFlag,
			sortFlag,
			debugFlag,
			// Hide the explicit help flag - WithHelpFlags adds split entries instead.
			&cli.BoolFlag{
				Name:    "help",
				Aliases: []string{"h"},
				Usage:   "Print help",
				Local:   true,
				Hidden:  true,
			},
		},
		Action: func(_ context.Context, _ *cli.Command) error {
			fmt.Println(examples.DemoMessage())
			return nil
		},
		HideHelp: true,
	}

	// Completion flags (hidden).
	comp := clib.NewCompletion(root)

	// Custom help using clib themed renderer, with split -h / --help and long-only examples.
	cli.HelpPrinter = clib.HelpPrinter(r, clib.Sections,
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
	)

	// Handle completions in Before hook.
	root.Before = func(ctx context.Context, cmd *cli.Command) (context.Context, error) {
		gen := complete.NewGenerator("catalog").FromFlags(clib.FlagMeta(cmd))
		gen.Specs = append(gen.Specs,
			complete.Spec{ShortFlag: "h", Terse: "Print short help"},
			complete.Spec{LongFlag: "help", Terse: "Print long help with examples"},
		)
		handled, err := comp.Handle(gen, nil, clib.WithQuiet(false))
		if err != nil {
			return ctx, err
		}
		if handled {
			os.Exit(0)
		}
		return ctx, nil
	}

	if err := root.Run(context.Background(), os.Args); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
