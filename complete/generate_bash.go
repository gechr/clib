package complete

import (
	"fmt"
	"strings"
)

// GenerateBash generates a bash shell completion script.
func GenerateBash(g *Generator) (string, error) {
	if err := ValidateGenerator(g); err != nil {
		return "", err
	}

	var sb strings.Builder

	command := g.AppName
	funcName := "_" + strings.ReplaceAll(command, "-", "_")
	cmdName := bashCmdNameFromApp(command)
	rootSpecs := SortVisibleSpecs(g.Specs)
	inheritedSpecs := persistentSpecs(g.Specs)
	fwdFn := bashForwardHelperName(g)
	dynFn := bashDynamicHelperName(g)

	needsPrev := bashNeedsPrev(g.Specs, g.Subs)

	localVars := "local i cur opts cmd"
	if needsPrev {
		localVars = "local i cur prev opts cmd"
	}
	fmt.Fprintf(&sb, `# %s bash completion
%s() {
    %s
    COMPREPLY=()
    if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
        cur="$2"
    else
        cur="${COMP_WORDS[COMP_CWORD]}"
    fi
`, command, funcName, localVars)
	if needsPrev {
		fmt.Fprint(&sb, "    prev=\"$3\"\n")
	}
	if fwdFn != "" {
		// Collect forwardable context flags once; dynamic value completions
		// pass them to the handler via "${__fwd[@]}".
		fmt.Fprintf(&sb, "    local -a __fwd=()\n    %s\n", fwdFn)
	}
	fmt.Fprintf(&sb, `    cmd=""
    opts=""

    for i in "${COMP_WORDS[@]:0:COMP_CWORD}"; do
        case "${cmd},${i}" in
            ",$1")
                cmd="%s"
                ;;
`, cmdName)

	if len(g.Subs) > 0 {
		bashWriteSubcmdTransitions(&sb, g.Subs, cmdName)
	}

	fmt.Fprint(&sb, `            *)
                ;;
        esac
    done

    case "${cmd}" in
`)

	bashWriteCmdCase(
		g,
		&sb,
		cmdName,
		rootSpecs,
		g.Subs,
		false,
		g.DynamicArgs,
		g.HasMaxPositionalArgs,
		g.MaxPositionalArgs,
		1,
	)
	if len(g.Subs) > 0 {
		bashWriteSubcmdCases(
			g,
			&sb,
			g.Subs,
			cmdName,
			inheritedSpecs,
			2, //nolint:mnd // depth 2 = first subcommand level
		)
	}

	fmt.Fprint(&sb, "    esac\n}\n")

	if dynFn != "" {
		bashWriteDynamicHelper(&sb, dynFn)
	}

	if fwdFn != "" {
		bashWriteForwardedFlagsHelper(&sb, fwdFn, allForwardableSpecs(g))
	}

	fmt.Fprintf(&sb, `
if [[ "${BASH_VERSINFO[0]}" -eq 4 && "${BASH_VERSINFO[1]}" -ge 4 || "${BASH_VERSINFO[0]}" -gt 4 ]]; then
    complete -F %s -o nosort -o bashdefault -o default %s
else
    complete -F %s -o bashdefault -o default %s
fi
`, funcName, command, funcName, command)

	return sb.String(), nil
}

// bashForwardHelperName returns the name of the shared forwarded-flags helper
// for g, or "" when no flag-value dynamic completion would consume it.
func bashForwardHelperName(g *Generator) string {
	if !forwardingActive(g) {
		return ""
	}
	return "_" + strings.ReplaceAll(g.AppName, "-", "_") + "_forwarded_flags"
}

// bashDynamicHelperName returns the helper used to append newline-delimited
// dynamic completion values, or "" when no generated call needs it.
func bashDynamicHelperName(g *Generator) string {
	if !bashNeedsDynamicHelper(g) {
		return ""
	}
	return "_" + strings.ReplaceAll(g.AppName, "-", "_") + "_dynamic_values"
}

