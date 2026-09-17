package bash

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

func genNested() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp", Specs: []complete.Spec{{LongFlag: "verbose", Terse: "Verbose"}},
		Subs: []complete.SubSpec{
			{
				Name: "auth", Terse: "Manage authentication",
				Specs: []complete.Spec{{LongFlag: "token", Terse: "Auth token", HasArg: true}},
				Subs: []complete.SubSpec{
					{
						Name:  "login",
						Terse: "Log in",
						Specs: []complete.Spec{{LongFlag: "browser", Terse: "Open browser"}},
					},
					{Name: "logout", Terse: "Log out"},
				},
			},
			{Name: "run", Terse: "Run command"},
		},
	}
}

func hintsGen() *complete.Generator {
	return &complete.Generator{AppName: "testapp", Specs: []complete.Spec{
		{LongFlag: "config", Terse: "Config file", HasArg: true, Extension: "yaml,yml"},
		{LongFlag: "output", Terse: "Output path", HasArg: true, ValueHint: complete.HintFile},
		{LongFlag: "dir", Terse: "Directory", HasArg: true, ValueHint: complete.HintDir},
		{LongFlag: "shell", Terse: "Shell command", HasArg: true, ValueHint: complete.HintCommand},
		{LongFlag: "user", Terse: "User name", HasArg: true, ValueHint: complete.HintUser},
		{LongFlag: "host", Terse: "Host name", HasArg: true, ValueHint: complete.HintHost},
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

func commaGen() *complete.Generator {
	return &complete.Generator{AppName: "testapp", Specs: []complete.Spec{
		{
			LongFlag:  "tags",
			Terse:     "Tags",
			HasArg:    true,
			CommaList: true,
			Values:    []string{"bug", "feature", "docs"},
		},
		{LongFlag: "labels", Terse: "Labels", HasArg: true, CommaList: true, Dynamic: "labels"},
		{LongFlag: "repo", Terse: "Repository", HasArg: true, Dynamic: "repos"},
	}}
}

func pathArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp", Specs: []complete.Spec{{LongFlag: "verbose", Terse: "Verbose"}},
		Subs: []complete.SubSpec{
			{
				Name:     "edit",
				Terse:    "Edit files",
				PathArgs: true,
				Specs:    []complete.Spec{{LongFlag: "editor", Terse: "Editor", HasArg: true}},
			},
		},
	}
}

func dynamicArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "testapp", DynamicArgs: []string{"items", "subitems"},
		Specs: []complete.Spec{{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose"}},
	}
}

func limitedDynamicArgsGen() *complete.Generator {
	return &complete.Generator{
		AppName:              "testapp",
		DynamicArgs:          []string{"email"},
		HasMaxPositionalArgs: true,
		MaxPositionalArgs:    1,
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose"},
		},
	}
}

func hyphenatedGen() *complete.Generator {
	return &complete.Generator{
		AppName: "my-app", Specs: []complete.Spec{{LongFlag: "verbose", Terse: "Verbose"}},
		Subs: []complete.SubSpec{{Name: "build", Terse: "Build"}},
	}
}

