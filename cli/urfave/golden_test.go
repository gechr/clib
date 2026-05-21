package urfave_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	urfavecli "github.com/gechr/clib/cli/urfave"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	"github.com/stretchr/testify/require"
	clilib "github.com/urfave/cli/v3"
)

var update = flag.Bool("update", false, "update golden files")

func sectionsBasic() []help.Section {
	return sectionsPlaceholderCommand(urfavecli.Sections)
}

func sectionsPreservePlaceholders() []help.Section {
	return sectionsPlaceholderCommand(
		urfavecli.SectionsWithOptions(urfavecli.WithPreservePlaceholders()),
	)
}

func sectionsPlaceholderCommand(sections func(*clilib.Command) []help.Section) []help.Section {
	cmd := &clilib.Command{
		Name: "catalog",
		Flags: []clilib.Flag{
			&clilib.StringFlag{
				Name:    "channel",
				Aliases: []string{"c"},
				Usage:   "Channel ID, name, or alias",
			},
			&clilib.IntFlag{
				Name:    "max-items",
				Aliases: []string{"M"},
				Usage:   "Maximum messages to return",
			},
			&clilib.BoolFlag{Name: "verbose", Aliases: []string{"v"}, Usage: "Verbose output"},
		},
	}

	urfavecli.Extend(cmd.Flags[0], urfavecli.FlagExtra{Placeholder: "CHANNEL"})
	urfavecli.Extend(cmd.Flags[1], urfavecli.FlagExtra{Placeholder: "N"})
	return sections(cmd)
}

func sectionsGrouped() []help.Section {
	queryFlag := &clilib.StringFlag{
		Name:    "query",
		Aliases: []string{"q"},
		Usage:   "Filter results by query",
	}
	tagsFlag := &clilib.GenericFlag{
		Name:  "tags",
		Usage: "Filter by tags",
		Value: &urfavecli.CSVFlag{},
	}
	formatFlag := &clilib.StringFlag{
		Name:    "format",
		Aliases: []string{"f"},
		Usage:   "Output format",
		Value:   "table",
	}

	cmd := &clilib.Command{
		Name: "catalog",
		Flags: []clilib.Flag{
			queryFlag,
			tagsFlag,
			formatFlag,
		},
	}

	urfavecli.Extend(queryFlag, urfavecli.FlagExtra{Group: "Filters", Placeholder: "text"})
	urfavecli.Extend(tagsFlag, urfavecli.FlagExtra{Group: "Filters", Placeholder: "tag"})
	urfavecli.Extend(formatFlag, urfavecli.FlagExtra{
		Group:       "Output",
		Placeholder: "format",
		Enum:        []string{"table", "json"},
		EnumDefault: "table",
	})
	return urfavecli.Sections(cmd)
}

// sectionsLongDescription exercises a command-level long description
// (cmd.Description), surfaced as a help.Description blurb below the Usage
// line and rendered by the shared renderer (paragraph breaks, backticks).
func sectionsLongDescription() []help.Section {
	cmd := &clilib.Command{
		Name:      "deploy",
		ArgsUsage: "<env>",
		Usage:     "Deploy the app",
		Description: "Deploy the application to the named environment.\n\n" +
			"Pre-flight checks run before cutover; the previous release\n" +
			"is retained so `--rollback` can restore it on failure.",
		Flags: []clilib.Flag{
			&clilib.BoolFlag{Name: "rollback", Usage: "Restore the previous release"},
		},
	}
	return urfavecli.Sections(cmd)
}

func TestGolden(t *testing.T) {
	r := help.NewRenderer(theme.Default())

	scenarios := map[string][]help.Section{
		"basic":                 sectionsBasic(),
		"grouped":               sectionsGrouped(),
		"preserve_placeholders": sectionsPreservePlaceholders(),
		"long_description":      sectionsLongDescription(),
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
