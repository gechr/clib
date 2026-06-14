package complete

import (
	"fmt"
	"slices"
	"strings"
)

// GeneratePwsh generates a PowerShell completion script.
//
// PowerShell native completers receive the raw command-line AST rather than a
// pre-tokenized word list, so the generated script walks $commandAst itself: it
// first resolves a canonical "app;sub;subsub" command path (canonicalizing
// aliases at every depth), then switches on that path to emit candidates. Flag
// values are completed by inspecting the token preceding the cursor.
func GeneratePwsh(g *Generator) (string, error) {
	if err := ValidateGenerator(g); err != nil {
		return "", err
	}

	var sb strings.Builder
	id := pwshID(g.AppName)

	fmt.Fprintf(&sb, "# %s PowerShell completion\n\n", g.AppName)
	sb.WriteString("using namespace System.Management.Automation\n")
	sb.WriteString("using namespace System.Management.Automation.Language\n")

	needTokens := hasDynamicArgs(g) || forwardingActive(g)
	if needTokens {
		pwshWriteTokensHelper(&sb, id)
	}
	if forwardingActive(g) {
		pwshWriteForwardedHelper(&sb, id, allForwardableSpecs(g))
	}
	if hasDynamicArgs(g) {
		pwshWritePositionalsHelper(&sb, id)
	}

	fmt.Fprintf(
		&sb,
		"\nRegister-ArgumentCompleter -Native -CommandName '%s' -ScriptBlock {\n",
		g.AppName,
	)
	sb.WriteString("    param($wordToComplete, $commandAst, $cursorPosition)\n\n")
	sb.WriteString("    $commandElements = $commandAst.CommandElements\n")
	fmt.Fprintf(&sb, "    $command = '%s'\n", g.AppName)

	if len(g.Subs) > 0 {
		pwshWriteResolution(&sb, g)
	}
	if hasAnyValueFlag(g) {
		pwshWritePrev(&sb)
	}

	needFiles := hasFileCompletion(g)
	sb.WriteString("\n")
	if needFiles {
		// File/dir/command hints are collected here rather than emitted into
		// $completions: a `return` inside the @(switch ...) assignment is
		// swallowed, and their results must bypass the prefix filter applied to
		// flag and value candidates below.
		sb.WriteString("    $fileCompletions = $null\n")
	}
	sb.WriteString("    $completions = @(switch ($command) {\n")
	pwshWriteCommandCase(
		&sb,
		g,
		g.AppName,
		combineVisibleSpecs(g.Specs),
		g.Subs,
		false,
		g.DynamicArgs,
		g.HasMaxPositionalArgs,
		g.MaxPositionalArgs,
		1,
	)
	if len(g.Subs) > 0 {
		pwshWriteSubCases(
			&sb,
			g,
			g.AppName,
			g.Subs,
			persistentSpecs(g.Specs),
			2, //nolint:mnd // depth 2 = first subcommand level
		)
	}
	sb.WriteString("    })\n\n")

	if needFiles {
		sb.WriteString("    if ($null -ne $fileCompletions) {\n")
		sb.WriteString("        $fileCompletions\n")
		sb.WriteString("    }\n")
	}
	sb.WriteString(
		"    $completions | Where-Object { $_.CompletionText -like \"$wordToComplete*\" } |\n",
	)
	sb.WriteString("        Sort-Object -Property ListItemText\n")
	sb.WriteString("}\n")

	return sb.String(), nil
}