func TestGenerate_Flat(t *testing.T) {
	out, err := Generate(flatGen())
	require.NoError(t, err)
	expected := `_testapp() {
    if [[ -z ${COMP_LINE+x} ]]; then
        _testapp_complete "$@"
        return
    fi
    local remaining="$COMP_LINE" fragment index joined=0 consumed suffix
    local -a words=()
    for ((index=0; index<=COMP_CWORD; index++)); do
        fragment=${COMP_WORDS[index]}
        if (( index > 0 )) && [[ $remaining == [[:space:]]* ]]; then
            joined=$((joined + 1))
        fi
        words[joined]=${words[joined]-}$fragment
        remaining=${remaining#*"$fragment"}
        # Some Bash versions omit the surrounding quotes from COMP_WORDS.
        remaining=${remaining#[\'\"]}
    done
    # Ignore text after the cursor, including the rest of the current word.
    consumed=$((${#COMP_LINE} - ${#remaining}))
    suffix=$((consumed - ${COMP_POINT:-$consumed}))
    if (( suffix > 0 && suffix <= ${#words[joined]} )); then
        words[joined]=${words[joined]:0:${#words[joined]}-suffix}
    fi
    local current=${words[joined]} keep="" previous=""
    if [[ -z $2 ]]; then
        keep=$current
    elif [[ $current == *"$2" ]]; then
        keep=${current%"$2"}
    fi
    if (( joined > 0 )); then
        previous=${words[joined-1]}
    fi
    # Dynamic scope gives the helpers reconstructed words without changing
    # the caller's completion state or the shell's word-break settings.
    local -a COMP_WORDS=("${words[@]}")
    local COMP_CWORD=$joined
    _testapp_complete "$1" "$current" "$previous"
    local candidate
    local -a replies=()
    for candidate in "${COMPREPLY[@]}"; do
        if [[ -n $keep && $candidate == "$keep"* ]]; then
            candidate=${candidate#"$keep"}
        fi
        if [[ -n $candidate ]]; then
            replies+=("$candidate")
        fi
    done
    COMPREPLY=("${replies[@]}")
}

# testapp bash completion
_testapp_complete() {
    local i cur prev opts cmd
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
    prev="$3"
    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="testapp"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        testapp)
            opts="--output -o --verbose -v"
            local value_prefix=""
            case "${cur}" in
                --output=*|-o=*)
                    prev=${cur%%=*}
                    value_prefix="${prev}="
                    cur=${cur#*=}
                    ;;
            esac
            case "${prev}" in
                --output|-o)
                    COMPREPLY=()
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
    esac
}

if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F _testapp -o nosort -o bashdefault -o default testapp
else
    complete -F _testapp -o bashdefault -o default testapp
fi
`
	require.Equal(t, expected, out)
}

func TestGenerate_Subcommands(t *testing.T) {
	out, err := Generate(genSubcommands())
	require.NoError(t, err)
	expected := `_testapp() {
    if [[ -z ${COMP_LINE+x} ]]; then
        _testapp_complete "$@"
        return
    fi
    local remaining="$COMP_LINE" fragment index joined=0 consumed suffix
    local -a words=()
    for ((index=0; index<=COMP_CWORD; index++)); do
        fragment=${COMP_WORDS[index]}
        if (( index > 0 )) && [[ $remaining == [[:space:]]* ]]; then
            joined=$((joined + 1))
        fi
        words[joined]=${words[joined]-}$fragment
        remaining=${remaining#*"$fragment"}
        # Some Bash versions omit the surrounding quotes from COMP_WORDS.
        remaining=${remaining#[\'\"]}
    done
    # Ignore text after the cursor, including the rest of the current word.
    consumed=$((${#COMP_LINE} - ${#remaining}))
    suffix=$((consumed - ${COMP_POINT:-$consumed}))
    if (( suffix > 0 && suffix <= ${#words[joined]} )); then
        words[joined]=${words[joined]:0:${#words[joined]}-suffix}
    fi
    local current=${words[joined]} keep="" previous=""
    if [[ -z $2 ]]; then
        keep=$current
    elif [[ $current == *"$2" ]]; then
        keep=${current%"$2"}
    fi
    if (( joined > 0 )); then
        previous=${words[joined-1]}
    fi
    # Dynamic scope gives the helpers reconstructed words without changing
    # the caller's completion state or the shell's word-break settings.
    local -a COMP_WORDS=("${words[@]}")
    local COMP_CWORD=$joined
    _testapp_complete "$1" "$current" "$previous"
    local candidate
    local -a replies=()
    for candidate in "${COMPREPLY[@]}"; do
        if [[ -n $keep && $candidate == "$keep"* ]]; then
            candidate=${candidate#"$keep"}
        fi
        if [[ -n $candidate ]]; then
            replies+=("$candidate")
        fi
    done
    COMPREPLY=("${replies[@]}")
}

# testapp bash completion
_testapp_complete() {
    local i cur prev opts cmd
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
    prev="$3"
    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="testapp"
                ;;
            testapp,build)
                cmd="testapp__build"
                ;;
            testapp,test|testapp,t)
                cmd="testapp__test"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        testapp)
            opts="--verbose -v build test"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        testapp__build)
            opts="--output -o --release"
            local value_prefix=""
            case "${cur}" in
                --output=*|-o=*)
                    prev=${cur%%=*}
                    value_prefix="${prev}="
                    cur=${cur#*=}
                    ;;
            esac
            case "${prev}" in
                --output|-o)
                    COMPREPLY=()
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        testapp__test)
            opts=""
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
    esac
}

if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F _testapp -o nosort -o bashdefault -o default testapp
else
    complete -F _testapp -o bashdefault -o default testapp
fi
`
	require.Equal(t, expected, out)
}

