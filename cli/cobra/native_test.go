package cobra_test

import (
	"bytes"
	"io"
	"strings"
	"testing"

	clibcobra "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func nativeTestRoot() *cobralib.Command {
	root := &cobralib.Command{Use: "app"}
	root.PersistentFlags().String("global", "", "Global text")
	root.Flags().Bool("root-only", false, "Root only")
	parent := &cobralib.Command{Use: "group"}
	parent.Flags().Bool("parent-only", false, "Parent only")
	cmd := &cobralib.Command{
		Use: "login", Aliases: []string{"l"}, Run: noop,
		Annotations: map[string]string{"clib": "dynamic-args='first, second'"},
	}
	cmd.Flags().StringP("format", "f", "", "Format")
	clibcobra.Extend(cmd.Flags().Lookup("format"), clibcobra.FlagExtra{
		Enum: []string{"json", "yaml"}, EnumTerse: []string{"JSON output", "YAML output"},
	})
	cmd.Flags().String("text", "", "Text")
	cmd.Flags().Int("number", 0, "Number")
	cmd.Flags().Bool("enabled", false, "Enabled")
	for name, hint := range map[string]string{"dir": "dir", "file": "file"} {
		cmd.Flags().String(name, "", name)
		clibcobra.Extend(cmd.Flags().Lookup(name), clibcobra.FlagExtra{Hint: hint})
	}
	cmd.Flags().String("config", "", "Config")
	clibcobra.Extend(cmd.Flags().Lookup("config"), clibcobra.FlagExtra{Extension: ".yaml, yml"})
	cmd.Flags().String("items", "", "Items")
	clibcobra.Extend(
		cmd.Flags().Lookup("items"),
		clibcobra.FlagExtra{Complete: "values=a b c,comma"},
	)
	cmd.Flags().Var(&clibcobra.CSVFlag{}, "csv", "CSV")
	clibcobra.Extend(cmd.Flags().Lookup("csv"), clibcobra.FlagExtra{Enum: []string{"a", "b", "c"}})
	cmd.Flags().StringSlice("slice", nil, "Slice")
	clibcobra.Extend(cmd.Flags().Lookup("slice"), clibcobra.FlagExtra{Complete: "values=a b c"})
	cmd.Flags().String("dynamic", "", "Dynamic")
	clibcobra.Extend(
		cmd.Flags().Lookup("dynamic"),
		clibcobra.FlagExtra{Complete: "predictor=choice"},
	)
	cmd.Flags().String("skip", "", "Skip")
	clibcobra.Extend(cmd.Flags().Lookup("skip"), clibcobra.FlagExtra{Complete: "-"})
	cmd.Flags().String("ordered", "", "Ordered")
	clibcobra.Extend(cmd.Flags().Lookup("ordered"), clibcobra.FlagExtra{
		Enum: []string{"b", "a"}, Order: complete.OrderKeep,
	})
	sibling := &cobralib.Command{Use: "other", Run: noop}
	sibling.Flags().Bool("sibling-only", false, "Sibling only")
	parent.AddCommand(cmd, sibling)
	root.AddCommand(parent)
	return root
}

func executeNative(t *testing.T, root *cobralib.Command, args ...string) string {
	t.Helper()
	var out bytes.Buffer
	root.SetOut(&out)
	root.SetErr(io.Discard)
	root.SetArgs(append([]string{cobralib.ShellCompRequestCmd}, args...))
	require.NoError(t, root.Execute())
	return out.String()
}

func TestNativeMappings(t *testing.T) {
	for _, tc := range []struct {
		name string
		args []string
		want string
	}{
		{"enum descriptions", []string{"--format", ""}, "json\tJSON output\nyaml\tYAML output\n:4\n"},
		{"short flag", []string{"-f", ""}, "json\tJSON output\nyaml\tYAML output\n:4\n"},
		{"text", []string{"--text", ""}, ":4\n"},
		{"number", []string{"--number", ""}, ":4\n"},
		{"bool", []string{"--enabled="}, "true\nfalse\n:4\n"},
		{"directory", []string{"--dir", ""}, ":16\n"},
		{"file", []string{"--file", ""}, ":0\n"},
		{"extensions", []string{"--config", ""}, "yaml\nyml\n:8\n"},
		{"comma", []string{"--items", "a,"}, "a,b\na,c\n:6\n"},
		{"CSV", []string{"--csv", "a,"}, "a,b\na,c\n:6\n"},
		{"slice", []string{"--slice", "a,"}, "a,b\na,c\n:6\n"},
		{"multiple used", []string{"--items", "a,b,"}, "a,b,c\n:6\n"},
		{"disabled", []string{"--skip", ""}, ":4\n"},
		{"order", []string{"--ordered", ""}, "b\na\n:36\n"},
		{"unknown current flag", []string{"--unknown="}, ":4\n"},
		{"unknown previous flag", []string{"--unknown", ""}, ":4\n"},
		{"unknown prior flag", []string{"--unknown=value", "--text", ""}, ":4\n"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			root := nativeTestRoot()
			_, err := clibcobra.RegisterCompletions(root)
			require.NoError(t, err)
			args := append([]string{"group", "l"}, tc.args...)
			require.Equal(t, tc.want, executeNative(t, root, args...))
		})
	}
}

