package zsh

import (
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

// --- Generator helpers ---

func flatGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Enable verbose output"},
			{LongFlag: "output", ShortFlag: "o", Terse: "Output file", HasArg: true},
		},
	}
}

func genSubcommands() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Enable verbose output"},
		},
		Subs: []complete.SubSpec{
			{Name: "build", Terse: "Build the project"},
			{Name: "test", Terse: "Run tests", Aliases: []string{"t"}},
		},
	}
}

func genNested() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp",
		Subs: []complete.SubSpec{
			{
				Name:  "remote",
				Terse: "Manage remotes",
				Subs: []complete.SubSpec{
					{Name: "add", Terse: "Add a remote"},
					{Name: "remove", Terse: "Remove a remote"},
				},
			},
		},
	}
}

func hintsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp",
		Specs: []complete.Spec{
			{LongFlag: "file", Terse: "A file", HasArg: true, ValueHint: complete.HintFile},
			{LongFlag: "dir", Terse: "A directory", HasArg: true, ValueHint: complete.HintDir},
			{LongFlag: "cmd", Terse: "A command", HasArg: true, ValueHint: complete.HintCommand},
			{LongFlag: "user", Terse: "A user", HasArg: true, ValueHint: complete.HintUser},
			{LongFlag: "host", Terse: "A host", HasArg: true, ValueHint: complete.HintHost},
			{LongFlag: "ext", Terse: "Extension file", HasArg: true, Extension: "yaml,yml"},
		},
	}
}

func valuesGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp",
		Specs: []complete.Spec{
			{
				LongFlag: "format",
				Terse:    "Output format",
				HasArg:   true,
				Values:   []string{"json", "yaml", "text"},
			},
			{LongFlag: "level", Terse: "Log level", HasArg: true, ValueDescs: []complete.ValueDesc{
				{Value: "debug", Desc: "Debug level"},
				{Value: "info", Desc: "Info level"},
			}},
		},
	}
}

func commaGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp",
		Specs: []complete.Spec{
			{
				LongFlag:  "columns",
				Terse:     "Columns to show",
				HasArg:    true,
				CommaList: true,
				Values:    []string{"name", "age", "email"},
			},
		},
	}
}

func pathArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp",
		Subs: []complete.SubSpec{
			{Name: "edit", Terse: "Edit files", PathArgs: true},
		},
	}
}

func dynamicArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName:     "testapp",
		DynamicArgs: []string{"kind", "name"},
	}
}

func hyphenatedGen() *complete.Generator {
	return &complete.Generator{
		AppName: "my-app",
		Specs: []complete.Spec{
			{LongFlag: "verbose", Terse: "Enable verbose output"},
		},
	}
}

// --- Generate tests ---

