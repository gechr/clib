package complete_test

import (
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

var update = flag.Bool("update", false, "update golden files")

func genDynamicArgs() *complete.Generator {
	return &complete.Generator{
		AppName:     "myapp",
		DynamicArgs: []string{"items", "subitems"},
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func genEnum() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: complete.SortVisibleSpecs(
			append(
				complete.SpecsFromFlagMeta(complete.FlagMeta{
					Name: "method", Terse: "Clone method", HasArg: true,
					Enum: []string{"ssh", "https"},
				}),
				complete.SpecsFromFlagMeta(complete.FlagMeta{
					Name: "verbose", Short: "v", Terse: "Verbose output",
				})...,
			),
		),
	}
}

func genEnumSlice() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: complete.SortVisibleSpecs(
			complete.SpecsFromFlagMeta(complete.FlagMeta{
				Name: "verbose", Short: "v", Terse: "Verbose output",
			}),
		),
		Subs: []complete.SubSpec{
			{
				Name:  "list",
				Terse: "List items",
				Specs: complete.SortVisibleSpecs(
					append(
						complete.SpecsFromFlagMeta(complete.FlagMeta{
							Name: "include", Terse: "Fields to include", HasArg: true,
							IsSlice:   true,
							Enum:      []string{"name", "email", "avatar"},
							EnumTerse: []string{"Display name", "Email address", "Profile picture"},
						}),
						complete.SpecsFromFlagMeta(complete.FlagMeta{
							Name: "limit", Terse: "Max results", HasArg: true,
						})...,
					),
				),
			},
			{
				Name:  "get",
				Terse: "Get item",
			},
		},
	}
}

func genExt() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{
				LongFlag:  "config",
				ShortFlag: "c",
				Terse:     "Config file",
				HasArg:    true,
				Extension: "yaml",
			},
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func genHide() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: complete.SortVisibleSpecs(
			append(
				complete.SpecsFromFlagMeta(complete.FlagMeta{
					Name: "include-pattern", Short: "i", Terse: "Filter by regex",
					HasArg: true, HideLong: true,
				}),
				append(
					complete.SpecsFromFlagMeta(complete.FlagMeta{
						Name: "include", Terse: "Include by name", HasArg: true,
					}),
					complete.SpecsFromFlagMeta(complete.FlagMeta{
						Name: "verbose", Short: "v", Terse: "Verbose output", HideShort: true,
					})...,
				)...,
			),
		),
	}
}

func genHints() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{
				LongFlag:  "config",
				ShortFlag: "c",
				Terse:     "Config file",
				HasArg:    true,
				Extension: "yaml,yml",
			},
			{
				LongFlag:  "output",
				ShortFlag: "o",
				Terse:     "Output path",
				HasArg:    true,
				ValueHint: complete.HintFile,
			},
			{
				LongFlag:  "dir",
				ShortFlag: "d",
				Terse:     "Directory",
				HasArg:    true,
				ValueHint: complete.HintDir,
			},
			{
				LongFlag:  "shell",
				Terse:     "Shell command",
				HasArg:    true,
				ValueHint: complete.HintCommand,
			},
			{LongFlag: "user", Terse: "User name", HasArg: true, ValueHint: complete.HintUser},
			{LongFlag: "host", Terse: "Host name", HasArg: true, ValueHint: complete.HintHost},
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func genHyphenated() *complete.Generator {
	return &complete.Generator{
		AppName: "my-app",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "build",
				Terse: "Build the project",
				Specs: []complete.Spec{
					{LongFlag: "output", ShortFlag: "o", Terse: "Output path", HasArg: true},
				},
			},
			{
				Name:  "test",
				Terse: "Run tests",
			},
		},
	}
}

func genPathArgs() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
		Subs: []complete.SubSpec{
			{
				Name:     "edit",
				Terse:    "Edit files",
				PathArgs: true,
				Specs: []complete.Spec{
					{LongFlag: "editor", Terse: "Editor command", HasArg: true},
				},
			},
			{
				Name:  "list",
				Terse: "List items",
			},
		},
	}
}

