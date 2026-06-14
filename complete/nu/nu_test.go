package nu

import (
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

func flatGen() *complete.Generator {
	return &complete.Generator{AppName: "testapp", Specs: []complete.Spec{
		{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
		{LongFlag: "output", ShortFlag: "o", Terse: "Output path", HasArg: true},
	}}
}

func genSubcommands() *complete.Generator {
	return &complete.Generator{AppName: "testapp", Specs: []complete.Spec{
		{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose"},
	}, Subs: []complete.SubSpec{
		{Name: "build", Terse: "Build the project", Specs: []complete.Spec{
			{LongFlag: "output", ShortFlag: "o", Terse: "Output", HasArg: true},
			{LongFlag: "release", Terse: "Release build"},
		}},
		{Name: "test", Aliases: []string{"t"}, Terse: "Run tests"},
	}}
}

func valuesGen() *complete.Generator {
	return &complete.Generator{AppName: "testapp", Specs: []complete.Spec{
		{
			LongFlag: "format",
			Terse:    "Format",
			HasArg:   true,
			Values:   []string{"json", "yaml", "text"},
		},
		{LongFlag: "level", Terse: "Level", HasArg: true, ValueDescs: []complete.ValueDesc{
			{Value: "info", Desc: "Information"}, {Value: "warn", Desc: "Warning"},
		}},
	}}
}

func hintsGen() *complete.Generator {
	return &complete.Generator{AppName: "testapp", Specs: []complete.Spec{
		{LongFlag: "config", Terse: "Config file", HasArg: true, Extension: "yaml,yml"},
		{LongFlag: "output", Terse: "Output path", HasArg: true, ValueHint: complete.HintFile},
		{LongFlag: "dir", Terse: "Directory", HasArg: true, ValueHint: complete.HintDir},
	}}
}

func dynamicArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp", DynamicArgs: []string{"items", "subitems"},
		Specs: []complete.Spec{{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose"}},
	}
}

// Generate is a thin wrapper over complete.GenerateNu; assert the delegation
// holds across a representative spread of generators. The exact script text is
// locked by the package-level golden tests.
func TestGenerate_DelegatesToComplete(t *testing.T) {
	for name, gen := range map[string]*complete.Generator{
		"flat":        flatGen(),
		"subcommands": genSubcommands(),
		"values":      valuesGen(),
		"hints":       hintsGen(),
		"dynamicargs": dynamicArgsGen(),
	} {
		t.Run(name, func(t *testing.T) {
			out, err := Generate(gen)
			require.NoError(t, err)

			expected, err := complete.GenerateNu(gen)
			require.NoError(t, err)
			require.Equal(t, expected, out)
		})
	}
}

func TestGenerate_Structure(t *testing.T) {
	out, err := Generate(genSubcommands())
	require.NoError(t, err)

	require.Contains(t, out, `extern "testapp" [`)
	require.Contains(t, out, "--verbose(-v)  # Verbose")
	// Each subcommand is its own extern; aliases get a duplicate extern.
	require.Contains(t, out, `extern "testapp build" [`)
	require.Contains(t, out, `extern "testapp test" [`)
	require.Contains(t, out, `extern "testapp t" [`)
}

func TestGenerate_Values(t *testing.T) {
	out, err := Generate(valuesGen())
	require.NoError(t, err)

	require.Contains(t, out, `def "nu-complete testapp format" [] {`)
	require.Contains(t, out, "['json' 'yaml' 'text']")
	// Value descriptions render as records consumed by Nushell's menu.
	require.Contains(t, out, "{value: 'info', description: 'Information'}")
}

func TestGenerate_Hints(t *testing.T) {
	out, err := Generate(hintsGen())
	require.NoError(t, err)

	// File/extension hints map to the native `path` shape; dirs to `directory`.
	require.Contains(t, out, "--config: path")
	require.Contains(t, out, "--output: path")
	require.Contains(t, out, "--dir: directory")
}

func TestGenerate_DynamicArgs(t *testing.T) {
	out, err := Generate(dynamicArgsGen())
	require.NoError(t, err)

	require.Contains(
		t,
		out,
		`def "_testapp_positionals" [context: string, cmdskip: int, valueflags: list<string>] {`,
	)
	require.Contains(t, out, `...rest: string@"nu-complete testapp args"`)
	require.Contains(t, out, `^testapp $"--@complete=($kind)"`)
}

func TestGenerate_Hyphenated(t *testing.T) {
	out, err := Generate(&complete.Generator{
		AppName:     "my-app",
		DynamicArgs: []string{"items"},
	})
	require.NoError(t, err)

	// Hyphens are munged out of Nushell helper names but kept in the command
	// name, the extern name, and the external invocation.
	require.Contains(
		t,
		out,
		`def "_my_app_positionals" [context: string, cmdskip: int, valueflags: list<string>] {`,
	)
	require.Contains(t, out, `extern "my-app" [`)
	require.Contains(t, out, "^my-app $\"--@complete=($kind)\"")
}

func TestGenerate_ErrorOnUnsafeAppName(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "bad;name"})
	require.EqualError(t, err, `AppName contains unsafe characters: "bad;name"`)
}

func TestGenerate_ErrorOnUnsafeDynamic(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "app", DynamicArgs: []string{"bad;arg"}})
	require.EqualError(t, err, `DynamicArgs contains unsafe characters: "bad;arg"`)
}
