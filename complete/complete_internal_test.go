package complete

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPrint_BuiltInShellsWithoutRegistration(t *testing.T) {
	gen := NewGenerator("myapp")

	tests := []struct {
		shell string
		want  string
	}{
		{
			shell: "bash",
			want: `# myapp bash completion
_myapp() {
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
                cmd="myapp"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        myapp)
            opts=""
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
    complete -F _myapp -o nosort -o bashdefault -o default myapp
else
    complete -F _myapp -o bashdefault -o default myapp
fi
`,
		},
		{
			shell: "zsh",
			want: `#compdef myapp

autoload -U is-at-least

_myapp() {
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
    && ret=0
}

if [ "$funcstack[1]" = "_myapp" ]; then
    _myapp "$@"
else
    compdef _myapp myapp
fi
`,
		},
		{
			shell: "fish",
			want:  "complete -c myapp -f\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.shell, func(t *testing.T) {
			var buf strings.Builder
			err := gen.Print(&buf, tt.shell)
			require.NoError(t, err)
			require.Equal(t, tt.want, buf.String())
		})
	}
}

func TestBashHelpers(t *testing.T) {
	require.Equal(t, "my__app", bashCmdNameFromApp("my-app"))
	require.Equal(t, "simple", bashCmdNameFromApp("simple"))

	specs := SortVisibleSpecs([]Spec{
		{LongFlag: "verbose", ShortFlag: "v"},
		{LongFlag: "output", ShortFlag: "o"},
	})
	subs := SortSubSpecs([]SubSpec{
		{Name: "build"},
		{Name: "test", Aliases: []string{"t"}},
	})
	require.Equal(t, "--output -o --verbose -v build test", bashOptsString(specs, subs))

	require.Equal(t, `local oldifs
if [ -n "${IFS+x}" ]; then
    oldifs="$IFS"
fi
IFS=$'\n'
COMPREPLY=($(compgen -d -- "${cur}") $(compgen -f -X '!*.yaml' -- "${cur}"))
if [ -n "${oldifs+x}" ]; then
    IFS="$oldifs"
fi
if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
    compopt -o filenames
fi
`, bashExtCompletionBlock("yaml"))
	require.Equal(t, `local oldifs
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
`, bashExtCompletionBlock("yaml,yml"))
}

func TestFishHelpers(t *testing.T) {
	require.Equal(t, "my_app", fishFuncName("my-app"))
	require.Equal(t, "simple", fishFuncName("simple"))
	require.Equal(t, `it\"s`, fishEscapeString(`it"s`))
	require.Equal(t, `cost \$5`, fishEscapeString("cost $5"))
	require.Equal(t, `c:\\tmp`, fishEscapeString(`c:\tmp`))
	require.Equal(t, []string{".yaml", ".yml"}, fishExtToSuffixes("yaml,yml"))
	require.Equal(t, []string{".txt"}, fishExtToSuffixes("txt"))
	require.Equal(t, `"json" "yaml"`, fishQuotedWords([]string{"json", "yaml"}))
}

func TestMatchCase(t *testing.T) {
	tests := []struct {
		name string
		orig string
		repl string
		want string
	}{
		{"empty_orig", "", "foo", "foo"},
		{"empty_repl", "bar", "", ""},
		{"all_upper", "FOO", "bar", "BAR"},
		{"title_case", "Foo", "bar", "Bar"},
		{"lowercase", "foo", "bar", "bar"},
		{"lowercase_orig_upper_repl", "foo", "BAZ", "BAZ"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, matchCase(tt.orig, tt.repl))
		})
	}
}

func TestBuiltInShellFunc(t *testing.T) {
	tests := []struct {
		name  string
		shell string
		ok    bool
	}{
		{"bash", "bash", true},
		{"fish", "fish", true},
		{"zsh", "zsh", true},
		{"empty", "", false},
		{"powershell", "powershell", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fn, ok := builtInShellFunc(tt.shell)
			require.Equal(t, tt.ok, ok)
			if tt.ok {
				require.NotNil(t, fn)
			} else {
				require.Nil(t, fn)
			}
		})
	}
}