func genPersistentFlags() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "root-local", Terse: "Root local"},
			{LongFlag: "root-persistent", Terse: "Root persistent", Persistent: true},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "parent",
				Terse: "Parent command",
				Specs: []complete.Spec{
					{LongFlag: "parent-local", Terse: "Parent local"},
					{LongFlag: "parent-persistent", Terse: "Parent persistent", Persistent: true},
				},
				Subs: []complete.SubSpec{
					{
						Name:  "child",
						Terse: "Child command",
						Specs: []complete.Spec{
							{LongFlag: "child-local", Terse: "Child local"},
						},
					},
				},
			},
		},
	}
}

func genValueDesc() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{
				LongFlag:  "format",
				ShortFlag: "f",
				Terse:     "Output format",
				HasArg:    true,
				ValueDescs: []complete.ValueDesc{
					{Value: "json", Desc: "JSON output"},
					{Value: "yaml", Desc: "YAML output"},
					{Value: "text", Desc: "Plain text"},
				},
			},
			{
				LongFlag:  "tags",
				ShortFlag: "t",
				Terse:     "Filter tags",
				HasArg:    true,
				CommaList: true,
				Values:    []string{"bug", "feature", "docs"},
			},
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func genSharedFlags() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Subs: []complete.SubSpec{
			{
				Name:  "list",
				Terse: "List items",
				Specs: []complete.Spec{
					// --include: same values in both subcommands → shared function
					{
						LongFlag:  "include",
						Terse:     "Fields to include",
						HasArg:    true,
						CommaList: true,
						Values:    []string{"name", "email"},
					},
					// --status: different values per subcommand → path-scoped functions
					{
						LongFlag:  "status",
						Terse:     "Filter by status",
						HasArg:    true,
						CommaList: true,
						Values:    []string{"active", "inactive"},
					},
					{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
				},
			},
			{
				Name:  "get",
				Terse: "Get item",
				Specs: []complete.Spec{
					{
						LongFlag:  "include",
						Terse:     "Fields to include",
						HasArg:    true,
						CommaList: true,
						Values:    []string{"name", "email"},
					},
					{
						LongFlag:  "status",
						Terse:     "Filter by status",
						HasArg:    true,
						CommaList: true,
						Values:    []string{"draft", "published"},
					},
				},
			},
		},
	}
}

func genSubDynamicArgs() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
		Subs: []complete.SubSpec{
			{
				Name:        "resolve",
				Terse:       "Resolve alerts",
				DynamicArgs: []string{"incident", "alert"},
				Specs: []complete.Spec{
					{LongFlag: "force", ShortFlag: "f", Terse: "Force resolution"},
				},
			},
			{
				Name:  "list",
				Terse: "List items",
			},
		},
	}
}

func genCollidingSubDynamicArgs() *complete.Generator {
	return &complete.Generator{
		AppName: "pdc",
		Subs: []complete.SubSpec{
			{
				Name:  "incident",
				Terse: "Manage incidents",
				Subs: []complete.SubSpec{
					{
						Name:        "show",
						Terse:       "Show an incident",
						DynamicArgs: []string{"incident"},
					},
				},
			},
			{
				Name:  "user",
				Terse: "Manage users",
				Subs: []complete.SubSpec{
					{
						Name:        "show",
						Terse:       "Show a user",
						DynamicArgs: []string{"user"},
					},
				},
			},
		},
	}
}

func TestGolden(t *testing.T) {
	scenarios := map[string]*complete.Generator{
		"collidingsubdynamicargs": genCollidingSubDynamicArgs(),
		"dynamicargs":             genDynamicArgs(),
		"enum":                    genEnum(),
		"enumslice":               genEnumSlice(),
		"ext":                     genExt(),
		"flat":                    genFlat(),
		"globalflags":             genGlobalFlags(),
		"hide":                    genHide(),
		"hints":                   genHints(),
		"hyphenated":              genHyphenated(),
		"nested":                  genNested(),
		"pathargs":                genPathArgs(),
		"persistentflags":         genPersistentFlags(),
		"sharedflags":             genSharedFlags(),
		"subcommands":             genSubcommands(),
		"subdynamicargs":          genSubDynamicArgs(),
		"valuedesc":               genValueDesc(),
	}

	shells := []string{"bash", "zsh", "fish"}

	for name, gen := range scenarios {
		for _, sh := range shells {
			t.Run(name+"."+sh, func(t *testing.T) {
				var buf strings.Builder
				err := gen.Print(&buf, sh)
				require.NoError(t, err)

				got := buf.String()
				goldenFile := filepath.Join("testdata", sh, name+".golden")

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
}
