//nolint:dupword // fish script keywords (e.g. "end\nend") trigger false positives
package complete

import (
	"fmt"
	"strings"
)

// GenerateFish generates a fish shell completion script.
func GenerateFish(g *Generator) (string, error) {
	if err := ValidateGenerator(g); err != nil {
		return "", err
	}

	var sb strings.Builder

	funcID := fishFuncName(g.AppName)
	rootSpecs := SortVisibleSpecs(g.Specs)
	rootPersistent := SortVisibleSpecs(persistentSpecs(g.Specs))
	rootLocal := fishNonPersistentSpecs(g.Specs)

	fmt.Fprintf(&sb, "complete -c %s -f\n", g.AppName)

	fishWriteCommaFunctions(g, &sb, rootSpecs, funcID)
	fishWriteCommaFunctionsTree(g, &sb, g.Subs, funcID)

	if len(g.Subs) > 0 {
		needsCmd := fmt.Sprintf("__%s_needs_command", funcID)
		usingSub := fmt.Sprintf("__%s_using_subcommand", funcID)

		fishWriteHelpers(g, &sb, funcID)

		fmt.Fprint(&sb, "\n")
		fishWriteSubEntries(g, &sb, g.Subs, needsCmd)

		if len(rootPersistent) > 0 || len(rootLocal) > 0 {
			fmt.Fprint(&sb, "\n")
			for _, spec := range rootPersistent {
				fishWriteSpec(g, &sb, spec, "")
			}
			for _, spec := range rootLocal {
				fishWriteSpec(g, &sb, spec, needsCmd)
			}
		}

		fishWriteSubTree(g, &sb, g.Subs, usingSub, "", funcID, persistentSpecs(g.Specs), 1)
	} else {
		fmt.Fprint(&sb, "\n")
		for _, spec := range rootSpecs {
			fishWriteSpec(g, &sb, spec, "")
		}
		if len(g.DynamicArgs) > 0 {
			helperName := fmt.Sprintf("__%s_dynamic_args", funcID)
			fishWriteDynamicArgsHelper(g, &sb, helperName, g.Specs, g.DynamicArgs, 0)
			fmt.Fprintf(&sb, "\ncomplete -c %s -a \"(%s)\" -f\n", g.AppName, helperName)
		}
	}

	return sb.String(), nil
}

