package urfave_test

import (
	"bytes"
	"context"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	urfavecli "github.com/gechr/clib/cli/urfave"
	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
	clilib "github.com/urfave/cli/v3"
)

func TestNewCompletion(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	before := len(cmd.Flags)
	c := urfavecli.NewCompletion(cmd)
	require.NotNil(t, c)
	// Should add 5 hidden flags.
	require.Len(t, cmd.Flags, before+5)
}

func TestHandle_NoFlags(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	c := urfavecli.NewCompletion(cmd)

	var result bool
	var resultErr error
	cmd.Action = func(_ context.Context, _ *clilib.Command) error {
		result, resultErr = c.Handle(nil, nil)
		return nil
	}
	err := cmd.Run(context.Background(), []string{"app"})
	require.NoError(t, err)
	require.False(t, result)
	require.NoError(t, resultErr)
}

func TestHandle_Complete_WithHandler(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	c := urfavecli.NewCompletion(cmd)

	var handlerShell, handlerKind string
	cmd.Action = func(_ context.Context, _ *clilib.Command) error {
		ok, err := c.Handle(nil, func(shell, kind string, _ []string) {
			handlerShell = shell
			handlerKind = kind
		})
		require.True(t, ok)
		require.NoError(t, err)
		return nil
	}
	err := cmd.Run(
		context.Background(),
		[]string{
			"app",
			"--" + complete.FlagComplete + "=flags",
			"--" + complete.FlagShell + "=fish",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "fish", handlerShell)
	require.Equal(t, "flags", handlerKind)
}

func TestHandle_Complete_NilHandler(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	c := urfavecli.NewCompletion(cmd)

	var result bool
	cmd.Action = func(_ context.Context, _ *clilib.Command) error {
		var err error
		result, err = c.Handle(nil, nil)
		require.NoError(t, err)
		return nil
	}
	err := cmd.Run(context.Background(), []string{"app", "--" + complete.FlagComplete + "=flags"})
	require.NoError(t, err)
	require.True(t, result)
}

func TestHandle_Complete_FromOSArgs(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	c := urfavecli.NewCompletion(cmd)

	oldArgs := os.Args
	t.Cleanup(func() { os.Args = oldArgs })
	os.Args = []string{
		"app",
		"--" + complete.FlagComplete + "=flags",
		"--" + complete.FlagShell + "=fish",
	}

	gen := complete.NewGenerator("app")
	var gotShell, gotKind string
	handled, err := c.Handle(gen, func(shell, kind string, _ []string) {
		gotShell = shell
		gotKind = kind
	})
	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "flags", gotKind)
}

func TestHandle_PrintCompletion(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	c := urfavecli.NewCompletion(cmd)

	gen := complete.NewGenerator("app")

	// Redirect stdout to capture output.
	r, w, err := os.Pipe()
	require.NoError(t, err)
	oldStdout := os.Stdout
	os.Stdout = w

	var result bool
	var handleErr error
	cmd.Action = func(_ context.Context, _ *clilib.Command) error {
		result, handleErr = c.Handle(gen, nil, urfavecli.WithQuiet(true))
		return nil
	}
	err = cmd.Run(
		context.Background(),
		[]string{"app", "--" + complete.FlagPrintCompletion, "--" + complete.FlagShell + "=fish"},
	)
	require.NoError(t, err)

	w.Close()
	os.Stdout = oldStdout
	var buf bytes.Buffer
	_, _ = buf.ReadFrom(r)
	r.Close()

	require.True(t, result)
	require.NoError(t, handleErr)
	require.Equal(t, "complete -c app -f\n\n", buf.String())
}

func TestHandle_InstallCompletion_FromOSArgs(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	c := urfavecli.NewCompletion(cmd)

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
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cmd := &clilib.Command{Name: "clibapp"}
	c := urfavecli.NewCompletion(cmd)
	gen := complete.NewGenerator("clibapp")

	var result bool
	var handleErr error
	cmd.Action = func(_ context.Context, _ *clilib.Command) error {
		result, handleErr = c.Handle(gen, nil, urfavecli.WithQuiet(true))
		return nil
	}
	err := cmd.Run(
		context.Background(),
		[]string{
			"clibapp",
			"--" + complete.FlagInstallCompletion,
			"--" + complete.FlagShell + "=fish",
		},
	)
	require.NoError(t, err)
	require.True(t, result)
	require.NoError(t, handleErr)

	// Verify the file was created.
	completionFile := filepath.Join(tmpDir, "fish", "completions", "clibapp.fish")
	_, err = os.Stat(completionFile)
	require.NoError(t, err)
}

