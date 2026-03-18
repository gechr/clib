package complete_test

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/bash"
	_ "github.com/gechr/clib/complete/fish"
	_ "github.com/gechr/clib/complete/zsh"
	"github.com/stretchr/testify/require"
)

func testFlags() []complete.FlagMeta {
	return []complete.FlagMeta{
		{
			Name:     "author",
			Short:    "a",
			Terse:    "Author",
			Help:     "Filter by author",
			HasArg:   true,
			Complete: "predictor=author",
		},
		{
			Name:   "state",
			Short:  "s",
			Terse:  "State",
			Help:   "Filter by state",
			HasArg: true,
			Enum:   []string{"open", "closed", "merged", "all"},
		},
		{Name: "verbose", Short: "v", Terse: "Verbose", Help: "Enable verbose output"},
		{
			Name:     "columns",
			Terse:    "Table columns",
			Help:     "Table columns",
			HasArg:   true,
			Complete: "predictor=columns,comma",
		},
		{Name: "hidden-flag", Help: "A hidden flag", HasArg: true, Hidden: true},
		{Name: "limit", Short: "L", Terse: "Max results", Help: "Max results", HasArg: true},
		{
			Name:     "ci",
			Terse:    "CI status",
			Help:     "Filter by CI status",
			HasArg:   true,
			Complete: "values=success failure pending",
		},
		{Name: "merge", Terse: "Auto-merge", Help: "Enable auto-merge", Negatable: true},
		{Name: "debug", Terse: "Debug", Help: "Enable debug logging", Complete: "-"},
		{Name: "query", Help: "Search query", IsArg: true, HasArg: true, IsSlice: true},
	}
}

func newTestGen() *complete.Generator {
	return complete.NewGenerator("clibapp").FromFlags(testFlags())
}

func TestGenerator_FromFlags(t *testing.T) {
	gen := newTestGen()

	// Should have 9 specs (10 fields minus 1 arg minus 1 complete:"-", plus 1 negatable --no- variant).
	require.Len(t, gen.Specs, 9)
}

func TestGenerator_FromFlags_IgnoresAliases(t *testing.T) {
	gen := complete.NewGenerator("test").FromFlags([]complete.FlagMeta{
		{
			Name:    "output",
			Short:   "o",
			Aliases: []string{"out", "O"},
			HasArg:  true,
			Help:    "Output format",
		},
	})

	require.Len(t, gen.Specs, 1)
	require.Equal(t, "output", gen.Specs[0].LongFlag)
	require.Equal(t, "o", gen.Specs[0].ShortFlag)
	require.Equal(t, "Output format", gen.Specs[0].Terse)
	require.True(t, gen.Specs[0].HasArg)
}

func TestNegatableSpecs(t *testing.T) {
	spec := complete.Spec{
		LongFlag:  "merge",
		ShortFlag: "m",
		Terse:     "Enable auto-merge",
	}
	pos, neg := complete.NegatableSpecs(spec, "", "", "")

	require.Equal(t, "merge", pos.LongFlag)
	require.Equal(t, "m", pos.ShortFlag)
	require.Equal(t, "Enable auto-merge", pos.Terse)

	require.Equal(t, "no-merge", neg.LongFlag)
	require.Empty(t, neg.ShortFlag)
	require.Equal(t, "Disable auto-merge", neg.Terse)
}

func TestNegatableSpecs_MultiByteTerse(t *testing.T) {
	spec := complete.Spec{
		LongFlag: "uber",
		Terse:    "\u00dcber fast mode",
	}
	pos, neg := complete.NegatableSpecs(spec, "", "", "")

	require.Equal(t, "\u00dcber fast mode", pos.Terse)
	// The negative description should lowercase the first rune correctly,
	// not corrupt the multi-byte UTF-8 character.
	require.Equal(t, "Disable \u00fcber fast mode", neg.Terse)
}

