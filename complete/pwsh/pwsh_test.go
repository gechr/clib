package pwsh

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

// Generate is a thin wrapper over complete.GeneratePwsh; assert the delegation
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

			expected, err := complete.GeneratePwsh(gen)
			require.NoError(t, err)
			require.Equal(t, expected, out)
		})
	}
}

func TestGenerate_Structure(t *testing.T) {
	out, err := Generate(genSubcommands())
	require.NoError(t, err)

	require.Contains(t, out, "Register-ArgumentCompleter -Native -CommandName 'testapp'")
	require.Contains(t, out, "using namespace System.Management.Automation")
	// Aliases canonicalize to the primary subcommand path.
	require.Contains(t, out, "'testapp;t' { 'testapp;test'; break }")
	// Value-flag completion routes through the previous token.
	require.Contains(t, out, "switch ($prev)")
}

func TestGenerate_Values(t *testing.T) {
	out, err := Generate(valuesGen())
	require.NoError(t, err)

	require.Contains(
		t,
		out,
		"[CompletionResult]::new('json', 'json', [CompletionResultType]::ParameterValue, 'json')",
	)
	// Empty terse falls back to the value so CompletionResult never gets an
	// empty tooltip.
	require.Contains(
		t,
		out,
		"[CompletionResult]::new('info', 'info', [CompletionResultType]::ParameterValue, 'Information')",
	)
}

func TestGenerate_Hints(t *testing.T) {
	out, err := Generate(hintsGen())
	require.NoError(t, err)

	require.Contains(t, out, "[CompletionCompleters]::CompleteFilename($wordToComplete)")
	require.Contains(t, out, `$_.ListItemText -match '\.(yaml|yml)$'`)
	require.Contains(t, out, "$_.ResultType -eq [CompletionResultType]::ProviderContainer")
}

func TestGenerate_DynamicArgs(t *testing.T) {
	out, err := Generate(dynamicArgsGen())
	require.NoError(t, err)

	require.Contains(t, out, "function __testapp_Tokens {")
	require.Contains(t, out, "function __testapp_Positionals {")
	require.Contains(t, out, `$callArgs = @("--@complete=$kind")`)
}

func TestGenerate_Hyphenated(t *testing.T) {
	out, err := Generate(&complete.Generator{
		AppName:     "my-app",
		DynamicArgs: []string{"items"},
	})
	require.NoError(t, err)

	// Hyphens are stripped from PowerShell function names but kept in the
	// command name and call operator.
	require.Contains(t, out, "function __my_app_Tokens {")
	require.Contains(t, out, "Register-ArgumentCompleter -Native -CommandName 'my-app'")
	require.Contains(t, out, "& 'my-app' @callArgs")
}

func TestGenerate_ErrorOnUnsafeAppName(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "bad;name"})
	require.EqualError(t, err, `AppName contains unsafe characters: "bad;name"`)
}

func TestGenerate_ErrorOnUnsafeDynamic(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "app", DynamicArgs: []string{"bad;arg"}})
	require.EqualError(t, err, `DynamicArgs contains unsafe characters: "bad;arg"`)
}