func TestGenerate_LimitedDynamicArgs(t *testing.T) {
	out, err := Generate(limitedDynamicArgsGen())
	require.NoError(t, err)
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "if [[ ${#__dyn_pos[@]} -ge 1 ]]; then")
	//nolint:gocritic // fragment check against generated/styled output; not worth pinning as an exact literal
	require.Contains(t, out, "--@complete=email")
}

func TestGenerate_Nested(t *testing.T) {
	out, err := Generate(genNested())
	require.NoError(t, err)
	expected := `_testapp() {
    if [[ -z ${COMP_LINE+x} ]]; then
        _testapp_complete "$@"
        return
    fi
    local remaining="$COMP_LINE" fragment index joined=0 consumed suffix
    local -a words=()
    for ((index=0; index<=COMP_CWORD; index++)); do
        fragment=${COMP_WORDS[index]}
        if (( index > 0 )) && [[ $remaining == [[:space:]]* ]]; then
            joined=$((joined + 1))
        fi
        words[joined]=${words[joined]-}$fragment
        remaining=${remaining#*"$fragment"}
        # Some Bash versions omit the surrounding quotes from COMP_WORDS.
        remaining=${remaining#[\'\"]}
    done
    # Ignore text after the cursor, including the rest of the current word.
    consumed=$((${#COMP_LINE} - ${#remaining}))
    suffix=$((consumed - ${COMP_POINT:-$consumed}))
    if (( suffix > 0 && suffix <= ${#words[joined]} )); then
        words[joined]=${words[joined]:0:${#words[joined]}-suffix}
    fi
    local current=${words[joined]} keep="" previous=""
    if [[ -z $2 ]]; then
        keep=$current
    elif [[ $current == *"$2" ]]; then
        keep=${current%"$2"}
    fi
    if (( joined > 0 )); then
        previous=${words[joined-1]}
    fi
    # Dynamic scope gives the helpers reconstructed words without changing
    # the caller's completion state or the shell's word-break settings.
    local -a COMP_WORDS=("${words[@]}")
    local COMP_CWORD=$joined
    _testapp_complete "$1" "$current" "$previous"
    local candidate
    local -a replies=()
    for candidate in "${COMPREPLY[@]}"; do
        if [[ -n $keep && $candidate == "$keep"* ]]; then
            candidate=${candidate#"$keep"}
        fi
        if [[ -n $candidate ]]; then
            replies+=("$candidate")
        fi
    done
    COMPREPLY=("${replies[@]}")
}

# testapp bash completion
_testapp_complete() {
    local i cur prev opts cmd
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
    prev="$3"
    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="testapp"
                ;;
            testapp,auth)
                cmd="testapp__auth"
                ;;
            testapp__auth,login)
                cmd="testapp__auth__login"
                ;;
            testapp__auth,logout)
                cmd="testapp__auth__logout"
                ;;
            testapp,run)
                cmd="testapp__run"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        testapp)
            opts="--verbose auth run"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        testapp__auth)
            opts="--token login logout"
            local value_prefix=""
            case "${cur}" in
                --token=*)
                    prev=${cur%%=*}
                    value_prefix="${prev}="
                    cur=${cur#*=}
                    ;;
            esac
            case "${prev}" in
                --token)
                    COMPREPLY=()
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        testapp__auth__login)
            opts="--browser"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        testapp__auth__logout)
            opts=""
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 3 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        testapp__run)
            opts=""
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
    esac
}

if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F _testapp -o nosort -o bashdefault -o default testapp
else
    complete -F _testapp -o bashdefault -o default testapp
fi
`
	require.Equal(t, expected, out)
}