func TestNegatableSpecs_ExplicitDescs(t *testing.T) {
	spec := complete.Spec{
		LongFlag: "draft",
		Terse:    "Filter by draft",
	}
	pos, neg := complete.NegatableSpecs(spec, "Include drafts", "Exclude drafts", "")

	require.Equal(t, "Include drafts", pos.Terse)
	require.Equal(t, "no-draft", neg.LongFlag)
	require.Equal(t, "Exclude drafts", neg.Terse)
}

func TestNegatableSpecs_PreservesOtherFields(t *testing.T) {
	spec := complete.Spec{
		LongFlag:  "debug",
		ShortFlag: "d",
		HasArg:    false,
		Terse:     "Enable debug",
		Extension: "log",
	}
	pos, neg := complete.NegatableSpecs(spec, "", "", "")

	require.Equal(t, "d", pos.ShortFlag)
	require.Equal(t, "log", pos.Extension)
	require.False(t, pos.HasArg)

	// Negative variant only gets LongFlag and Terse.
	require.Empty(t, neg.ShortFlag)
	require.Empty(t, neg.Extension)
}

func TestParseClibTag(t *testing.T) {
	tests := []struct {
		name     string
		tag      string
		wantDesc string
		wantComp string
		wantGrp  string
	}{
		{"empty", "", "", "", ""},
		{"terse only", "terse='Draft filter'", "Draft filter", "", ""},
		{"complete only", "complete='predictor=repo'", "", "predictor=repo", ""},
		{"group only", "group='output'", "", "", "output"},
		{
			"all keys",
			"terse='Author',complete='predictor=author',group='people'",
			"Author",
			"predictor=author",
			"people",
		},
		{
			"complete with commas",
			"complete='predictor=columns,comma'",
			"",
			"predictor=columns,comma",
			"",
		},
		{"unquoted values", "terse=Simple,group=misc", "Simple", "", "misc"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var f complete.FlagMeta
			require.NoError(t, f.ParseClibTag(tt.tag))
			require.Equal(t, tt.wantDesc, f.Terse)
			require.Equal(t, tt.wantComp, f.Complete)
			require.Equal(t, tt.wantGrp, f.Group)
		})
	}
}

func TestParseClibTag_NegatableDescs(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("negatable,positive='Show output',negative='Hide output'"))
	require.True(t, f.Negatable)
	require.Equal(t, "Show output", f.PositiveDesc)
	require.Equal(t, "Hide output", f.NegativeDesc)
}

func TestParseCompleteTag(t *testing.T) {
	tests := []struct {
		tag           string
		wantPredictor string
		wantComma     bool
		wantValues    []string
	}{
		{"predictor=author", "author", false, nil},
		{"comma", "", true, nil},
		{"predictor=columns,comma", "columns", true, nil},
		{"comma,predictor=foo", "foo", true, nil},
		{"", "", false, nil},
		{"values=success failure pending", "", false, []string{"success", "failure", "pending"}},
		{"values=a b c,comma", "", true, []string{"a", "b", "c"}},
	}

	for _, tt := range tests {
		// Use FromFlags with manually constructed FlagMeta to test parseCompleteTag indirectly.
		gen := complete.NewGenerator("test").FromFlags([]complete.FlagMeta{
			{Name: "test", HasArg: true, Complete: tt.tag},
		})

		if len(gen.Specs) == 0 {
			continue
		}
		spec := gen.Specs[0]
		require.Equal(t, tt.wantPredictor, spec.Dynamic, "tag=%q dynamic", tt.tag)
		require.Equal(t, tt.wantComma, spec.CommaList, "tag=%q comma", tt.tag)
		if tt.wantValues != nil {
			require.Equal(t, tt.wantValues, spec.Values, "tag=%q values", tt.tag)
		}
	}
}

// --- Print tests ---

