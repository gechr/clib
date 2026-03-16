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

func hyphenatedGen() *complete.Generator {
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

func extGen() *complete.Generator {
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

func hintsGen() *complete.Generator {
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

func valueDescGen() *complete.Generator {
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

func pathArgsGen() *complete.Generator {
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

func dynamicArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName:     "myapp",
		DynamicArgs: []string{"items", "subitems"},
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		},
	}
}

func persistentFlagsGen() *complete.Generator {
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

func TestGolden(t *testing.T) {
	scenarios := map[string]*complete.Generator{
		"dynamicargs":     dynamicArgsGen(),
		"ext":             extGen(),
		"flat":            newTestGen(),
		"globalflags":     globalFlagsGen(),
		"hints":           hintsGen(),
		"hyphenated":      hyphenatedGen(),
		"nested":          nestedGen(),
		"pathargs":        pathArgsGen(),
		"persistentflags": persistentFlagsGen(),
		"subcommands":     subcommandGen(),
		"valuedesc":       valueDescGen(),
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
