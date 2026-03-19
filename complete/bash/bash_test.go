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

func hyphenatedGen() *complete.Generator {
	return &complete.Generator{
		AppName: "my-app", Specs: []complete.Spec{{LongFlag: "verbose", Terse: "Verbose"}},
		Subs: []complete.SubSpec{{Name: "build", Terse: "Build"}},
	}
}

func TestGenerate_Flat(t *testing.T) {
	out, err := Generate(flatGen())
	require.NoError(t, err)
	expected := `# testapp bash completion
_testapp() {
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
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            case "${prev}" in
                --output|-o)
                    COMPREPLY=()
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
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
	expected := `# testapp bash completion
_testapp() {
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
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            case "${prev}" in
                --output|-o)
                    COMPREPLY=()
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
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

func TestGenerate_Nested(t *testing.T) {
	out, err := Generate(genNested())
	require.NoError(t, err)
	expected := `# testapp bash completion
_testapp() {
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
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            case "${prev}" in
                --token)
                    COMPREPLY=()
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
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
	expected := `# testapp bash completion
_testapp() {
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
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
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
                    return 0
                    ;;
                --dir)
                    COMPREPLY=()
                    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
                        compopt -o plusdirs
                    fi
                    return 0
                    ;;
                --host)
                    COMPREPLY=($(compgen -A hostname -- "${cur}"))
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
                    return 0
                    ;;
                --shell)
                    COMPREPLY=($(compgen -c -- "${cur}"))
                    return 0
                    ;;
                --user)
                    COMPREPLY=($(compgen -u -- "${cur}"))
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
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
	expected := `# testapp bash completion
_testapp() {
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
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            case "${prev}" in
                --format)
                    COMPREPLY=($(compgen -W 'json yaml text' -- "${cur}"))
                    return 0
                    ;;
                --level)
                    COMPREPLY=($(compgen -W 'info warn' -- "${cur}"))
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
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
	expected := `# testapp bash completion
_testapp() {
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
            opts="--labels --repo --tags"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            case "${prev}" in
                --labels)
                    local prefix=""
                    local cur_val="${cur}"
                    local all_vals=($(testapp --@complete=labels 2>/dev/null))
                    local -a avail=()
                    if [[ "${cur}" == *,* ]]; then
                        prefix="${cur%,*},"
                        cur_val="${cur##*,}"
                        IFS=',' read -ra selected <<< "${prefix}"
                        for val in "${all_vals[@]}"; do
                            local found=0
                            for sel in "${selected[@]}"; do
                                if [[ "${val}" == "${sel}" ]]; then
                                    found=1
                                    break
                                fi
                            done
                            if [[ "${found}" -eq 0 ]]; then
                                avail+=("${val}")
                            fi
                        done
                    else
                        avail=("${all_vals[@]}")
                    fi
                    COMPREPLY=($(compgen -W "${avail[*]}" -- "${cur_val}"))
                    if [[ -n "${prefix}" ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${prefix}}")
                    fi
                    compopt -o nospace
                    return 0
                    ;;
                --repo)
                    COMPREPLY=($(compgen -W "$(testapp --@complete=repos 2>/dev/null)" -- "${cur}"))
                    return 0
                    ;;
                --tags)
                    local prefix=""
                    local cur_val="${cur}"
                    local all_vals=(bug feature docs)
                    local -a avail=()
                    if [[ "${cur}" == *,* ]]; then
                        prefix="${cur%,*},"
                        cur_val="${cur##*,}"
                        IFS=',' read -ra selected <<< "${prefix}"
                        for val in "${all_vals[@]}"; do
                            local found=0
                            for sel in "${selected[@]}"; do
                                if [[ "${val}" == "${sel}" ]]; then
                                    found=1
                                    break
                                fi
                            done
                            if [[ "${found}" -eq 0 ]]; then
                                avail+=("${val}")
                            fi
                        done
                    else
                        avail=("${all_vals[@]}")
                    fi
                    COMPREPLY=($(compgen -W "${avail[*]}" -- "${cur_val}"))
                    if [[ -n "${prefix}" ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${prefix}}")
                    fi
                    compopt -o nospace
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
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

func TestGenerate_PathArgs(t *testing.T) {
	out, err := Generate(pathArgsGen())
	require.NoError(t, err)
	expected := `# testapp bash completion
_testapp() {
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
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 2 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            case "${prev}" in
                --editor)
                    COMPREPLY=()
                    return 0
                    ;;
                *)
                    COMPREPLY=()
                    ;;
            esac
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
	expected := `# testapp bash completion
_testapp() {
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
            opts="--verbose -v"
            if [[ ${cur} == -* ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            local -a __dyn_pos=()
            local __skip_next=0
            local __after_dd=0
            for ((j=1; j<COMP_CWORD; j++)); do
                if [[ "${__after_dd}" -eq 1 ]]; then
                    __dyn_pos+=("${COMP_WORDS[j]}")
                    continue
                fi
                if [[ "${__skip_next}" -eq 1 ]]; then
                    __skip_next=0
                    continue
                fi
                if [[ "${COMP_WORDS[j]}" == "--" ]]; then
                    __after_dd=1
                    continue
                fi
                case "${COMP_WORDS[j]}" in
                    -*)
                        ;;
                    *)
                        __dyn_pos+=("${COMP_WORDS[j]}")
                        ;;
                esac
            done
            case ${#__dyn_pos[@]} in
                0)
                    COMPREPLY=($(compgen -W "${opts} $(testapp --@complete=items 2>/dev/null)" -- "${cur}"))
                    ;;
                1)
                    COMPREPLY=($(compgen -W "$(testapp --@complete=subitems -- "${__dyn_pos[@]}" 2>/dev/null)" -- "${cur}"))
                    ;;
            esac
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

func TestGenerate_Hyphenated(t *testing.T) {
	out, err := Generate(hyphenatedGen())
	require.NoError(t, err)
	expected := `# my-app bash completion
_my_app() {
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