func TestGenerator_Print_DefaultShell(t *testing.T) {
	gen := newTestGen()
	var buf strings.Builder
	err := gen.Print(&buf, "")
	require.NoError(t, err)
	//nolint:dupword // fish script naturally contains repeated "end" keywords
	expected := `complete -c clibapp -f

# Comma-separated columns completion
function __clibapp_complete_columns
    set -l value (string replace -r '^--columns=' '' -- (commandline -ct))
    set -l columns (clibapp --@complete=columns)
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $columns
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $columns
    end
end

complete -c clibapp -s a -l author -x -a "(clibapp --@complete=author)" -d "Author"
complete -c clibapp -l ci -x -a "success failure pending" -d "CI status"
complete -c clibapp -l columns -x -kra "(__clibapp_complete_columns)" -d "Table columns"
complete -c clibapp -s L -l limit -r -d "Max results"
complete -c clibapp -l merge -d "Auto-merge"
complete -c clibapp -l no-merge -d "Disable auto-merge"
complete -c clibapp -s s -l state -x -a "open closed merged all" -d "State"
complete -c clibapp -s v -l verbose -d "Verbose"
`
	require.Equal(t, expected, buf.String())
}

func TestGenerator_Print_UnsupportedShell(t *testing.T) {
	gen := newTestGen()
	var buf strings.Builder
	err := gen.Print(&buf, "elvish")
	require.EqualError(t, err, `unsupported shell "elvish" (supported: bash, zsh, fish)`)
}

// --- Install tests ---

func TestGenerator_Install_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("fish", true)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "fish", "completions", "clibapp.fish")
	content, err := os.ReadFile(completionFile)
	require.NoError(t, err)
	//nolint:dupword // fish script naturally contains repeated "end" keywords
	expected := `complete -c clibapp -f

# Comma-separated columns completion
function __clibapp_complete_columns
    set -l value (string replace -r '^--columns=' '' -- (commandline -ct))
    set -l columns (clibapp --@complete=columns)
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $columns
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $columns
    end
end

complete -c clibapp -s a -l author -x -a "(clibapp --@complete=author)" -d "Author"
complete -c clibapp -l ci -x -a "success failure pending" -d "CI status"
complete -c clibapp -l columns -x -kra "(__clibapp_complete_columns)" -d "Table columns"
complete -c clibapp -s L -l limit -r -d "Max results"
complete -c clibapp -l merge -d "Auto-merge"
complete -c clibapp -l no-merge -d "Disable auto-merge"
complete -c clibapp -s s -l state -x -a "open closed merged all" -d "State"
complete -c clibapp -s v -l verbose -d "Verbose"
`
	require.Equal(t, expected, string(content))
}

func TestGenerator_Install_DefaultShell(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("", true)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "fish", "completions", "clibapp.fish")
	_, err = os.Stat(completionFile)
	require.NoError(t, err)
}

func TestGenerator_Install_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("bash", true)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "bash-completion", "completions", "clibapp")
	content, err := os.ReadFile(completionFile)
	require.NoError(t, err)
	expected := `# clibapp bash completion
_clibapp() {
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
                cmd="clibapp"
                ;;
            *)
                ;;
        esac
    done

    case "${cmd}" in
        clibapp)
            opts="--author -a --ci --columns --limit -L --merge --no-merge --state -s --verbose -v"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq 1 ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
            case "${prev}" in
                --author|-a)
                    COMPREPLY=($(compgen -W "$(clibapp --@complete=author 2>/dev/null)" -- "${cur}"))
                    return 0
                    ;;
                --ci)
                    COMPREPLY=($(compgen -W 'success failure pending' -- "${cur}"))
                    return 0
                    ;;
                --columns)
                    local prefix=""
                    local cur_val="${cur}"
                    local all_vals=($(clibapp --@complete=columns 2>/dev/null))
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
                --limit|-L)
                    COMPREPLY=()
                    return 0
                    ;;
                --state|-s)
                    COMPREPLY=($(compgen -W 'open closed merged all' -- "${cur}"))
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
    complete -F _clibapp -o nosort -o bashdefault -o default clibapp
else
    complete -F _clibapp -o bashdefault -o default clibapp
fi
`
	require.Equal(t, expected, string(content))
}

