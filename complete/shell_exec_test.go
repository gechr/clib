package complete_test

import (
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

// These tests execute the generated completion scripts through the real shells
// and assert how the dynamic-completion handler is invoked. They complement the
// golden tests (which lock the generated text) by locking the runtime behavior:
// context forwarding, the "--" terminator, positional slot selection, and the
// cross-scope short-alias guard. Each shell is skipped when it is not installed.

// genForwardShortCollision builds two sibling subcommands that both use -p for
// different long flags, exercising the cross-scope short-alias guard in
// allForwardableSpecs.
func genForwardShortCollision() *complete.Generator {
	return &complete.Generator{
		AppName: "myapp",
		Subs: []complete.SubSpec{
			{
				Name:  "one",
				Terse: "First group",
				Specs: []complete.Spec{
					{
						LongFlag:  "profile",
						ShortFlag: "p",
						Terse:     "Profile",
						HasArg:    true,
						Forward:   true,
					},
					{LongFlag: "target", Terse: "Target", HasArg: true, Dynamic: "target"},
				},
			},
			{
				Name:  "two",
				Terse: "Second group",
				Specs: []complete.Spec{
					{
						LongFlag:  "project",
						ShortFlag: "p",
						Terse:     "Project",
						HasArg:    true,
						Forward:   true,
					},
					{LongFlag: "target", Terse: "Target", HasArg: true, Dynamic: "target"},
				},
			},
		},
	}
}

// shellQuote single-quotes s for safe interpolation into a shell command.
func shellQuote(s string) string {
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

func lookShell(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skipf("%s not installed; skipping shell execution test", name)
	}
	return path
}

// completionEnv writes a stub binary that records the handler invocation to a
// log file plus the generated completion script, and returns the temp dir and
// log path. The stub is named after the generator's app so the script's
// "<app> --@complete=..." calls resolve to it via PATH.
// completionEnv returns (dir, scriptPath, logPath).
func completionEnv(t *testing.T, gen *complete.Generator, shell string) (string, string, string) {
	t.Helper()
	dir := t.TempDir()
	logPath := filepath.Join(dir, "handler.log")

	stub := "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \"$CLIB_COMPLETE_LOG\"\necho candidate\n"
	require.NoError(t, os.WriteFile(filepath.Join(dir, gen.AppName), []byte(stub), 0o755))

	var buf strings.Builder
	require.NoError(t, gen.Print(&buf, shell))
	scriptPath := filepath.Join(dir, "completion."+shell)
	require.NoError(t, os.WriteFile(scriptPath, []byte(buf.String()), 0o644))
	return dir, scriptPath, logPath
}

func shellEnv(dir, logPath string) []string {
	return append(os.Environ(),
		"PATH="+dir+string(os.PathListSeparator)+os.Getenv("PATH"),
		"CLIB_COMPLETE_LOG="+logPath,
	)
}

func readHandlerLog(t *testing.T, path string) string {
	t.Helper()
	data, err := os.ReadFile(path)
	if os.IsNotExist(err) {
		return ""
	}
	require.NoError(t, err)
	return strings.TrimSpace(string(data))
}

// driveBash sources the bash completion, simulates the given COMP_WORDS with the
// cursor at cword, invokes the completion function, and returns the recorded
// handler invocation.
func driveBash(t *testing.T, gen *complete.Generator, words []string, cword int) string {
	t.Helper()
	bash := lookShell(t, "bash")
	dir, scriptPath, logPath := completionEnv(t, gen, "bash")

	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = shellQuote(w)
	}
	cur, prev := "", ""
	if cword < len(words) {
		cur = words[cword]
	}
	if cword > 0 && cword-1 < len(words) {
		prev = words[cword-1]
	}
	compFunc := "_" + strings.ReplaceAll(gen.AppName, "-", "_")

	driver := strings.Join([]string{
		"source " + shellQuote(scriptPath),
		"COMP_WORDS=(" + strings.Join(quoted, " ") + ")",
		"COMP_CWORD=" + strconv.Itoa(cword),
		compFunc + " " + gen.AppName + " " + shellQuote(cur) + " " + shellQuote(prev),
	}, "\n")

	cmd := exec.Command(bash, "--norc", "-c", driver)
	cmd.Env = shellEnv(dir, logPath)
	_ = cmd.Run() // completion may exit non-zero; the recorded log is what matters
	return readHandlerLog(t, logPath)
}

// driveFish sources the fish completion and asks fish to complete the given
// command line (a trailing space requests completion of the next token).
func driveFish(t *testing.T, gen *complete.Generator, line string) string {
	t.Helper()
	fish := lookShell(t, "fish")
	dir, scriptPath, logPath := completionEnv(t, gen, "fish")

	driver := "source " + shellQuote(
		scriptPath,
	) + "\ncomplete -C " + shellQuote(
		line,
	) + " >/dev/null\n"
	cmd := exec.Command(fish, "--no-config", "-c", driver)
	cmd.Env = shellEnv(dir, logPath)
	_ = cmd.Run()
	return readHandlerLog(t, logPath)
}

