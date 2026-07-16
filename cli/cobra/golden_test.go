package cobra_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	cobracli "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/help"
	"github.com/gechr/clib/theme"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func sectionsLowercasePlaceholders() []help.Section {
	cmd := &cobralib.Command{Use: "catalog"}
	cmd.Flags().StringP("channel", "c", "", "Channel `CHANNEL`, name, or alias")
	cmd.Flags().IntP("max-items", "M", 0, "Maximum messages to return")
	cobracli.Extend(cmd.Flags().Lookup("max-items"), cobracli.FlagExtra{
		Placeholder: "N",
	})
	return cobracli.Sections(cmd)
}

func sectionsPreservePlaceholders() []help.Section {
	cmd := &cobralib.Command{Use: "catalog"}
	cmd.Flags().StringP("channel", "c", "", "Channel `CHANNEL`, name, or alias")
	cmd.Flags().IntP("max-items", "M", 0, "Maximum messages to return")
	cobracli.Extend(cmd.Flags().Lookup("max-items"), cobracli.FlagExtra{
		Placeholder: "N",
	})
	return cobracli.SectionsWithOptions(cobracli.WithPreservePlaceholders())(cmd)
}

// sectionsFlagRefsInBackticks exercises backticked content in flag
// descriptions. Backticked flag references (`--foo`) and backticked
// identifiers on bool flags must survive to the renderer as inline code
// markers, never as value placeholders. Non-bool flags continue to honour
// the standard pflag placeholder convention.
func sectionsFlagRefsInBackticks() []help.Section {
	cmd := &cobralib.Command{Use: "run"}
	cmd.Flags().Bool("alpha", false, "Alias for `--foo=0`")
	cmd.Flags().Bool("bravo", false, "Print totals (implies `--alpha`)")
	cmd.Flags().Bool("charlie", false, "Use the `embedded` backend")
	cmd.Flags().String("delta", "", "User `name` to set")
	return cobracli.Sections(cmd)
}

// sectionsLongDescription exercises a command-level long description
// (cmd.Long), surfaced as a help.Description blurb below the Usage line and
// rendered by the shared renderer (paragraph breaks, backtick styling).
func sectionsLongDescription() []help.Section {
	cmd := &cobralib.Command{
		Use:   "deploy <env>",
		Short: "Deploy the app",
		Long: "Deploy the application to the named environment.\n\n" +
			"Pre-flight checks run before cutover; the previous release\n" +
			"is retained so `--rollback` can restore it on failure.",
	}
	cmd.Flags().Bool("rollback", false, "Restore the previous release")
	return cobracli.Sections(cmd)
}

// sectionsEnumRefs exercises cross-referenced enum values: ValidArgs become
// the <provider> positional's enum set, so `github` in the command's long
// description renders in the positional-arg color; a flag enum value (`debug`)
// renders in the flag color. The reference lives in cmd.Long rather than a
// value flag's usage string because pflag would otherwise consume the first
// backticked word as the flag's value placeholder.
func sectionsEnumRefs() []help.Section {
	cmd := &cobralib.Command{
		Use:       "login <provider>",
		Short:     "Authenticate with a provider",
		Long:      "Authenticate with a provider.\n\nThe `github` provider uses an OAuth device flow.",
		ValidArgs: []string{"github", "gitlab", "gitea"},
	}
	cmd.Flags().String("log-level", "", "Logging verbosity")
	cobracli.Extend(cmd.Flags().Lookup("log-level"), cobracli.FlagExtra{
		Enum: []string{"debug", "info", "warn"},
	})
	cmd.Flags().Bool("verbose", false, "Shorthand for `debug` logging")
	return cobracli.Sections(cmd)
}

// sectionsNegatable exercises the three negatable renderings: the default
// bracketed [no-] form, a PositiveOnly extra advertising only --prerelease,
// and a NegativeOnly extra advertising only --no-cache.
func sectionsNegatable() []help.Section {
	cmd := &cobralib.Command{Use: "clover"}
	cmd.Flags().Bool("downgrade", false, "Allow selecting versions older than the current one")
	cmd.Flags().Bool("prerelease", false, "Allow selecting prerelease versions")
	cmd.Flags().Bool("cache", true, "Reuse cached HTTP responses across runs")
	cobracli.Extend(cmd.Flags().Lookup("downgrade"), cobracli.FlagExtra{
		Negatable: true,
	})
	cobracli.Extend(cmd.Flags().Lookup("prerelease"), cobracli.FlagExtra{
		Negatable:    true,
		PositiveOnly: true,
	})
	cobracli.Extend(cmd.Flags().Lookup("cache"), cobracli.FlagExtra{
		Negatable:    true,
		NegativeOnly: true,
	})
	return cobracli.Sections(cmd)
}

func TestGolden(t *testing.T) {
	r := help.NewRenderer(theme.Dark())

	scenarios := map[string][]help.Section{
		"lowercase_placeholders": sectionsLowercasePlaceholders(),
		"preserve_placeholders":  sectionsPreservePlaceholders(),
		"flag_refs_in_backticks": sectionsFlagRefsInBackticks(),
		"long_description":       sectionsLongDescription(),
		"enum_refs":              sectionsEnumRefs(),
		"negatable":              sectionsNegatable(),
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
