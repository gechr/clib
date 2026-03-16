package urfave_test

import (
	"testing"

	urfavecli "github.com/gechr/clib/cli/urfave"
	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
	clilib "github.com/urfave/cli/v3"
)

func TestFlagMeta_Basic(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output format"},
			&clilib.BoolFlag{
				Name:    "verbose",
				Aliases: []string{"v"},
				Usage:   "Enable verbose output",
			},
		},
	}

	flags := urfavecli.FlagMeta(cmd)

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
	repoFlag := &clilib.StringFlag{Name: "repo", Aliases: []string{"R"}, Usage: "Filter by repo"}
	urfavecli.Extend(repoFlag, urfavecli.FlagExtra{
		Group:       "Filters",
		Placeholder: "owner/repo",
		Terse:       "Repository",
		Complete:    "predictor=repo",
	})

	cmd := &clilib.Command{
		Flags: []clilib.Flag{repoFlag},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)

	f := flags[0]
	require.Equal(t, "repo", f.Name)
	require.Equal(t, "Filters", f.Group)
	require.Equal(t, "owner/repo", f.Placeholder)
	require.Equal(t, "Repository", f.Terse)
	require.Equal(t, "predictor=repo", f.Complete)
}

func TestFlagMeta_CSVFlag(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.GenericFlag{
				Name:    "columns",
				Aliases: []string{"c"},
				Usage:   "Table columns",
				Value:   &urfavecli.CSVFlag{},
			},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].IsCSV)
	require.True(t, flags[0].HasArg)
}

func TestFlagMeta_Hidden(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "secret", Usage: "Secret flag", Hidden: true},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].Hidden)
}

func TestFlagMeta_SliceType(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.StringSliceFlag{Name: "tags", Usage: "Tags"},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].IsSlice)
	require.True(t, flags[0].HasArg)
}

func TestFlagMeta_Negatable(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.BoolWithInverseFlag{Name: "draft", Usage: "Include drafts"},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.True(t, flags[0].Negatable)
	require.Equal(t, "draft", flags[0].Name)
}

func TestFlagMeta_NilCommand(t *testing.T) {
	flags := urfavecli.FlagMeta(nil)
	require.Nil(t, flags)
}

func TestFlagMeta_NegatableWithAliases(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.BoolWithInverseFlag{
				Name:    "draft",
				Aliases: []string{"d", "dr"},
				Usage:   "Include drafts",
			},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)

	f := flags[0]
	require.Equal(t, "draft", f.Name)
	require.True(t, f.Negatable)
	require.Equal(t, "d", f.Short)
	require.Equal(t, []string{"dr"}, f.Aliases)
}

func TestFlagMeta_ShortOnlyNames(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.BoolFlag{Name: "v", Usage: "Verbose"},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	// Single-char name: should be set as Name (fallback when no multi-char).
	require.Equal(t, "v", flags[0].Name)
}

func TestFlagMeta_CategoryGroup(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Usage: "Output format", Category: "Display"},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "Display", flags[0].Group)
}

func TestFlagMeta_EnumAndHighlight(t *testing.T) {
	flag := &clilib.StringFlag{Name: "format", Usage: "Format"}
	urfavecli.Extend(flag, urfavecli.FlagExtra{
		Enum:          []string{"json", "yaml"},
		EnumHighlight: []string{"j", "y"},
	})

	cmd := &clilib.Command{
		Flags: []clilib.Flag{flag},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, []string{"json", "yaml"}, flags[0].Enum)
	require.Equal(t, []string{"j", "y"}, flags[0].EnumHighlight)
}

func TestFlagMeta_MultipleAliases(t *testing.T) {
	cmd := &clilib.Command{
		Flags: []clilib.Flag{
			&clilib.StringFlag{
				Name:    "output",
				Aliases: []string{"o", "out", "format"},
				Usage:   "Output format",
			},
		},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "output", flags[0].Name)
	require.Equal(t, "o", flags[0].Short)
	require.Equal(t, []string{"out", "format"}, flags[0].Aliases)
}

func TestFlagMeta_Extension(t *testing.T) {
	configFlag := &clilib.StringFlag{Name: "config", Usage: "Config file"}
	urfavecli.Extend(configFlag, urfavecli.FlagExtra{
		Extension: "yaml",
	})

	cmd := &clilib.Command{
		Flags: []clilib.Flag{configFlag},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "yaml", flags[0].Extension)
}

func TestFlagMeta_Hint(t *testing.T) {
	outputFlag := &clilib.StringFlag{Name: "output", Usage: "Output path"}
	urfavecli.Extend(outputFlag, urfavecli.FlagExtra{
		Hint: "dir",
	})

	cmd := &clilib.Command{
		Flags: []clilib.Flag{outputFlag},
	}

	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "dir", flags[0].ValueHint)
}

func TestFlagMeta_PlaceholderOverride(t *testing.T) {
	repoFlag := &clilib.StringFlag{Name: "repo", Usage: "Repository"}
	urfavecli.Extend(repoFlag, urfavecli.FlagExtra{
		Placeholder: "owner/repo",
	})

	cmd := &clilib.Command{Flags: []clilib.Flag{repoFlag}}
	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.Equal(t, "owner/repo", flags[0].Placeholder)
	require.True(t, flags[0].PlaceholderOverride)
}

func TestFlagMeta_PlaceholderOverride_Empty(t *testing.T) {
	repoFlag := &clilib.StringFlag{Name: "repo", Usage: "Repository"}
	urfavecli.Extend(repoFlag, urfavecli.FlagExtra{
		Terse: "Repository",
	})

	cmd := &clilib.Command{Flags: []clilib.Flag{repoFlag}}
	flags := urfavecli.FlagMeta(cmd)
	require.Len(t, flags, 1)
	require.False(t, flags[0].PlaceholderOverride)
}

func TestFlagMeta_PositiveNegativeDesc(t *testing.T) {
	draftFlag := &clilib.BoolWithInverseFlag{Name: "draft", Usage: "Include drafts"}
	urfavecli.Extend(draftFlag, urfavecli.FlagExtra{
		PositiveDesc: "Include drafts",
		NegativeDesc: "Exclude drafts",
	})

	cmd := &clilib.Command{Flags: []clilib.Flag{draftFlag}}
	flags := urfavecli.FlagMeta(cmd)
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