func TestGenerate_Hints(t *testing.T) {
	out, err := Generate(hintsGen())
	require.NoError(t, err)
	expected := `_testapp() {
    if [[ -z ${COMP_LINE+x} ]]; then
        _testapp_complete "$@"
        return
    fi
    local remaining="$COMP_LINE" fragment index joined=0 consumed suffix
    local -a words=()
    for ((index=0; index<=COMP_CWORD; index++)); do
        fragment=${COMP_WORDS[index]}
        if (( index > 0 )) && [[ $remaining == [[:space:]]* ]]; then
            joined=$((joined + 1))
        fi
        words[joined]=${words[joined]-}$fragment
        remaining=${remaining#*"$fragment"}
        # Some Bash versions omit the surrounding quotes from COMP_WORDS.
        remaining=${remaining#[\'\"]}
    done
    # Ignore text after the cursor, including the rest of the current word.
    consumed=$((${#COMP_LINE} - ${#remaining}))
    suffix=$((consumed - ${COMP_POINT:-$consumed}))
    if (( suffix > 0 && suffix <= ${#words[joined]} )); then
        words[joined]=${words[joined]:0:${#words[joined]}-suffix}
    fi
    local current=${words[joined]} keep="" previous=""
    if [[ -z $2 ]]; then
        keep=$current
    elif [[ $current == *"$2" ]]; then
        keep=${current%"$2"}
    fi
    if (( joined > 0 )); then
        previous=${words[joined-1]}
    fi
    # Dynamic scope gives the helpers reconstructed words without changing
    # the caller's completion state or the shell's word-break settings.
    local -a COMP_WORDS=("${words[@]}")
    local COMP_CWORD=$joined
    _testapp_complete "$1" "$current" "$previous"
    local candidate
    local -a replies=()
    for candidate in "${COMPREPLY[@]}"; do
        if [[ -n $keep && $candidate == "$keep"* ]]; then
            candidate=${candidate#"$keep"}
        fi
        if [[ -n $candidate ]]; then
            replies+=("$candidate")
        fi
    done
    COMPREPLY=("${replies[@]}")
}

# testapp bash completion
_testapp_complete() {
    local i cur prev opts cmd
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
    prev="$3"
    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="testapp"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        testapp)
            opts="--config --dir --host --output --shell --user"
            local value_prefix=""
            case "${cur}" in
                --config=*|--dir=*|--host=*|--output=*|--shell=*|--user=*)
                    prev=${cur%%=*}
                    value_prefix="${prev}="
                    cur=${cur#*=}
                    ;;
            esac
            case "${prev}" in
                --config)
                    local oldifs
                    if [ -n "${IFS+x}" ]; then
                        oldifs="$IFS"
                    fi
                    IFS=$'\n'
                    COMPREPLY=($(compgen -d -- "${cur}") $(compgen -f -X '!@(*.yaml|*.yml)' -- "${cur}"))
                    if [ -n "${oldifs+x}" ]; then
                        IFS="$oldifs"
                    fi
                    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
                        compopt -o filenames
                    fi
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                --dir)
                    COMPREPLY=()
                    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
                        compopt -o plusdirs
                    fi
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                --host)
                    COMPREPLY=($(compgen -A hostname -- "${cur}"))
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                --output)
                    local oldifs
                    if [ -n "${IFS+x}" ]; then
                        oldifs="$IFS"
                    fi
                    IFS=$'\n'
                    COMPREPLY=($(compgen -f -- "${cur}"))
                    if [ -n "${oldifs+x}" ]; then
                        IFS="$oldifs"
                    fi
                    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
                        compopt -o filenames
                    fi
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                --shell)
                    COMPREPLY=($(compgen -c -- "${cur}"))
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                --user)
                    COMPREPLY=($(compgen -u -- "${cur}"))
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
    esac
}

if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F _testapp -o nosort -o bashdefault -o default testapp
else
    complete -F _testapp -o bashdefault -o default testapp
fi
`
	require.Equal(t, expected, out)
}