func TestHandle_UninstallCompletion(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	// Create the completion file first.
	completionDir := filepath.Join(tmpDir, "fish", "completions")
	require.NoError(t, os.MkdirAll(completionDir, 0o755))
	completionFile := filepath.Join(completionDir, "clibapp.fish")
	require.NoError(t, os.WriteFile(completionFile, []byte("# test"), 0o600))

	cmd := &clilib.Command{Name: "clibapp"}
	c := urfavecli.NewCompletion(cmd)
	gen := complete.NewGenerator("clibapp")

	var result bool
	var handleErr error
	cmd.Action = func(_ context.Context, _ *clilib.Command) error {
		result, handleErr = c.Handle(gen, nil)
		return nil
	}
	err := cmd.Run(
		context.Background(),
		[]string{
			"clibapp",
			"--" + complete.FlagUninstallCompletion,
			"--" + complete.FlagShell + "=fish",
		},
	)
	require.NoError(t, err)
	require.True(t, result)
	require.NoError(t, handleErr)

	// Verify the file was removed.
	_, err = os.Stat(completionFile)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

// --- Subcommands tests ---

func TestSubcommands(t *testing.T) {
	child := &clilib.Command{
		Name:    "build",
		Aliases: []string{"b"},
		Usage:   "Build the project",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Aliases: []string{"o"}, Usage: "Output path"},
		},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
	require.Len(t, subs, 1)
	require.Equal(t, "build", subs[0].Name)
	require.Equal(t, []string{"b"}, subs[0].Aliases)
	require.Equal(t, "Build the project", subs[0].Terse)
	require.Len(t, subs[0].Specs, 1)
	require.Equal(t, "output", subs[0].Specs[0].LongFlag)
	require.Equal(t, "o", subs[0].Specs[0].ShortFlag)
	require.True(t, subs[0].Specs[0].HasArg)
}

func TestSubcommands_SkipsHidden(t *testing.T) {
	root := &clilib.Command{
		Name: "myapp",
		Commands: []*clilib.Command{
			{Name: "visible", Usage: "Visible command"},
			{Name: "hidden", Usage: "Hidden command", Hidden: true},
		},
	}

	subs := urfavecli.Subcommands(root)
	require.Len(t, subs, 1)
	require.Equal(t, "visible", subs[0].Name)
}

func TestSubcommands_SkipsHelpFlag(t *testing.T) {
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "output", Usage: "Output path"},
			&clilib.BoolFlag{Name: "help", Usage: "Show help"},
		},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
	require.Len(t, subs, 1)
	for _, spec := range subs[0].Specs {
		require.NotEqual(t, "help", spec.LongFlag, "should skip help flag")
	}
}

func TestSubcommands_EnumValues(t *testing.T) {
	formatFlag := &clilib.StringFlag{Name: "format", Usage: "Output format"}
	urfavecli.Extend(formatFlag, urfavecli.FlagExtra{
		Enum: []string{"json", "yaml", "text"},
	})
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{formatFlag},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
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
	subs := urfavecli.Subcommands(nil)
	require.Nil(t, subs)
}

func TestSubcommands_PathArgs(t *testing.T) {
	child := &clilib.Command{
		Name:  "open",
		Usage: "Open a file",
	}
	urfavecli.ExtendCommand(child, urfavecli.CommandExtra{PathArgs: true})
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
	require.Len(t, subs, 1)
	require.True(t, subs[0].PathArgs)
}