func TestGenerator_Install_Zsh(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("zsh", true)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "zsh", "site-functions", "_clibapp")
	content, err := os.ReadFile(completionFile)
	require.NoError(t, err)
	expected := `#compdef clibapp

autoload -U is-at-least

_clibapp() {
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
        '(-a --author)-a+[Author]:author:($(clibapp --@complete=author))' \
        '(-a --author)--author=[Author]:author:($(clibapp --@complete=author))' \
        '--ci=[CI status]:ci:(success failure pending)' \
        '--columns=[Table columns]:columns:{_sequence compadd - $(clibapp --@complete=columns)}' \
        '(-L --limit)-L+[Max results]: :_default' \
        '(-L --limit)--limit=[Max results]: :_default' \
        '--merge[Auto-merge]' \
        '--no-merge[Disable auto-merge]' \
        '(-s --state)-s+[State]:state:(open closed merged all)' \
        '(-s --state)--state=[State]:state:(open closed merged all)' \
        '(-v --verbose)-v[Verbose]' \
        '(-v --verbose)--verbose[Verbose]' \
    && ret=0
}

if [ "$funcstack[1]" = "_clibapp" ]; then
    _clibapp "$@"
else
    compdef _clibapp clibapp
fi
`
	require.Equal(t, expected, string(content))
}

func TestGenerator_Install_UnsupportedShell(t *testing.T) {
	gen := newTestGen()
	err := gen.Install("elvish", true)
	require.EqualError(t, err, `unsupported shell "elvish" (supported: bash, zsh, fish)`)
}

func TestGenerator_Install_NotQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("fish", false)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "fish", "completions", "clibapp.fish")
	_, err = os.Stat(completionFile)
	require.NoError(t, err)
}

// --- Uninstall tests ---

func TestGenerator_Uninstall_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := newTestGen()
	// Install first.
	err := gen.Install("fish", true)
	require.NoError(t, err)

	// Then uninstall.
	err = gen.Uninstall("fish", false)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "fish", "completions", "clibapp.fish")
	_, err = os.Stat(completionFile)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestGenerator_Uninstall_DefaultShell(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("", true)
	require.NoError(t, err)

	err = gen.Uninstall("", false)
	require.NoError(t, err)
}

func TestGenerator_Uninstall_NotFound(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := complete.NewGenerator("nonexistent")
	err := gen.Uninstall("fish", false)
	require.NoError(t, err) // Should not error when file doesn't exist.
}

func TestGenerator_Uninstall_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("bash", true)
	require.NoError(t, err)

	err = gen.Uninstall("bash", false)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "bash-completion", "completions", "clibapp")
	_, err = os.Stat(completionFile)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestGenerator_Uninstall_Zsh(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	gen := newTestGen()
	err := gen.Install("zsh", true)
	require.NoError(t, err)

	err = gen.Uninstall("zsh", false)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "zsh", "site-functions", "_clibapp")
	_, err = os.Stat(completionFile)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestGenerator_Uninstall_UnsupportedShell(t *testing.T) {
	gen := newTestGen()
	err := gen.Uninstall("elvish", false)
	require.EqualError(t, err, `unsupported shell "elvish"`)
}

// Test fishCompletionFile with default config dir (XDG_CONFIG_HOME unset).
func TestGenerator_Uninstall_DefaultConfigDir(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	gen := complete.NewGenerator("__clib_test_nonexistent__")
	err := gen.Uninstall("fish", false)
	require.NoError(t, err) // File doesn't exist -> no error.
}

// --- Additional ParseClibTag key tests ---

func TestParseClibTag_Placeholder(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("placeholder='<value>'"))
	require.Equal(t, "<value>", f.Placeholder)
	require.True(t, f.PlaceholderOverride)
}

