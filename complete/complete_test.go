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

func genFlat() *complete.Generator {
	return complete.NewGenerator("clibapp").FromFlags(testFlags())
}

func TestGenerator_FromFlags(t *testing.T) {
	gen := genFlat()

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
		{"order keep", "order=keep", "", "", ""},
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

func TestParseClibTag_Order(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("order=keep"))
	require.Equal(t, complete.OrderKeep, f.Order)
}

func TestParseClibTag_OrderShell(t *testing.T) {
	var f complete.FlagMeta
	require.NoError(t, f.ParseClibTag("order=shell"))
	require.Equal(t, complete.OrderShell, f.Order)
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
	gen := genFlat()
	var buf strings.Builder
	err := gen.Print(&buf, "")
	require.NoError(t, err)
	//nolint:dupword // fish script naturally contains repeated "end" keywords
	expected := `complete -c clibapp -f

# Comma-separated columns completion
function __clibapp_complete_columns
    set -l value (string replace -r '^--columns=' '' -- (commandline -ct))
    set -l _flag_columns (clibapp --@complete=columns)
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $_flag_columns
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $_flag_columns
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
	gen := genFlat()
	var buf strings.Builder
	err := gen.Print(&buf, "tcsh")
	require.EqualError(
		t,
		err,
		`unsupported shell "tcsh" (supported: bash, zsh, fish, pwsh, elvish, nu)`,
	)
}

func TestGenerator_Print_FishOrderKeep(t *testing.T) {
	gen := complete.NewGenerator("ordered").FromFlags([]complete.FlagMeta{
		{
			Name:     "item",
			HasArg:   true,
			Complete: "predictor=item",
			Order:    complete.OrderKeep,
			Help:     "Item",
		},
		{
			Name:   "mode",
			HasArg: true,
			Enum:   []string{"fast", "safe"},
			Order:  complete.OrderKeep,
			Help:   "Mode",
		},
	})

	var buf strings.Builder
	err := gen.Print(&buf, "fish")
	require.NoError(t, err)
	require.Equal(t, `complete -c ordered -f

complete -c ordered -l item -k -x -a "(ordered --@complete=item)" -d "Item"
complete -c ordered -l mode -k -x -a "fast safe" -d "Mode"
`, buf.String())
}

func TestGenerator_Print_FishDefaultOrderKeep(t *testing.T) {
	gen := complete.NewGenerator("ordered", complete.WithOrder(complete.OrderKeep)).
		FromFlags([]complete.FlagMeta{
			{
				Name:     "item",
				HasArg:   true,
				Complete: "predictor=item",
				Help:     "Item",
			},
			{
				Name:   "mode",
				HasArg: true,
				Enum:   []string{"fast", "safe"},
				Help:   "Mode",
			},
		})

	var buf strings.Builder
	err := gen.Print(&buf, "fish")
	require.NoError(t, err)
	require.Equal(t, `complete -c ordered -f

complete -c ordered -l item -k -x -a "(ordered --@complete=item)" -d "Item"
complete -c ordered -l mode -k -x -a "fast safe" -d "Mode"
`, buf.String())
}

func TestGenerator_Print_FishOrderShellOverridesDefault(t *testing.T) {
	gen := complete.NewGenerator("ordered", complete.WithOrder(complete.OrderKeep)).
		FromFlags([]complete.FlagMeta{
			{
				Name:     "item",
				HasArg:   true,
				Complete: "predictor=item",
				Help:     "Item",
			},
			{
				Name:   "mode",
				HasArg: true,
				Enum:   []string{"fast", "safe"},
				Order:  complete.OrderShell,
				Help:   "Mode",
			},
		})

	var buf strings.Builder
	err := gen.Print(&buf, "fish")
	require.NoError(t, err)
	require.Equal(t, `complete -c ordered -f

complete -c ordered -l item -k -x -a "(ordered --@complete=item)" -d "Item"
complete -c ordered -l mode -x -a "fast safe" -d "Mode"
`, buf.String())
}

func TestGenerator_HiddenFlagCompletions(t *testing.T) {
	flags := []complete.FlagMeta{
		{Name: "visible", Help: "Visible"},
		{Name: "secret", Help: "Secret", Hidden: true},
		{Name: "debug", Help: "Debug", Hidden: true, CompleteWhenHidden: true},
	}

	// Default: hidden flags are omitted, except the per-flag complete-hidden
	// opt-in, which is still offered.
	var def strings.Builder
	require.NoError(t, complete.NewGenerator("hid").FromFlags(flags).Print(&def, "fish"))
	require.Contains(t, def.String(), "-l visible")
	require.Contains(t, def.String(), "-l debug")
	require.NotContains(t, def.String(), "-l secret")

	// WithIncludeHidden: every hidden flag is offered.
	var all strings.Builder
	require.NoError(t, complete.NewGenerator("hid", complete.WithIncludeHidden()).
		FromFlags(flags).Print(&all, "fish"))
	require.Contains(t, all.String(), "-l visible")
	require.Contains(t, all.String(), "-l debug")
	require.Contains(t, all.String(), "-l secret")
}

// --- Install tests ---

func TestGenerator_Install_Fish(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := genFlat()
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
    set -l _flag_columns (clibapp --@complete=columns)
    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $_flag_columns
            if not contains -- $col $selected
                printf '%s\n' "$prefix$col"
            end
        end
    else
        printf '%s\n' $_flag_columns
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

	gen := genFlat()
	err := gen.Install("", true)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "fish", "completions", "clibapp.fish")
	_, err = os.Stat(completionFile)
	require.NoError(t, err)
}

func TestGenerator_Install_Bash(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	gen := genFlat()
	err := gen.Install("bash", true)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "bash-completion", "completions", "clibapp")
	content, err := os.ReadFile(completionFile)
	require.NoError(t, err)
	var buf strings.Builder
	require.NoError(t, gen.Print(&buf, "bash"))
	require.Equal(t, buf.String(), string(content))
}

func TestGenerator_Install_Zsh(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_DATA_HOME", tmpDir)

	gen := genFlat()
	err := gen.Install("zsh", true)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "zsh", "site-functions", "_clibapp")
	content, err := os.ReadFile(completionFile)
	require.NoError(t, err)
	var buf strings.Builder
	require.NoError(t, gen.Print(&buf, "zsh"))
	require.Equal(t, buf.String(), string(content))
}

func TestGenerator_Install_UnsupportedShell(t *testing.T) {
	gen := genFlat()
	err := gen.Install("tcsh", true)
	require.EqualError(
		t,
		err,
		`unsupported shell "tcsh" (supported: bash, zsh, fish, pwsh, elvish, nu)`,
	)
}

func TestGenerator_Install_NotQuiet(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	gen := genFlat()
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

	gen := genFlat()
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

	gen := genFlat()
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

	gen := genFlat()
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

	gen := genFlat()
	err := gen.Install("zsh", true)
	require.NoError(t, err)

	err = gen.Uninstall("zsh", false)
	require.NoError(t, err)

	completionFile := filepath.Join(tmpDir, "zsh", "site-functions", "_clibapp")
	_, err = os.Stat(completionFile)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestGenerator_Uninstall_UnsupportedShell(t *testing.T) {
	gen := genFlat()
	err := gen.Uninstall("tcsh", false)
	require.EqualError(t, err, `unsupported shell "tcsh"`)
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

func TestFlagMeta_Desc_MultilineHelp(t *testing.T) {
	f := complete.FlagMeta{Help: "First line.\nSecond line.\nThird line."}
	require.Equal(t, "First line.", f.Desc())
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

func genSubcommands() *complete.Generator {
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

func genGlobalFlags() *complete.Generator {
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

func genNested() *complete.Generator {
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

// TestForwardFlagValue_ForwardsContext verifies that dynamic flag-value
// completions forward collected context flags to the handler, and that the
// forwarded-flags helper is emitted, for every shell.
func TestForwardFlagValue_ForwardsContext(t *testing.T) {
	cases := map[string]struct {
		helper    string
		plainCall string
		commaCall string
	}{
		"fish": {
			helper:    "function __myapp_forwarded_flags",
			plainCall: "myapp --@complete=target -- (__myapp_forwarded_flags)",
			commaCall: "myapp --@complete=tag -- (__myapp_forwarded_flags)",
		},
		"bash": {
			helper:    "_myapp_forwarded_flags() {",
			plainCall: `myapp --@complete=target -- "${__fwd[@]}"`,
			commaCall: `myapp --@complete=tag -- "${__fwd[@]}"`,
		},
		"zsh": {
			helper:    "_myapp_forwarded_flags() {",
			plainCall: `myapp --@complete=target -- "${__fwd[@]}" 2>/dev/null`,
			commaCall: `myapp --@complete=tag -- "${__fwd[@]}" 2>/dev/null`,
		},
	}

	for sh, want := range cases {
		t.Run(sh, func(t *testing.T) {
			var buf strings.Builder
			require.NoError(t, genForwardFlagValue().Print(&buf, sh))
			got := buf.String()
			require.Contains(t, got, want.helper, "forwarded-flags helper should be defined")
			require.Contains(t, got, want.plainCall, "dynamic flag value should forward context")
			require.Contains(
				t,
				got,
				want.commaCall,
				"comma dynamic flag value should forward context",
			)
		})
	}
}

// TestForwardFlagValue_StopsAtTerminator verifies that the forwarded-flags
// helper stops scanning at the "--" terminator, so tokens after it (which are
// positional, not flags) are never collected as forwarded context.
func TestForwardFlagValue_StopsAtTerminator(t *testing.T) {
	terminatorBreak := map[string]string{
		"fish": "        else if test \"$t\" = --\n            break\n",
		"bash": "        if [[ \"${COMP_WORDS[j]}\" == \"--\" ]]; then\n            break\n",
		"zsh":  "        if [[ $token == -- ]]; then\n            break\n",
	}

	for sh, want := range terminatorBreak {
		t.Run(sh, func(t *testing.T) {
			var buf strings.Builder
			require.NoError(t, genForwardFlagValue().Print(&buf, sh))
			require.Contains(t, buf.String(), want,
				"forwarded-flags helper must break on the -- terminator")
		})
	}
}

// TestForwardDynamicArgs_NotCountedAsPositional verifies that a forwarded flag
// value is not stored in the positional-counting array (which would shift the
// active completion slot) and that forwarded context still reaches the first
// slot's handler call.
func TestForwardDynamicArgs_NotCountedAsPositional(t *testing.T) {
	cases := map[string]struct {
		firstSlotForwards string // handler call for the first positional slot
		positionalAppend  string // old (buggy) append of the forwarded value
	}{
		"fish": {
			firstSlotForwards: "myapp --@complete=items -- $forwarded",
			positionalAppend:  "set -a positional (string replace -r '^-c=' '--category='",
		},
		"bash": {
			firstSlotForwards: `myapp --@complete=items -- "${__fwd[@]}"`,
			positionalAppend:  `__dyn_pos+=("--category=`,
		},
		"zsh": {
			firstSlotForwards: `myapp --@complete=items -- "${__fwd[@]}"`,
			positionalAppend:  `__pos+=("--category=`,
		},
	}

	for sh, want := range cases {
		t.Run(sh, func(t *testing.T) {
			var buf strings.Builder
			require.NoError(t, genForwardDynamicArgs().Print(&buf, sh))
			got := buf.String()
			require.Contains(t, got, want.firstSlotForwards,
				"first slot must forward context, proving the forwarded flag is not counted")
			require.NotContains(t, got, want.positionalAppend,
				"forwarded flag value must not be appended to the positional array")
		})
	}
}

