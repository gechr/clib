package cobra_test

import (
	"bytes"
	"os"
	"testing"

	cobracli "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestNewCompletion_RegistersFlags(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)
	require.NotNil(t, c)

	pf := cmd.PersistentFlags()
	for _, name := range []string{
		complete.FlagComplete,
		complete.FlagShell,
		complete.FlagInstallCompletion,
		complete.FlagUninstallCompletion,
		complete.FlagPrintCompletion,
	} {
		f := pf.Lookup(name)
		require.NotNil(t, f, "flag %q should be registered", name)
	}
}

func TestHandle_NoAction(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)
	gen := complete.NewGenerator("app")

	handled, err := c.Handle(gen, nil)
	require.NoError(t, err)
	require.False(t, handled)
}

func TestHandle_Complete_WithHandler(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	// Simulate --@complete=author --@shell=fish
	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagComplete, "author"))
	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagShell, "fish"))

	gen := complete.NewGenerator("app")
	var gotShell, gotKind string
	handler := func(shell, kind string, _ []string) {
		gotShell = shell
		gotKind = kind
	}

	handled, err := c.Handle(gen, handler)
	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "author", gotKind)
}

func TestHandle_Complete_NilHandler(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagComplete, "repo"))

	gen := complete.NewGenerator("app")
	handled, err := c.Handle(gen, nil)
	require.NoError(t, err)
	require.True(t, handled)
}

func TestHandle_Complete_FromOSArgs(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{
		"app",
		"--" + complete.FlagComplete + "=repo",
		"--" + complete.FlagShell + "=fish",
	}

	gen := complete.NewGenerator("app")
	var gotShell, gotKind string
	handler := func(shell, kind string, _ []string) {
		gotShell = shell
		gotKind = kind
	}

	handled, err := c.Handle(gen, handler)
	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "repo", gotKind)
}

func TestHandle_Complete_DefaultShell(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagComplete, "repo"))
	// Don't set --shell; it should detect the shell.

	gen := complete.NewGenerator("app")
	var gotShell string
	handler := func(shell, _ string, _ []string) {
		gotShell = shell
	}

	handled, err := c.Handle(gen, handler)
	require.NoError(t, err)
	require.True(t, handled)
	require.NotEmpty(t, gotShell)
}

func TestHandle_InstallCompletion_FromOSArgs(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{
		"app",
		"--" + complete.FlagInstallCompletion,
		"--" + complete.FlagShell + "=elvish",
	}

	gen := complete.NewGenerator("app")
	handled, err := c.Handle(gen, nil)
	require.True(t, handled)
	require.EqualError(t, err, `unsupported shell "elvish" (supported: bash, zsh, fish)`)
}

func TestHandle_InstallCompletion(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagInstallCompletion, "true"))
	// Use unsupported shell to avoid filesystem writes.
	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagShell, "elvish"))

	gen := complete.NewGenerator("app")
	handled, err := c.Handle(gen, nil)
	require.True(t, handled)
	require.Equal(t, `unsupported shell "elvish" (supported: bash, zsh, fish)`, err.Error())
}

func TestHandle_UninstallCompletion(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagUninstallCompletion, "true"))
	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagShell, "elvish"))

	gen := complete.NewGenerator("app")
	handled, err := c.Handle(gen, nil)
	require.True(t, handled)
	require.Equal(t, `unsupported shell "elvish"`, err.Error())
}

func TestHandle_PrintCompletion(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagPrintCompletion, "true"))
	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagShell, "elvish"))

	gen := complete.NewGenerator("app")
	handled, err := c.Handle(gen, nil)
	require.True(t, handled)
	require.Equal(t, `unsupported shell "elvish" (supported: bash, zsh, fish)`, err.Error())
}

func TestHandle_WithQuiet(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagInstallCompletion, "true"))
	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagShell, "elvish"))

	gen := complete.NewGenerator("app")
	// WithQuiet exercises the options.go path.
	handled, err := c.Handle(gen, nil, cobracli.WithQuiet(true))
	require.True(t, handled)
	require.Error(t, err) // elvish unsupported, but WithQuiet was applied.
}

// --- Subcommands tests ---

// noop is a no-op Run function to make cobra commands "runnable"
// (required by IsAvailableCommand).
var noop = func(*cobralib.Command, []string) {}