// driveZshForwarded sources the zsh completion and runs its forwarded-flags
// helper over the given words, returning the collected __fwd entries. Driving
// zsh's full _arguments flow headlessly is unreliable, so this exercises the
// clib-authored collection logic (forwarding, the "--" terminator, and the
// short-alias guard) directly.
func driveZshForwarded(t *testing.T, gen *complete.Generator, words []string, current int) string {
	t.Helper()
	zsh := lookShell(t, "zsh")
	dir, scriptPath, logPath := completionEnv(t, gen, "zsh")

	quoted := make([]string, len(words))
	for i, w := range words {
		quoted[i] = shellQuote(w)
	}
	helper := "_" + strings.ReplaceAll(gen.AppName, "-", "_") + "_forwarded_flags"

	driver := strings.Join([]string{
		"source " + shellQuote(scriptPath) + " 2>/dev/null",
		"words=(" + strings.Join(quoted, " ") + ")",
		"CURRENT=" + strconv.Itoa(current),
		helper,
		`print -r -- "${__fwd[*]}"`,
	}, "\n")

	cmd := exec.Command(zsh, "-f", "-c", driver)
	cmd.Env = shellEnv(dir, logPath)
	out, _ := cmd.Output()
	return strings.TrimSpace(string(out))
}

// execCase is a single command-line scenario expressed for every shell driver.
type execCase struct {
	name string
	gen  *complete.Generator

	words []string // bash COMP_WORDS / zsh words (incl. app name and trailing "")
	cword int      // bash cursor index
	zsh   int      // zsh CURRENT (1-based, points at the cursor token)
	line  string   // fish command line (trailing space = complete next token)

	wantHandler string // exact handler invocation ("$*" the handler is called with)
	wantFwd     string // expected zsh "${__fwd[*]}" (zsh helper level)
}

func execCases() []execCase {
	fv := genForwardFlagValue()
	da := genForwardDynamicArgs()
	col := genForwardShortCollision()

	return []execCase{
		{
			name:        "flag_value_forwards_context",
			gen:         fv,
			words:       []string{"myapp", "-p", "prod", "deploy", "--target", ""},
			cword:       5,
			zsh:         6,
			line:        "myapp -p prod deploy --target ",
			wantHandler: "--@complete=target -- --profile=prod",
			wantFwd:     "--profile=prod",
		},
		{
			name:        "positional_slot0_no_context",
			gen:         da,
			words:       []string{"myapp", ""},
			cword:       1,
			zsh:         2,
			line:        "myapp ",
			wantHandler: "--@complete=items --",
			wantFwd:     "",
		},
		{
			name:        "positional_slot0_forwarded_equals",
			gen:         da,
			words:       []string{"myapp", "--category=alpha", ""},
			cword:       2,
			zsh:         3,
			line:        "myapp --category=alpha ",
			wantHandler: "--@complete=items -- --category=alpha",
			wantFwd:     "--category=alpha",
		},
		{
			name:        "positional_real_arg_advances_slot",
			gen:         da,
			words:       []string{"myapp", "--category=alpha", "widget", ""},
			cword:       3,
			zsh:         4,
			line:        "myapp --category=alpha widget ",
			wantHandler: "--@complete=values -- --category=alpha widget",
			wantFwd:     "--category=alpha",
		},
		{
			name:        "terminator_stops_forwarding",
			gen:         da,
			words:       []string{"myapp", "--", "--category", "alpha", ""},
			cword:       4,
			zsh:         5,
			line:        "myapp -- --category alpha ",
			wantHandler: "--@complete=values -- --category alpha",
			wantFwd:     "",
		},
		{
			name:        "collision_short_not_forwarded",
			gen:         col,
			words:       []string{"myapp", "two", "-p", "acme", "--target", ""},
			cword:       5,
			zsh:         6,
			line:        "myapp two -p acme --target ",
			wantHandler: "--@complete=target --",
			wantFwd:     "",
		},
		{
			name:        "collision_long_forwarded",
			gen:         col,
			words:       []string{"myapp", "two", "--project", "acme", "--target", ""},
			cword:       5,
			zsh:         6,
			line:        "myapp two --project acme --target ",
			wantHandler: "--@complete=target -- --project=acme",
			wantFwd:     "--project=acme",
		},
	}
}

func TestShellExec_Bash(t *testing.T) {
	for _, tc := range execCases() {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantHandler, driveBash(t, tc.gen, tc.words, tc.cword))
		})
	}
}

func TestShellExec_Fish(t *testing.T) {
	for _, tc := range execCases() {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantHandler, driveFish(t, tc.gen, tc.line))
		})
	}
}

func TestShellExec_ZshForwardedFlags(t *testing.T) {
	for _, tc := range execCases() {
		t.Run(tc.name, func(t *testing.T) {
			require.Equal(t, tc.wantFwd, driveZshForwarded(t, tc.gen, tc.words, tc.zsh),
				"zsh forwarded-flags helper produced unexpected context")
		})
	}
}