// TestForwardDynamicArgs_StopsAtTerminator verifies that positional dynamic
// completions, which route forwarded context through the shared helper, do not
// collect forwardable flags that appear after the "--" terminator.
func TestForwardDynamicArgs_StopsAtTerminator(t *testing.T) {
	terminatorBreak := map[string]string{
		"fish": "        else if test \"$t\" = --\n            break\n",
		"bash": "        if [[ \"${COMP_WORDS[j]}\" == \"--\" ]]; then\n            break\n",
		"zsh":  "        if [[ $token == -- ]]; then\n            break\n",
	}
	for sh, want := range terminatorBreak {
		t.Run(sh, func(t *testing.T) {
			var buf strings.Builder
			require.NoError(t, genForwardDynamicArgs().Print(&buf, sh))
			require.Contains(t, buf.String(), want,
				"positional forwarded-flags helper must break on the -- terminator")
		})
	}
}

// TestForwardFlagValue_NoForwardSpecs verifies that without any forwardable
// flag, dynamic completions stay context-free and no helper is emitted.
func TestForwardFlagValue_NoForwardSpecs(t *testing.T) {
	gen := &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "target", Terse: "Target name", HasArg: true, Dynamic: "target"},
		},
	}

	for _, sh := range []string{"fish", "bash", "zsh"} {
		t.Run(sh, func(t *testing.T) {
			var buf strings.Builder
			require.NoError(t, gen.Print(&buf, sh))
			got := buf.String()
			require.NotContains(t, got, "forwarded_flags")
			require.NotContains(t, got, "--@complete=target --")
		})
	}
}

func TestZshDynamicValuesSplitOnNewlines(t *testing.T) {
	gen := &complete.Generator{
		AppName: "myapp",
		Specs: []complete.Spec{
			{LongFlag: "labels", Terse: "Labels", HasArg: true, Dynamic: "labels", CommaList: true},
			{LongFlag: "status", Terse: "Status", HasArg: true, Dynamic: "status"},
		},
	}

	var buf strings.Builder
	require.NoError(t, gen.Print(&buf, "zsh"))
	got := buf.String()

	require.Contains(
		t,
		got,
		`'--labels=[Labels]:labels:{ local -a items; items=(${(f)"$(myapp --@complete=labels)"}); _sequence compadd - "${(@)items}" }'`,
	)
	require.Contains(
		t,
		got,
		`'--status=[Status]:status:{ local -a items; items=(${(f)"$(myapp --@complete=status)"}); compadd -a items }'`,
	)
}
