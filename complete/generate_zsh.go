package complete

import (
	"fmt"
	"strings"
)

// GenerateZsh generates a zsh shell completion script.
func GenerateZsh(g *Generator) (string, error) {
	if err := ValidateGenerator(g); err != nil {
		return "", err
	}

	var sb strings.Builder
	funcName := zshFuncName(g.AppName)
	rootSpecs := SortVisibleSpecs(g.Specs)
	inheritedSpecs := persistentSpecs(g.Specs)
	sortedSubs := SortSubSpecs(g.Subs)

	fmt.Fprintf(&sb, `#compdef %[1]s

autoload -U is-at-least

_%[2]s() {
    typeset -A opt_args
    typeset -a _arguments_options
    local ret=1

    if is-at-least 5.2; then
        _arguments_options=(-s -S -C)
    else
        _arguments_options=(-s -C)
    fi

    local context curcontext="$curcontext" state line
`, g.AppName, funcName)

	if len(sortedSubs) > 0 {
		zshWriteArguments(
			g,
			&sb,
			rootSpecs,
			false,
			g.DynamicArgs,
			g.HasMaxPositionalArgs,
			g.MaxPositionalArgs,
			funcName,
			g.AppName,
			"    ",
		)
		zshWriteSubcommandCase(g, &sb, sortedSubs, inheritedSpecs, funcName, g.AppName, "    ")
	} else {
		zshWriteArguments(
			g,
			&sb,
			rootSpecs,
			false,
			g.DynamicArgs,
			g.HasMaxPositionalArgs,
			g.MaxPositionalArgs,
			"",
			"",
			"    ",
		)
	}

	fmt.Fprint(&sb, "}\n")

	if len(sortedSubs) > 0 {
		zshWriteDescribeTree(&sb, sortedSubs, funcName, g.AppName)
	}

	fmt.Fprintf(&sb, `
if [ "$funcstack[1]" = "_%[1]s" ]; then
    _%[1]s "$@"
else
    compdef _%[1]s %[2]s
fi
`, funcName, g.AppName)

	return sb.String(), nil
}

func zshWriteArguments(
	g *Generator,
	sb *strings.Builder,
	specs []Spec,
	pathArgs bool,
	dynamicArgs []string,
	hasMaxPositionalArgs bool,
	maxPositionalArgs int,
	funcName, stateName, indent string,
) {
	fmt.Fprintf(sb, "%s_arguments \"${_arguments_options[@]}\" : \\\n", indent)

	specIndent := indent + "    "
	for _, spec := range specs {
		zshWriteSpec(g, sb, spec, specIndent)
	}

	switch {
	case funcName != "":
		fmt.Fprintf(
			sb,
			"%[1]s':: :_%[2]s_commands' \\\n%[1]s'*::: :->%[3]s' \\\n",
			specIndent,
			funcName,
			stateName,
		)
	case pathArgs:
		if hasMaxPositionalArgs {
			for i := range maxPositionalArgs {
				fmt.Fprintf(sb, "%s'%d:file:_files' \\\n", specIndent, i+1)
			}
		} else {
			fmt.Fprintf(sb, "%s'*:file:_files' \\\n", specIndent)
		}
	case len(dynamicArgs) > 0:
		limit := len(dynamicArgs)
		if hasMaxPositionalArgs && maxPositionalArgs < limit {
			limit = maxPositionalArgs
		}
		for i := range limit {
			fmt.Fprintf(sb, "%s'%d: :->dyn_%d' \\\n", specIndent, i+1, i+1)
		}
		if !hasMaxPositionalArgs {
			fmt.Fprintf(sb, "%s'*: :->dyn_rest' \\\n", specIndent)
		}
	}

	fmt.Fprintf(sb, "%s&& ret=0\n", indent)

	if funcName == "" && !pathArgs && len(dynamicArgs) > 0 {
		zshWriteDynamicArgsCases(
			g,
			sb,
			specs,
			dynamicArgs,
			hasMaxPositionalArgs,
			maxPositionalArgs,
			indent,
		)
	}
}