func pwshID(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

// pwshQuote renders s as a PowerShell single-quoted string literal, doubling
// embedded single quotes and folding whitespace so the result stays on one line.
func pwshQuote(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	s = strings.ReplaceAll(s, "'", "''")
	return "'" + s + "'"
}

// pwshTooltip returns desc when set, else text; CompletionResult rejects empty
// list-item and tooltip strings, so callers must always pass a non-empty value.
func pwshTooltip(text, desc string) string {
	if strings.TrimSpace(desc) != "" {
		return desc
	}
	return text
}

func pwshResult(sb *strings.Builder, indent, completion, listItem, resultType, tooltip string) {
	fmt.Fprintf(
		sb,
		"%s[CompletionResult]::new(%s, %s, [CompletionResultType]::%s, %s)\n",
		indent,
		pwshQuote(completion),
		pwshQuote(listItem),
		resultType,
		pwshQuote(tooltip),
	)
}

// hasAnyValueFlag reports whether any flag in the tree takes a value, i.e.
// whether the script needs a $prev-based value-completion switch.
func hasAnyValueFlag(g *Generator) bool {
	var walk func([]Spec, []SubSpec) bool
	walk = func(specs []Spec, subs []SubSpec) bool {
		for _, spec := range specs {
			if spec.HasArg {
				return true
			}
		}
		for _, sub := range subs {
			if walk(sub.Specs, sub.Subs) {
				return true
			}
		}
		return false
	}
	return walk(g.Specs, g.Subs)
}

// hasFileCompletion reports whether any flag value or positional uses
// PowerShell's built-in filesystem/command completers (extension filters, file,
// dir, or command hints, or PathArgs). These results are routed through the
// $fileCompletions variable so they bypass the prefix filter.
func hasFileCompletion(g *Generator) bool {
	specHasFile := func(spec Spec) bool {
		return spec.HasArg && spec.Dynamic == "" && !spec.CommaList &&
			len(spec.Values) == 0 && len(spec.ValueDescs) == 0 &&
			(spec.Extension != "" || spec.ValueHint == HintFile ||
				spec.ValueHint == HintDir || spec.ValueHint == HintCommand)
	}
	var walk func([]Spec, []SubSpec) bool
	walk = func(specs []Spec, subs []SubSpec) bool {
		if slices.ContainsFunc(specs, specHasFile) {
			return true
		}
		for _, sub := range subs {
			if sub.PathArgs || walk(sub.Specs, sub.Subs) {
				return true
			}
		}
		return false
	}
	return walk(g.Specs, g.Subs)
}

// pwshWriteResolution emits the command-path resolution loop. It walks bareword
// elements, skips flags, and uses a transition switch to canonicalize each
// subcommand (including aliases) into the running "app;sub" path.
func pwshWriteResolution(sb *strings.Builder, g *Generator) {
	sb.WriteString("    for ($i = 1; $i -lt $commandElements.Count; $i++) {\n")
	sb.WriteString("        $element = $commandElements[$i]\n")
	sb.WriteString("        if ($element -isnot [StringConstantExpressionAst] -or\n")
	sb.WriteString(
		"            $element.StringConstantType -ne [StringConstantType]::BareWord) {\n",
	)
	sb.WriteString("            break\n")
	sb.WriteString("        }\n")
	sb.WriteString("        $value = $element.Value\n")
	sb.WriteString("        if ($value -eq $wordToComplete) { break }\n")
	sb.WriteString("        if ($value.StartsWith('-')) { continue }\n")
	sb.WriteString("        $next = switch (\"$command;$value\") {\n")
	pwshWriteTransitions(sb, g.AppName, g.Subs)
	sb.WriteString("            default { '' }\n")
	sb.WriteString("        }\n")
	sb.WriteString("        if ($next -eq '') { break }\n")
	sb.WriteString("        $command = $next\n")
	sb.WriteString("    }\n")
}

func pwshWriteTransitions(sb *strings.Builder, parentPath string, subs []SubSpec) {
	for _, sub := range SortSubSpecs(subs) {
		childPath := parentPath + ";" + sub.Name
		for _, name := range append([]string{sub.Name}, sub.Aliases...) {
			fmt.Fprintf(
				sb,
				"            '%s;%s' { '%s'; break }\n",
				parentPath,
				name,
				childPath,
			)
		}
		if len(sub.Subs) > 0 {
			pwshWriteTransitions(sb, childPath, sub.Subs)
		}
	}
}

// pwshWritePrev emits the logic that captures the token immediately before the
// cursor, used to decide whether a flag value is being completed.
func pwshWritePrev(sb *strings.Builder) {
	sb.WriteString("    $prev = ''\n")
	sb.WriteString("    if ($commandElements.Count -ge 1) {\n")
	sb.WriteString("        $lastText = $commandElements[-1].Extent.Text\n")
	sb.WriteString("        if ($wordToComplete -ne '' -and $lastText -eq $wordToComplete) {\n")
	sb.WriteString("            if ($commandElements.Count -ge 2) {\n")
	sb.WriteString("                $prev = $commandElements[-2].Extent.Text\n")
	sb.WriteString("            }\n")
	sb.WriteString("        } else {\n")
	sb.WriteString("            $prev = $lastText\n")
	sb.WriteString("        }\n")
	sb.WriteString("    }\n")
}

func pwshWriteSubCases(
	sb *strings.Builder,
	g *Generator,
	parentPath string,
	subs []SubSpec,
	inheritedSpecs []Spec,
	depth int,
) {
	for _, sub := range SortSubSpecs(subs) {
		childPath := parentPath + ";" + sub.Name
		visibleSpecs := combineVisibleSpecs(inheritedSpecs, sub.Specs)

		pwshWriteCommandCase(
			sb,
			g,
			childPath,
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
		pwshWriteSubCases(sb, g, childPath, sub.Subs, nextInherited, depth+1)
	}
}

// pwshWriteCommandCase emits a single switch arm for the given command path.
// The arm routes flag-value completion through a $prev switch (when the command
// has value-taking flags) and otherwise lists flags, subcommands, and positional
// completions.
func pwshWriteCommandCase(
	sb *strings.Builder,
	g *Generator,
	path string,
	specs []Spec,
	subs []SubSpec,
	pathArgs bool,
	dynamicArgs []string,
	hasMax bool,
	maxPos int,
	depth int,
) {
	hasContent := len(specs) > 0 || len(subs) > 0 || len(dynamicArgs) > 0 || pathArgs
	if !hasContent {
		return
	}

	fmt.Fprintf(sb, "        '%s' {\n", path)

	var valueFlags []Spec
	for _, spec := range specs {
		if spec.HasArg {
			valueFlags = append(valueFlags, spec)
		}
	}

	if len(valueFlags) > 0 {
		sb.WriteString("            switch ($prev) {\n")
		for _, spec := range valueFlags {
			pwshWriteValueClause(sb, g, spec, "                ")
		}
		sb.WriteString("                default {\n")
		pwshWriteListing(
			sb,
			g,
			specs,
			subs,
			pathArgs,
			dynamicArgs,
			hasMax,
			maxPos,
			depth,
			"                    ",
		)
		sb.WriteString("                }\n")
		sb.WriteString("            }\n")
	} else {
		pwshWriteListing(
			sb,
			g,
			specs,
			subs,
			pathArgs,
			dynamicArgs,
			hasMax,
			maxPos,
			depth,
			"            ",
		)
	}

	sb.WriteString("            break\n")
	sb.WriteString("        }\n")
}

func pwshFlagPatterns(spec Spec) []string {
	var patterns []string
	if spec.LongFlag != "" {
		patterns = append(patterns, pwshQuote("--"+spec.LongFlag))
	}
	if spec.ShortFlag != "" {
		patterns = append(patterns, pwshQuote("-"+spec.ShortFlag))
	}
	return patterns
}

// pwshWriteValueClause emits one $prev arm completing the value of a single
// flag. File/dir/command hints assign PowerShell's built-in completers to
// $fileCompletions (emitted after the switch, bypassing the prefix filter);
// enum, value-description, comma-list, and dynamic completions emit
// CompletionResult objects for the shared prefix filter.
func pwshWriteValueClause(sb *strings.Builder, g *Generator, spec Spec, indent string) {
	patterns := pwshFlagPatterns(spec)
	if len(patterns) == 0 {
		return
	}
	fmt.Fprintf(sb, "%s{ $_ -in %s } {\n", indent, strings.Join(patterns, ", "))
	inner := indent + "    "

	switch {
	case spec.CommaList && spec.Dynamic != "":
		pwshWriteCommaDynamic(sb, g, spec, inner)
	case spec.CommaList && len(spec.ValueDescs) > 0:
		pwshWriteCommaStatic(sb, valueDescPairs(spec.ValueDescs), inner)
	case spec.CommaList && len(spec.Values) > 0:
		pairs := make([][2]string, len(spec.Values))
		for i, v := range spec.Values {
			pairs[i] = [2]string{v, v}
		}
		pwshWriteCommaStatic(sb, pairs, inner)
	case spec.Dynamic != "":
		pwshWriteDynamicValues(sb, g, spec.Dynamic, inner)
	case len(spec.ValueDescs) > 0:
		for _, vd := range spec.ValueDescs {
			pwshResult(
				sb,
				inner,
				vd.Value,
				vd.Value,
				"ParameterValue",
				pwshTooltip(vd.Value, vd.Desc),
			)
		}
	case len(spec.Values) > 0:
		for _, v := range spec.Values {
			pwshResult(sb, inner, v, v, "ParameterValue", v)
		}
	case spec.Extension != "":
		fmt.Fprintf(
			sb,
			"%s$fileCompletions = [CompletionCompleters]::CompleteFilename($wordToComplete) | Where-Object {\n",
			inner,
		)
		fmt.Fprintf(
			sb,
			"%s    $_.ResultType -eq [CompletionResultType]::ProviderContainer -or $_.ListItemText -match '%s'\n",
			inner,
			pwshExtRegex(spec.Extension),
		)
		fmt.Fprintf(sb, "%s}\n", inner)
	case spec.ValueHint == HintFile:
		fmt.Fprintf(
			sb,
			"%s$fileCompletions = [CompletionCompleters]::CompleteFilename($wordToComplete)\n",
			inner,
		)
	case spec.ValueHint == HintDir:
		fmt.Fprintf(
			sb,
			"%s$fileCompletions = [CompletionCompleters]::CompleteFilename($wordToComplete) | Where-Object {\n",
			inner,
		)
		fmt.Fprintf(
			sb,
			"%s    $_.ResultType -eq [CompletionResultType]::ProviderContainer\n",
			inner,
		)
		fmt.Fprintf(sb, "%s}\n", inner)
	case spec.ValueHint == HintCommand:
		fmt.Fprintf(
			sb,
			"%s$fileCompletions = [CompletionCompleters]::CompleteCommand($wordToComplete)\n",
			inner,
		)
	}

	fmt.Fprintf(sb, "%sbreak\n", inner)
	fmt.Fprintf(sb, "%s}\n", indent)
}

func valueDescPairs(descs []ValueDesc) [][2]string {
	pairs := make([][2]string, len(descs))
	for i, vd := range descs {
		pairs[i] = [2]string{vd.Value, pwshTooltip(vd.Value, vd.Desc)}
	}
	return pairs
}

// pwshWriteListing emits flag-name and subcommand candidates, plus file or
// dynamic positional completions, for a command's positional context.
func pwshWriteListing(
	sb *strings.Builder,
	g *Generator,
	specs []Spec,
	subs []SubSpec,
	pathArgs bool,
	dynamicArgs []string,
	hasMax bool,
	maxPos int,
	depth int,
	indent string,
) {
	for _, spec := range specs {
		if spec.Hidden {
			continue
		}
		if spec.LongFlag != "" {
			flag := "--" + spec.LongFlag
			pwshResult(sb, indent, flag, flag, "ParameterName", pwshTooltip(flag, spec.Terse))
		}
		if spec.ShortFlag != "" {
			flag := "-" + spec.ShortFlag
			pwshResult(sb, indent, flag, flag, "ParameterName", pwshTooltip(flag, spec.Terse))
		}
	}
	for _, sub := range SortSubSpecs(subs) {
		pwshResult(
			sb,
			indent,
			sub.Name,
			sub.Name,
			"ParameterValue",
			pwshTooltip(sub.Name, sub.Terse),
		)
	}

	if pathArgs {
		fmt.Fprintf(
			sb,
			"%s$fileCompletions = [CompletionCompleters]::CompleteFilename($wordToComplete)\n",
			indent,
		)
	}
	if len(dynamicArgs) > 0 {
		pwshWriteDynArgs(sb, g, specs, dynamicArgs, hasMax, maxPos, depth, indent)
	}
}

// pwshWriteDynArgs emits the positional dynamic-completion logic: it counts the
// real positionals on the command line, selects the matching slot, and invokes
// the handler with forwarded context flags and preceding positionals.
func pwshWriteDynArgs(
	sb *strings.Builder,
	g *Generator,
	specs []Spec,
	dynamicArgs []string,
	hasMax bool,
	maxPos int,
	depth int,
	indent string,
) {
	id := pwshID(g.AppName)
	forward := forwardingActive(g)
	cmdSkip := depth - 1

	fmt.Fprintf(
		sb,
		"%s$tokens = @(__%s_Tokens $commandAst $wordToComplete)\n",
		indent,
		id,
	)
	fmt.Fprintf(
		sb,
		"%s$positional = @(__%s_Positionals $tokens %d %s)\n",
		indent,
		id,
		cmdSkip,
		pwshValueFlagArray(specs),
	)
	fmt.Fprintf(sb, "%s$n = $positional.Count\n", indent)

	body := indent
	if hasMax {
		fmt.Fprintf(sb, "%sif ($n -lt %d) {\n", indent, maxPos)
		body = indent + "    "
	}

	fmt.Fprintf(sb, "%s$kind = ''\n", body)
	fmt.Fprintf(sb, "%s$first = $false\n", body)
	fmt.Fprintf(sb, "%sswitch ($n) {\n", body)
	for i, da := range dynamicArgs {
		first := "$false"
		if i == 0 {
			first = "$true"
		}
		fmt.Fprintf(sb, "%s    %d { $kind = %s; $first = %s }\n", body, i, pwshQuote(da), first)
	}
	fmt.Fprintf(
		sb,
		"%s    default { $kind = %s }\n",
		body,
		pwshQuote(dynamicArgs[len(dynamicArgs)-1]),
	)
	fmt.Fprintf(sb, "%s}\n", body)

	if forward {
		fmt.Fprintf(sb, "%s$fwd = @(__%s_ForwardedFlags $tokens)\n", body, id)
	}
	fmt.Fprintf(sb, "%s$callArgs = @(\"--%s=$kind\")\n", body, FlagComplete)
	switch {
	case forward:
		fmt.Fprintf(sb, "%s$callArgs += '--'\n", body)
		fmt.Fprintf(sb, "%s$callArgs += $fwd\n", body)
		fmt.Fprintf(sb, "%sif (-not $first) { $callArgs += $positional }\n", body)
	default:
		fmt.Fprintf(sb, "%sif (-not $first) {\n", body)
		fmt.Fprintf(sb, "%s    $callArgs += '--'\n", body)
		fmt.Fprintf(sb, "%s    $callArgs += $positional\n", body)
		fmt.Fprintf(sb, "%s}\n", body)
	}
	fmt.Fprintf(
		sb,
		"%s& '%s' @callArgs 2>$null | Where-Object { $_ } | ForEach-Object {\n",
		body,
		g.AppName,
	)
	fmt.Fprintf(
		sb,
		"%s    [CompletionResult]::new($_, $_, [CompletionResultType]::ParameterValue, $_)\n",
		body,
	)
	fmt.Fprintf(sb, "%s}\n", body)

	if hasMax {
		fmt.Fprintf(sb, "%s}\n", indent)
	}
}

// pwshValueFlagArray renders the value-taking flags of specs as a PowerShell
// string array, used by the positional scanner to skip a flag and its value.
func pwshValueFlagArray(specs []Spec) string {
	var tokens []string
	for _, spec := range specs {
		if !spec.HasArg {
			continue
		}
		if spec.LongFlag != "" {
			tokens = append(tokens, pwshQuote("--"+spec.LongFlag))
		}
		if spec.ShortFlag != "" {
			tokens = append(tokens, pwshQuote("-"+spec.ShortFlag))
		}
	}
	if len(tokens) == 0 {
		return "@()"
	}
	return "@(" + strings.Join(tokens, ", ") + ")"
}

func pwshDynamicCall(g *Generator, predictor string) string {
	if forwardingActive(g) {
		return fmt.Sprintf(
			"& '%s' --%s=%s -- $fwd 2>$null",
			g.AppName,
			FlagComplete,
			predictor,
		)
	}
	return fmt.Sprintf("& '%s' --%s=%s 2>$null", g.AppName, FlagComplete, predictor)
}

func pwshWriteDynamicValues(sb *strings.Builder, g *Generator, predictor, indent string) {
	if forwardingActive(g) {
		fmt.Fprintf(
			sb,
			"%s$fwd = @(__%s_ForwardedFlags @(__%s_Tokens $commandAst $wordToComplete))\n",
			indent,
			pwshID(g.AppName),
			pwshID(g.AppName),
		)
	}
	fmt.Fprintf(
		sb,
		"%s%s | Where-Object { $_ } | ForEach-Object {\n",
		indent,
		pwshDynamicCall(g, predictor),
	)
	fmt.Fprintf(
		sb,
		"%s    [CompletionResult]::new($_, $_, [CompletionResultType]::ParameterValue, $_)\n",
		indent,
	)
	fmt.Fprintf(sb, "%s}\n", indent)
}

// pwshWriteCommaStatic emits comma-separated value completion: it splits the
// in-progress word on commas, drops already-selected values, and re-emits each
// remaining candidate carrying the accumulated prefix.
func pwshWriteCommaStatic(sb *strings.Builder, pairs [][2]string, indent string) {
	pwshWriteCommaPrefix(sb, indent)
	fmt.Fprintf(sb, "%sforeach ($pair in @(\n", indent)
	for _, p := range pairs {
		fmt.Fprintf(sb, "%s    @(%s, %s)\n", indent, pwshQuote(p[0]), pwshQuote(p[1]))
	}
	fmt.Fprintf(sb, "%s)) {\n", indent)
	fmt.Fprintf(sb, "%s    if ($selected -notcontains $pair[0]) {\n", indent)
	fmt.Fprintf(
		sb,
		"%s        [CompletionResult]::new(\"$prefix$($pair[0])\", \"$prefix$($pair[0])\", [CompletionResultType]::ParameterValue, $pair[1])\n",
		indent,
	)
	fmt.Fprintf(sb, "%s    }\n", indent)
	fmt.Fprintf(sb, "%s}\n", indent)
}

func pwshWriteCommaDynamic(sb *strings.Builder, g *Generator, spec Spec, indent string) {
	pwshWriteCommaPrefix(sb, indent)
	if forwardingActive(g) {
		fmt.Fprintf(
			sb,
			"%s$fwd = @(__%s_ForwardedFlags @(__%s_Tokens $commandAst $wordToComplete))\n",
			indent,
			pwshID(g.AppName),
			pwshID(g.AppName),
		)
	}
	fmt.Fprintf(
		sb,
		"%s%s | Where-Object { $_ } | ForEach-Object {\n",
		indent,
		pwshDynamicCall(g, spec.Dynamic),
	)
	fmt.Fprintf(sb, "%s    if ($selected -notcontains $_) {\n", indent)
	fmt.Fprintf(
		sb,
		"%s        [CompletionResult]::new(\"$prefix$_\", \"$prefix$_\", [CompletionResultType]::ParameterValue, $_)\n",
		indent,
	)
	fmt.Fprintf(sb, "%s    }\n", indent)
	fmt.Fprintf(sb, "%s}\n", indent)
}

func pwshWriteCommaPrefix(sb *strings.Builder, indent string) {
	fmt.Fprintf(sb, "%s$prefix = ''\n", indent)
	fmt.Fprintf(
		sb,
		"%sif ($wordToComplete -match '^(.*,)([^,]*)$') { $prefix = $Matches[1] }\n",
		indent,
	)
	fmt.Fprintf(sb, "%s$selected = @($prefix -split ',' | Where-Object { $_ })\n", indent)
}

// pwshExtRegex builds a case-insensitive suffix regex for an extension filter
// such as "yaml,yml" -> "\.(yaml|yml)$".
func pwshExtRegex(ext string) string {
	parts := strings.Split(ext, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return `\.(` + strings.Join(parts, "|") + `)$`
}

func pwshWriteTokensHelper(sb *strings.Builder, id string) {
	fmt.Fprintf(sb, "\nfunction __%s_Tokens {\n", id)
	sb.WriteString("    param($CommandAst, $WordToComplete)\n")
	sb.WriteString("    $elements = $CommandAst.CommandElements\n")
	sb.WriteString("    $tokens = @()\n")
	sb.WriteString("    for ($i = 1; $i -lt $elements.Count; $i++) {\n")
	sb.WriteString("        $text = $elements[$i].Extent.Text\n")
	sb.WriteString(
		"        if ($WordToComplete -ne '' -and $i -eq ($elements.Count - 1) -and $text -eq $WordToComplete) {\n",
	)
	sb.WriteString("            continue\n")
	sb.WriteString("        }\n")
	sb.WriteString("        $tokens += $text\n")
	sb.WriteString("    }\n")
	sb.WriteString("    , $tokens\n")
	sb.WriteString("}\n")
}

// pwshWriteForwardedHelper emits a helper that scans the tokens for forwardable
// context flags and returns them normalized as --name=value, stopping at "--".
func pwshWriteForwardedHelper(sb *strings.Builder, id string, fwd []forwardSpec) {
	fmt.Fprintf(sb, "\nfunction __%s_ForwardedFlags {\n", id)
	sb.WriteString("    param([string[]]$Tokens)\n")
	sb.WriteString("    $forwarded = @()\n")
	sb.WriteString("    $skipNext = $false\n")
	sb.WriteString("    $name = ''\n")
	sb.WriteString("    foreach ($t in $Tokens) {\n")
	sb.WriteString("        if ($skipNext) {\n")
	sb.WriteString("            if ($name -ne '') { $forwarded += \"--$name=$t\"; $name = '' }\n")
	sb.WriteString("            $skipNext = $false\n")
	sb.WriteString("            continue\n")
	sb.WriteString("        }\n")
	sb.WriteString("        if ($t -eq '--') { break }\n")
	sb.WriteString("        switch -regex ($t) {\n")
	for _, f := range fwd {
		var bare []string
		if f.LongFlag != "" {
			bare = append(bare, "--"+f.LongFlag)
		}
		if f.ShortFlag != "" {
			bare = append(bare, "-"+f.ShortFlag)
		}
		fmt.Fprintf(
			sb,
			"            '^(%s)$' { $skipNext = $true; $name = '%s'; break }\n",
			strings.Join(bare, "|"),
			f.LongFlag,
		)
		if f.LongFlag != "" {
			fmt.Fprintf(sb, "            '^--%s=' { $forwarded += $t; break }\n", f.LongFlag)
		}
		if f.ShortFlag != "" && f.LongFlag != "" {
			prefixLen := len("-" + f.ShortFlag + "=")
			fmt.Fprintf(
				sb,
				"            '^-%s=' { $forwarded += ('--%s=' + $t.Substring(%d)); break }\n",
				f.ShortFlag,
				f.LongFlag,
				prefixLen,
			)
		}
	}
	sb.WriteString("            default { }\n")
	sb.WriteString("        }\n")
	sb.WriteString("    }\n")
	sb.WriteString("    , $forwarded\n")
	sb.WriteString("}\n")
}

// pwshWritePositionalsHelper emits a helper that extracts real positional
// arguments from the tokens, skipping flags, their values, leading subcommand
// tokens (CmdSkip), and honoring the "--" terminator.
func pwshWritePositionalsHelper(sb *strings.Builder, id string) {
	fmt.Fprintf(sb, "\nfunction __%s_Positionals {\n", id)
	sb.WriteString("    param([string[]]$Tokens, [int]$CmdSkip, [string[]]$ValueFlags)\n")
	sb.WriteString("    $positional = @()\n")
	sb.WriteString("    $skipNext = $false\n")
	sb.WriteString("    $afterDoubleDash = $false\n")
	sb.WriteString("    $skip = $CmdSkip\n")
	sb.WriteString("    foreach ($t in $Tokens) {\n")
	sb.WriteString("        if ($afterDoubleDash) { $positional += $t; continue }\n")
	sb.WriteString("        if ($skipNext) { $skipNext = $false; continue }\n")
	sb.WriteString("        if ($t -eq '--') { $afterDoubleDash = $true; continue }\n")
	sb.WriteString("        if ($ValueFlags -contains $t) { $skipNext = $true; continue }\n")
	sb.WriteString("        if ($t -match '^-.+=') { continue }\n")
	sb.WriteString("        if ($t -like '-*') { continue }\n")
	sb.WriteString("        if ($skip -gt 0) { $skip--; continue }\n")
	sb.WriteString("        $positional += $t\n")
	sb.WriteString("    }\n")
	sb.WriteString("    , $positional\n")
	sb.WriteString("}\n")
}
