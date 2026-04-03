package complete_test

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"testing"

	"github.com/gechr/clib/complete"
	_ "github.com/gechr/clib/complete/fish"
	"github.com/stretchr/testify/require"
)

func TestPreflight_NoMatch(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "run", "--verbose"}
	_, _, ok := complete.Preflight()
	require.False(t, ok)
}

func TestPreflight_InstallCompletion(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--install-completion"}
	f, args, ok := complete.Preflight()
	require.True(t, ok)
	require.True(t, f.InstallCompletion)
	require.Nil(t, args)
}

func TestPreflight_PrintCompletion(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--print-completion"}
	f, args, ok := complete.Preflight()
	require.True(t, ok)
	require.True(t, f.PrintCompletion)
	require.Nil(t, args)
}

func TestPreflight_UninstallCompletion(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--uninstall-completion"}
	f, args, ok := complete.Preflight()
	require.True(t, ok)
	require.True(t, f.UninstallCompletion)
	require.Nil(t, args)
}

func TestPreflight_Complete(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--@complete=author", "--@shell=fish"}
	f, args, ok := complete.Preflight()
	require.True(t, ok)
	require.Equal(t, "author", f.Complete)
	require.Equal(t, "fish", f.Shell)
	require.Nil(t, args)
}

func TestPreflight_CompleteWithPositionalArgs(t *testing.T) {
	t.Setenv("_TEST_ARGS", "1")
	os.Args = []string{"myapp", "--@complete=resolve-kind", "--@shell=fish", "--", "team"}
	f, args, ok := complete.Preflight()
	require.True(t, ok)
	require.Equal(t, "resolve-kind", f.Complete)
	require.Equal(t, []string{"team"}, args)
}

func TestCompletionFlags_Handle_NoAction(t *testing.T) {
	f := complete.CompletionFlags{}
	handled, err := f.Handle(testPreflightGenerator(), nil)
	require.NoError(t, err)
	require.False(t, handled)
}

func TestCompletionFlags_Handle_Complete(t *testing.T) {
	var gotShell, gotKind string
	f := complete.CompletionFlags{Complete: "author", Shell: "fish"}
	handled, err := f.Handle(testPreflightGenerator(), func(shell, kind string, _ []string) {
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
	f := complete.CompletionFlags{Complete: "author"}
	handled, err := f.Handle(testPreflightGenerator(), func(shell, _ string, _ []string) {
		gotShell = shell
	})
	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
}

func TestCompletionFlags_Handle_CompleteNilHandler(t *testing.T) {
	f := complete.CompletionFlags{Complete: "author", Shell: "fish"}
	handled, err := f.Handle(testPreflightGenerator(), nil)
	require.NoError(t, err)
	require.True(t, handled)
}

func TestCompletionFlags_Handle_PrintCompletion(t *testing.T) {
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w

	f := complete.CompletionFlags{PrintCompletion: true, Shell: "fish"}
	handled, hErr := f.Handle(testPreflightGenerator(), nil)

	w.Close()
	os.Stdout = oldStdout

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.NoError(t, hErr)
	require.True(t, handled)
	require.Contains(t, buf.String(), "complete -c clibapp")
}

func TestCompletionFlags_Handle_InstallCompletion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	f := complete.CompletionFlags{InstallCompletion: true, Shell: "fish"}
	handled, err := f.Handle(testPreflightGenerator(), nil)

	require.NoError(t, err)
	require.True(t, handled)

	completionFile := filepath.Join(dir, "fish", "completions", "clibapp.fish")
	_, err = os.Stat(completionFile)
	require.NoError(t, err)
}

func TestCompletionFlags_Handle_InstallCompletionWithQuiet(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	oldStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w

	f := complete.CompletionFlags{InstallCompletion: true, Shell: "fish"}
	handled, hErr := f.Handle(testPreflightGenerator(), nil, complete.WithQuiet(true))

	w.Close()
	os.Stderr = oldStderr

	var buf bytes.Buffer
	_, _ = io.Copy(&buf, r)

	require.NoError(t, hErr)
	require.True(t, handled)
	require.Empty(t, buf.String(), "quiet mode should suppress install message")
}

func TestCompletionFlags_Handle_UninstallCompletion(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", dir)

	completionDir := filepath.Join(dir, "fish", "completions")
	require.NoError(t, os.MkdirAll(completionDir, 0o755))
	completionFile := filepath.Join(completionDir, "clibapp.fish")
	require.NoError(t, os.WriteFile(completionFile, []byte("# old"), 0o600))

	f := complete.CompletionFlags{UninstallCompletion: true, Shell: "fish"}
	handled, err := f.Handle(testPreflightGenerator(), nil)

	require.NoError(t, err)
	require.True(t, handled)

	_, err = os.Stat(completionFile)
	require.ErrorIs(t, err, fs.ErrNotExist)
}

func TestCompletionFlags_Handle_WithArgs(t *testing.T) {
	var gotShell, gotKind string
	var gotArgs []string
	f := complete.CompletionFlags{Complete: "namespaces", Shell: "fish"}
	handled, err := f.Handle(testPreflightGenerator(), func(shell, kind string, args []string) {
		gotShell = shell
		gotKind = kind
		gotArgs = args
	}, complete.WithArgs([]string{"colima"}))

	require.NoError(t, err)
	require.True(t, handled)
	require.Equal(t, "fish", gotShell)
	require.Equal(t, "namespaces", gotKind)
	require.Equal(t, []string{"colima"}, gotArgs)
}

func testPreflightGenerator() *complete.Generator {
	return complete.NewGenerator("clibapp").FromFlags([]complete.FlagMeta{
		{Name: "verbose", Short: "v", Help: "Verbose output"},
		{Name: "limit", Short: "L", Help: "Max results", HasArg: true},
	})
}