func fishFuncName(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func fishEscapeString(s string) string {
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	s = strings.ReplaceAll(s, `$`, `\$`)
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return s
}

func fishEscapeDesc(s string) string {
	return fishEscapeString(s)
}

func fishWriteDynamicArgsHelper(
	g *Generator,
	sb *strings.Builder,
	helperName string,
	specs []Spec,
	dynamicArgs []string,
	cmdSkip int,
) {
	exact, equals := argValuePatterns(specs)
	fmt.Fprintf(sb, "\nfunction %s\n", helperName)
	fmt.Fprint(sb, "    set -l tokens (commandline -xpc)\n")
	fmt.Fprint(sb, "    set -e tokens[1]\n")
	fmt.Fprint(sb, "    set -l positional\n")
	fmt.Fprint(sb, "    set -l skip_next 0\n")
	fmt.Fprint(sb, "    set -l dashdash 0\n")
	if cmdSkip > 0 {
		fmt.Fprintf(sb, "    set -l cmd_skip %d\n", cmdSkip)
	}
	fmt.Fprint(sb, "    for t in $tokens\n")
	fmt.Fprint(sb, "        if test $dashdash -eq 1\n")
	fmt.Fprint(sb, "            set -a positional $t\n")
	fmt.Fprint(sb, "        else if test $skip_next -eq 1\n")
	fmt.Fprint(sb, "            set skip_next 0\n")
	fmt.Fprint(sb, "        else if test \"$t\" = --\n")
	fmt.Fprint(sb, "            set dashdash 1\n")
	if len(exact) > 0 {
		fmt.Fprintf(sb, "        else if contains -- $t %s\n", strings.Join(exact, " "))
		fmt.Fprint(sb, "            set skip_next 1\n")
	}
	if len(equals) > 0 {
		fmt.Fprintf(sb, "        else if %s\n", fishMatchPatterns("$t", equals))
		fmt.Fprint(sb, "            true\n")
	}
	fmt.Fprint(sb, "        else if not string match -q -- '-*' $t\n")
	if cmdSkip > 0 {
		fmt.Fprint(sb, "            if test $cmd_skip -gt 0\n")
		fmt.Fprint(sb, "                set cmd_skip (math $cmd_skip - 1)\n")
		fmt.Fprint(sb, "            else\n")
		fmt.Fprint(sb, "                set -a positional $t\n")
		fmt.Fprint(sb, "            end\n")
	} else {
		fmt.Fprint(sb, "            set -a positional $t\n")
	}
	fmt.Fprint(sb, "        end\n")
	fmt.Fprint(sb, "    end\n")
	fmt.Fprint(sb, "    set -l nargs (count $positional)\n")
	fmt.Fprint(sb, "    switch $nargs\n")
	for i, da := range dynamicArgs {
		fmt.Fprintf(sb, "        case %d\n", i)
		if i == 0 {
			fmt.Fprintf(sb, "            %s --%s=%s\n", g.AppName, FlagComplete, da)
			continue
		}

		fmt.Fprintf(sb, "            %s --%s=%s -- $positional\n", g.AppName, FlagComplete, da)
	}
	fmt.Fprint(sb, "    end\nend\n")
}

func fishMatchPatterns(token string, patterns []string) string {
	parts := make([]string, 0, len(patterns))
	for _, pattern := range patterns {
		parts = append(parts, fmt.Sprintf("string match -q -- '%s' %s", pattern, token))
	}
	return strings.Join(parts, "; or ")
}

func fishWriteCommaFunctions(
	g *Generator,
	sb *strings.Builder,
	specs []Spec,
	funcID string,
) {
	for _, spec := range specs {
		if !spec.CommaList || spec.LongFlag == "" {
			continue
		}
		if spec.Dynamic == "" && len(spec.Values) == 0 {
			continue
		}
		fmt.Fprint(sb, "\n")
		fishWriteCommaFunction(g, sb, spec, funcID)
	}
}

func fishWriteCommaFunctionsTree(
	g *Generator,
	sb *strings.Builder,
	subs []SubSpec,
	funcID string,
) {
	for _, sub := range subs {
		fishWriteCommaFunctions(g, sb, SortVisibleSpecs(sub.Specs), funcID)
		fishWriteCommaFunctionsTree(g, sb, sub.Subs, funcID)
	}
}

func fishWriteHelpers(g *Generator, sb *strings.Builder, funcID string) {
	optspecsFn := fmt.Sprintf("__%s_global_optspecs", funcID)
	needsFn := fmt.Sprintf("__%s_needs_command", funcID)
	usingFn := fmt.Sprintf("__%s_using_subcommand", funcID)

	fmt.Fprintf(sb, "\nfunction %s\n    string join \\n --", optspecsFn)
	for _, spec := range g.Specs {
		if spec.Hidden {
			continue
		}
		var optspec string
		switch {
		case spec.ShortFlag != "" && spec.LongFlag != "":
			optspec = spec.ShortFlag + "/" + spec.LongFlag
		case spec.LongFlag != "":
			optspec = spec.LongFlag
		case spec.ShortFlag != "":
			optspec = spec.ShortFlag
		default:
			continue
		}
		if spec.HasArg {
			optspec += "="
		}
		fmt.Fprintf(sb, " '%s'", optspec)
	}
	fmt.Fprint(sb, "\nend\n")

	fmt.Fprintf(sb, `
function %[1]s
    set -l cmd (commandline -xpc)
    set -e cmd[1]
    argparse -s (%[2]s) -- $cmd 2>/dev/null
    or return
    if set -q argv[1]
        echo $argv[1]
        return 1
    end
    return 0
end

function %[3]s
    set -l cmd (%[1]s)
    test -z "$cmd"
    and return 1
    contains -- $cmd[1] $argv
end
`, needsFn, optspecsFn, usingFn)
}

func fishWriteSubEntries(
	g *Generator,
	sb *strings.Builder,
	subs []SubSpec,
	condition string,
) {
	for _, sub := range SortSubSpecs(subs) {
		fishWriteSubcommand(g, sb, sub.Name, sub.Terse, condition)
	}
}

func fishWriteSubTree(
	g *Generator,
	sb *strings.Builder,
	subs []SubSpec,
	usingSub, parentCondition, funcID string,
	inheritedSpecs []Spec,
	depth int,
) {
	for _, sub := range SortSubSpecs(subs) {
		allNames := append([]string{sub.Name}, sub.Aliases...)
		subSeen := usingSub + " " + strings.Join(allNames, " ")

		seenCondition := subSeen
		if parentCondition != "" {
			seenCondition = parentCondition + "; and " + subSeen
		}

		leafCondition := seenCondition
		if len(sub.Subs) > 0 {
			var childNames []string
			for _, child := range sub.Subs {
				childNames = append(childNames, child.Name)
				childNames = append(childNames, child.Aliases...)
			}
			leafCondition += "; and not " + usingSub + " " + strings.Join(childNames, " ")
		}

		subPersistent := SortVisibleSpecs(persistentSpecs(sub.Specs))
		subLocal := fishNonPersistentSpecs(sub.Specs)
		hasDynArgs := len(sub.DynamicArgs) > 0
		if len(subPersistent) == 0 && len(subLocal) == 0 && len(sub.Subs) == 0 && !sub.PathArgs &&
			!hasDynArgs {
			continue
		}

		fmt.Fprint(sb, "\n")

		if len(sub.Subs) > 0 {
			fishWriteSubEntries(g, sb, sub.Subs, leafCondition)
		}
		for _, spec := range subPersistent {
			fishWriteSpec(g, sb, spec, seenCondition)
		}
		for _, spec := range subLocal {
			fishWriteSpec(g, sb, spec, leafCondition)
		}
		if sub.PathArgs {
			fmt.Fprintf(sb, "complete -c %s -n '%s' -F\n", g.AppName, leafCondition)
		}
		if hasDynArgs {
			helperName := fmt.Sprintf("__%s_%s_dynamic_args", funcID, fishFuncName(sub.Name))
			allSpecs := appendSpecs(inheritedSpecs, sub.Specs)
			fishWriteDynamicArgsHelper(g, sb, helperName, allSpecs, sub.DynamicArgs, depth)
			fmt.Fprintf(
				sb,
				"\ncomplete -c %s -n '%s' -a \"(%s)\" -f\n",
				g.AppName,
				leafCondition,
				helperName,
			)
		}
		if len(sub.Subs) > 0 {
			nextInherited := appendSpecs(inheritedSpecs, persistentSpecs(sub.Specs))
			fishWriteSubTree(
				g,
				sb,
				sub.Subs,
				usingSub,
				seenCondition,
				funcID,
				nextInherited,
				depth+1,
			)
		}
	}
}

func fishWriteCommaFunction(
	g *Generator,
	sb *strings.Builder,
	spec Spec,
	funcID string,
) {
	funcName := fmt.Sprintf("__%s_complete_%s", funcID, fishFuncName(spec.LongFlag))
	varName := fishFuncName(spec.LongFlag)

	fmt.Fprintf(sb, "# Comma-separated %[1]s completion\nfunction %[2]s\n", spec.LongFlag, funcName)
	fmt.Fprintf(
		sb,
		"    set -l value (string replace -r '^--%s=' '' -- (commandline -ct))\n",
		spec.LongFlag,
	)
	if spec.Dynamic != "" {
		fmt.Fprintf(
			sb,
			"    set -l %s (%s --%s=%s)\n",
			varName,
			g.AppName,
			FlagComplete,
			spec.Dynamic,
		)
	} else {
		fmt.Fprintf(sb, "    set -l %s %s\n", varName, fishQuotedWords(spec.Values))
	}
	fmt.Fprintf(sb, `    if string match -qr '^(?<prefix>.*,)' -- $value
        set -l selected (string split ',' -- $prefix)
        for col in $%[1]s
            if not contains -- $col $selected
                printf '%%s\n' "$prefix$col"
            end
        end
    else
        printf '%%s\n' $%[1]s
    end
end
`, varName)
}

func fishWriteSubcommand(g *Generator, sb *strings.Builder, name, terse, condition string) {
	var parts []string
	parts = append(parts, fmt.Sprintf("complete -c %s", g.AppName))
	if condition != "" {
		parts = append(parts, fmt.Sprintf("-n '%s'", condition))
	}
	parts = append(parts, fmt.Sprintf("-a %s", name))
	if terse != "" {
		parts = append(parts, fmt.Sprintf("-d %q", terse))
	}
	fmt.Fprintf(sb, "%s\n", strings.Join(parts, " "))
}

func fishWriteSpec(g *Generator, sb *strings.Builder, spec Spec, condition string) {
	var parts []string
	parts = append(parts, fmt.Sprintf("complete -c %s", g.AppName))

	if condition != "" {
		parts = append(parts, fmt.Sprintf("-n '%s'", condition))
	}
	if spec.ShortFlag != "" {
		parts = append(parts, fmt.Sprintf("-s %s", spec.ShortFlag))
	}
	if spec.LongFlag != "" {
		parts = append(parts, fmt.Sprintf("-l %s", spec.LongFlag))
	}

	if spec.HasArg {
		switch {
		case spec.CommaList && (spec.Dynamic != "" || len(spec.Values) > 0):
			funcName := fmt.Sprintf(
				"__%s_complete_%s",
				fishFuncName(g.AppName),
				fishFuncName(spec.LongFlag),
			)
			parts = append(parts, "-x", fmt.Sprintf("-kra \"(%s)\"", funcName))
		case spec.Dynamic != "":
			parts = append(
				parts,
				"-x",
				fmt.Sprintf("-a \"(%s --%s=%s)\"", g.AppName, FlagComplete, spec.Dynamic),
			)
		case len(spec.ValueDescs) > 0:
			var vals []string
			for _, vd := range spec.ValueDescs {
				value := fishEscapeString(vd.Value)
				if vd.Desc != "" {
					vals = append(vals, fmt.Sprintf("%s\t%s", value, fishEscapeDesc(vd.Desc)))
				} else {
					vals = append(vals, value)
				}
			}
			parts = append(parts, "-x", fmt.Sprintf("-a \"%s\"", strings.Join(vals, "\n")))
		case len(spec.Values) > 0:
			values := make([]string, len(spec.Values))
			for i, value := range spec.Values {
				values[i] = fishEscapeString(value)
			}
			parts = append(parts, "-x", fmt.Sprintf("-a \"%s\"", strings.Join(values, " ")))
		case spec.Extension != "":
			suffixes := fishExtToSuffixes(spec.Extension)
			parts = append(
				parts,
				"-k",
				fmt.Sprintf("-xa \"(__fish_complete_suffix %s)\"", strings.Join(suffixes, " ")),
			)
		case spec.ValueHint != "":
			switch spec.ValueHint {
			case HintFile:
				parts = append(parts, "-r", "-F")
			case HintDir:
				parts = append(parts, "-r", "-f", "-a \"(__fish_complete_directories)\"")
			case HintCommand:
				parts = append(parts, "-r", "-f", "-a \"(__fish_complete_command)\"")
			case HintUser:
				parts = append(parts, "-r", "-f", "-a \"(__fish_complete_users)\"")
			case HintHost:
				parts = append(parts, "-r", "-f", "-a \"(__fish_print_hostnames)\"")
			default:
				parts = append(parts, "-r", "-f")
			}
		default:
			parts = append(parts, "-r")
		}
	}

	if spec.Terse != "" {
		parts = append(parts, fmt.Sprintf("-d %q", spec.Terse))
	}

	fmt.Fprintf(sb, "%s\n", strings.Join(parts, " "))
}

func fishQuotedWords(values []string) string {
	quoted := make([]string, len(values))
	for i, value := range values {
		quoted[i] = fmt.Sprintf(`"%s"`, fishEscapeString(value))
	}
	return strings.Join(quoted, " ")
}

func fishExtToSuffixes(ext string) []string {
	parts := strings.Split(ext, ",")
	suffixes := make([]string, len(parts))
	for i, p := range parts {
		suffixes[i] = "." + strings.TrimSpace(p)
	}
	return suffixes
}

func fishNonPersistentSpecs(specs []Spec) []Spec {
	var local []Spec
	for _, spec := range specs {
		if !spec.Persistent {
			local = append(local, spec)
		}
	}
	return SortVisibleSpecs(local)
}
