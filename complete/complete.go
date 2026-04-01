package complete

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"unicode"
	"unicode/utf8"

	"github.com/gechr/clib/shell"
)

// Handler is called when a dynamic completion is requested.
// It receives the completion type, the detected shell name,
// and any preceding positional args passed from the shell.
type Handler func(shell, kind string, args []string)

// Action describes which completion action was requested.
type Action struct {
	Shell               string   // resolved shell name
	Complete            string   // dynamic completion type (--@complete value)
	Args                []string // preceding positional args from the shell
	InstallCompletion   bool
	UninstallCompletion bool
	PrintCompletion     bool
}

// HandleAction dispatches the given completion action against gen.
// Returns true if an action was handled (caller should exit).
func HandleAction(a Action, gen *Generator, handler Handler, quiet bool) (bool, error) {
	if a.Complete != "" {
		if handler != nil {
			handler(a.Shell, a.Complete, a.Args)
		}
		return true, nil
	}
	if a.InstallCompletion {
		return true, gen.Install(a.Shell, quiet)
	}
	if a.UninstallCompletion {
		return true, gen.Uninstall(a.Shell, quiet)
	}
	if a.PrintCompletion {
		return true, gen.Print(os.Stdout, a.Shell)
	}
	return false, nil
}

// ApplyMeta populates spec fields from a FlagMeta's completion-related
// annotations (Complete, Extension, ValueHint, Terse, Enum).
func ApplyMeta(spec *Spec, meta *FlagMeta) {
	if meta.Terse != "" {
		spec.Terse = meta.Terse
	}
	if len(meta.Enum) > 0 {
		if len(meta.EnumTerse) == len(meta.Enum) {
			spec.ValueDescs = make([]ValueDesc, len(meta.Enum))
			for i, v := range meta.Enum {
				spec.ValueDescs[i] = ValueDesc{Value: v, Desc: meta.EnumTerse[i]}
			}
		} else {
			spec.Values = meta.Enum
		}
		if meta.IsSlice || meta.IsCSV {
			spec.CommaList = true
		}
	}
	if meta.Extension != "" {
		spec.Extension = meta.Extension
	}
	if meta.ValueHint != "" {
		spec.ValueHint = meta.ValueHint
	}
	if meta.Complete != "" {
		predictor, commaList, staticValues := ParseCompleteTag(meta.Complete)
		if predictor != "" {
			spec.Dynamic = predictor
		}
		spec.CommaList = commaList
		if len(staticValues) > 0 {
			spec.Values = staticValues
		}
	}
}

// Spec describes a single flag for shell completion generation.
type Spec struct {
	CommaList  bool        // comma-separated multi-value (e.g. --columns)
	Dynamic    string      // dynamic completion type (e.g. "author" -> "<app> --@complete=author")
	Extension  string      // file extension filter for completion (e.g. "yaml" or "yaml,yml")
	HasArg     bool        // flag takes a value
	Hidden     bool        // hidden from completions
	LongFlag   string      // e.g. "author" (no dashes)
	Persistent bool        // true if the flag remains available on descendant subcommands
	ShortFlag  string      // e.g. "a" (no dash)
	Terse      string      // very short description for tab completion
	ValueDescs []ValueDesc // static values with descriptions (takes precedence over Values)
	ValueHint  string      // value type hint: file, dir, command, user, host, url, email
	Values     []string    // static completion values (from enum)
}

// SubSpec describes a subcommand for shell completion generation.
type SubSpec struct {
	Name                 string    // subcommand name (e.g. "bump")
	Aliases              []string  // command aliases (e.g. ["up"] for "update")
	Terse                string    // short description for tab completion
	Specs                []Spec    // subcommand-specific flag specs
	Subs                 []SubSpec // nested subcommands
	PathArgs             bool      // enable file completion for positional args
	DynamicArgs          []string  // per-position dynamic completion; final entry repeats for additional positional args
	MaxPositionalArgs    int
	HasMaxPositionalArgs bool
}

// Value hint constants for completion.
const (
	HintFile    = "file"
	HintDir     = "dir"
	HintCommand = "command"
	HintUser    = "user"
	HintHost    = "host"
	HintURL     = "url"
	HintEmail   = "email"
)

// ValueDesc pairs a completion value with an optional description.
type ValueDesc struct {
	Value string
	Desc  string
}

// Generator generates shell completion scripts.
type Generator struct {
	AppName              string
	DynamicArgs          []string // per-position dynamic completion; final entry repeats for additional positional args
	Specs                []Spec
	Subs                 []SubSpec
	MaxPositionalArgs    int
	HasMaxPositionalArgs bool
}

// NewGenerator creates a Generator for the named application.
func NewGenerator(command string) *Generator {
	return &Generator{AppName: command}
}

// FromFlags populates completion specs from pre-inspected flag metadata.
func (g *Generator) FromFlags(flags []FlagMeta) *Generator {
	for _, f := range flags {
		g.Specs = append(g.Specs, SpecsFromFlagMeta(f)...)
	}
	return g
}