func TestGenerate_Values(t *testing.T) {
	out, err := Generate(valuesGen())
	require.NoError(t, err)
	expected := `_testapp() {
    if [[ -z ${COMP_LINE+x} ]]; then
        _testapp_complete "$@"
        return
    fi
    local remaining="$COMP_LINE" fragment index joined=0 consumed suffix
    local -a words=()
    for ((index=0; index<=COMP_CWORD; index++)); do
        fragment=${COMP_WORDS[index]}
        if (( index > 0 )) && [[ $remaining == [[:space:]]* ]]; then
            joined=$((joined + 1))
        fi
        words[joined]=${words[joined]-}$fragment
        remaining=${remaining#*"$fragment"}
        # Some Bash versions omit the surrounding quotes from COMP_WORDS.
        remaining=${remaining#[\'\"]}
    done
    # Ignore text after the cursor, including the rest of the current word.
    consumed=$((${#COMP_LINE} - ${#remaining}))
    suffix=$((consumed - ${COMP_POINT:-$consumed}))
    if (( suffix > 0 && suffix <= ${#words[joined]} )); then
        words[joined]=${words[joined]:0:${#words[joined]}-suffix}
    fi
    local current=${words[joined]} keep="" previous=""
    if [[ -z $2 ]]; then
        keep=$current
    elif [[ $current == *"$2" ]]; then
        keep=${current%"$2"}
    fi
    if (( joined > 0 )); then
        previous=${words[joined-1]}
    fi
    # Dynamic scope gives the helpers reconstructed words without changing
    # the caller's completion state or the shell's word-break settings.
    local -a COMP_WORDS=("${words[@]}")
    local COMP_CWORD=$joined
    _testapp_complete "$1" "$current" "$previous"
    local candidate
    local -a replies=()
    for candidate in "${COMPREPLY[@]}"; do
        if [[ -n $keep && $candidate == "$keep"* ]]; then
            candidate=${candidate#"$keep"}
        fi
        if [[ -n $candidate ]]; then
            replies+=("$candidate")
        fi
    done
    COMPREPLY=("${replies[@]}")
}

# testapp bash completion
_testapp_complete() {
    local i cur prev opts cmd
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
    prev="$3"
    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="testapp"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        testapp)
            opts="--format --level"
            local value_prefix=""
            case "${cur}" in
                --format=*|--level=*)
                    prev=${cur%%=*}
                    value_prefix="${prev}="
                    cur=${cur#*=}
                    ;;
            esac
            case "${prev}" in
                --format)
                    COMPREPLY=($(compgen -W 'json yaml text' -- "${cur}"))
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                --level)
                    COMPREPLY=($(compgen -W 'info warn' -- "${cur}"))
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
    esac
}

if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F _testapp -o nosort -o bashdefault -o default testapp
else
    complete -F _testapp -o bashdefault -o default testapp
fi
`
	require.Equal(t, expected, out)
}

