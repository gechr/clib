package kong_test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/cli/kong"
	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

const expectedFishCompletion = `complete -c clibapp -f

complete -c clibapp -s L -l limit -r -d "Max results"
complete -c clibapp -s v -l verbose -d "Verbose output"
`

func testGenerator() *complete.Generator {
	return complete.NewGenerator("clibapp").FromFlags([]complete.FlagMeta{
		{Name: "verbose", Short: "v", Help: "Verbose output"},
		{Name: "limit", Short: "L", Help: "Max results", HasArg: true},
	})
}

func TestPreflight_NoMatch(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "complete", "author"}
	_, _, ok := kong.Preflight()
	require.False(t, ok)
}

func TestPreflight_InstallCompletion(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--install-completion"}
	f, args, ok := kong.Preflight()
	require.True(t, ok)
	require.True(t, f.InstallCompletion)
	require.Nil(t, args)
}

func TestPreflight_PrintCompletion(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--print-completion"}
	f, args, ok := kong.Preflight()
	require.True(t, ok)
	require.True(t, f.PrintCompletion)
	require.Nil(t, args)
}

func TestPreflight_UninstallCompletion(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--uninstall-completion"}
	f, args, ok := kong.Preflight()
	require.True(t, ok)
	require.True(t, f.UninstallCompletion)
	require.Nil(t, args)
}

func TestPreflight_Complete(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--@complete=author", "--@shell=fish"}
	f, args, ok := kong.Preflight()
	require.True(t, ok)
	require.Equal(t, "author", f.Complete)
	require.Equal(t, "fish", f.Shell)
	require.Nil(t, args)
}

func TestPreflight_CompleteWithPositionalArgs(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--@complete=resolve-kind", "--@shell=fish", "--", "team"}
	f, args, ok := kong.Preflight()
	require.True(t, ok)
	require.Equal(t, "resolve-kind", f.Complete)
	require.Equal(t, []string{"team"}, args)
}

func TestCompletionFlags_Handle_NoAction(t *testing.T) {
	f := kong.CompletionFlags{}
	handled, err := f.Handle(testGenerator(), nil)

	require.NoError(t, err)
	require.False(t, handled)
}

func TestCompletionFlags_Handle_Complete(t *testing.T) {
	var gotShell, gotKind string
	f := kong.CompletionFlags{Complete: "author", Shell: "fish"}
	handled, err := f.Handle(testGenerator(), func(shell, kind string, _ []string) {
		gotShell = shell
		gotKind = kind
	})

	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "author", gotKind)
}

func TestCompletionFlags_Handle_CompleteDetectsShell(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "fish")

	var gotShell string
	f := kong.CompletionFlags{Complete: "author"}
	handled, err := f.Handle(testGenerator(), func(shell, _ string, _ []string) {
		gotShell = shell
	})

	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
}

func TestCompletionFlags_Handle_CompleteNilHandler(t *testing.T) {
	f := kong.CompletionFlags{Complete: "author", Shell: "fish"}
	handled, err := f.Handle(testGenerator(), nil)

	require.NoError(t, err)
	require.True(t, handled)
}

func TestCompletionFlags_Handle_PrintCompletion(t *testing.T) {
	// Capture stdout since Handle writes to os.Stdout.
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	f := kong.CompletionFlags{PrintCompletion: true, Shell: "fish"}
	handled, hErr := f.Handle(testGenerator(), nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.NoError(t, hErr)
	require.True(t, handled)
	require.Equal(t, expectedFishCompletion, buf.String())
}

func TestCompletionFlags_Handle_PrintCompletionDefaultShell(t *testing.T) {
	// Empty shell should auto-detect; force fish via COMPLETE_SHELL.
	t.Setenv("COMPLETE_SHELL", "fish")

	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	f := kong.CompletionFlags{PrintCompletion: true}
	handled, hErr := f.Handle(testGenerator(), nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.NoError(t, hErr)
	require.True(t, handled)
	require.Equal(t, expectedFishCompletion, buf.String())
}

func TestCompletionFlags_Handle_InstallCompletion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	f := kong.CompletionFlags{InstallCompletion: true, Shell: "fish"}
	handled, err := f.Handle(testGenerator(), nil)

	require.NoError(t, err)
	require.True(t, handled)

	completionFile := filepath.Join(dir, "fish", "completions", "clibapp.fish")
	content, err := os.ReadFile(completionFile)
	require.NoError(t, err)
	require.Equal(t, expectedFishCompletion, string(content))
}

func TestCompletionFlags_Handle_InstallCompletionWithQuiet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Capture stderr to verify quiet suppresses the install message.
	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	f := kong.CompletionFlags{InstallCompletion: true, Shell: "fish"}
	handled, hErr := f.Handle(testGenerator(), nil, kong.WithQuiet(true))

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.NoError(t, hErr)
	require.True(t, handled)
	require.Empty(t, buf.String(), "quiet mode should suppress install message")

	// File should still be created.
	completionFile := filepath.Join(dir, "fish", "completions", "clibapp.fish")
	_, err = os.Stat(completionFile)
	require.NoError(t, err)
}

