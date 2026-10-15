package cobra_test

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	clibcobra "github.com/gechr/clib/cli/cobra"
	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
	"github.com/stretchr/testify/require"
)

func TestGoldenNativeShells(t *testing.T) {
	root := &cobralib.Command{Use: "nativeapp"}
	gen, err := clibcobra.RegisterCompletions(root)
	require.NoError(t, err)
	for _, sh := range []string{"bash", "fish", "zsh"} {
		t.Run(sh, func(t *testing.T) {
			var buf bytes.Buffer
			require.NoError(t, gen.Print(&buf, sh))
			path := filepath.Join("testdata", "native_"+sh+".golden")
			if *update {
				require.NoError(t, os.WriteFile(path, buf.Bytes(), 0o600))
				return
			}
			want, readErr := os.ReadFile(path)
			require.NoError(t, readErr)
			require.Equal(t, string(want), buf.String())
		})
	}
}

func TestNativeGeneratorIsolation(t *testing.T) {
	root := &cobralib.Command{Use: "nativeapp"}
	_, err := clibcobra.RegisterCompletions(root)
	require.NoError(t, err)
	var out bytes.Buffer
	require.NoError(t, complete.NewGenerator("legacyapp").Print(&out, "bash"))
	//nolint:gocritic // check the protocol independently of the complete legacy script
	require.NotContains(t, out.String(), "__complete")
}

func TestNativeShellDrivers(t *testing.T) {
	dir := t.TempDir()
	binary := filepath.Join(dir, "nativeapp")
	build := exec.CommandContext(
		t.Context(),
		"go",
		"build",
		"-buildvcs=false",
		"-o",
		binary,
		"./testdata/native",
	)
	out, err := build.CombinedOutput()
	require.NoError(t, err, "%s", out)
	for _, name := range []string{"directory", "directory with spaces"} {
		require.NoError(t, os.Mkdir(filepath.Join(dir, name), 0o700))
	}
	for _, name := range []string{"directory.json", "file.txt", "file.yaml", "two words.yaml"} {
		require.NoError(t, os.WriteFile(filepath.Join(dir, name), nil, 0o600))
	}
	// Only the fixture and system programs are visible to the clients.
	environment := []string{
		"PATH=" + dir + ":/usr/bin:/bin",
		"HOME=" + dir,
		"XDG_CONFIG_HOME=" + dir,
	}
	scripts := make(map[string]string)
	for _, sh := range []string{"bash", "fish", "zsh"} {
		cmd := exec.CommandContext(t.Context(), binary, "completion", sh)
		cmd.Env = environment
		script, printErr := cmd.Output()
		require.NoError(t, printErr)
		path := filepath.Join(dir, "completion."+sh)
		require.NoError(t, os.WriteFile(path, script, 0o600))
		scripts[sh] = path
		t.Run(sh+" syntax", func(t *testing.T) {
			check := exec.CommandContext(t.Context(), nativeShell(t, sh), "-n", path)
			output, checkErr := check.CombinedOutput()
			require.NoError(t, checkErr, "%s", output)
		})
	}
	t.Run("bash", func(t *testing.T) { testNativeBash(t, dir, environment, scripts["bash"]) })
	t.Run("fish", func(t *testing.T) { testNativeFish(t, dir, environment, scripts["fish"]) })
	t.Run("zsh", func(t *testing.T) { testNativeZsh(t, dir, environment, scripts["zsh"]) })
	t.Run("zsh literal prefixes", func(t *testing.T) {
		testNativeZshLiteralPrefixes(t, dir, environment, scripts["zsh"])
	})
}