func TestNativeContext(t *testing.T) {
	for _, tc := range []struct {
		args      []string
		kind      string
		flag      string
		word      string
		positions []string
		afterDash bool
	}{
		{args: []string{"group", "login", "p"}, kind: "first", word: "p"},
		{args: []string{"group", "login", "one", "t"}, kind: "second", positions: []string{"one"}, word: "t"},
		{args: []string{"group", "login", "one", "two", ""}, kind: "second", positions: []string{"one", "two"}},
		{args: []string{"--global=outer", "group", "l", "one", "--dynamic=x"}, kind: "choice", flag: "dynamic", positions: []string{"one"}, word: "x"},
		{args: []string{"group", "login", "one", "--", "--text"}, kind: "second", positions: []string{"one"}, word: "--text", afterDash: true},
		{args: []string{"group", "login", "--"}, word: "--"},
	} {
		t.Run(strings.Join(tc.args, " "), func(t *testing.T) {
			root := nativeTestRoot()
			called := false
			_, err := clibcobra.RegisterCompletions(
				root,
				clibcobra.WithCompletionHandler(
					func(c clibcobra.CompletionContext) ([]string, cobralib.ShellCompDirective) {
						called = true
						require.Equal(t, "app group login", c.Command.CommandPath())
						require.Equal(t, tc.kind, c.Kind)
						require.Equal(t, tc.word, c.ToComplete)
						require.Equal(t, append([]string{}, tc.positions...), c.Args)
						require.Equal(t, tc.afterDash, c.AfterDash)
						if tc.flag == "" {
							require.Nil(t, c.Flag)
						} else {
							require.Equal(t, tc.flag, c.Flag.Name)
							global, getErr := c.Command.Flags().GetString("global")
							require.NoError(t, getErr)
							require.Equal(t, "outer", global)
						}
						if c.AfterDash {
							return nil, cobralib.ShellCompDirectiveDefault
						}
						return []string{
							"answer\tdescription",
						}, cobralib.ShellCompDirectiveNoFileComp
					},
				),
			)
			require.NoError(t, err)
			out := executeNative(t, root, tc.args...)
			require.Equal(t, tc.kind != "", called)
			if tc.afterDash {
				require.Equal(t, ":0\n", out)
			}
		})
	}
}

func TestNativeCallbacksWin(t *testing.T) {
	for _, flag := range []bool{false, true} {
		root := nativeTestRoot()
		cmd, _, err := root.Find([]string{"group", "login"})
		require.NoError(t, err)
		callback := func(*cobralib.Command, []string, string) ([]string, cobralib.ShellCompDirective) {
			return []string{"native"}, cobralib.ShellCompDirectiveNoSpace
		}
		cmd.ValidArgsFunction = callback
		require.NoError(t, cmd.RegisterFlagCompletionFunc("format", callback))
		_, err = clibcobra.RegisterCompletions(root)
		require.NoError(t, err)
		args := []string{"group", "login"}
		if flag {
			args = append(args, "--format")
		}
		args = append(args, "")
		require.Equal(t, "native\n:2\n", executeNative(t, root, args...))
	}
}

func TestNativeFlagScopes(t *testing.T) {
	root := nativeTestRoot()
	_, err := clibcobra.RegisterCompletions(root)
	require.NoError(t, err)
	out := executeNative(t, root, "group", "login", "--text", "other", "--")
	names := strings.FieldsFunc(out, func(r rune) bool { return r == '\t' || r == '\n' })
	require.Contains(t, names, "--format")
	require.Contains(t, names, "--global")
	for _, excluded := range []string{"--sibling-only", "--parent-only", "--root-only"} {
		require.NotContains(t, names, excluded)
	}
}

func TestNativeUnknownCommand(t *testing.T) {
	root := nativeTestRoot()
	_, err := clibcobra.RegisterCompletions(root)
	require.NoError(t, err)
	require.Equal(t, ":4\n", executeNative(t, root, "bogus", ""))
	require.Nil(t, root.Args)
	require.False(t, root.FParseErrWhitelist.UnknownFlags)
	root.SetArgs([]string{"bogus"})
	require.Error(t, root.Execute(), "normal parsing must still reject unknown commands")
}