func bashNeedsDynamicHelper(g *Generator) bool {
	if len(g.DynamicArgs) > 0 {
		return true
	}

	var walk func([]Spec, []SubSpec) bool
	walk = func(specs []Spec, subs []SubSpec) bool {
		for _, spec := range specs {
			if spec.Dynamic != "" && !spec.CommaList {
				return true
			}
		}
		for _, sub := range subs {
			if len(sub.DynamicArgs) > 0 || walk(sub.Specs, sub.Subs) {
				return true
			}
		}
		return false
	}
	return walk(g.Specs, g.Subs)
}

// bashDynamicCmd returns the command and arguments that produce dynamic flag
// completions, forwarding collected context flags when forwarding is active.
func bashDynamicCmd(g *Generator, predictor string) string {
	if bashForwardHelperName(g) != "" {
		return fmt.Sprintf(
			`%s --%s=%s -- "${__fwd[@]}"`,
			g.AppName,
			FlagComplete,
			predictor,
		)
	}
	return fmt.Sprintf("%s --%s=%s", g.AppName, FlagComplete, predictor)
}

// bashDynamicArgsCmd builds the handler command for a positional
// dynamic completion slot. first==true means no preceding positionals; when
// forwarding is active, collected context flags precede the real positionals.
func bashDynamicArgsCmd(g *Generator, da string, first bool) string {
	forward := bashForwardHelperName(g) != ""
	switch {
	case forward && first:
		return fmt.Sprintf(
			`%s --%s=%s -- "${__fwd[@]}"`,
			g.AppName,
			FlagComplete,
			da,
		)
	case forward:
		return fmt.Sprintf(
			`%s --%s=%s -- "${__fwd[@]}" "${__dyn_pos[@]}"`,
			g.AppName,
			FlagComplete,
			da,
		)
	case first:
		return fmt.Sprintf("%s --%s=%s", g.AppName, FlagComplete, da)
	default:
		return fmt.Sprintf(
			`%s --%s=%s -- "${__dyn_pos[@]}"`,
			g.AppName,
			FlagComplete,
			da,
		)
	}
}

// bashWriteDynamicHelper emits a bash 3-compatible helper that reads dynamic
// completion output as a newline-delimited list. The local IFS protects
// multi-word values when bash splits compgen's command-substitution output.
func bashWriteDynamicHelper(sb *strings.Builder, helperName string) {
	fmt.Fprintf(sb, "\n%s() {\n", helperName)
	fmt.Fprint(sb, "    local cur=\"$1\"\n")
	fmt.Fprint(sb, "    shift\n")
	fmt.Fprint(sb, "    local line\n")
	fmt.Fprint(sb, "    local -a vals=()\n")
	fmt.Fprint(sb, "    while IFS= read -r line || [[ -n \"${line}\" ]]; do\n")
	fmt.Fprint(sb, "        vals+=(\"${line}\")\n")
	fmt.Fprint(sb, "    done < <(\"$@\" 2>/dev/null)\n")
	fmt.Fprint(sb, "    local IFS=$'\\n'\n")
	fmt.Fprint(sb, "    COMPREPLY+=($(compgen -W \"${vals[*]}\" -- \"${cur}\"))\n")
	fmt.Fprint(sb, "}\n")
}