// SpecsFromFlagMeta expands a single FlagMeta into completion specs,
// including negated variants where applicable.
func SpecsFromFlagMeta(f FlagMeta) []Spec {
	if f.IsArg || f.Complete == "-" {
		return nil
	}

	longFlag := f.Name
	shortFlag := f.Short
	if f.HideLong {
		longFlag = ""
	}
	if f.HideShort {
		shortFlag = ""
	}

	spec := Spec{
		LongFlag:   longFlag,
		ShortFlag:  shortFlag,
		Terse:      f.Desc(),
		HasArg:     f.HasArg,
		Hidden:     f.Hidden,
		Extension:  f.Extension,
		Persistent: f.Persistent,
		ValueHint:  f.ValueHint,
	}
	ApplyMeta(&spec, &f)

	if f.Negatable && f.Name != "" {
		pos, neg := NegatableSpecs(spec, f.PositiveDesc, f.NegativeDesc, f.InversePrefix)
		return []Spec{pos, neg}
	}
	return []Spec{spec}
}

// negatableDescs returns the positive and negative descriptions for a negatable flag.
func negatableDescs(desc string) (string, string) {
	if desc == "" {
		return "", ""
	}
	for _, pair := range [][2]string{
		{"enable", "disable"},
		{"disable", "enable"},
		{"show", "hide"},
		{"hide", "show"},
	} {
		if after, ok := cutPrefixFold(desc, pair[0]+" "); ok {
			return desc, matchCase(desc[:len(pair[0])], pair[1]) + " " + after
		}
	}
	if after, ok := cutPrefixFold(desc, "toggle "); ok {
		return matchCase(desc[:6], "enable") + " " + after,
			matchCase(desc[:6], "disable") + " " + after
	}
	r, size := utf8.DecodeRuneInString(desc)
	return desc, "Disable " + string(unicode.ToLower(r)) + desc[size:]
}

// NegatableSpecs returns the positive and negative Spec pair for a negatable
// flag. It rewrites spec.Terse using automatic prefix detection (e.g.
// "Enable X" → "Disable X") and creates a --<inversePrefix><name> variant.
// An empty inversePrefix defaults to "no-".
// Explicit positiveDesc / negativeDesc override the auto-derived text.
func NegatableSpecs(spec Spec, positiveDesc, negativeDesc, inversePrefix string) (Spec, Spec) {
	if inversePrefix == "" {
		inversePrefix = "no-"
	}
	posDesc, negDesc := negatableDescs(spec.Terse)
	if positiveDesc != "" {
		posDesc = positiveDesc
	}
	if negativeDesc != "" {
		negDesc = negativeDesc
	}
	spec.Terse = posDesc
	negative := spec
	negative.CommaList = false
	negative.Dynamic = ""
	negative.Extension = ""
	negative.HasArg = false
	negative.LongFlag = inversePrefix + spec.LongFlag
	negative.ShortFlag = ""
	negative.Terse = negDesc
	negative.ValueDescs = nil
	negative.ValueHint = ""
	negative.Values = nil
	return spec, negative
}

// cutPrefixFold is like strings.CutPrefix but case-insensitive.
func cutPrefixFold(s, prefix string) (string, bool) {
	if len(s) < len(prefix) {
		return s, false
	}
	if strings.EqualFold(s[:len(prefix)], prefix) {
		return s[len(prefix):], true
	}
	return s, false
}

// matchCase returns repl with the same casing pattern as orig.
func matchCase(orig, repl string) string {
	if orig == "" || repl == "" {
		return repl
	}
	if orig == strings.ToUpper(orig) {
		return strings.ToUpper(repl)
	}
	r, _ := utf8.DecodeRuneInString(orig)
	if unicode.IsUpper(r) {
		replR, replSize := utf8.DecodeRuneInString(repl)
		return string(unicode.ToUpper(replR)) + repl[replSize:]
	}
	return repl
}

// ParseCompleteTag parses the complete struct tag value.
// Format: comma-separated parts of "predictor=<name>", "comma", and/or "values=<space-separated values>".
func ParseCompleteTag(tag string) (string, bool, []string) {
	var predictor string
	var comma bool
	var values []string
	for part := range strings.SplitSeq(tag, ",") {
		part = strings.TrimSpace(part)
		if after, ok := strings.CutPrefix(part, "predictor="); ok {
			predictor = after
		} else if part == "comma" {
			comma = true
		} else if after, ok := strings.CutPrefix(part, "values="); ok {
			values = strings.Fields(after)
		}
	}
	return predictor, comma, values
}

// ShellFunc generates a completion script for a given shell.
type ShellFunc func(g *Generator) (string, error)

var shellRegistry = map[string]ShellFunc{}

// RegisterShell registers a shell completion generator.
// Shell subpackages call this from init().
func RegisterShell(name string, fn ShellFunc) {
	shellRegistry[name] = fn
}