func zshWriteDynamicArgsCases(
	g *Generator,
	sb *strings.Builder,
	specs []Spec,
	dynamicArgs []string,
	hasMaxPositionalArgs bool,
	maxPositionalArgs int,
	indent string,
) {
	inner := indent + "    "
	exact, equals := argValuePatterns(specs)
	fmt.Fprintf(sb, "\n%scase $state in\n", indent)
	limit := len(dynamicArgs)
	if hasMaxPositionalArgs && maxPositionalArgs < limit {
		limit = maxPositionalArgs
	}
	for i := range limit {
		da := dynamicArgs[i]
		fmt.Fprintf(sb, "%s(dyn_%d)\n", indent, i+1)
		fmt.Fprintf(sb, "%slocal -a __pos=()\n", inner)
		fmt.Fprintf(sb, "%slocal __skip_next=0\n", inner)
		fmt.Fprintf(sb, "%slocal __after_dd=0\n", inner)
		fmt.Fprintf(sb, "%slocal token\n", inner)
		fmt.Fprintf(sb, "%sfor ((i=2; i<CURRENT; i++)); do\n", inner)
		fmt.Fprintf(sb, "%s    token=${words[i]}\n", inner)
		fmt.Fprintf(sb, "%s    if (( __after_dd )); then\n", inner)
		fmt.Fprintf(sb, "%s        __pos+=(\"$token\")\n", inner)
		fmt.Fprintf(sb, "%s        continue\n", inner)
		fmt.Fprintf(sb, "%s    fi\n", inner)
		fmt.Fprintf(sb, "%s    if (( __skip_next )); then\n", inner)
		fmt.Fprintf(sb, "%s        __skip_next=0\n", inner)
		fmt.Fprintf(sb, "%s        continue\n", inner)
		fmt.Fprintf(sb, "%s    fi\n", inner)
		fmt.Fprintf(sb, "%s    if [[ $token == -- ]]; then\n", inner)
		fmt.Fprintf(sb, "%s        __after_dd=1\n", inner)
		fmt.Fprintf(sb, "%s        continue\n", inner)
		fmt.Fprintf(sb, "%s    fi\n", inner)
		fmt.Fprintf(sb, "%s    case $token in\n", inner)
		if len(exact) > 0 {
			fmt.Fprintf(
				sb,
				"%s        (%s)\n%s            __skip_next=1\n%s            ;;\n",
				inner,
				strings.Join(exact, "|"),
				inner,
				inner,
			)
		}
		if len(equals) > 0 {
			fmt.Fprintf(
				sb,
				"%s        (%s)\n%s            ;;\n",
				inner,
				strings.Join(equals, "|"),
				inner,
			)
		}
		fmt.Fprintf(sb, "%s        (-*)\n%s            ;;\n", inner, inner)
		fmt.Fprintf(
			sb,
			"%s        (*)\n%s            __pos+=(\"$token\")\n%s            ;;\n",
			inner,
			inner,
			inner,
		)
		fmt.Fprintf(sb, "%s    esac\n", inner)
		fmt.Fprintf(sb, "%sdone\n", inner)
		fmt.Fprintf(sb, "%slocal -a items\n", inner)
		if i == 0 {
			fmt.Fprintf(
				sb,
				"%sitems=(${(f)\"$(%s --%s=%s 2>/dev/null)\"})\n",
				inner,
				g.AppName,
				FlagComplete,
				da,
			)
		} else {
			fmt.Fprintf(
				sb,
				"%sitems=(${(f)\"$(%s --%s=%s -- \"${__pos[@]}\" 2>/dev/null)\"})\n",
				inner,
				g.AppName,
				FlagComplete,
				da,
			)
		}
		fmt.Fprintf(sb, "%scompadd -a items\n", inner)
		fmt.Fprintf(sb, "%s;;\n", indent)
	}
	if !hasMaxPositionalArgs {
		fmt.Fprintf(sb, "%s(dyn_rest)\n", indent)
		fmt.Fprintf(sb, "%slocal -a __pos=()\n", inner)
		fmt.Fprintf(sb, "%slocal __skip_next=0\n", inner)
		fmt.Fprintf(sb, "%slocal __after_dd=0\n", inner)
		fmt.Fprintf(sb, "%slocal token\n", inner)
		fmt.Fprintf(sb, "%sfor ((i=2; i<CURRENT; i++)); do\n", inner)
		fmt.Fprintf(sb, "%s    token=${words[i]}\n", inner)
		fmt.Fprintf(sb, "%s    if (( __after_dd )); then\n", inner)
		fmt.Fprintf(sb, "%s        __pos+=(\"$token\")\n", inner)
		fmt.Fprintf(sb, "%s        continue\n", inner)
		fmt.Fprintf(sb, "%s    fi\n", inner)
		fmt.Fprintf(sb, "%s    if (( __skip_next )); then\n", inner)
		fmt.Fprintf(sb, "%s        __skip_next=0\n", inner)
		fmt.Fprintf(sb, "%s        continue\n", inner)
		fmt.Fprintf(sb, "%s    fi\n", inner)
		fmt.Fprintf(sb, "%s    if [[ $token == -- ]]; then\n", inner)
		fmt.Fprintf(sb, "%s        __after_dd=1\n", inner)
		fmt.Fprintf(sb, "%s        continue\n", inner)
		fmt.Fprintf(sb, "%s    fi\n", inner)
		fmt.Fprintf(sb, "%s    case $token in\n", inner)
		if len(exact) > 0 {
			fmt.Fprintf(
				sb,
				"%s        (%s)\n%s            __skip_next=1\n%s            ;;\n",
				inner,
				strings.Join(exact, "|"),
				inner,
				inner,
			)
		}
		if len(equals) > 0 {
			fmt.Fprintf(
				sb,
				"%s        (%s)\n%s            ;;\n",
				inner,
				strings.Join(equals, "|"),
				inner,
			)
		}
		fmt.Fprintf(sb, "%s        (-*)\n%s            ;;\n", inner, inner)
		fmt.Fprintf(
			sb,
			"%s        (*)\n%s            __pos+=(\"$token\")\n%s            ;;\n",
			inner,
			inner,
			inner,
		)
		fmt.Fprintf(sb, "%s    esac\n", inner)
		fmt.Fprintf(sb, "%sdone\n", inner)
		fmt.Fprintf(sb, "%slocal -a items\n", inner)
		fmt.Fprintf(
			sb,
			"%sitems=(${(f)\"$(%s --%s=%s -- \"${__pos[@]}\" 2>/dev/null)\"})\n",
			inner,
			g.AppName,
			FlagComplete,
			dynamicArgs[len(dynamicArgs)-1],
		)
		fmt.Fprintf(sb, "%scompadd -a items\n", inner)
		fmt.Fprintf(sb, "%s;;\n", indent)
	}
	fmt.Fprintf(sb, "%sesac\n", indent)
}