func TestCompletionFlags_Handle_UninstallCompletion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Pre-create the completion file.
	completionDir := filepath.Join(dir, "fish", "completions")
	require.NoError(t, os.MkdirAll(completionDir, 0o755))
	completionFile := filepath.Join(completionDir, "clibapp.fish")
	require.NoError(t, os.WriteFile(completionFile, []byte("# old"), 0o600))

	f := kong.CompletionFlags{UninstallCompletion: true, Shell: "fish"}
	handled, err := f.Handle(testGenerator(), nil)

	require.NoError(t, err)
	require.True(t, handled)

	_, err = os.Stat(completionFile)
	require.ErrorIs(t, err, fs.ErrNotExist, "completion file should be removed")
}

func TestCompletionFlags_Handle_UninstallCompletionNotFound(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	// Uninstalling when no file exists should not error.
	f := kong.CompletionFlags{UninstallCompletion: true, Shell: "fish"}
	handled, err := f.Handle(testGenerator(), nil)

	require.NoError(t, err)
	require.True(t, handled)
}

// --- Subcommands tests ---

func TestSubcommands(t *testing.T) {
	type BuildCmd struct {
		Output string `name:"output" help:"Output path" short:"o"`
	}
	type TestCmd struct {
		Coverage bool `name:"coverage" help:"Enable coverage"`
	}
	var cli struct {
		Build   BuildCmd `help:"Build the project" cmd:""`
		Test    TestCmd  `help:"Run tests"         aliases:"t"           cmd:""`
		Verbose bool     `name:"verbose"           help:"Verbose output" short:"v"`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 2)

	// Find subcommands by name (order not guaranteed).
	subMap := map[string]complete.SubSpec{}
	for _, s := range subs {
		subMap[s.Name] = s
	}

	build := subMap["build"]
	require.Equal(t, "Build the project", build.Terse)
	require.Len(t, build.Specs, 1)
	require.Equal(t, "output", build.Specs[0].LongFlag)
	require.Equal(t, "o", build.Specs[0].ShortFlag)
	require.True(t, build.Specs[0].HasArg)

	test := subMap["test"]
	require.Equal(t, "Run tests", test.Terse)
	require.Equal(t, []string{"t"}, test.Aliases)
	require.Len(t, test.Specs, 1)
	require.Equal(t, "coverage", test.Specs[0].LongFlag)
	require.False(t, test.Specs[0].HasArg)
}

