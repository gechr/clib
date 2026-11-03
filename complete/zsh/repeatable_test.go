package zsh

import (
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

func TestGenerate_Repeatable(t *testing.T) {
	gen := flatGen()
	for i := range gen.Specs {
		gen.Specs[i].Repeatable = true
	}
	gen.Specs = append(gen.Specs, complete.Spec{
		LongFlag: "quiet", ShortFlag: "q", Terse: "Suppress output",
	})
	expected := `#compdef testapp

autoload -U is-at-least

_testapp() {
    typeset -A opt_args
    typeset -a _arguments_options
    local ret=1

    if is-at-least 5.2; then
        _arguments_options=(-s -S -C)
    else
        _arguments_options=(-s -C)
    fi

    local context curcontext="$curcontext" state line
    _arguments "${_arguments_options[@]}" : \
        '*-o+[Output file]: :_default' \
        '*--output=[Output file]: :_default' \
        '(-q --quiet)-q[Suppress output]' \
        '(-q --quiet)--quiet[Suppress output]' \
        '*-v[Enable verbose output]' \
        '*--verbose[Enable verbose output]' \
    && ret=0
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(gen)
	require.NoError(t, err)
	require.Equal(t, expected, out)
}
