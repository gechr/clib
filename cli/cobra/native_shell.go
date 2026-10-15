package cobra

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
)

//go:embed native_bash.sh
var nativeBashScript string

func nativeBash(g *complete.Generator) (string, error) {
	if err := complete.ValidateShellSafe(g.AppName, "AppName"); err != nil {
		return "", err
	}
	return strings.ReplaceAll(nativeBashScript, "__CLIB_APP__", g.AppName), nil
}

func nativeFish(g *complete.Generator) (string, error) {
	if err := complete.ValidateShellSafe(g.AppName, "AppName"); err != nil {
		return "", err
	}
	var script strings.Builder
	cmd := &cobralib.Command{Use: g.AppName}
	if err := cmd.GenFishCompletion(&script, true); err != nil {
		return "", err
	}
	const directoryFilter = `if test $dirfilter -eq 1
        set -l flagPrefix (string match -r -- '^-[^=]*=' (commandline -ct))
        set --global __CLIB_APP___comp_results
        for dir in (__fish_complete_directories)
            set --append --global __CLIB_APP___comp_results "$flagPrefix$dir"
        end
        return 0
    end
    if test $filefilter -eq 1`
	return patchCompletionScript(script.String(),
		"set -l results (eval $requestComp 2> /dev/null)",
		`set -l current (commandline -ct)
    set -l program $args[1]
    if string match -q -- '~/*' "$program"
        set program "$HOME/"(string sub -s 3 -- "$program")
    end
    set -l results (env COBRA_ACTIVE_HELP=0 "$program" __complete $args[2..-1] "$current" 2> /dev/null)`,
		"if test $filefilter -eq 1; or test $dirfilter -eq 1",
		strings.ReplaceAll(directoryFilter, "__CLIB_APP__", "__"+g.AppName),
		"# For Fish, when completing a flag with an = (e.g."+", <program> -n=<TAB>)",
		"# Prefix inline flag assignments with the flag name.",
	)
}

func nativeZsh(g *complete.Generator) (string, error) {
	if err := complete.ValidateShellSafe(g.AppName, "AppName"); err != nil {
		return "", err
	}
	var script strings.Builder
	cmd := &cobralib.Command{Use: g.AppName}
	if err := cmd.GenZshCompletion(&script); err != nil {
		return "", err
	}
	return patchCompletionScript(
		script.String(),
		`words=("${=words[1,CURRENT]}")`,
		`words=("${(@)words[1,CURRENT]}")`,
		fmt.Sprintf(
			"# For zsh, when completing a flag with an = (e.g."+", %s -n=<TAB>)",
			g.AppName,
		),
		"# Prefix inline flag assignments with the flag name.",
		// _describe and _arguments evaluate their option strings. Quote the
		// prefix once so literal shell syntax in the current word stays data.
		`flagPrefix="-P ${BASH_REMATCH}"`,
		`flagPrefix="-P ${(q)BASH_REMATCH}"`,
		`out=$(eval ${requestComp} 2>/dev/null)`,
		`local program="${(Q)words[1]}"
    if [[ "$program" == '~/'* ]]; then
        program="$HOME/${program#\~/}"
    fi
    out=$("$program" __complete "${(@Q)words[2,-1]}" 2>/dev/null)`,
	)
}

// Require exactly one match per patch: upstream changes must fail generation,
// rather than silently emitting an unsafe or unsupported client.
func patchCompletionScript(script string, oldNew ...string) (string, error) {
	if len(oldNew)%2 != 0 {
		return "", fmt.Errorf("completion script patch requires old/new pairs")
	}
	for i := 0; i+1 < len(oldNew); i += 2 {
		if count := strings.Count(script, oldNew[i]); count != 1 {
			return "", fmt.Errorf(
				"cobra completion template changed: expected one patch match, got %d",
				count,
			)
		}
		script = strings.Replace(script, oldNew[i], oldNew[i+1], 1)
	}
	return script, nil
}