func TestParseClibTag_Highlight(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("highlight='foo,bar'"))
	require.Equal(t, []string{"foo", "bar"}, f.EnumHighlight)
}

func TestParseClibTag_HighlightEmpty(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("highlight=''"))
	require.Nil(t, f.EnumHighlight)
}

func TestParseClibTag_Default(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("default='open'"))
	require.Equal(t, "open", f.EnumDefault)
}

func TestParseClibTag_AllKeys(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag(
		"terse='My flag',complete='predictor=author',group='people',placeholder='<name>',negatable,positive='Enable it',negative='Disable it',highlight='a,b',default='x'",
	))
	require.Equal(t, "My flag", f.Terse)
	require.Equal(t, "predictor=author", f.Complete)
	require.Equal(t, "people", f.Group)
	require.Equal(t, "<name>", f.Placeholder)
	require.True(t, f.PlaceholderOverride)
	require.True(t, f.Negatable)
	require.Equal(t, "Enable it", f.PositiveDesc)
	require.Equal(t, "Disable it", f.NegativeDesc)
	require.Equal(t, []string{"a", "b"}, f.EnumHighlight)
	require.Equal(t, "x", f.EnumDefault)
}

// --- Desc fallback ---

func TestFlagMeta_Desc_FallbackToHelp(t *testing.T) {
	f := complete.FlagMeta{Help: "Help text"}
	require.Equal(t, "Help text", f.Desc())
}

func TestFlagMeta_Desc_PreferDescription(t *testing.T) {
	f := complete.FlagMeta{Terse: "Description", Help: "Help text"}
	require.Equal(t, "Description", f.Desc())
}

// --- ParseClibTag ext ---

func TestParseClibTag_Ext(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("ext='yaml'"))
	require.Equal(t, "yaml", f.Extension)
}

func TestParseClibTag_ExtMultiple(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("ext='yaml,yml'"))
	require.Equal(t, "yaml,yml", f.Extension)
}

func TestFromFlags_Extension(t *testing.T) {
	gen := complete.NewGenerator("test").FromFlags([]complete.FlagMeta{
		{Name: "config", HasArg: true, Extension: "yaml"},
	})

	require.Len(t, gen.Specs, 1)
	require.Equal(t, "yaml", gen.Specs[0].Extension)
}

// --- ParseClibTag hint ---

func TestParseClibTag_Hint(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("hint='file'"))
	require.Equal(t, "file", f.ValueHint)
}

// --- ApplyMeta guard tests ---

func TestApplyMeta_DoesNotOverwriteExtension(t *testing.T) {
	spec := complete.Spec{Extension: "yaml"}
	complete.ApplyMeta(&spec, &complete.FlagMeta{})
	require.Equal(t, "yaml", spec.Extension, "empty meta should not clear pre-set Extension")
}

func TestApplyMeta_DoesNotOverwriteValueHint(t *testing.T) {
	spec := complete.Spec{ValueHint: "file"}
	complete.ApplyMeta(&spec, &complete.FlagMeta{})
	require.Equal(t, "file", spec.ValueHint, "empty meta should not clear pre-set ValueHint")
}

func TestApplyMeta_OverwritesExtensionWhenSet(t *testing.T) {
	spec := complete.Spec{Extension: "yaml"}
	complete.ApplyMeta(&spec, &complete.FlagMeta{Extension: "json"})
	require.Equal(t, "json", spec.Extension)
}

func TestApplyMeta_OverwritesValueHintWhenSet(t *testing.T) {
	spec := complete.Spec{ValueHint: "file"}
	complete.ApplyMeta(&spec, &complete.FlagMeta{ValueHint: "dir"})
	require.Equal(t, "dir", spec.ValueHint)
}

// --- Subcommand test generator helpers ---