func TestGenerate_Flat(t *testing.T) {
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
        '(-o --output)-o+[Output file]: :_default' \
        '(-o --output)--output=[Output file]: :_default' \
        '(-v --verbose)-v[Enable verbose output]' \
        '(-v --verbose)--verbose[Enable verbose output]' \
    && ret=0
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(flatGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_Subcommands(t *testing.T) {
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
        '(-v --verbose)-v[Enable verbose output]' \
        '(-v --verbose)--verbose[Enable verbose output]' \
        ':: :_testapp_commands' \
        '*::: :->testapp' \
    && ret=0

    case $state in
    (testapp)
        words=($line[1] "${words[@]}")
        (( CURRENT += 1 ))
        curcontext="${curcontext%%:*:*}:testapp-command-$line[1]:"
        case $line[1] in
            (build)
                _arguments "${_arguments_options[@]}" : \
                && ret=0
            ;;
            (test|t)
                _arguments "${_arguments_options[@]}" : \
                && ret=0
            ;;
        esac
    ;;
    esac
}

(( $+functions[_testapp_commands] )) ||
_testapp_commands() {
    local commands; commands=(
        'build:Build the project' \
        'test:Run tests' \
    )
    _describe -t commands 'testapp commands' commands "$@"
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(genSubcommands())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_Nested(t *testing.T) {
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
        ':: :_testapp_commands' \
        '*::: :->testapp' \
    && ret=0

    case $state in
    (testapp)
        words=($line[1] "${words[@]}")
        (( CURRENT += 1 ))
        curcontext="${curcontext%%:*:*}:testapp-command-$line[1]:"
        case $line[1] in
            (remote)
                _arguments "${_arguments_options[@]}" : \
                    ':: :_testapp__remote_commands' \
                    '*::: :->remote' \
                && ret=0

                case $state in
                (remote)
                    words=($line[1] "${words[@]}")
                    (( CURRENT += 1 ))
                    curcontext="${curcontext%%:*:*}:remote-command-$line[1]:"
                    case $line[1] in
                        (add)
                            _arguments "${_arguments_options[@]}" : \
                            && ret=0
                        ;;
                        (remove)
                            _arguments "${_arguments_options[@]}" : \
                            && ret=0
                        ;;
                    esac
                ;;
                esac
            ;;
        esac
    ;;
    esac
}

(( $+functions[_testapp_commands] )) ||
_testapp_commands() {
    local commands; commands=(
        'remote:Manage remotes' \
    )
    _describe -t commands 'testapp commands' commands "$@"
}

(( $+functions[_testapp__remote_commands] )) ||
_testapp__remote_commands() {
    local commands; commands=(
        'add:Add a remote' \
        'remove:Remove a remote' \
    )
    _describe -t commands 'testapp remote commands' commands "$@"
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(genNested())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_Hints(t *testing.T) {
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
        '--cmd=[A command]:cmd:_command_names -e' \
        '--dir=[A directory]:dir:_files -/' \
        '--ext=[Extension file]:ext:_files -g "*.{yaml,yml}"' \
        '--file=[A file]:file:_files' \
        '--host=[A host]:host:_hosts' \
        '--user=[A user]:user:_users' \
    && ret=0
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(hintsGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_Values(t *testing.T) {
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
        '--format=[Output format]:format:(json yaml text)' \
        '--level=[Log level]:level:((debug\:Debug\ level info\:Info\ level))' \
    && ret=0
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(valuesGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_CommaList(t *testing.T) {
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
        '--columns=[Columns to show]:columns:{_sequence compadd - name age email}' \
    && ret=0
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(commaGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_PathArgs(t *testing.T) {
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
        ':: :_testapp_commands' \
        '*::: :->testapp' \
    && ret=0

    case $state in
    (testapp)
        words=($line[1] "${words[@]}")
        (( CURRENT += 1 ))
        curcontext="${curcontext%%:*:*}:testapp-command-$line[1]:"
        case $line[1] in
            (edit)
                _arguments "${_arguments_options[@]}" : \
                    '*:file:_files' \
                && ret=0
            ;;
        esac
    ;;
    esac
}

(( $+functions[_testapp_commands] )) ||
_testapp_commands() {
    local commands; commands=(
        'edit:Edit files' \
    )
    _describe -t commands 'testapp commands' commands "$@"
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(pathArgsGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_DynamicArgs(t *testing.T) {
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
        '1: :->dyn_1' \
        '2: :->dyn_2' \
    && ret=0

    case $state in
    (dyn_1)
        local -a __pos=()
        local __skip_next=0
        local __after_dd=0
        local token
        for ((i=2; i<CURRENT; i++)); do
            token=${words[i]}
            if (( __after_dd )); then
                __pos+=("$token")
                continue
            fi
            if (( __skip_next )); then
                __skip_next=0
                continue
            fi
            if [[ $token == -- ]]; then
                __after_dd=1
                continue
            fi
            case $token in
                (-*)
                    ;;
                (*)
                    __pos+=("$token")
                    ;;
            esac
        done
        local -a items
        items=(${(f)"$(testapp --@complete=kind 2>/dev/null)"})
        compadd -a items
    ;;
    (dyn_2)
        local -a __pos=()
        local __skip_next=0
        local __after_dd=0
        local token
        for ((i=2; i<CURRENT; i++)); do
            token=${words[i]}
            if (( __after_dd )); then
                __pos+=("$token")
                continue
            fi
            if (( __skip_next )); then
                __skip_next=0
                continue
            fi
            if [[ $token == -- ]]; then
                __after_dd=1
                continue
            fi
            case $token in
                (-*)
                    ;;
                (*)
                    __pos+=("$token")
                    ;;
            esac
        done
        local -a items
        items=(${(f)"$(testapp --@complete=name -- "${__pos[@]}" 2>/dev/null)"})
        compadd -a items
    ;;
    esac
}

if [ "$funcstack[1]" = "_testapp" ]; then
    _testapp "$@"
else
    compdef _testapp testapp
fi
`
	out, err := Generate(dynamicArgsGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_Hyphenated(t *testing.T) {
	expected := `#compdef my-app

autoload -U is-at-least

_my_app() {
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
        '--verbose[Enable verbose output]' \
    && ret=0
}

if [ "$funcstack[1]" = "_my_app" ]; then
    _my_app "$@"
else
    compdef _my_app my-app
fi
`
	out, err := Generate(hyphenatedGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_ErrorOnUnsafeAppName(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "bad;name"})
	require.EqualError(t, err, `AppName contains unsafe characters: "bad;name"`)
}

func TestGenerate_ErrorOnUnsafeDynamic(t *testing.T) {
	_, err := Generate(&complete.Generator{AppName: "app", DynamicArgs: []string{"bad;arg"}})
	require.EqualError(t, err, `DynamicArgs contains unsafe characters: "bad;arg"`)
}