func TestSubcommands(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	build := &cobralib.Command{
		Use:     "build",
		Aliases: []string{"b"},
		Short:   "Build the project",
		Run:     noop,
	}
	build.Flags().StringP("output", "o", "", "Output path")
	root.AddCommand(build)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.Equal(t, "build", subs[0].Name)
	require.Equal(t, []string{"b"}, subs[0].Aliases)
	require.Equal(t, "Build the project", subs[0].Terse)
	require.Len(t, subs[0].Specs, 1)
	require.Equal(t, "output", subs[0].Specs[0].LongFlag)
	require.Equal(t, "o", subs[0].Specs[0].ShortFlag)
	require.True(t, subs[0].Specs[0].HasArg)
	require.Equal(t, "Output path", subs[0].Specs[0].Terse)
}

func TestSubcommands_SkipsHidden(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	visible := &cobralib.Command{Use: "visible", Short: "Visible command", Run: noop}
	hidden := &cobralib.Command{Use: "hidden", Short: "Hidden command", Hidden: true, Run: noop}
	root.AddCommand(visible, hidden)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.Equal(t, "visible", subs[0].Name)
}

func TestSubcommands_SkipsHelpFlag(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("output", "", "Output path")
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	for _, spec := range subs[0].Specs {
		require.NotEqual(t, "help", spec.LongFlag, "should skip built-in help flag")
	}
}