// bashWriteForwardedFlagsHelper emits a function that scans the command line and
// populates __fwd with each forwardable context flag as --name=value.
func bashWriteForwardedFlagsHelper(sb *strings.Builder, helperName string, fwd []forwardSpec) {
	fmt.Fprintf(sb, "\n%s() {\n", helperName)
	fmt.Fprint(sb, "    __fwd=()\n")
	fmt.Fprint(sb, "    local j __skip_next=0 __fwd_name=\"\"\n")
	fmt.Fprint(sb, "    for ((j=1; j<COMP_CWORD; j++)); do\n")
	fmt.Fprint(sb, `        if [[ "${__skip_next}" -eq 1 ]]; then
            if [[ -n "${__fwd_name}" ]]; then
                __fwd+=("--${__fwd_name}=${COMP_WORDS[j]}")
                __fwd_name=""
            fi
            __skip_next=0
            continue
        fi
        if [[ "${COMP_WORDS[j]}" == "--" ]]; then
            break
        fi
        case "${COMP_WORDS[j]}" in
`)
	for _, f := range fwd {
		var patterns []string
		if f.LongFlag != "" {
			patterns = append(patterns, "--"+f.LongFlag)
		}
		if f.ShortFlag != "" {
			patterns = append(patterns, "-"+f.ShortFlag)
		}
		fmt.Fprintf(
			sb,
			"            %s)\n                __skip_next=1\n                __fwd_name=%q\n                ;;\n",
			strings.Join(patterns, "|"),
			f.LongFlag,
		)
		if f.ShortFlag != "" && f.LongFlag != "" {
			fmt.Fprintf(
				sb,
				"            --%s=*)\n                __fwd+=(\"${COMP_WORDS[j]}\")\n                ;;\n",
				f.LongFlag,
			)
			fmt.Fprintf(
				sb,
				"            -%s=*)\n                __fwd+=(\"--%s=${COMP_WORDS[j]#-%s=}\")\n                ;;\n",
				f.ShortFlag,
				f.LongFlag,
				f.ShortFlag,
			)
		} else if f.LongFlag != "" {
			fmt.Fprintf(
				sb,
				"            --%s=*)\n                __fwd+=(\"${COMP_WORDS[j]}\")\n                ;;\n",
				f.LongFlag,
			)
		}
	}
	fmt.Fprint(sb, "            *)\n                ;;\n        esac\n    done\n}\n")
}

func bashNeedsPrev(specs []Spec, subs []SubSpec) bool {
	for _, spec := range specs {
		if spec.HasArg {
			return true
		}
	}
	for _, sub := range subs {
		if bashNeedsPrev(sub.Specs, sub.Subs) {
			return true
		}
	}
	return false
}

func bashCmdNameFromApp(name string) string {
	return strings.ReplaceAll(name, "-", "__")
}

func bashWriteSubcmdTransitions(sb *strings.Builder, subs []SubSpec, parentCmd string) {
	for _, sub := range SortSubSpecs(subs) {
		childCmd := parentCmd + "__" + bashCmdNameFromApp(sub.Name)

		patterns := []string{fmt.Sprintf("%s,%s", parentCmd, sub.Name)}
		for _, alias := range sub.Aliases {
			patterns = append(patterns, fmt.Sprintf("%s,%s", parentCmd, alias))
		}

		fmt.Fprintf(sb, "            %s)\n                cmd=\"%s\"\n                ;;\n",
			strings.Join(patterns, "|"), childCmd)

		if len(sub.Subs) > 0 {
			bashWriteSubcmdTransitions(sb, sub.Subs, childCmd)
		}
	}
}

func bashOptsString(specs []Spec, subs []SubSpec) string {
	var parts []string
	for _, spec := range specs {
		if spec.LongFlag != "" {
			parts = append(parts, "--"+spec.LongFlag)
		}
		if spec.ShortFlag != "" {
			parts = append(parts, "-"+spec.ShortFlag)
		}
	}
	for _, sub := range SortSubSpecs(subs) {
		parts = append(parts, sub.Name)
	}
	return strings.Join(parts, " ")
}

