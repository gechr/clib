package complete_test

import (
	"os/exec"
	"strconv"
	"strings"
	"testing"

	"github.com/gechr/clib/complete"
	"github.com/stretchr/testify/require"
)

func TestShellExec_BashWordBreaks(t *testing.T) {
	gen := &complete.Generator{
		AppName:     "myapp",
		DynamicArgs: []string{"item", "next"},
		Specs: []complete.Spec{
			{LongFlag: "target", ShortFlag: "t", HasArg: true, Dynamic: "item"},
			{LongFlag: "profile", HasArg: true, Forward: true},
			{LongFlag: "format", HasArg: true, Values: []string{"text:plain", "text:html"}},
		},
	}
	tests := []struct {
		name    string
		line    string
		words   []string
		cur     string
		tail    string
		want    []string
		handler string
	}{
		{
			name:    "at",
			line:    "myapp branch@",
			words:   []string{"myapp", "branch", "@"},
			cur:     "@",
			want:    []string{"@remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "at prefix",
			line:    "myapp branch@r",
			words:   []string{"myapp", "branch", "@", "r"},
			cur:     "@r",
			want:    []string{"@remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "at suffix",
			line:    "myapp branch@r",
			words:   []string{"myapp", "branch", "@", "r"},
			cur:     "r",
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "empty suffix",
			line:    "myapp branch@",
			words:   []string{"myapp", "branch", "@", ""},
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "colon",
			line:    "myapp group::",
			words:   []string{"myapp", "group", "::"},
			want:    []string{"item"},
			handler: "--@complete=item --",
		},
		{
			name:    "assignment",
			line:    "myapp key=branch@r",
			words:   []string{"myapp", "key", "=", "branch", "@", "r"},
			cur:     "r",
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "flag value",
			line:    "myapp --target branch@r",
			words:   []string{"myapp", "--target", "branch", "@", "r"},
			cur:     "r",
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "inline flag",
			line:    "myapp --target=branch@r",
			words:   []string{"myapp", "--target", "=", "branch", "@", "r"},
			cur:     "r",
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "short inline flag",
			line:    "myapp -t=branch@r",
			words:   []string{"myapp", "-t", "=", "branch", "@", "r"},
			cur:     "r",
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:  "inline enum",
			line:  "myapp --format=text:",
			words: []string{"myapp", "--format", "=", "text", ":"},
			want:  []string{"plain", "html"},
		},
		{
			name:  "enum",
			line:  "myapp --format text:p",
			words: []string{"myapp", "--format", "text", ":", "p"},
			cur:   "p",
			want:  []string{"plain"},
		},
		{
			name:    "preceding positional",
			line:    "myapp branch@remote n",
			words:   []string{"myapp", "branch", "@", "remote", "n"},
			cur:     "n",
			want:    []string{"next"},
			handler: "--@complete=next -- branch@remote",
		},
		{
			name:    "whitespace",
			line:    "myapp branch@ n",
			words:   []string{"myapp", "branch", "@", "n"},
			cur:     "n",
			want:    []string{"next"},
			handler: "--@complete=next -- branch@",
		},
		{
			name:    "tab",
			line:    "myapp branch@\tn",
			words:   []string{"myapp", "branch", "@", "n"},
			cur:     "n",
			want:    []string{"next"},
			handler: "--@complete=next -- branch@",
		},
		{
			name:    "forwarded",
			line:    "myapp --profile=team@host branch@r",
			words:   []string{"myapp", "--profile", "=", "team", "@", "host", "branch", "@", "r"},
			cur:     "r",
			want:    []string{"remote"},
			handler: "--@complete=item -- --profile=team@host",
		},
		{
			name:    "quoted preceding",
			line:    "myapp 'two words' n",
			words:   []string{"myapp", "'two words'", "n"},
			cur:     "n",
			want:    []string{"next"},
			handler: "--@complete=next -- 'two words'",
		},
		{
			name:    "stripped quotes",
			line:    "myapp 'two words' n",
			words:   []string{"myapp", "two words", "n"},
			cur:     "n",
			want:    []string{"next"},
			handler: "--@complete=next -- two words",
		},
		{
			name:    "middle of word",
			line:    "myapp branch@r",
			tail:    "ubbish",
			words:   []string{"myapp", "branch", "@", "rubbish"},
			cur:     "r",
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "complete suffix",
			line:    "myapp branch@remote",
			words:   []string{"myapp", "branch", "@", "remote"},
			cur:     "remote",
			want:    []string{"remote"},
			handler: "--@complete=item --",
		},
		{
			name:    "fully typed insertion",
			line:    "myapp branch@remote",
			words:   []string{"myapp", "branch", "@", "remote", ""},
			handler: "--@complete=item --",
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			stub := "#!/usr/bin/env bash\nprintf '%s\\n' \"$*\" >> \"$CLIB_COMPLETE_LOG\"\nprintf '%s\\n' branch@remote group::item key=branch@remote next\n"
			dir, scriptPath, logPath := completionEnvWithStub(t, gen, "bash", stub)
			quoted := make([]string, len(tc.words))
			for i, word := range tc.words {
				quoted[i] = shellQuote(word)
			}
			driver := strings.Join([]string{
				"source " + shellQuote(scriptPath),
				"COMP_LINE=" + shellQuote(tc.line+tc.tail),
				"COMP_POINT=" + strconv.Itoa(len(tc.line)),
				"COMP_WORDS=(" + strings.Join(quoted, " ") + ")",
				"COMP_CWORD=" + strconv.Itoa(len(tc.words)-1),
				`before=$(declare -p COMP_WORDS COMP_CWORD COMP_WORDBREAKS; shopt -p hostcomplete)`,
				"_myapp myapp " + shellQuote(tc.cur) + " ''",
				`after=$(declare -p COMP_WORDS COMP_CWORD COMP_WORDBREAKS; shopt -p hostcomplete)`,
				`[[ $before == "$after" ]] || exit 1`,
				`if ((${#COMPREPLY[@]})); then printf '%s\n' "${COMPREPLY[@]}"; fi`,
			}, "\n")
			cmd := exec.Command(lookShell(t, "bash"), "--norc", "-c", driver)
			cmd.Env = shellEnv(dir, logPath)
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", out)
			require.Equal(t, strings.Join(tc.want, "\n"), strings.TrimSuffix(string(out), "\n"))
			require.Equal(t, tc.handler, readHandlerLog(t, logPath))
		})
	}
}