func TestNativeUnknownCommandDoesNotCompleteRootArgs(t *testing.T) {
	for _, annotation := range []string{"dynamic-args='rootitem'", "complete='path'"} {
		root := &cobralib.Command{Use: "app", Annotations: map[string]string{"clib": annotation}}
		root.AddCommand(&cobralib.Command{Use: "login", Run: noop})
		_, err := clibcobra.RegisterCompletions(root, clibcobra.WithCompletionHandler(
			func(clibcobra.CompletionContext) ([]string, cobralib.ShellCompDirective) {
				t.Fatal("unknown subcommand reached root predictor")
				return nil, cobralib.ShellCompDirectiveDefault
			}))
		require.NoError(t, err)
		require.Equal(t, ":4\n", executeNative(t, root, "bogus", ""))
	}
}

func TestNativePositionals(t *testing.T) {
	for _, tc := range []struct {
		annotation string
		valid      []string
		args       []string
		want       string
	}{
		{annotation: "complete='path'", args: []string{""}, want: ":0\n"},
		{annotation: "complete='path'", args: []string{"one", ""}, want: ":4\n"},
		{valid: []string{"one", "two"}, args: []string{""}, want: "one\ntwo\n:4\n"},
		{args: []string{""}, want: ":4\n"},
	} {
		root := &cobralib.Command{
			Use: "app <item>", Run: noop, Args: cobralib.ExactArgs(1),
			ValidArgs: tc.valid, Annotations: map[string]string{"clib": tc.annotation},
		}
		_, err := clibcobra.RegisterCompletions(root)
		require.NoError(t, err)
		require.Equal(t, tc.want, executeNative(t, root, tc.args...))
	}
}

func TestCommandMeta(t *testing.T) {
	require.Empty(t, clibcobra.CommandMeta(nil))
	root := &cobralib.Command{Use: "app", Annotations: map[string]string{
		"clib": "complete='path',dynamic-args='first, second'",
	}}
	require.Equal(
		t,
		clibcobra.CommandMetadata{PathArgs: true, DynamicArgs: []string{"first", "second"}},
		clibcobra.CommandMeta(root),
	)
}

func TestNativeHooksAndNoDescriptions(t *testing.T) {
	for _, useErrorHook := range []bool{false, true} {
		root := nativeTestRoot()
		calls := 0
		if useErrorHook {
			root.PersistentPreRunE = func(*cobralib.Command, []string) error { calls++; return nil }
		} else {
			root.PersistentPreRun = func(*cobralib.Command, []string) { calls++ }
		}
		for range 2 {
			_, err := clibcobra.RegisterCompletions(root)
			require.NoError(t, err)
		}
		var out bytes.Buffer
		root.SetOut(&out)
		root.SetErr(io.Discard)
		root.SetArgs([]string{"__completeNoDesc", "group", "login", "--format", ""})
		require.NoError(t, root.Execute())
		require.Equal(t, "json\nyaml\n:4\n", out.String())
		require.Equal(t, 1, calls)
		require.Nil(t, root.Args)
		root.SetArgs([]string{"group", "login"})
		require.NoError(t, root.Execute())
		require.Equal(t, 2, calls)
	}
}

func TestNativeAfterDashForNativeCallback(t *testing.T) {
	for _, request := range []string{"__complete", "__completeNoDesc"} {
		for _, after := range []bool{false, true} {
			root := nativeTestRoot()
			cmd, _, err := root.Find([]string{"group", "login"})
			require.NoError(t, err)
			called := false
			cmd.ValidArgsFunction = func(cmd *cobralib.Command, _ []string, _ string) ([]string, cobralib.ShellCompDirective) {
				called = true
				require.Equal(t, after, clibcobra.AfterDash(cmd))
				return nil, cobralib.ShellCompDirectiveNoFileComp
			}
			_, err = clibcobra.RegisterCompletions(root)
			require.NoError(t, err)
			args := []string{request, "group", "login"}
			if after {
				args = append(args, "--")
			}
			root.SetArgs(append(args, ""))
			root.SetOut(io.Discard)
			root.SetErr(io.Discard)
			require.NoError(t, root.Execute())
			require.True(t, called)
		}
	}
	require.False(t, clibcobra.AfterDash(nil))
}

func TestNativeDynamicCommaDescriptions(t *testing.T) {
	root := &cobralib.Command{Use: "app", Run: noop}
	root.Flags().StringSlice("items", nil, "Items")
	clibcobra.Extend(root.Flags().Lookup("items"), clibcobra.FlagExtra{Complete: "predictor=item"})
	_, err := clibcobra.RegisterCompletions(root, clibcobra.WithCompletionHandler(
		func(c clibcobra.CompletionContext) ([]string, cobralib.ShellCompDirective) {
			require.Equal(t, "a,b", c.ToComplete)
			return []string{"a\tFirst", "b\tSecond"}, cobralib.ShellCompDirectiveNoFileComp
		}))
	require.NoError(t, err)
	require.Equal(t, "a,b\tSecond\n:6\n", executeNative(t, root, "--items=a,b"))
}