func bashWriteCmdCase(
	g *Generator,
	sb *strings.Builder,
	cmdName string,
	specs []Spec,
	subs []SubSpec,
	pathArgs bool,
	dynamicArgs []string,
	hasMaxPositionalArgs bool,
	maxPositionalArgs int,
	depth int,
) {
	opts := bashOptsString(specs, subs)

	if len(dynamicArgs) > 0 {
		fmt.Fprintf(sb, `        %s)
            opts="%s"
            if [[ ${cur} == -* ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
`, cmdName, opts)
	} else {
		fmt.Fprintf(sb, `        %s)
            opts="%s"
            if [[ ${cur} == -* || ${COMP_CWORD} -eq %d ]]; then
                COMPREPLY=($(compgen -W "${opts}" -- "${cur}"))
                return 0
            fi
`, cmdName, opts, depth)
	}

	var hasArgSpecs []Spec
	for _, spec := range specs {
		if spec.HasArg {
			hasArgSpecs = append(hasArgSpecs, spec)
		}
	}

	if len(hasArgSpecs) > 0 {
		fmt.Fprint(sb, "            case \"${prev}\" in\n")
		for _, spec := range hasArgSpecs {
			bashWritePrevCase(g, sb, spec)
		}
		fmt.Fprint(sb, `                *)
                    COMPREPLY=()
                    ;;
            esac
`)
	}

	switch {
	case pathArgs:
		if hasMaxPositionalArgs {
			bashWriteDynamicArgsParser(sb, specs, depth)
			fmt.Fprintf(
				sb,
				"            if [[ ${#__dyn_pos[@]} -ge %d ]]; then\n",
				maxPositionalArgs,
			)
			fmt.Fprint(sb, "                COMPREPLY=($(compgen -W \"${opts}\" -- \"${cur}\"))\n")
			fmt.Fprint(sb, "                return 0\n")
			fmt.Fprint(sb, "            fi\n")
		}
		WriteIndented(sb, "            ", bashFileCompletionBlock)
	case len(dynamicArgs) > 0:
		bashWriteDynamicArgsParser(sb, specs, depth)
		if hasMaxPositionalArgs {
			fmt.Fprintf(
				sb,
				"            if [[ ${#__dyn_pos[@]} -ge %d ]]; then\n",
				maxPositionalArgs,
			)
			fmt.Fprint(sb, "                COMPREPLY=($(compgen -W \"${opts}\" -- \"${cur}\"))\n")
			fmt.Fprint(sb, "                return 0\n")
			fmt.Fprint(sb, "            fi\n")
		}
		fmt.Fprint(sb, "            case ${#__dyn_pos[@]} in\n")
		for i, da := range dynamicArgs {
			if i == 0 {
				fmt.Fprintf(
					sb,
					"                %d)\n                    COMPREPLY=($(compgen -W \"${opts}\" -- \"${cur}\"))\n                    %s \"${cur}\" %s\n                    ;;\n",
					i,
					bashDynamicHelperName(g),
					bashDynamicArgsCmd(g, da, true),
				)
				continue
			}

			fmt.Fprintf(
				sb,
				"                %d)\n                    COMPREPLY=()\n                    %s \"${cur}\" %s\n                    ;;\n",
				i,
				bashDynamicHelperName(g),
				bashDynamicArgsCmd(g, da, false),
			)
		}
		fmt.Fprintf(
			sb,
			"                *)\n                    COMPREPLY=()\n                    %s \"${cur}\" %s\n                    ;;\n",
			bashDynamicHelperName(g),
			bashDynamicArgsCmd(g, dynamicArgs[len(dynamicArgs)-1], false),
		)
		fmt.Fprint(sb, "            esac\n")
	default:
		fmt.Fprint(sb, "            COMPREPLY=($(compgen -W \"${opts}\" -- \"${cur}\"))\n")
	}
	fmt.Fprint(sb, "            return 0\n            ;;\n")
}