func TestArgValuePatterns(t *testing.T) {
	tests := []struct {
		name       string
		specs      []Spec
		wantExact  []string
		wantEquals []string
	}{
		{"empty", nil, nil, nil},
		{"no_arg", []Spec{{HasArg: false, LongFlag: "verbose"}}, nil, nil},
		{
			"long_and_short",
			[]Spec{{HasArg: true, LongFlag: "output", ShortFlag: "o"}},
			[]string{"--output", "-o"},
			[]string{"--output=*", "-o=*"},
		},
		{
			"long_only",
			[]Spec{{HasArg: true, LongFlag: "file"}},
			[]string{"--file"},
			[]string{"--file=*"},
		},
		{
			"short_only",
			[]Spec{{HasArg: true, ShortFlag: "f"}},
			[]string{"-f"},
			[]string{"-f=*"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exact, equals := argValuePatterns(tt.specs)
			require.Equal(t, tt.wantExact, exact)
			require.Equal(t, tt.wantEquals, equals)
		})
	}
}

func TestFishMatchPatterns(t *testing.T) {
	got := fishMatchPatterns("$tok", []string{"*.go"})
	require.Equal(t, "string match -q -- '*.go' $tok", got)

	got = fishMatchPatterns("$tok", []string{"*.go", "*.txt"})
	require.Equal(t, "string match -q -- '*.go' $tok; or string match -q -- '*.txt' $tok", got)
}

func TestZshHelpers(t *testing.T) {
	require.Equal(t, "my_app", zshFuncName("my-app"))
	require.Equal(t, "simple", zshFuncName("simple"))
	require.Equal(t, "*.yaml", zshExtGlob("yaml"))
	require.Equal(t, "*.{yaml,yml}", zshExtGlob("yaml,yml"))
	require.Equal(t, `\\`, zshEscapeHelp(`\`))
	require.Equal(t, `'\''`, zshEscapeHelp(`'`))
	require.Equal(t, `\[`, zshEscapeHelp(`[`))
	require.Equal(t, `\]`, zshEscapeHelp(`]`))
	require.Equal(t, `\:`, zshEscapeHelp(`:`))
	require.Equal(t, `\$`, zshEscapeHelp(`$`))
	require.Equal(t, "\\`", zshEscapeHelp("`"))
	require.Equal(t, " ", zshEscapeHelp("\n"))
	require.Equal(t, `\\`, zshEscapeValue(`\`))
	require.Equal(t, `'\''`, zshEscapeValue(`'`))
	require.Equal(t, `\[`, zshEscapeValue(`[`))
	require.Equal(t, `\]`, zshEscapeValue(`]`))
	require.Equal(t, `\:`, zshEscapeValue(`:`))
	require.Equal(t, `\$`, zshEscapeValue(`$`))
	require.Equal(t, "\\`", zshEscapeValue("`"))
	require.Equal(t, `\ `, zshEscapeValue("\n"))
	require.Equal(t, `\(`, zshEscapeValue(`(`))
	require.Equal(t, `\)`, zshEscapeValue(`)`))
	require.Equal(t, `\ `, zshEscapeValue(` `))
	require.Equal(t, "(-v --verbose)", zshExclusion(Spec{ShortFlag: "v", LongFlag: "verbose"}))
	require.Empty(t, zshExclusion(Spec{LongFlag: "verbose"}))

	tests := []struct {
		hint string
		want string
	}{
		{HintFile, "_files"},
		{HintDir, "_files -/"},
		{HintCommand, "_command_names -e"},
		{HintUser, "_users"},
		{HintHost, "_hosts"},
		{HintURL, "_urls"},
		{HintEmail, "_email_addresses"},
		{"unknown", "_default"},
	}
	for _, tt := range tests {
		t.Run(tt.hint, func(t *testing.T) {
			require.Equal(t, tt.want, zshHintCompleter(tt.hint))
		})
	}
}