func builtInShellFunc(name string) (ShellFunc, bool) {
	switch name {
	case shell.Bash:
		return GenerateBash, true
	case shell.Fish:
		return GenerateFish, true
	case shell.Zsh:
		return GenerateZsh, true
	default:
		return nil, false
	}
}

func resolveShellFunc(name string) (ShellFunc, bool) {
	if fn, ok := shellRegistry[name]; ok {
		return fn, true
	}
	return builtInShellFunc(name)
}

func supportedShells() string {
	names := []string{shell.Bash, shell.Zsh, shell.Fish}
	var custom []string
	for name := range shellRegistry {
		if !slices.Contains(names, name) {
			custom = append(custom, name)
		}
	}
	slices.Sort(custom)
	names = append(names, custom...)
	return strings.Join(names, ", ")
}

// SortVisibleSpecs returns non-hidden specs sorted by long flag name,
// falling back to short flag for short-only flags.
func SortVisibleSpecs(specs []Spec) []Spec {
	var result []Spec
	for _, s := range specs {
		if !s.Hidden {
			result = append(result, s)
		}
	}
	slices.SortStableFunc(result, func(a, b Spec) int {
		ak, bk := a.LongFlag, b.LongFlag
		if ak == "" {
			ak = a.ShortFlag
		}
		if bk == "" {
			bk = b.ShortFlag
		}
		return strings.Compare(ak, bk)
	})
	return result
}

func persistentSpecs(specs []Spec) []Spec {
	var result []Spec
	for _, spec := range specs {
		if spec.Persistent {
			result = append(result, spec)
		}
	}
	return result
}

func combineVisibleSpecs(specSets ...[]Spec) []Spec {
	total := 0
	for _, specs := range specSets {
		total += len(specs)
	}

	combined := make([]Spec, 0, total)
	for _, specs := range specSets {
		combined = append(combined, specs...)
	}
	return SortVisibleSpecs(combined)
}

func appendSpecs(specs ...[]Spec) []Spec {
	total := 0
	for _, group := range specs {
		total += len(group)
	}
	result := make([]Spec, 0, total)
	for _, group := range specs {
		result = append(result, group...)
	}
	return result
}

func argValuePatterns(specs []Spec) ([]string, []string) {
	var exact []string
	var equals []string
	for _, spec := range specs {
		if !spec.HasArg {
			continue
		}
		if spec.LongFlag != "" {
			exact = append(exact, "--"+spec.LongFlag)
			equals = append(equals, "--"+spec.LongFlag+"=*")
		}
		if spec.ShortFlag != "" {
			exact = append(exact, "-"+spec.ShortFlag)
			equals = append(equals, "-"+spec.ShortFlag+"=*")
		}
	}
	return exact, equals
}

// SortSubSpecs returns a copy of subs sorted by name.
func SortSubSpecs(subs []SubSpec) []SubSpec {
	sorted := make([]SubSpec, len(subs))
	copy(sorted, subs)
	slices.SortStableFunc(sorted, func(a, b SubSpec) int {
		return strings.Compare(a.Name, b.Name)
	})
	return sorted
}

// Print writes the completion script for the given shell to w.
// An empty shell defaults to fish.
func (g *Generator) Print(w io.Writer, sh string) error {
	if sh == "" {
		sh = shell.Fish
	}
	fn, ok := resolveShellFunc(sh)
	if !ok {
		return fmt.Errorf("unsupported shell %q (supported: %s)", sh, supportedShells())
	}
	script, err := fn(g)
	if err != nil {
		return err
	}
	_, err = io.WriteString(w, script)
	return err
}

// Install writes the completion script to the appropriate shell config directory.
// An empty shell defaults to fish.
func (g *Generator) Install(sh string, quiet bool) error {
	if sh == "" {
		sh = shell.Fish
	}

	var buf strings.Builder
	if err := g.Print(&buf, sh); err != nil {
		return err
	}

	path, err := g.completionFile(sh)
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("failed to create directory %s: %w", dir, err)
	}

	if err := os.WriteFile(path, []byte(buf.String()), 0o600); err != nil {
		return fmt.Errorf("failed to write %s: %w", path, err)
	}

	if !quiet {
		fmt.Fprintf(os.Stderr, "Completion installed to %s\n", path)
	}
	return nil
}

// Uninstall removes the completion script for the given shell.
// An empty shell defaults to fish.
func (g *Generator) Uninstall(sh string, quiet bool) error {
	if sh == "" {
		sh = shell.Fish
	}

	path, err := g.completionFile(sh)
	if err != nil {
		return err
	}

	if err := os.Remove(path); err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			if !quiet {
				fmt.Fprintf(os.Stderr, "No completion found (file %s does not exist)\n", path)
			}
			return nil
		}
		return fmt.Errorf("failed to remove %s: %w", path, err)
	}
	if !quiet {
		fmt.Fprintf(os.Stderr, "Completion removed from %s\n", path)
	}
	return nil
}

func (g *Generator) completionFile(sh string) (string, error) {
	return shell.CompletionFile(g.AppName, sh)
}