func testNativeZshLiteralPrefixes(t *testing.T, dir string, environment []string, script string) {
	t.Helper()
	for _, flag := range []string{"format", "config", "dir", "file"} {
		t.Run(flag, func(t *testing.T) {
			driver := `compdef() { :; }
source "$1"
shift
words=("$@")
CURRENT=${#words[@]}
_describe() { [[ $words[-1] != --file=* ]] && printf '%s\n' "$@"; }
_files() { printf '%s\n' "$@"; }
_arguments() { eval "${1#*:*:}"; }
_nativeapp`
			word := "--" + flag + "=x$(touch evaluated)="
			cmd := exec.CommandContext(t.Context(), nativeShell(t, "zsh"), "-f", "-c",
				driver, "zsh", script, "nativeapp", "login", word)
			cmd.Dir, cmd.Env = dir, environment
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", out)
			require.NoFileExists(t, filepath.Join(dir, "evaluated"))
			require.Contains(t, strings.Split(string(out), "\n"), word)
		})
	}
}

func nativeShell(t *testing.T, name string) string {
	t.Helper()
	path, err := exec.LookPath(name)
	if err != nil {
		t.Skip(name + " unavailable")
	}
	return path
}

func testNativeBash(t *testing.T, dir string, environment []string, script string) {
	t.Helper()
	bashes := []string{nativeShell(t, "bash")}
	if _, err := os.Stat("/bin/bash"); err == nil && !slices.Contains(bashes, "/bin/bash") {
		bashes = append(bashes, "/bin/bash")
	}
	for _, bash := range bashes {
		t.Run(bash, func(t *testing.T) {
			for _, tc := range []struct {
				name  string
				line  string
				words []string
				want  []string
				tail  string
			}{
				{name: "equals split", line: "nativeapp login --format=j", words: []string{"nativeapp", "login", "--format", "=", "j"}, want: []string{"json "}},
				{name: "quoted colon", line: `nativeapp login "ns:v`, want: []string{"ns:value"}},
				{name: "single quoted colon", line: `nativeapp login 'ns:v`, want: []string{"ns:value"}},
				{name: "escaped colon", line: `nativeapp login ns\:v`, want: []string{"ns:value "}},
				{name: "quoted flag colon", line: `nativeapp login --target="ns:v`, want: []string{"ns:value"}},
				{name: "tilde executable", line: "~/nativeapp login --format=j", want: []string{"json "}},
				{name: "colon split", line: "nativeapp login ns:v", words: []string{"nativeapp", "login", "ns", ":", "v"}, want: []string{"value "}},
				{name: "equals and colon", line: "nativeapp login --target=ns:v", want: []string{"value "}},
				{name: "mixed quoted prefix", line: `nativeapp login pro"fi`, want: []string{"file"}},
				{name: "quoted prefix", line: `nativeapp login "two w`, want: []string{"two words"}},
				{name: "escaped prefix", line: `nativeapp login two\ w`, want: []string{`two\ words `}},
				{name: "quoted preceding", line: `nativeapp login --text "two words" --format j`, want: []string{"json "}},
				{name: "text no files", line: "nativeapp login --text "},
				{name: "directories", line: "nativeapp login --dir dir", want: []string{`directory/`, `directory\ with\ spaces/`}},
				{name: "directory no space", line: `nativeapp login --dir "directory with`, want: []string{"directory with spaces/"}},
				{name: "extensions", line: "nativeapp login --config file", want: []string{"file.yaml "}},
				{name: "file spaces", line: `nativeapp login --file "two w`, want: []string{"two words.yaml"}},
				{name: "after dash files", line: "nativeapp login profile -- file.t", want: []string{"file.txt "}},
				{name: "nospace", line: "nativeapp login hold", want: []string{"holding"}},
				{name: "comma keeps cursor", line: "nativeapp login --items=a,b,", want: []string{`a\,b\,c`}},
				{name: "middle cursor", line: "nativeapp login --format=j", tail: "unk --text ignored", want: []string{"json "}},
				{name: "empty assignment then space", line: "nativeapp login --text= pro", want: []string{"profile "}},
				{name: "unknown suppresses files", line: "nativeapp login --unknown="},
			} {
				t.Run(tc.name, func(t *testing.T) {
					words := tc.words
					if words == nil {
						words = strings.Split(tc.line+tc.tail, " ")
					}
					driver := `source "$1"
COMP_LINE=$2
COMP_POINT=$3
shift 3
COMP_WORDS=("$@")
COMP_CWORD=$((${#COMP_WORDS[@]}-1))
__start_nativeapp
if ((${#COMPREPLY[@]})); then printf '%s\n' "${COMPREPLY[@]}"; fi`
					args := append(
						[]string{
							"--noprofile",
							"--norc",
							"-c",
							driver,
							"bash",
							script,
							tc.line + tc.tail,
							strconv.Itoa(len(tc.line)),
						},
						words...)
					cmd := exec.CommandContext(t.Context(), bash, args...)
					cmd.Dir, cmd.Env = dir, environment
					out, err := cmd.CombinedOutput()
					require.NoError(t, err, "%s", out)
					var got []string
					if len(out) > 0 {
						got = strings.Split(strings.TrimSuffix(string(out), "\n"), "\n")
					}
					require.ElementsMatch(t, tc.want, got)
				})
			}
		})
	}
}