func TestSubcommands_MaxPositionalArgs(t *testing.T) {
	child := &clilib.Command{
		Name:      "find",
		Usage:     "Find a user",
		Arguments: []clilib.Argument{&clilib.StringArg{Name: "user"}},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
	require.Len(t, subs, 1)
	require.True(t, subs[0].HasMaxPositionalArgs)
	require.Equal(t, 1, subs[0].MaxPositionalArgs)
}

func TestSubcommands_MaxPositionalArgs_UnlimitedSlice(t *testing.T) {
	child := &clilib.Command{
		Name:      "tail",
		Usage:     "Tail files",
		Arguments: []clilib.Argument{&clilib.StringArgs{Name: "file", Max: -1}},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
	require.Len(t, subs, 1)
	require.False(t, subs[0].HasMaxPositionalArgs)
}

func TestSubcommands_Extension(t *testing.T) {
	configFlag := &clilib.StringFlag{Name: "config", Usage: "Config file"}
	urfavecli.Extend(configFlag, urfavecli.FlagExtra{
		Extension: "yaml",
	})
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{configFlag},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
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
	outputFlag := &clilib.StringFlag{Name: "output", Usage: "Output path"}
	urfavecli.Extend(outputFlag, urfavecli.FlagExtra{
		Hint: "file",
	})
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{outputFlag},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
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
	repoFlag := &clilib.StringFlag{Name: "repo", Usage: "Repository"}
	urfavecli.Extend(repoFlag, urfavecli.FlagExtra{
		Complete: "predictor=repo",
	})
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{repoFlag},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
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
	colFlag := &clilib.StringFlag{Name: "columns", Usage: "Columns"}
	urfavecli.Extend(colFlag, urfavecli.FlagExtra{
		Complete: "values=a b c,comma",
	})
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{colFlag},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
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
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{
			&clilib.BoolWithInverseFlag{
				Name:  "merge",
				Usage: "Enable auto-merge",
			},
		},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
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
	draftFlag := &clilib.BoolWithInverseFlag{
		Name:  "draft",
		Usage: "Filter by draft",
	}
	urfavecli.Extend(draftFlag, urfavecli.FlagExtra{
		PositiveDesc: "Include drafts",
		NegativeDesc: "Exclude drafts",
	})
	child := &clilib.Command{
		Name:  "child",
		Usage: "A child command",
		Flags: []clilib.Flag{draftFlag},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{child},
	}

	subs := urfavecli.Subcommands(root)
	require.Len(t, subs, 1)

	specMap := map[string]complete.Spec{}
	for _, s := range subs[0].Specs {
		specMap[s.LongFlag] = s
	}
	require.Len(t, specMap, 2)
	require.Equal(t, "Include drafts", specMap["draft"].Terse)
	require.Equal(t, "Exclude drafts", specMap["no-draft"].Terse)
}

func TestHandle_Complete_WithArgs(t *testing.T) {
	cmd := &clilib.Command{Name: "app"}
	c := urfavecli.NewCompletion(cmd)

	var gotShell, gotKind string
	var gotArgs []string
	cmd.Action = func(_ context.Context, _ *clilib.Command) error {
		ok, err := c.Handle(nil, func(shell, kind string, args []string) {
			gotShell = shell
			gotKind = kind
			gotArgs = args
		}, urfavecli.WithArgs([]string{"colima"}))
		require.True(t, ok)
		require.NoError(t, err)
		return nil
	}
	err := cmd.Run(
		context.Background(),
		[]string{
			"app",
			"--" + complete.FlagComplete + "=namespaces",
			"--" + complete.FlagShell + "=fish",
		},
	)
	require.NoError(t, err)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "namespaces", gotKind)
	require.Equal(t, []string{"colima"}, gotArgs)
}

func TestSubcommands_Nested(t *testing.T) {
	login := &clilib.Command{
		Name:  "login",
		Usage: "Log in",
		Flags: []clilib.Flag{
			&clilib.BoolFlag{Name: "browser", Usage: "Open browser"},
		},
	}
	logout := &clilib.Command{
		Name:  "logout",
		Usage: "Log out",
	}
	auth := &clilib.Command{
		Name:     "auth",
		Usage:    "Manage authentication",
		Commands: []*clilib.Command{login, logout},
		Flags: []clilib.Flag{
			&clilib.StringFlag{Name: "token", Usage: "Auth token"},
		},
	}
	root := &clilib.Command{
		Name:     "myapp",
		Commands: []*clilib.Command{auth},
	}

	subs := urfavecli.Subcommands(root)
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