func TestSubcommands_SkipsHidden(t *testing.T) {
	type PublicCmd struct{}
	type HiddenCmd struct{}
	var cli struct {
		Public PublicCmd `help:"Public command" cmd:""`
		Hidden HiddenCmd `help:"Hidden command" hidden:"" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.Equal(t, "public", subs[0].Name)
}

func TestSubcommands_SkipsHelpFlag(t *testing.T) {
	type Cmd struct {
		Output string `name:"output" help:"Output path"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	for _, spec := range subs[0].Specs {
		require.NotEqual(t, "help", spec.LongFlag, "should skip kong's built-in help flag")
	}
}

func TestSubcommands_EnumValues(t *testing.T) {
	type Cmd struct {
		Format string `name:"format" help:"Output format" default:"text" enum:"json,yaml,text"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
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

func TestSubcommands_NilParser(t *testing.T) {
	subs := kong.Subcommands(nil)
	require.Nil(t, subs)
}

func TestSubcommands_PathArgs_Predictor(t *testing.T) {
	type Cmd struct {
		Paths []string `arg:"" optional:"" predictor:"path"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.True(t, subs[0].PathArgs)
}

func TestSubcommands_PathArgs_ClibAnnotation(t *testing.T) {
	type Cmd struct{}
	var cli struct {
		Build Cmd `help:"Build" clib:"complete='path'" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.True(t, subs[0].PathArgs)
}

func TestSubcommands_MaxPositionalArgs(t *testing.T) {
	type Cmd struct {
		User string `name:"user" arg:""`
	}
	var cli struct {
		Find Cmd `help:"Find" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.True(t, subs[0].HasMaxPositionalArgs)
	require.Equal(t, 1, subs[0].MaxPositionalArgs)
}

func TestSubcommands_MaxPositionalArgs_UnlimitedSlice(t *testing.T) {
	type Cmd struct {
		Files []string `name:"file" arg:"" optional:""`
	}
	var cli struct {
		Tail Cmd `help:"Tail" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.False(t, subs[0].HasMaxPositionalArgs)
}

func TestSubcommands_FlagTerse(t *testing.T) {
	type Cmd struct {
		Output string `name:"output"  help:"Output format for results"                      short:"o" clib:"terse='Output format'"`
		DryRun bool   `name:"dry-run" help:"Show what would be done without making changes" short:"n" clib:"terse='Dry run'"`
		Limit  int    `name:"limit"   help:"Maximum number of results"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)

	specMap := map[string]complete.Spec{}
	for _, s := range subs[0].Specs {
		specMap[s.LongFlag] = s
	}

	// Flags with clib terse should use it.
	require.Equal(t, "Output format", specMap["output"].Terse)
	require.Equal(t, "Dry run", specMap["dry-run"].Terse)

	// Flags without clib terse should fall back to help.
	require.Equal(t, "Maximum number of results", specMap["limit"].Terse)
}

func TestSubcommands_Extension(t *testing.T) {
	type Cmd struct {
		Config string `name:"config" help:"Config file" clib:"ext='yaml'"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
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
	type Cmd struct {
		Output string `name:"output" help:"Output path" clib:"hint='file'"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
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
	type Cmd struct {
		Repo string `name:"repo" help:"Repository" clib:"complete='predictor=repo'"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
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

func TestSubcommands_NegatableFlag_NativeTag(t *testing.T) {
	type Cmd struct {
		Merge bool `name:"merge" help:"Enable auto-merge" negatable:""`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)

	specMap := map[string]complete.Spec{}
	for _, s := range subs[0].Specs {
		specMap[s.LongFlag] = s
	}
	require.Equal(t, map[string]complete.Spec{
		"merge":    {LongFlag: "merge", Persistent: true, Terse: "Enable auto-merge"},
		"no-merge": {LongFlag: "no-merge", Persistent: true, Terse: "Disable auto-merge"},
	}, specMap)
}

func TestSubcommands_NegatableFlag_ClibTag(t *testing.T) {
	type Cmd struct {
		Draft bool `name:"draft" help:"Filter by draft" clib:"negatable,positive='Include drafts',negative='Exclude drafts'"`
	}
	var cli struct {
		Build Cmd `help:"Build" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)

	specMap := map[string]complete.Spec{}
	for _, s := range subs[0].Specs {
		specMap[s.LongFlag] = s
	}
	require.Equal(t, map[string]complete.Spec{
		"draft":    {LongFlag: "draft", Persistent: true, Terse: "Include drafts"},
		"no-draft": {LongFlag: "no-draft", Persistent: true, Terse: "Exclude drafts"},
	}, specMap)
}

func TestCompletionFlags_Handle_Complete_WithArgs(t *testing.T) {
	var gotShell, gotKind string
	var gotArgs []string
	f := kong.CompletionFlags{Complete: "namespaces", Shell: "fish"}
	handled, err := f.Handle(testGenerator(), func(shell, kind string, args []string) {
		gotShell = shell
		gotKind = kind
		gotArgs = args
	}, kong.WithArgs([]string{"colima"}))

	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "namespaces", gotKind)
	require.Equal(t, []string{"colima"}, gotArgs)
}

func TestSubcommands_DynamicArgs_Predictor(t *testing.T) {
	type CompleteCmd struct {
		Kind string `arg:"" predictor:"complete-kind"`
	}
	type ResolveCmd struct {
		Kind  string `arg:"" predictor:"resolve-kind"`
		Value string `arg:""`
	}
	var cli struct {
		Complete CompleteCmd `help:"Tab completions" cmd:""`
		Resolve  ResolveCmd  `help:"Resolution"      cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	subMap := map[string]complete.SubSpec{}
	for _, s := range subs {
		subMap[s.Name] = s
	}

	// complete <kind>: single positional with predictor.
	comp := subMap["complete"]
	require.Equal(t, []string{"complete-kind"}, comp.DynamicArgs)
	require.True(t, comp.HasMaxPositionalArgs)
	require.Equal(t, 1, comp.MaxPositionalArgs)

	// resolve <kind> <value>: first positional has predictor, second does not.
	// DynamicArgs should stop at the first arg without a predictor.
	res := subMap["resolve"]
	require.Equal(t, []string{"resolve-kind"}, res.DynamicArgs)
	require.True(t, res.HasMaxPositionalArgs)
	require.Equal(t, 2, res.MaxPositionalArgs)
}

func TestSubcommands_Nested(t *testing.T) {
	type LoginCmd struct {
		Browser bool `name:"browser" help:"Open browser"`
	}
	type LogoutCmd struct{}
	type AuthCmd struct {
		Login  LoginCmd  `help:"Log in"  cmd:""`
		Logout LogoutCmd `help:"Log out" cmd:""`
		Token  string    `name:"token"   help:"Auth token"`
	}
	var cli struct {
		Auth AuthCmd `help:"Manage authentication" cmd:""`
	}
	parser, err := konglib.New(&cli, konglib.Name("myapp"))
	require.NoError(t, err)

	subs := kong.Subcommands(parser)
	require.Len(t, subs, 1)
	require.Equal(t, "auth", subs[0].Name)
	require.Len(t, subs[0].Subs, 2)

	// Find nested subcommands by name.
	childMap := map[string]complete.SubSpec{}
	for _, s := range subs[0].Subs {
		childMap[s.Name] = s
	}
	login := childMap["login"]
	require.Equal(t, "Log in", login.Terse)
	require.Len(t, login.Specs, 1)
	require.Equal(t, "browser", login.Specs[0].LongFlag)

	logout := childMap["logout"]
	require.Equal(t, "Log out", logout.Terse)
}

// --- CompletionCommand tests ---

func TestCompletionCommand_Fish(t *testing.T) {
	gen := complete.NewGenerator("myapp")

	var cli struct{}
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	parser, err := konglib.New(&cli,
		konglib.Name("myapp"),
		kong.CompletionCommand(func() *complete.Generator { return gen }),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	kctx, err := parser.Parse([]string{"completion", "fish"})
	require.NoError(t, err)
	require.NoError(t, kctx.Run())

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.Contains(t, buf.String(), "complete -c myapp")
}

func TestCompletionCommand_Bash(t *testing.T) {
	gen := complete.NewGenerator("myapp")

	var cli struct{}
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	parser, err := konglib.New(&cli,
		konglib.Name("myapp"),
		kong.CompletionCommand(func() *complete.Generator { return gen }),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	kctx, err := parser.Parse([]string{"completion", "bash"})
	require.NoError(t, err)
	require.NoError(t, kctx.Run())

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.Contains(t, buf.String(), "_myapp()")
}

func TestCompletionCommand_Zsh(t *testing.T) {
	gen := complete.NewGenerator("myapp")

	var cli struct{}
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	parser, err := konglib.New(&cli,
		konglib.Name("myapp"),
		kong.CompletionCommand(func() *complete.Generator { return gen }),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	kctx, err := parser.Parse([]string{"completion", "zsh"})
	require.NoError(t, err)
	require.NoError(t, kctx.Run())

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.Contains(t, buf.String(), "#compdef myapp")
}

func TestCompletionCommand_InvalidShell(t *testing.T) {
	var cli struct{}
	parser, err := konglib.New(&cli,
		konglib.Name("myapp"),
		kong.CompletionCommand(func() *complete.Generator {
			return complete.NewGenerator("myapp")
		}),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	_, err = parser.Parse([]string{"completion", "elvish"})
	require.Error(t, err)
}

func TestCompletionCommand_Hidden(t *testing.T) {
	var cli struct{}
	parser, err := konglib.New(&cli,
		konglib.Name("myapp"),
		kong.CompletionCommand(func() *complete.Generator {
			return complete.NewGenerator("myapp")
		}),
		konglib.Exit(func(int) {}),
	)
	require.NoError(t, err)

	// The completion command should not appear in visible children.
	for _, child := range parser.Model.Children {
		if child.Name == "completion" {
			require.True(t, child.Hidden)
		}
	}
}