func testNativeFish(t *testing.T, dir string, environment []string, script string) {
	t.Helper()
	for _, tc := range []struct{ line, want, exclude string }{
		{"~/nativeapp login --format j", "json", "file.txt"},
		{`nativeapp login --text '(touch evaluated)' --format j`, "json", "file.txt"},
		{`nativeapp login --text 'two words' --echo-text `, "two words", "file.txt"},
		{`nativeapp login --text '(touch evaluated)' --echo-text `, "(touch evaluated)", "file.txt"},
		{"nativeapp login --dir dir", "directory/", "directory.json"},
		{"nativeapp login --dir=dir", "directory/", "directory.json"},
		{`nativeapp login --dir "dir`, "directory/", "directory.json"},
		{"nativeapp login --dir ~/dir", "directory/", "directory.json"},
		{"nativeapp login --text ", "", "file.txt"},
		{"nativeapp login --", "--format", "--sibling-only"},
		{"nativeapp login --text other --", "--format", "--ancestor-only"},
		{"nativeapp login profile -- file.t", "file.txt", "profile"},
	} {
		t.Run(tc.line, func(t *testing.T) {
			cmd := exec.CommandContext(
				t.Context(),
				nativeShell(t, "fish"),
				"-N",
				"-c",
				`source $argv[1]; complete -C $argv[2]`,
				script,
				tc.line,
			)
			cmd.Dir, cmd.Env = dir, environment
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", out)
			//nolint:gocritic // shell descriptions and absolute paths vary
			require.Contains(t, string(out), tc.want)
			if tc.want == "json" {
				require.Equal(t, "json\n", string(out))
			}
			//nolint:gocritic // check forbidden candidates regardless of descriptions
			require.NotContains(t, string(out), tc.exclude)
			require.NoFileExists(t, filepath.Join(dir, "evaluated"))
		})
	}
}

func testNativeZsh(t *testing.T, dir string, environment []string, script string) {
	t.Helper()
	for _, words := range [][]string{
		{"nativeapp", "login", "--format", "j"},
		{"nativeapp", "login", `--format=x$(touch evaluated)=`},
		{"~/nativeapp", "login", "--format", "j"},
		{"nativeapp", "login", "--text", `"$(touch evaluated)"`, "--format", "j"},
		{"nativeapp", "login", "--text", `"two words"`, "--format", "j"},
	} {
		driver := `compdef() { :; }
source "$1"
shift
words=("$@")
CURRENT=${#words[@]}
_describe() { printf '%s\n' "${completions[@]}"; }
_nativeapp`
		args := append([]string{"-f", "-c", driver, "zsh", script}, words...)
		cmd := exec.CommandContext(t.Context(), nativeShell(t, "zsh"), args...)
		cmd.Dir, cmd.Env = dir, environment
		out, err := cmd.CombinedOutput()
		require.NoError(t, err, "%s", out)
		require.Equal(t, "json\nyaml\n", string(out))
		require.NoFileExists(t, filepath.Join(dir, "evaluated"))
	}
}