func TestSubcommands_EnumValues(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("format", "text", "Output format")
	cobracli.Extend(child.Flags().Lookup("format"), cobracli.FlagExtra{
		Enum: []string{"json", "yaml", "text"},
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	var formatSpec *complete.Spec
	for i := range subs[0].Specs {
		if subs[0].Specs[i].LongFlag == "format" {
			formatSpec = &subs[0].Specs[i]
			break
		}
	}
	require.NotNil(t, formatSpec)
	require.True(t, formatSpec.HasArg)
	require.ElementsMatch(t, []string{"json", "yaml", "text"}, formatSpec.Values)
}

func TestSubcommands_NilCommand(t *testing.T) {
	subs := cobracli.Subcommands(nil)
	require.Nil(t, subs)
}

func TestSubcommands_PathArgs(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{
		Use:   "open",
		Short: "Open a file",
		Run:   noop,
		Annotations: map[string]string{
			"clib": "complete='path'",
		},
	}
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.True(t, subs[0].PathArgs)
}

func TestSubcommands_DynamicArgs(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{
		Use:   "resolve",
		Short: "Resolve alerts",
		Run:   noop,
		Annotations: map[string]string{
			"clib": "dynamic-args='incident, alert'",
		},
	}
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.Equal(t, []string{"incident", "alert"}, subs[0].DynamicArgs)
}

func TestSubcommands_MaxPositionalArgs_ExactArgs(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{
		Use:   "find <user>",
		Short: "Find a user",
		Run:   noop,
		Args:  cobralib.ExactArgs(1),
	}
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.True(t, subs[0].HasMaxPositionalArgs)
	require.Equal(t, 1, subs[0].MaxPositionalArgs)
}

func TestSubcommands_MaxPositionalArgs_MaximumNArgs(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{
		Use:   "tail [files]",
		Short: "Tail files",
		Run:   noop,
		Args:  cobralib.MaximumNArgs(2),
	}
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.True(t, subs[0].HasMaxPositionalArgs)
	require.Equal(t, 2, subs[0].MaxPositionalArgs)
}

func TestSubcommands_Extension(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("config", "", "Config file")
	cobracli.Extend(child.Flags().Lookup("config"), cobracli.FlagExtra{
		Extension: "yaml",
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	var configSpec *complete.Spec
	for i := range subs[0].Specs {
		if subs[0].Specs[i].LongFlag == "config" {
			configSpec = &subs[0].Specs[i]
			break
		}
	}
	require.NotNil(t, configSpec)
	require.Equal(t, "yaml", configSpec.Extension)
}

func TestSubcommands_ValueHint(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("output", "", "Output path")
	cobracli.Extend(child.Flags().Lookup("output"), cobracli.FlagExtra{
		Hint: "file",
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	var outputSpec *complete.Spec
	for i := range subs[0].Specs {
		if subs[0].Specs[i].LongFlag == "output" {
			outputSpec = &subs[0].Specs[i]
			break
		}
	}
	require.NotNil(t, outputSpec)
	require.Equal(t, "file", outputSpec.ValueHint)
}

func TestSubcommands_Complete(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("repo", "", "Repository")
	cobracli.Extend(child.Flags().Lookup("repo"), cobracli.FlagExtra{
		Complete: "predictor=repo",
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	var repoSpec *complete.Spec
	for i := range subs[0].Specs {
		if subs[0].Specs[i].LongFlag == "repo" {
			repoSpec = &subs[0].Specs[i]
			break
		}
	}
	require.NotNil(t, repoSpec)
	require.Equal(t, "repo", repoSpec.Dynamic)
}

func TestSubcommands_CommaList(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("columns", "", "Columns")
	cobracli.Extend(child.Flags().Lookup("columns"), cobracli.FlagExtra{
		Complete: "values=a b c,comma",
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	var colSpec *complete.Spec
	for i := range subs[0].Specs {
		if subs[0].Specs[i].LongFlag == "columns" {
			colSpec = &subs[0].Specs[i]
			break
		}
	}
	require.NotNil(t, colSpec)
	require.True(t, colSpec.CommaList)
	require.Equal(t, []string{"a", "b", "c"}, colSpec.Values)
}

func TestSubcommands_NegatableFlag(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().Bool("merge", false, "Enable auto-merge")
	cobracli.Extend(child.Flags().Lookup("merge"), cobracli.FlagExtra{
		Negatable: true,
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	specMap := map[string]complete.Spec{}
	for _, s := range subs[0].Specs {
		specMap[s.LongFlag] = s
	}
	require.Len(t, specMap, 2)
	require.Equal(t, "Enable auto-merge", specMap["merge"].Terse)
	require.Equal(t, "Disable auto-merge", specMap["no-merge"].Terse)
}

func TestSubcommands_NegatableFlag_ExplicitDescs(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().Bool("draft", false, "Filter by draft")
	cobracli.Extend(child.Flags().Lookup("draft"), cobracli.FlagExtra{
		Negatable:    true,
		PositiveDesc: "Include drafts",
		NegativeDesc: "Exclude drafts",
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	specMap := map[string]complete.Spec{}
	for _, s := range subs[0].Specs {
		specMap[s.LongFlag] = s
	}
	require.Equal(t, "Include drafts", specMap["draft"].Terse)
	require.Equal(t, "Exclude drafts", specMap["no-draft"].Terse)
}

func TestSubcommands_Nested(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	auth := &cobralib.Command{Use: "auth", Short: "Manage authentication"}
	auth.Flags().String("token", "", "Auth token")

	login := &cobralib.Command{Use: "login", Short: "Log in", Run: noop}
	login.Flags().Bool("browser", false, "Open browser")

	logout := &cobralib.Command{Use: "logout", Short: "Log out", Run: noop}
	auth.AddCommand(login, logout)
	root.AddCommand(auth)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.Equal(t, "auth", subs[0].Name)
	require.Len(t, subs[0].Subs, 2)

	childMap := map[string]complete.SubSpec{}
	for _, s := range subs[0].Subs {
		childMap[s.Name] = s
	}
	loginSub := childMap["login"]
	require.Equal(t, "Log in", loginSub.Terse)
	require.Len(t, loginSub.Specs, 1)
	require.Equal(t, "browser", loginSub.Specs[0].LongFlag)

	logoutSub := childMap["logout"]
	require.Equal(t, "Log out", logoutSub.Terse)
}

func TestSubcommands_SkipsDeprecated(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	visible := &cobralib.Command{Use: "visible", Short: "Visible command", Run: noop}
	deprecated := &cobralib.Command{
		Use:        "old",
		Short:      "Old command",
		Deprecated: "use visible instead",
		Run:        noop,
	}
	root.AddCommand(visible, deprecated)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)
	require.Equal(t, "visible", subs[0].Name)
}

func TestSubcommands_SkipsDeprecatedFlags(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("current", "", "Current flag")
	child.Flags().String("old-flag", "", "Old flag")
	_ = child.Flags().MarkDeprecated("old-flag", "use --current instead")
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	var names []string
	for _, spec := range subs[0].Specs {
		names = append(names, spec.LongFlag)
	}
	require.Equal(t, []string{"current"}, names)
}

func TestHandle_Complete_WithArgs(t *testing.T) {
	cmd := &cobralib.Command{Use: "app"}
	c := cobracli.NewCompletion(cmd)

	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagComplete, "namespaces"))
	require.NoError(t, cmd.PersistentFlags().Set(complete.FlagShell, "fish"))

	gen := complete.NewGenerator("app")
	var gotShell, gotKind string
	var gotArgs []string
	handler := func(shell, kind string, args []string) {
		gotShell = shell
		gotKind = kind
		gotArgs = args
	}

	handled, err := c.Handle(gen, handler, cobracli.WithArgs([]string{"colima"}))
	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "namespaces", gotKind)
	require.Equal(t, []string{"colima"}, gotArgs)
}

func TestSubcommands_AliasesDoNotCreateSpecs(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	child := &cobralib.Command{Use: "child", Short: "A child command", Run: noop}
	child.Flags().String("format", "", "Output format")
	cobracli.Extend(child.Flags().Lookup("format"), cobracli.FlagExtra{
		Aliases:   []string{"fmt"},
		Enum:      []string{"json", "yaml", "text"},
		Extension: "json",
		Hint:      "file",
		Complete:  "predictor=format,comma",
	})
	root.AddCommand(child)

	subs := cobracli.Subcommands(root)
	require.Len(t, subs, 1)

	require.Len(t, subs[0].Specs, 1)
	spec := subs[0].Specs[0]
	require.Equal(t, "format", spec.LongFlag)
	require.True(t, spec.HasArg)
	require.Equal(t, "Output format", spec.Terse)
	require.Equal(t, []string{"json", "yaml", "text"}, spec.Values)
	require.Equal(t, "json", spec.Extension)
	require.Equal(t, "file", spec.ValueHint)
	require.Equal(t, "format", spec.Dynamic)
	require.True(t, spec.CommaList)
	require.False(t, spec.Hidden)
}

func TestGenerator_PersistentFlagsPropagateToNestedSubcommands(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	root.Flags().Bool("root-local", false, "Root local")
	root.PersistentFlags().Bool("root-persistent", false, "Root persistent")

	parent := &cobralib.Command{Use: "parent", Short: "Parent command", Run: noop}
	parent.Flags().Bool("parent-local", false, "Parent local")
	parent.PersistentFlags().Bool("parent-persistent", false, "Parent persistent")

	child := &cobralib.Command{Use: "child", Short: "Child command", Run: noop}
	child.Flags().Bool("child-local", false, "Child local")

	parent.AddCommand(child)
	root.AddCommand(parent)

	gen := complete.NewGenerator("myapp").FromFlags(cobracli.FlagMeta(root))
	gen.Subs = cobracli.Subcommands(root)

	require.Equal(t, []complete.Spec{
		{LongFlag: "root-local", Terse: "Root local", HasArg: false},
		{LongFlag: "root-persistent", Terse: "Root persistent", HasArg: false, Persistent: true},
	}, gen.Specs)
	require.Equal(t, []complete.SubSpec{
		{
			Name:  "parent",
			Terse: "Parent command",
			Specs: []complete.Spec{
				{LongFlag: "parent-local", Terse: "Parent local", HasArg: false},
				{
					LongFlag:   "parent-persistent",
					Terse:      "Parent persistent",
					HasArg:     false,
					Persistent: true,
				},
			},
			Subs: []complete.SubSpec{
				{
					Name:  "child",
					Terse: "Child command",
					Specs: []complete.Spec{
						{LongFlag: "child-local", Terse: "Child local", HasArg: false},
					},
				},
			},
		},
	}, gen.Subs)
}

// --- CompletionCommand tests ---

func TestCompletionCommand_Fish(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	gen := complete.NewGenerator("myapp")
	cmd := cobracli.CompletionCommand(root, func() *complete.Generator { return gen })
	root.AddCommand(cmd)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"completion", "fish"})
	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "complete -c myapp")
}

func TestCompletionCommand_Bash(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	gen := complete.NewGenerator("myapp")
	cmd := cobracli.CompletionCommand(root, func() *complete.Generator { return gen })
	root.AddCommand(cmd)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"completion", "bash"})
	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "_myapp()")
}

func TestCompletionCommand_Zsh(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	gen := complete.NewGenerator("myapp")
	cmd := cobracli.CompletionCommand(root, func() *complete.Generator { return gen })
	root.AddCommand(cmd)

	var buf bytes.Buffer
	root.SetOut(&buf)
	root.SetArgs([]string{"completion", "zsh"})
	require.NoError(t, root.Execute())
	require.Contains(t, buf.String(), "#compdef myapp")
}

func TestCompletionCommand_DisablesDefault(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	gen := complete.NewGenerator("myapp")
	cobracli.CompletionCommand(root, func() *complete.Generator { return gen })

	require.True(t, root.CompletionOptions.DisableDefaultCmd)
}

func TestCompletionCommand_Hidden(t *testing.T) {
	root := &cobralib.Command{Use: "myapp"}
	gen := complete.NewGenerator("myapp")
	cmd := cobracli.CompletionCommand(root, func() *complete.Generator { return gen })

	require.True(t, cmd.Hidden)
}
