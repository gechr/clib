package complete

import (
	"fmt"
	"strings"
)

// GenerateNu generates a Nushell completion script.
//
// Nushell completions are declarative: every command is described by an `extern`
// signature whose flags and positionals carry optional `@"nu-complete …"`
// custom-completer attributes. Unlike the bash/zsh/fish/pwsh/elvish generators -
// each of which emits a single dispatcher that resolves the command path at
// runtime - this generator leans on Nushell's own parser to select the matching
// `extern`, and only emits helper closures for the cases Nushell cannot express
// declaratively: dynamic predictors (which shell out to `<app> --@complete=…`),
// comma-separated value lists, positional dynamic args, and forwarded context
// flags. File and directory value hints map to Nushell's native `path` and
// `directory` shapes.
func GenerateNu(g *Generator) (string, error) {
	if err := ValidateGenerator(g); err != nil {
		return "", err
	}

	var sb strings.Builder
	id := nuID(g.AppName)

	fmt.Fprintf(&sb, "# %s Nushell completion\n", g.AppName)

	if forwardingActive(g) {
		nuWriteForwardedHelper(&sb, id, allForwardableSpecs(g))
	}
	if hasDynamicArgs(g) {
		nuWritePositionalsHelper(&sb, id)
	}

	// Root command, then every subcommand depth-first. Inherited persistent
	// flags are merged into each subcommand's signature via combineVisibleSpecs.
	nuEmitCommand(
		&sb,
		g,
		[]string{g.AppName},
		nil,
		combineVisibleSpecs(g.Specs),
		false,
		g.DynamicArgs,
		g.HasMaxPositionalArgs,
		g.MaxPositionalArgs,
		1,
	)
	nuEmitSubs(
		&sb,
		g,
		[]string{g.AppName},
		g.Subs,
		persistentSpecs(g.Specs),
		2, //nolint:mnd // depth 2 = first subcommand level
	)

	return sb.String(), nil
}

// nuID munges an app name into a token usable in Nushell helper-command names,
// replacing every character outside [A-Za-z0-9_] with an underscore.
func nuID(name string) string {
	return strings.Map(func(r rune) rune {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '_':
			return r
		default:
			return '_'
		}
	}, name)
}