func zshWriteSpec(g *Generator, sb *strings.Builder, spec Spec, indent string) {
	help := zshEscapeHelp(spec.Terse)
	excl := zshExclusion(spec)

	if !spec.HasArg {
		if spec.ShortFlag != "" {
			fmt.Fprintf(sb, "%s'%s-%s[%s]' \\\n", indent, excl, spec.ShortFlag, help)
		}
		if spec.LongFlag != "" {
			fmt.Fprintf(sb, "%s'%s--%s[%s]' \\\n", indent, excl, spec.LongFlag, help)
		}
		return
	}

	completer := zshCompleter(g, spec)
	valueName := " "
	if spec.Dynamic != "" || len(spec.Values) > 0 || len(spec.ValueDescs) > 0 ||
		spec.Extension != "" ||
		spec.ValueHint != "" {
		valueName = spec.LongFlag
		if valueName == "" {
			valueName = spec.ShortFlag
		}
	}

	if spec.ShortFlag != "" {
		fmt.Fprintf(
			sb,
			"%s'%s-%s+[%s]:%s:%s' \\\n",
			indent,
			excl,
			spec.ShortFlag,
			help,
			valueName,
			completer,
		)
	}
	if spec.LongFlag != "" {
		fmt.Fprintf(
			sb,
			"%s'%s--%s=[%s]:%s:%s' \\\n",
			indent,
			excl,
			spec.LongFlag,
			help,
			valueName,
			completer,
		)
	}
}

func zshCompleter(g *Generator, spec Spec) string {
	switch {
	case spec.CommaList && spec.Dynamic != "":
		return fmt.Sprintf(
			"{_sequence compadd - $(%s --%s=%s)}",
			g.AppName,
			FlagComplete,
			spec.Dynamic,
		)
	case spec.CommaList && len(spec.Values) > 0:
		escaped := make([]string, len(spec.Values))
		for i, v := range spec.Values {
			escaped[i] = zshEscapeValue(v)
		}
		return "{_sequence compadd - " + strings.Join(escaped, " ") + "}"
	case spec.Dynamic != "":
		return fmt.Sprintf("($(%s --%s=%s))", g.AppName, FlagComplete, spec.Dynamic)
	case len(spec.ValueDescs) > 0:
		var parts []string
		for _, vd := range spec.ValueDescs {
			if vd.Desc != "" {
				parts = append(parts, zshEscapeValue(vd.Value)+`\:`+zshEscapeValue(vd.Desc))
			} else {
				parts = append(parts, zshEscapeValue(vd.Value))
			}
		}
		return "((" + strings.Join(parts, " ") + "))"
	case len(spec.Values) > 0:
		escaped := make([]string, len(spec.Values))
		for i, v := range spec.Values {
			escaped[i] = zshEscapeValue(v)
		}
		return "(" + strings.Join(escaped, " ") + ")"
	case spec.Extension != "":
		return `_files -g "` + zshExtGlob(spec.Extension) + `"`
	case spec.ValueHint != "":
		return zshHintCompleter(spec.ValueHint)
	default:
		return "_default"
	}
}