func subcommandGen() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
			{
				LongFlag: "color",
				Terse:    "Color mode",
				HasArg:   true,
				Values:   []string{"auto", "always", "never"},
			},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "build",
				Terse: "Build the project",
				Specs: []complete.Spec{
					{LongFlag: "output", ShortFlag: "o", Terse: "Output path", HasArg: true},
					{LongFlag: "release", Terse: "Release build"},
				},
			},
			{
				Name:    "test",
				Aliases: []string{"t"},
				Terse:   "Run tests",
				Specs: []complete.Spec{
					{LongFlag: "coverage", Terse: "Enable coverage"},
					{LongFlag: "run", ShortFlag: "r", Terse: "Test pattern", HasArg: true},
				},
			},
		},
	}
}

func globalFlagsGen() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "config", Terse: "Config file", HasArg: true},
			{LongFlag: "no-config", Terse: "Disable config"},
			{LongFlag: "no-proxy", Terse: "Ignore proxies"},
			{LongFlag: "verbose", ShortFlag: "v", Terse: "Verbose output"},
			{
				LongFlag: "color",
				Terse:    "Color mode",
				HasArg:   true,
				Values:   []string{"auto", "always", "never"},
			},
			{ShortFlag: "h", Terse: "Show help"},
			{LongFlag: "help", Terse: "Show help"},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "run",
				Terse: "Run command",
				Specs: []complete.Spec{
					{LongFlag: "dry-run", ShortFlag: "n", Terse: "Dry run"},
					{
						LongFlag:  "output",
						ShortFlag: "o",
						Terse:     "Output format",
						HasArg:    true,
						Values:    []string{"text", "json"},
					},
				},
			},
			{
				Name:  "version",
				Terse: "Show version",
			},
		},
	}
}

// --- HandleAction tests ---

func TestHandleAction_Complete_WithHandler(t *testing.T) {
	var gotShell, gotKind string
	var gotArgs []string
	handler := func(shell, kind string, args []string) {
		gotShell = shell
		gotKind = kind
		gotArgs = args
	}

	a := complete.Action{
		Shell:    "fish",
		Complete: "namespaces",
		Args:     []string{"colima", "start"},
	}
	handled, err := complete.HandleAction(a, nil, handler, false)

	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "namespaces", gotKind)
	require.Equal(t, []string{"colima", "start"}, gotArgs)
}

func TestHandleAction_Complete_EmptyArgs(t *testing.T) {
	var gotArgs []string
	handler := func(_, _ string, args []string) {
		gotArgs = args
	}

	a := complete.Action{
		Shell:    "zsh",
		Complete: "flags",
	}
	handled, err := complete.HandleAction(a, nil, handler, false)

	require.NoError(t, err)
	require.True(t, handled)
	require.Nil(t, gotArgs)
}

func TestHandleAction_Complete_NilHandler(t *testing.T) {
	a := complete.Action{
		Shell:    "fish",
		Complete: "namespaces",
		Args:     []string{"colima"},
	}
	handled, err := complete.HandleAction(a, nil, nil, false)

	require.NoError(t, err)
	require.True(t, handled)
}

func TestHandleAction_NoAction(t *testing.T) {
	a := complete.Action{}
	handled, err := complete.HandleAction(a, nil, nil, false)

	require.NoError(t, err)
	require.False(t, handled)
}

func nestedGen() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "verbose", Terse: "Verbose"},
		},
		Subs: []complete.SubSpec{
			{
				Name:  "auth",
				Terse: "Manage authentication",
				Specs: []complete.Spec{
					{LongFlag: "token", Terse: "Auth token", HasArg: true},
				},
				Subs: []complete.SubSpec{
					{
						Name:  "login",
						Terse: "Log in",
						Specs: []complete.Spec{
							{LongFlag: "browser", Terse: "Open browser"},
						},
					},
					{
						Name:  "logout",
						Terse: "Log out",
					},
				},
			},
			{
				Name:  "run",
				Terse: "Run command",
			},
		},
	}
}