func bashWriteDynamicArgsParser(sb *strings.Builder, specs []Spec, depth int) {
	cmdSkip := depth - 1
	// Forwardable flags are classified as ordinary value flags so their values
	// are skipped rather than counted as positionals; forwarded context is
	// collected separately in __fwd.
	exact, equals := argValuePatterns(specs)
	fmt.Fprint(sb, "            local -a __dyn_pos=()\n")
	fmt.Fprint(sb, "            local __skip_next=0\n")
	fmt.Fprint(sb, "            local __after_dd=0\n")
	if cmdSkip > 0 {
		fmt.Fprintf(sb, "            local __cmd_skip=%d\n", cmdSkip)
	}
	fmt.Fprint(sb, "            for ((j=1; j<COMP_CWORD; j++)); do\n")
	fmt.Fprint(sb, `                if [[ "${__after_dd}" -eq 1 ]]; then
                    __dyn_pos+=("${COMP_WORDS[j]}")
                    continue
                fi
                if [[ "${__skip_next}" -eq 1 ]]; then
                    __skip_next=0
                    continue
                fi
                if [[ "${COMP_WORDS[j]}" == "--" ]]; then
                    __after_dd=1
                    continue
                fi
                case "${COMP_WORDS[j]}" in
`)
	if len(exact) > 0 {
		fmt.Fprintf(
			sb,
			"                    %s)\n                        __skip_next=1\n                        ;;\n",
			strings.Join(exact, "|"),
		)
	}
	if len(equals) > 0 {
		fmt.Fprintf(
			sb,
			"                    %s)\n                        ;;\n",
			strings.Join(equals, "|"),
		)
	}
	fmt.Fprint(sb, "                    -*)\n                        ;;\n")
	if cmdSkip > 0 {
		fmt.Fprint(sb, `                    *)
                        if [[ $__cmd_skip -gt 0 ]]; then
                            ((__cmd_skip--))
                        else
                            __dyn_pos+=("${COMP_WORDS[j]}")
                        fi
                        ;;
`)
	} else {
		fmt.Fprint(
			sb,
			"                    *)\n                        __dyn_pos+=(\"${COMP_WORDS[j]}\")\n                        ;;\n",
		)
	}
	fmt.Fprint(sb, "                esac\n            done\n")
}

func bashWritePrevCase(g *Generator, sb *strings.Builder, spec Spec) {
	var patterns []string
	if spec.LongFlag != "" {
		patterns = append(patterns, "--"+spec.LongFlag)
	}
	if spec.ShortFlag != "" {
		patterns = append(patterns, "-"+spec.ShortFlag)
	}
	if len(patterns) == 0 {
		return
	}

	fmt.Fprintf(sb, "                %s)\n", strings.Join(patterns, "|"))

	switch {
	case spec.CommaList && spec.Dynamic != "":
		bashWriteCommaCompletion(sb, bashDynamicCmd(g, spec.Dynamic), true)
	case spec.CommaList && len(spec.Values) > 0:
		bashWriteCommaCompletion(sb, strings.Join(spec.Values, " "), false)
	case spec.CommaList && len(spec.ValueDescs) > 0:
		vals := make([]string, len(spec.ValueDescs))
		for i, vd := range spec.ValueDescs {
			vals[i] = vd.Value
		}
		bashWriteCommaCompletion(sb, strings.Join(vals, " "), false)
	case spec.Dynamic != "":
		fmt.Fprintf(
			sb,
			"                    COMPREPLY=()\n                    %s \"${cur}\" %s\n",
			bashDynamicHelperName(g),
			bashDynamicCmd(g, spec.Dynamic),
		)
	case len(spec.Values) > 0:
		quoted := make([]string, len(spec.Values))
		for i, v := range spec.Values {
			quoted[i] = strings.ReplaceAll(v, "'", "'\\''")
		}
		fmt.Fprintf(sb,
			"                    COMPREPLY=($(compgen -W '%s' -- \"${cur}\"))\n",
			strings.Join(quoted, " "))
	case len(spec.ValueDescs) > 0:
		vals := make([]string, len(spec.ValueDescs))
		for i, vd := range spec.ValueDescs {
			vals[i] = strings.ReplaceAll(vd.Value, "'", "'\\''")
		}
		fmt.Fprintf(sb,
			"                    COMPREPLY=($(compgen -W '%s' -- \"${cur}\"))\n",
			strings.Join(vals, " "))
	case spec.Extension != "":
		WriteIndented(sb, "                    ", bashExtCompletionBlock(spec.Extension))
	case spec.ValueHint == HintFile:
		WriteIndented(sb, "                    ", bashFileCompletionBlock)
	case spec.ValueHint == HintDir:
		WriteIndented(sb, "                    ", bashDirCompletionBlock)
	case spec.ValueHint == HintCommand:
		fmt.Fprint(sb, "                    COMPREPLY=($(compgen -c -- \"${cur}\"))\n")
	case spec.ValueHint == HintUser:
		fmt.Fprint(sb, "                    COMPREPLY=($(compgen -u -- \"${cur}\"))\n")
	case spec.ValueHint == HintHost:
		fmt.Fprint(sb, "                    COMPREPLY=($(compgen -A hostname -- \"${cur}\"))\n")
	default:
		fmt.Fprint(sb, "                    COMPREPLY=()\n")
	}

	fmt.Fprint(sb, "                    return 0\n                    ;;\n")
}