func TestGenerate_CommaList(t *testing.T) {
	out, err := Generate(commaGen())
	require.NoError(t, err)
	expected, err := complete.GenerateBash(commaGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_PathArgs(t *testing.T) {
	out, err := Generate(pathArgsGen())
	require.NoError(t, err)
	expected := `_testapp() {
    if [[ -z ${COMP_LINE+x} ]]; then
        _testapp_complete "$@"
        return
    fi
    local remaining="$COMP_LINE" fragment index joined=0 consumed suffix
    local -a words=()
    for ((index=0; index<=COMP_CWORD; index++)); do
        fragment=${COMP_WORDS[index]}
        if (( index > 0 )) && [[ $remaining == [[:space:]]* ]]; then
            joined=$((joined + 1))
        fi
        words[joined]=${words[joined]-}$fragment
        remaining=${remaining#*"$fragment"}
        # Some Bash versions omit the surrounding quotes from COMP_WORDS.
        remaining=${remaining#[\'\"]}
    done
    # Ignore text after the cursor, including the rest of the current word.
    consumed=$((${#COMP_LINE} - ${#remaining}))
    suffix=$((consumed - ${COMP_POINT:-$consumed}))
    if (( suffix > 0 && suffix <= ${#words[joined]} )); then
        words[joined]=${words[joined]:0:${#words[joined]}-suffix}
    fi
    local current=${words[joined]} keep="" previous=""
    if [[ -z $2 ]]; then
        keep=$current
    elif [[ $current == *"$2" ]]; then
        keep=${current%"$2"}
    fi
    if (( joined > 0 )); then
        previous=${words[joined-1]}
    fi
    # Dynamic scope gives the helpers reconstructed words without changing
    # the caller's completion state or the shell's word-break settings.
    local -a COMP_WORDS=("${words[@]}")
    local COMP_CWORD=$joined
    _testapp_complete "$1" "$current" "$previous"
    local candidate
    local -a replies=()
    for candidate in "${COMPREPLY[@]}"; do
        if [[ -n $keep && $candidate == "$keep"* ]]; then
            candidate=${candidate#"$keep"}
        fi
        if [[ -n $candidate ]]; then
            replies+=("$candidate")
        fi
    done
    COMPREPLY=("${replies[@]}")
}

# testapp bash completion
_testapp_complete() {
    local i cur prev opts cmd
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
    prev="$3"
    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="testapp"
                ;;
            testapp,edit)
                cmd="testapp__edit"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        testapp)
            opts="--verbose edit"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        testapp__edit)
            opts="--editor"
            local value_prefix=""
            case "${cur}" in
                --editor=*)
                    prev=${cur%%=*}
                    value_prefix="${prev}="
                    cur=${cur#*=}
                    ;;
            esac
            case "${prev}" in
                --editor)
                    COMPREPLY=()
                    if [[ -n $value_prefix ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${value_prefix}}")
                    fi
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            local oldifs
            if [ -n "${IFS+x}" ]; then
                oldifs="$IFS"
            fi
            IFS=$'\n'
            COMPREPLY=($(compgen -f -- "${cur}"))
            if [ -n "${oldifs+x}" ]; then
                IFS="$oldifs"
            fi
            if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
                compopt -o filenames
            fi
            return 0
            ;;
    esac
}

if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F _testapp -o nosort -o bashdefault -o default testapp
else
    complete -F _testapp -o bashdefault -o default testapp
fi
`
	require.Equal(t, expected, out)
}

func TestGenerate_DynamicArgs(t *testing.T) {
	out, err := Generate(dynamicArgsGen())
	require.NoError(t, err)
	expected, err := complete.GenerateBash(dynamicArgsGen())
	require.NoError(t, err)
	require.Equal(t, expected, out)
}

func TestGenerate_Hyphenated(t *testing.T) {
	out, err := Generate(hyphenatedGen())
	require.NoError(t, err)
	expected := `_my_app() {
    if [[ -z ${COMP_LINE+x} ]]; then
        _my_app_complete "$@"
        return
    fi
    local remaining="$COMP_LINE" fragment index joined=0 consumed suffix
    local -a words=()
    for ((index=0; index<=COMP_CWORD; index++)); do
        fragment=${COMP_WORDS[index]}
        if (( index > 0 )) && [[ $remaining == [[:space:]]* ]]; then
            joined=$((joined + 1))
        fi
        words[joined]=${words[joined]-}$fragment
        remaining=${remaining#*"$fragment"}
        # Some Bash versions omit the surrounding quotes from COMP_WORDS.
        remaining=${remaining#[\'\"]}
    done
    # Ignore text after the cursor, including the rest of the current word.
    consumed=$((${#COMP_LINE} - ${#remaining}))
    suffix=$((consumed - ${COMP_POINT:-$consumed}))
    if (( suffix > 0 && suffix <= ${#words[joined]} )); then
        words[joined]=${words[joined]:0:${#words[joined]}-suffix}
    fi
    local current=${words[joined]} keep="" previous=""
    if [[ -z $2 ]]; then
        keep=$current
    elif [[ $current == *"$2" ]]; then
        keep=${current%"$2"}
    fi
    if (( joined > 0 )); then
        previous=${words[joined-1]}
    fi
    # Dynamic scope gives the helpers reconstructed words without changing
    # the caller's completion state or the shell's word-break settings.
    local -a COMP_WORDS=("${words[@]}")
    local COMP_CWORD=$joined
    _my_app_complete "$1" "$current" "$previous"
    local candidate
    local -a replies=()
    for candidate in "${COMPREPLY[@]}"; do
        if [[ -n $keep && $candidate == "$keep"* ]]; then
            candidate=${candidate#"$keep"}
        fi
        if [[ -n $candidate ]]; then
            replies+=("$candidate")
        fi
    done
    COMPREPLY=("${replies[@]}")
}

# my-app bash completion
_my_app_complete() {
    local i cur opts cmd
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="my__app"
                ;;
            my__app,build)
                cmd="my__app__build"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        my__app)
            opts="--verbose build"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
        my__app__build)
            opts=""
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
            return 0
            ;;
    esac
}

if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F _my_app -o nosort -o bashdefault -o default my-app
else
    complete -F _my_app -o bashdefault -o default my-app
fi
`
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
