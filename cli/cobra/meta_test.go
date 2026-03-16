package cobra_test

import (
	"testing"

	cobracli "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestFlagMeta_Basic(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringP("output", "o", "", "Output format")
	cmd.Flags().BoolP("verbose", "v", false, "Enable verbose output")

	flags := cobracli.FlagMeta(cmd)

	require.Len(t, flags, 2)

	output := findMeta(flags, "output")
	require.NotNil(t, output)
	require.Equal(t, "o", output.Short)
	require.Equal(t, "Output format", output.Help)
	require.True(t, output.HasArg)
	require.False(t, output.Hidden)

	verbose := findMeta(flags, "verbose")
	require.NotNil(t, verbose)
	require.Equal(t, "v", verbose.Short)
	require.False(t, verbose.HasArg)
}

func TestFlagMeta_Annotations(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringP("repo", "R", "", "Filter by repo")
	cobracli.Extend(cmd.Flags().Lookup("repo"), cobracli.FlagExtra{
		Group:       "Filters",
		Placeholder: "owner/repo",
		Terse:       "Repository",
		Complete:    "predictor=repo",
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)

	f := flags[0]
	require.Equal(t, "repo", f.Name)
	require.Equal(t, "Filters", f.Group)
	require.Equal(t, "owner/repo", f.Placeholder)
	require.Equal(t, "Repository", f.Terse)
	require.Equal(t, "predictor=repo", f.Complete)
}

func TestFlagMeta_CSVFlag(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}

	csv := &cobracli.CSVFlag{}
	cmd.Flags().VarP(csv, "columns", "c", "Table columns")

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].IsCSV)
	require.True(t, flags[0].HasArg)
}

func TestFlagMeta_Hidden(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("secret", "", "Secret flag")
	_ = cmd.Flags().MarkHidden("secret")

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].Hidden)
}

func TestFlagMeta_PersistentFlags(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("local", "", "Local flag")
	cmd.PersistentFlags().Bool("verbose", false, "Verbose output")

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 2)

	local := findMeta(flags, "local")
	require.NotNil(t, local)
	require.False(t, local.Persistent)

	verbose := findMeta(flags, "verbose")
	require.NotNil(t, verbose)
	require.True(t, verbose.Persistent)
}

func TestFlagMeta_SliceType(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().StringSlice("tags", nil, "Tags")

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].IsSlice)
	require.True(t, flags[0].HasArg)
}

func TestFlagMeta_Negatable(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().Bool("draft", false, "Include drafts")
	cobracli.Extend(cmd.Flags().Lookup("draft"), cobracli.FlagExtra{
		Negatable: true,
		Group:     "Filters",
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].Negatable)
	require.Equal(t, "Filters", flags[0].Group)
}

func TestFlagMeta_NilCommand(t *testing.T) {
	flags := cobracli.FlagMeta(nil)
	require.Nil(t, flags)
}

func TestFlagMeta_EnumAnnotation(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("state", "open", "PR state")
	cobracli.Extend(cmd.Flags().Lookup("state"), cobracli.FlagExtra{
		Enum: []string{"open", "closed", "merged"},
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, []string{"open", "closed", "merged"}, flags[0].Enum)
}

func TestFlagMeta_HighlightAnnotation(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("state", "open", "PR state")
	cobracli.Extend(cmd.Flags().Lookup("state"), cobracli.FlagExtra{
		Enum:          []string{"open", "closed"},
		EnumHighlight: []string{"o", "c"},
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, []string{"open", "closed"}, flags[0].Enum)
	require.Equal(t, []string{"o", "c"}, flags[0].EnumHighlight)
}

func TestFlagMeta_NoAnnotations(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("name", "", "Your name")

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Empty(t, flags[0].Enum)
	require.Empty(t, flags[0].EnumHighlight)
	require.Empty(t, flags[0].Group)
	require.Empty(t, flags[0].Complete)
	require.Empty(t, flags[0].Terse)
	require.Empty(t, flags[0].Placeholder)
	require.False(t, flags[0].Negatable)
}

func TestFlagMeta_Extension(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("config", "", "Config file")
	cobracli.Extend(cmd.Flags().Lookup("config"), cobracli.FlagExtra{
		Extension: "yaml",
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "yaml", flags[0].Extension)
}

func TestFlagMeta_Hint(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("output", "", "Output path")
	cobracli.Extend(cmd.Flags().Lookup("output"), cobracli.FlagExtra{
		Hint: "file",
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "file", flags[0].ValueHint)
}

func TestFlagMeta_Aliases(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("output", "", "Output path")
	cobracli.Extend(cmd.Flags().Lookup("output"), cobracli.FlagExtra{
		Aliases: []string{"out", "o"},
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, []string{"out", "o"}, flags[0].Aliases)
}

func TestFlagMeta_PlaceholderOverride(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("repo", "", "Repository")
	cobracli.Extend(cmd.Flags().Lookup("repo"), cobracli.FlagExtra{
		Placeholder: "owner/repo",
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "owner/repo", flags[0].Placeholder)
	require.True(t, flags[0].PlaceholderOverride)
}

func TestFlagMeta_PlaceholderOverride_Empty(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().String("repo", "", "Repository")
	cobracli.Extend(cmd.Flags().Lookup("repo"), cobracli.FlagExtra{
		Terse: "Repository",
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.False(t, flags[0].PlaceholderOverride)
}

func TestFlagMeta_PositiveNegativeDesc(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	cmd.Flags().Bool("draft", false, "Include drafts")
	cobracli.Extend(cmd.Flags().Lookup("draft"), cobracli.FlagExtra{
		Negatable:    true,
		PositiveDesc: "Include drafts",
		NegativeDesc: "Exclude drafts",
	})

	flags := cobracli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "Include drafts", flags[0].PositiveDesc)
	require.Equal(t, "Exclude drafts", flags[0].NegativeDesc)
}

func findMeta(flags []complete.FlagMeta, name string) *complete.FlagMeta {
	for i := range flags {
		if flags[i].Name == name {
			return &flags[i]
		}
	}
	return nil
}
