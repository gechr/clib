package elvish

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

// Generate is a thin wrapper over complete.GenerateElvish; assert the delegation
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

			expected, err := complete.GenerateElvish(gen)
			require.NoError(t, err)
			require.Equal(t, expected, out)
		})
	}
}

func TestGenerate_Structure(t *testing.T) {
	out, err := Generate(genSubcommands())
	require.NoError(t, err)

	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "set edit:completion:arg-completer['testapp'] = {|@words|")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "use str")
	// Aliases canonicalize to the primary subcommand path alongside the name.
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "(or (eq $w 'test') (eq $w 't'))) { set next = 'testapp;test' }")
	// Value-flag completion routes through the previous token.
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "if (or (eq $prev '--output') (eq $prev '-o')) {")
}

func TestGenerate_Values(t *testing.T) {
	out, err := Generate(valuesGen())
	require.NoError(t, err)

	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "put 'json'")
	// Empty terse falls back to a bare put; descriptions use complex-candidate.
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "edit:complex-candidate 'info' &display='info  Information'")
}

func TestGenerate_Hints(t *testing.T) {
	out, err := Generate(hintsGen())
	require.NoError(t, err)

	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "edit:complete-filename $cur")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "edit:complete-dirname $cur")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "str:has-suffix $s '.yaml'")
}

func TestGenerate_DynamicArgs(t *testing.T) {
	out, err := Generate(dynamicArgsGen())
	require.NoError(t, err)

	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "fn _testapp_positionals {|cmdskip valueflags @tokens|")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "var callargs = ['--@complete='$kind]")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "(external 'testapp') $@callargs 2>/dev/null | from-lines")
}

func TestGenerate_Hyphenated(t *testing.T) {
	out, err := Generate(&complete.Generator{
		AppName:     "my-app",
		DynamicArgs: []string{"items"},
	})
	require.NoError(t, err)

	// Hyphens are munged out of Elvish function names but kept in the command
	// name and arg-completer key.
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "fn _my_app_positionals {|cmdskip valueflags @tokens|")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "set edit:completion:arg-completer['my-app'] = {|@words|")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "(external 'my-app') $@callargs")
}

func TestGenerate_ErrorOnUnsafeAppName(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "bad;name"})
	require.EqualError(t, err, `AppName contains unsafe characters: "bad;name"`)
}

func TestGenerate_ErrorOnUnsafeDynamic(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "app", DynamicArgs: []string{"bad;arg"}})
	require.EqualError(t, err, `DynamicArgs contains unsafe characters: "bad;arg"`)
}
