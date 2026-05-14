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

func TestGolden(t *testing.T) {
	r := help.NewRenderer(theme.Default())

	scenarios := map[string][]help.Section{
		"lowercase_placeholders": sectionsLowercasePlaceholders(),
		"preserve_placeholders":  sectionsPreservePlaceholders(),
		"flag_refs_in_backticks": sectionsFlagRefsInBackticks(),
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