func bashWriteCommaCompletion(sb *strings.Builder, valuesExpr string, dynamic bool) {
	fmt.Fprint(sb, `                    local prefix=""
                    local cur_val="${cur}"
`)
	if dynamic {
		fmt.Fprintf(sb, `                    local line
                    local -a all_vals=()
                    while IFS= read -r line || [[ -n "${line}" ]]; do
                        all_vals+=("${line}")
                    done < <(%s 2>/dev/null)
`, valuesExpr)
	} else {
		fmt.Fprintf(sb, "                    local all_vals=(%s)\n", valuesExpr)
	}
	fmt.Fprint(sb, `                    local IFS=$'\n'
                    local -a avail=()
                    if [[ "${cur}" == *,* ]]; then
                        prefix="${cur%,*},"
                        cur_val="${cur##*,}"
                        IFS=',' read -ra selected <<< "${prefix}"
                        for val in "${all_vals[@]}"; do
                            local found=0
                            for sel in "${selected[@]}"; do
                                if [[ "${val}" == "${sel}" ]]; then
                                    found=1
                                    break
                                fi
                            done
                            if [[ "${found}" -eq 0 ]]; then
                                avail+=("${val}")
                            fi
                        done
                    else
                        avail=("${all_vals[@]}")
                    fi
                    COMPREPLY=($(compgen -W "${avail[*]}" -- "${cur_val}"))
                    if [[ -n "${prefix}" ]]; then
                        COMPREPLY=("${COMPREPLY[@]/#/${prefix}}")
                    fi
                    compopt -o nospace
`)
}

const bashFileCompletionBlock = `local oldifs
if [ -n "${IFS+x}" ]; then
    oldifs="$IFS"
fi
IFS=$'\n'
COMPREPLY=($(compgen -f -- "${cur}"))
if [ -n "${oldifs+x}" ]; then
    IFS="$oldifs"
fi
if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
    compopt -o filenames
fi
`

func bashExtCompletionBlock(ext string) string {
	parts := strings.Split(ext, ",")
	for i, p := range parts {
		parts[i] = "*." + strings.TrimSpace(p)
	}
	filter := parts[0]
	if len(parts) > 1 {
		filter = "@(" + strings.Join(parts, "|") + ")"
	}
	return fmt.Sprintf(`local oldifs
if [ -n "${IFS+x}" ]; then
    oldifs="$IFS"
fi
IFS=$'\n'
COMPREPLY=($(compgen -d -- "${cur}") $(compgen -f -X '!%s' -- "${cur}"))
if [ -n "${oldifs+x}" ]; then
    IFS="$oldifs"
fi
if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
    compopt -o filenames
fi
`, filter)
}

const bashDirCompletionBlock = `COMPREPLY=()
if [[ "${BASH_VERSINFO[0]}" -ge 4 ]]; then
    compopt -o plusdirs
fi
`

func bashWriteSubcmdCases(
	g *Generator,
	sb *strings.Builder,
	subs []SubSpec,
	parentCmd string,
	inheritedSpecs []Spec,
	depth int,
) {
	for _, sub := range SortSubSpecs(subs) {
		childCmd := parentCmd + "__" + bashCmdNameFromApp(sub.Name)
		visibleSpecs := combineVisibleSpecs(inheritedSpecs, sub.Specs)

		bashWriteCmdCase(
			g,
			sb,
			childCmd,
			visibleSpecs,
			sub.Subs,
			sub.PathArgs,
			sub.DynamicArgs,
			sub.HasMaxPositionalArgs,
			sub.MaxPositionalArgs,
			depth,
		)

		if len(sub.Subs) == 0 {
			continue
		}

		nextInherited := appendSpecs(inheritedSpecs, persistentSpecs(sub.Specs))
		bashWriteSubcmdCases(g, sb, sub.Subs, childCmd, nextInherited, depth+1)
	}
}