func zshWriteSubcommandCase(
	g *Generator,
	sb *strings.Builder,
	subs []SubSpec,
	inheritedSpecs []Spec,
	parentFuncName, stateName, indent string,
) {
	inner := indent + "    "
	fmt.Fprintf(sb, `
%[1]scase $state in
%[1]s(%[2]s)
%[3]swords=($line[1] "${words[@]}")
%[3]s(( CURRENT += 1 ))
%[3]scurcontext="${curcontext%%%%:*:*}:%[2]s-command-$line[1]:"
%[3]scase $line[1] in
`, indent, stateName, inner)

	for _, sub := range subs {
		names := append([]string{sub.Name}, sub.Aliases...)
		pattern := strings.Join(names, "|")

		caseIndent := inner + "    "
		bodyIndent := caseIndent + "    "
		fmt.Fprintf(sb, "%s(%s)\n", caseIndent, pattern)

		allSpecs := combineVisibleSpecs(inheritedSpecs, sub.Specs)
		nextInherited := appendSpecs(inheritedSpecs, persistentSpecs(sub.Specs))
		sortedChildSubs := SortSubSpecs(sub.Subs)
		childFuncName := parentFuncName + "__" + zshFuncName(sub.Name)

		if len(sortedChildSubs) > 0 {
			zshWriteArguments(
				g,
				sb,
				allSpecs,
				false,
				nil,
				false,
				0,
				childFuncName,
				sub.Name,
				bodyIndent,
			)
			zshWriteSubcommandCase(
				g,
				sb,
				sortedChildSubs,
				nextInherited,
				childFuncName,
				sub.Name,
				bodyIndent,
			)
		} else {
			zshWriteArguments(
				g,
				sb,
				allSpecs,
				sub.PathArgs,
				sub.DynamicArgs,
				sub.HasMaxPositionalArgs,
				sub.MaxPositionalArgs,
				"",
				"",
				bodyIndent,
			)
		}

		fmt.Fprintf(sb, "%s;;\n", caseIndent)
	}

	fmt.Fprintf(sb, "%[1]sesac\n%[2]s;;\n%[2]sesac\n", inner, indent)
}

func zshWriteDescribeTree(
	sb *strings.Builder,
	subs []SubSpec,
	funcName, displayName string,
) {
	zshWriteDescribeFunc(sb, subs, funcName, displayName)

	for _, sub := range subs {
		if len(sub.Subs) == 0 {
			continue
		}

		childFuncName := funcName + "__" + zshFuncName(sub.Name)
		childDisplayName := displayName + " " + sub.Name
		zshWriteDescribeTree(sb, SortSubSpecs(sub.Subs), childFuncName, childDisplayName)
	}
}

func zshWriteDescribeFunc(sb *strings.Builder, subs []SubSpec, funcName, displayName string) {
	fmt.Fprintf(
		sb,
		"\n(( $+functions[_%[1]s_commands] )) ||\n_%[1]s_commands() {\n    local commands; commands=(\n",
		funcName,
	)

	for _, sub := range subs {
		fmt.Fprintf(sb, "        '%s:%s' \\\n", sub.Name, zshEscapeHelp(sub.Terse))
	}

	fmt.Fprintf(
		sb,
		"    )\n    _describe -t commands '%s commands' commands \"$@\"\n}\n",
		displayName,
	)
}

func zshFuncName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func zshExclusion(spec Spec) string {
	if spec.ShortFlag != "" && spec.LongFlag != "" {
		return fmt.Sprintf("(-%s --%s)", spec.ShortFlag, spec.LongFlag)
	}
	return ""
}

func zshHintCompleter(hint string) string {
	switch hint {
	case HintFile:
		return "_files"
	case HintDir:
		return "_files -/"
	case HintCommand:
		return "_command_names -e"
	case HintUser:
		return "_users"
	case HintHost:
		return "_hosts"
	case HintURL:
		return "_urls"
	case HintEmail:
		return "_email_addresses"
	default:
		return "_default"
	}
}

func zshEscapeHelp(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `'`, `'\''`)
	s = strings.ReplaceAll(s, `[`, `\[`)
	s = strings.ReplaceAll(s, `]`, `\]`)
	s = strings.ReplaceAll(s, `:`, `\:`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "\n", " ")
	return s
}

func zshEscapeValue(s string) string {
	s = zshEscapeHelp(s)
	s = strings.ReplaceAll(s, `(`, `\(`)
	s = strings.ReplaceAll(s, `)`, `\)`)
	s = strings.ReplaceAll(s, ` `, `\ `)
	return s
}

func zshExtGlob(ext string) string {
	if !strings.Contains(ext, ",") {
		return "*." + ext
	}
	return "*.{" + ext + "}"
}
