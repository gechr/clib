package cobra_test

import (
	"bytes"
	"flag"
	"os"
	"path/filepath"
	"testing"

	"github.com/charmbracelet/x/ansi"
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

func TestGolden(t *testing.T) {
	r := help.NewRenderer(theme.Default())

	scenarios := map[string][]help.Section{
		"lowercase_placeholders": sectionsLowercasePlaceholders(),
		"preserve_placeholders":  sectionsPreservePlaceholders(),
	}

	for name, sections := range scenarios {
		t.Run(name, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, r.Render(&buf, sections))

			got := ansi.Strip(buf.String())
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