// nuQuote renders s as a Nushell string literal, folding whitespace so the
// result stays on one line. Single quotes (which Nushell does not let you escape
// inside a single-quoted string) force a double-quoted literal instead.
func nuQuote(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	if !strings.Contains(s, "'") {
		return "'" + s + "'"
	}
	s = strings.ReplaceAll(s, `\`, `\\`)
	s = strings.ReplaceAll(s, `"`, `\"`)
	return `"` + s + `"`
}

// nuName renders a command/completer name (always shell-safe) as a double-quoted
// literal, the conventional spelling for `extern` and custom-completer names.
func nuName(s string) string {
	return `"` + s + `"`
}

// nuComment folds a description to a single line for use after `#`.
func nuComment(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.ReplaceAll(s, "\r", " ")
	s = strings.ReplaceAll(s, "\t", " ")
	return strings.TrimSpace(s)
}

func nuQuotedList(values []string) string {
	quoted := make([]string, len(values))
	for i, v := range values {
		quoted[i] = nuQuote(v)
	}
	return strings.Join(quoted, " ")
}

func nuForwardedName(id string) string { return "_" + id + "_forwarded_flags" }
func nuPositionalsName(id string) string {
	return "_" + id + "_positionals"
}

// nuFlagKey returns the stable per-flag suffix used in completer names: the long
// flag when present, else the short flag.
func nuFlagKey(spec Spec) string {
	if spec.LongFlag != "" {
		return spec.LongFlag
	}
	return spec.ShortFlag
}

// nuCompleterName builds the scoped custom-completer name for a flag value or
// positional, e.g. "nu-complete myapp list include".
func nuCompleterName(names []string, key string) string {
	return "nu-complete " + strings.Join(names, " ") + " " + key
}

// nuExternName renders the space-joined command path as a quoted `extern` name.
func nuExternName(names []string) string {
	return nuName(strings.Join(names, " "))
}

// nuCompleteArg renders the dynamic-completion request flag as a quoted literal,
// e.g. "--@complete=author".
func nuCompleteArg(predictor string) string {
	return nuName("--" + FlagComplete + "=" + predictor)
}

// nuNeedsValueCompleter reports whether a flag's value is completed by a custom
// `nu-complete` closure (as opposed to a native shape or no completion at all).
func nuNeedsValueCompleter(spec Spec) bool {
	if !spec.HasArg {
		return false
	}
	return spec.Dynamic != "" || len(spec.Values) > 0 || len(spec.ValueDescs) > 0
}

// nuEmitSubs walks the subcommand tree, emitting each command's completers and
// `extern` with inherited persistent flags folded in.
func nuEmitSubs(
	sb *strings.Builder,
	g *Generator,
	parentNames []string,
	subs []SubSpec,
	inherited []Spec,
	depth int,
) {
	for _, sub := range SortSubSpecs(subs) {
		names := append(append([]string{}, parentNames...), sub.Name)

		var aliasPaths [][]string
		for _, alias := range sub.Aliases {
			aliasPaths = append(aliasPaths, append(append([]string{}, parentNames...), alias))
		}

		nuEmitCommand(
			sb,
			g,
			names,
			aliasPaths,
			combineVisibleSpecs(inherited, sub.Specs),
			sub.PathArgs,
			sub.DynamicArgs,
			sub.HasMaxPositionalArgs,
			sub.MaxPositionalArgs,
			depth,
		)

		if len(sub.Subs) == 0 {
			continue
		}
		next := appendSpecs(inherited, persistentSpecs(sub.Specs))
		nuEmitSubs(sb, g, names, sub.Subs, next, depth+1)
	}
}

// nuEmitCommand emits the custom-completer defs for a command's value-taking
// flags and positional dynamic args, followed by its `extern` signature. The
// signature is duplicated under each alias path; alias externs reuse the primary
// path's completer names.
func nuEmitCommand(
	sb *strings.Builder,
	g *Generator,
	names []string,
	aliasPaths [][]string,
	specs []Spec,
	pathArgs bool,
	dynamicArgs []string,
	hasMax bool,
	maxPos int,
	depth int,
) {
	for _, spec := range specs {
		if nuNeedsValueCompleter(spec) {
			nuWriteValueCompleter(sb, g, names, spec)
		}
	}
	if len(dynamicArgs) > 0 {
		nuWriteArgsCompleter(sb, g, names, specs, dynamicArgs, hasMax, maxPos, depth)
	}

	nuWriteExtern(sb, names, names, specs, pathArgs, dynamicArgs)
	for _, ap := range aliasPaths {
		nuWriteExtern(sb, ap, names, specs, pathArgs, dynamicArgs)
	}
}

// nuWriteExtern emits a single `extern` signature. externNames is the command
// path the signature is declared under; completerNames is the primary path whose
// scoped completer names the flags reference (the two differ only for aliases).
func nuWriteExtern(
	sb *strings.Builder,
	externNames []string,
	completerNames []string,
	specs []Spec,
	pathArgs bool,
	dynamicArgs []string,
) {
	fmt.Fprintf(sb, "\nextern %s [\n", nuExternName(externNames))
	for _, spec := range specs {
		fmt.Fprintf(sb, "    %s\n", nuFlagLine(completerNames, spec))
	}
	switch {
	case len(dynamicArgs) > 0:
		fmt.Fprintf(sb, "    ...rest: string@%s\n", nuName(nuCompleterName(completerNames, "args")))
	case pathArgs:
		sb.WriteString("    ...rest: path\n")
	}
	sb.WriteString("]\n")
}

// nuFlagLine renders one flag's `extern` parameter, e.g.
// "--include(-i): string@\"nu-complete myapp include\"  # Fields to include".
func nuFlagLine(completerNames []string, spec Spec) string {
	line := nuFlagSpelling(spec) + nuTypeAnnotation(completerNames, spec)
	if spec.Terse != "" {
		line += "  # " + nuComment(spec.Terse)
	}
	return line
}

// nuFlagSpelling renders the flag name(s): "--long(-s)", "--long", or "-s".
func nuFlagSpelling(spec Spec) string {
	switch {
	case spec.LongFlag != "" && spec.ShortFlag != "":
		return fmt.Sprintf("--%s(-%s)", spec.LongFlag, spec.ShortFlag)
	case spec.LongFlag != "":
		return "--" + spec.LongFlag
	default:
		return "-" + spec.ShortFlag
	}
}

// nuTypeAnnotation renders the value-shape annotation for a flag. Boolean flags
// get none; custom-completed flags get `: string@"…"`; file/extension and
// directory hints map to Nushell's native `path` and `directory` shapes; every
// other value flag is a free-form `: string`.
func nuTypeAnnotation(completerNames []string, spec Spec) string {
	if !spec.HasArg {
		return ""
	}
	if nuNeedsValueCompleter(spec) {
		return ": string@" + nuName(nuCompleterName(completerNames, nuFlagKey(spec)))
	}
	switch {
	case spec.Extension != "" || spec.ValueHint == HintFile:
		return ": path"
	case spec.ValueHint == HintDir:
		return ": directory"
	default:
		return ": string"
	}
}

// nuWriteValueCompleter emits the `nu-complete` closure for a single flag value.
func nuWriteValueCompleter(sb *strings.Builder, g *Generator, names []string, spec Spec) {
	name := nuCompleterName(names, nuFlagKey(spec))
	switch {
	case spec.CommaList && spec.Dynamic != "":
		nuWriteCommaDynamic(sb, g, name, spec.Dynamic)
	case spec.CommaList && len(spec.ValueDescs) > 0:
		nuWriteCommaStaticDescs(sb, name, spec.ValueDescs)
	case spec.CommaList && len(spec.Values) > 0:
		nuWriteCommaStatic(sb, name, spec.Values)
	case spec.Dynamic != "":
		nuWriteDynamic(sb, g, name, spec.Dynamic)
	case len(spec.ValueDescs) > 0:
		nuWriteStaticDescs(sb, name, spec.ValueDescs)
	case len(spec.Values) > 0:
		nuWriteStatic(sb, name, spec.Values)
	}
}

func nuWriteStatic(sb *strings.Builder, name string, values []string) {
	fmt.Fprintf(sb, "\ndef %s [] {\n", nuName(name))
	fmt.Fprintf(sb, "    [%s]\n}\n", nuQuotedList(values))
}

func nuWriteStaticDescs(sb *strings.Builder, name string, descs []ValueDesc) {
	fmt.Fprintf(sb, "\ndef %s [] {\n    [\n", nuName(name))
	for _, vd := range descs {
		fmt.Fprintf(
			sb,
			"        {value: %s, description: %s}\n",
			nuQuote(vd.Value),
			nuQuote(vd.Desc),
		)
	}
	sb.WriteString("    ]\n}\n")
}

// nuWriteDynamic emits a closure that shells out to the app's dynamic-completion
// handler. When forwarding is active it threads forwarded context flags through.
func nuWriteDynamic(sb *strings.Builder, g *Generator, name, predictor string) {
	if forwardingActive(g) {
		fmt.Fprintf(sb, "\ndef %s [context: string] {\n", nuName(name))
		fmt.Fprintf(
			sb,
			"    try { ^%s %s -- ...(%s $context) | lines } catch { [] }\n}\n",
			g.AppName,
			nuCompleteArg(predictor),
			nuForwardedName(nuID(g.AppName)),
		)
		return
	}
	fmt.Fprintf(sb, "\ndef %s [] {\n", nuName(name))
	fmt.Fprintf(
		sb,
		"    try { ^%s %s | lines } catch { [] }\n}\n",
		g.AppName,
		nuCompleteArg(predictor),
	)
}

// nuWriteCommaPrefix emits the shared comma-list preamble: it captures the
// prefix already typed (everything up to the final comma of the current token)
// and the set of values already chosen, so they can be skipped and re-prefixed.
func nuWriteCommaPrefix(sb *strings.Builder) {
	sb.WriteString(
		"    let tok = (if ($context | str ends-with \" \") { \"\" } else { $context | split row \" \" | last })\n",
	)
	sb.WriteString("    let prefix = ($tok | str replace -r '[^,]*$' '')\n")
	sb.WriteString("    let selected = ($prefix | split row \",\" | where {|x| $x != \"\" })\n")
}

func nuWriteCommaStatic(sb *strings.Builder, name string, values []string) {
	fmt.Fprintf(sb, "\ndef %s [context: string] {\n", nuName(name))
	nuWriteCommaPrefix(sb)
	fmt.Fprintf(
		sb,
		"    [%s] | where {|v| $v not-in $selected } | each {|v| $\"($prefix)($v)\" }\n}\n",
		nuQuotedList(values),
	)
}

func nuWriteCommaStaticDescs(sb *strings.Builder, name string, descs []ValueDesc) {
	fmt.Fprintf(sb, "\ndef %s [context: string] {\n", nuName(name))
	nuWriteCommaPrefix(sb)
	sb.WriteString("    [\n")
	for _, vd := range descs {
		fmt.Fprintf(
			sb,
			"        {value: %s, description: %s}\n",
			nuQuote(vd.Value),
			nuQuote(vd.Desc),
		)
	}
	sb.WriteString(
		"    ] | where {|r| $r.value not-in $selected } | each {|r| {value: $\"($prefix)($r.value)\", description: $r.description} }\n}\n",
	)
}

func nuWriteCommaDynamic(sb *strings.Builder, g *Generator, name, predictor string) {
	fmt.Fprintf(sb, "\ndef %s [context: string] {\n", nuName(name))
	nuWriteCommaPrefix(sb)
	call := fmt.Sprintf("^%s %s", g.AppName, nuCompleteArg(predictor))
	if forwardingActive(g) {
		call += fmt.Sprintf(" -- ...(%s $context)", nuForwardedName(nuID(g.AppName)))
	}
	fmt.Fprintf(
		sb,
		"    try { %s | lines } catch { [] } | where {|v| $v not-in $selected } | each {|v| $\"($prefix)($v)\" }\n}\n",
		call,
	)
}

// nuWriteArgsCompleter emits the positional dynamic-args closure. It counts the
// real positionals already on the line, selects the matching slot's kind, and
// shells out to the handler with forwarded context flags and preceding
// positionals, mirroring the bash/fish/pwsh/elvish positional protocol.
func nuWriteArgsCompleter(
	sb *strings.Builder,
	g *Generator,
	names []string,
	specs []Spec,
	dynamicArgs []string,
	hasMax bool,
	maxPos int,
	depth int,
) {
	id := nuID(g.AppName)
	forward := forwardingActive(g)
	cmdSkip := depth - 1

	fmt.Fprintf(sb, "\ndef %s [context: string] {\n", nuName(nuCompleterName(names, "args")))
	fmt.Fprintf(
		sb,
		"    let positional = (%s $context %d [%s])\n",
		nuPositionalsName(id),
		cmdSkip,
		nuValueFlagList(specs),
	)
	sb.WriteString("    let n = ($positional | length)\n")
	if hasMax {
		fmt.Fprintf(sb, "    if $n >= %d { return [] }\n", maxPos)
	}

	last := dynamicArgs[len(dynamicArgs)-1]
	sb.WriteString("    let kind = (")
	for i, da := range dynamicArgs {
		if i == 0 {
			fmt.Fprintf(sb, "if $n == %d { %s }", i, nuQuote(da))
		} else {
			fmt.Fprintf(sb, " else if $n == %d { %s }", i, nuQuote(da))
		}
	}
	fmt.Fprintf(sb, " else { %s })\n", nuQuote(last))
	sb.WriteString("    let first = ($n == 0)\n")

	if forward {
		fmt.Fprintf(sb, "    let fwd = (%s $context)\n", nuForwardedName(id))
		sb.WriteString("    let extra = (if $first { [] } else { $positional })\n")
		fmt.Fprintf(
			sb,
			"    try { ^%s $\"--%s=($kind)\" -- ...$fwd ...$extra | lines } catch { [] }\n}\n",
			g.AppName,
			FlagComplete,
		)
		return
	}
	sb.WriteString("    if $first {\n")
	fmt.Fprintf(
		sb,
		"        try { ^%s $\"--%s=($kind)\" | lines } catch { [] }\n",
		g.AppName,
		FlagComplete,
	)
	sb.WriteString("    } else {\n")
	fmt.Fprintf(
		sb,
		"        try { ^%s $\"--%s=($kind)\" -- ...$positional | lines } catch { [] }\n",
		g.AppName,
		FlagComplete,
	)
	sb.WriteString("    }\n}\n")
}

// nuValueFlagList renders the value-taking flags of specs as the body of a
// Nushell list literal (space-separated quoted entries), used by the positional
// scanner to skip a flag and its value.
func nuValueFlagList(specs []Spec) string {
	var tokens []string
	for _, spec := range specs {
		if !spec.HasArg {
			continue
		}
		if spec.LongFlag != "" {
			tokens = append(tokens, nuQuote("--"+spec.LongFlag))
		}
		if spec.ShortFlag != "" {
			tokens = append(tokens, nuQuote("-"+spec.ShortFlag))
		}
	}
	return strings.Join(tokens, " ")
}

// nuWriteForwardedHelper emits a helper that scans the line for forwardable
// context flags and returns them normalized as --name=value, stopping at "--".
func nuWriteForwardedHelper(sb *strings.Builder, id string, fwd []forwardSpec) {
	fmt.Fprintf(sb, "\ndef %s [context: string] {\n", nuName(nuForwardedName(id)))
	// Materialize the token list with [ ...(...) ]: nu 0.114's type inference
	// rejects a lazy `where`-closure stream consumed by the `for` loop below
	// ("can't convert to oneof<table, binary, list<any>>").
	sb.WriteString(
		"    let toks = [ ...($context | split row \" \" | where {|x| $x != \"\" } | skip 1) ]\n",
	)
	sb.WriteString("    mut out = []\n")
	sb.WriteString("    mut skipnext = false\n")
	sb.WriteString("    mut name = \"\"\n")
	sb.WriteString("    for t in $toks {\n")
	sb.WriteString("        if $skipnext {\n")
	sb.WriteString(
		"            if $name != \"\" { $out = ($out | append $\"--($name)=($t)\"); $name = \"\" }\n",
	)
	sb.WriteString("            $skipnext = false\n")
	sb.WriteString("            continue\n")
	sb.WriteString("        }\n")
	sb.WriteString("        if $t == \"--\" { break }\n")
	for _, f := range fwd {
		var conds []string
		if f.LongFlag != "" {
			conds = append(conds, fmt.Sprintf("$t == \"--%s\"", f.LongFlag))
		}
		if f.ShortFlag != "" {
			conds = append(conds, fmt.Sprintf("$t == \"-%s\"", f.ShortFlag))
		}
		fmt.Fprintf(
			sb,
			"        if %s { $skipnext = true; $name = %s; continue }\n",
			strings.Join(conds, " or "),
			nuQuote(f.LongFlag),
		)
		if f.LongFlag != "" {
			fmt.Fprintf(
				sb,
				"        if ($t | str starts-with \"--%s=\") { $out = ($out | append $t); continue }\n",
				f.LongFlag,
			)
		}
		if f.ShortFlag != "" && f.LongFlag != "" {
			prefixLen := len("-" + f.ShortFlag + "=")
			fmt.Fprintf(
				sb,
				"        if ($t | str starts-with \"-%s=\") { $out = ($out | append $\"--%s=($t | str substring %d..)\"); continue }\n",
				f.ShortFlag,
				f.LongFlag,
				prefixLen,
			)
		}
	}
	sb.WriteString("    }\n")
	sb.WriteString("    $out\n")
	sb.WriteString("}\n")
}

// nuWritePositionalsHelper emits a helper that extracts the real positional
// arguments from the line, skipping flags, their values, the leading subcommand
// tokens (cmdskip), and the in-progress token, while honoring the "--"
// terminator.
// nuPositionalsHelper materializes tokens with [ ...(...) ]: see
// nuWriteForwardedHelper - a lazy `where`-closure stream trips nu 0.114 type
// inference in the `for` loop.
const nuPositionalsHelper = `
def %s [context: string, cmdskip: int, valueflags: list<string>] {
    let trailing = ($context | str ends-with " ")
    let toks0 = [ ...($context | split row " " | where {|x| $x != "" } | skip 1) ]
    let toks = (if $trailing { $toks0 } else { $toks0 | drop 1 })
    mut pos = []
    mut skipnext = false
    mut dashdash = false
    mut skip = $cmdskip
    for t in $toks {
        if $dashdash { $pos = ($pos | append $t); continue }
        if $skipnext { $skipnext = false; continue }
        if $t == "--" { $dashdash = true; continue }
        if ($t in $valueflags) { $skipnext = true; continue }
        if ($t | str starts-with "-") { continue }
        if $skip > 0 { $skip = ($skip - 1); continue }
        $pos = ($pos | append $t)
    }
    $pos
}
`

func nuWritePositionalsHelper(sb *strings.Builder, id string) {
	fmt.Fprintf(sb, nuPositionalsHelper, nuName(nuPositionalsName(id)))
}
