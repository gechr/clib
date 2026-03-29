package complete

import (
	"fmt"
	"regexp"
	"strings"
)

// shellSafeRe matches strings containing only shell-safe characters:
// alphanumeric, hyphens, underscores, dots, and @.
// These are safe for interpolation into generated shell scripts as command
// names, flag names, and completion type identifiers.
var shellSafeRe = regexp.MustCompile(`^[a-zA-Z0-9._@-]+$`)

// ValidateShellSafe checks that s contains only shell-safe characters
// (alphanumeric, hyphens, underscores, dots, and @). It returns an error
// if s is empty or contains unsafe characters.
func ValidateShellSafe(s, label string) error {
	if s == "" {
		return fmt.Errorf("%s must not be empty", label)
	}
	if !shellSafeRe.MatchString(s) {
		return fmt.Errorf("%s contains unsafe characters: %q", label, s)
	}
	return nil
}

func validateExtensionList(ext string) error {
	for part := range strings.SplitSeq(ext, ",") {
		part = strings.TrimSpace(part)
		if err := ValidateShellSafe(part, "Extension"); err != nil {
			return err
		}
	}
	return nil
}

// ValidateSpecs validates shell-sensitive fields in specs.
func ValidateSpecs(specs []Spec) error {
	for _, spec := range specs {
		if spec.Dynamic != "" {
			if err := ValidateShellSafe(spec.Dynamic, "Dynamic"); err != nil {
				return err
			}
		}
		if spec.LongFlag != "" {
			if err := ValidateShellSafe(spec.LongFlag, "LongFlag"); err != nil {
				return err
			}
		}
		if spec.ShortFlag != "" {
			if err := ValidateShellSafe(spec.ShortFlag, "ShortFlag"); err != nil {
				return err
			}
		}
		if spec.Extension != "" {
			if err := validateExtensionList(spec.Extension); err != nil {
				return err
			}
		}
	}
	return nil
}

// ValidateSubs recursively validates shell-sensitive subcommand fields.
func ValidateSubs(subs []SubSpec) error {
	for _, sub := range subs {
		if err := ValidateShellSafe(sub.Name, "SubcommandName"); err != nil {
			return err
		}
		for _, alias := range sub.Aliases {
			if err := ValidateShellSafe(alias, "SubcommandAlias"); err != nil {
				return err
			}
		}
		for _, da := range sub.DynamicArgs {
			if err := ValidateShellSafe(da, "DynamicArgs"); err != nil {
				return err
			}
		}
		if err := ValidateSpecs(sub.Specs); err != nil {
			return err
		}
		if err := ValidateSubs(sub.Subs); err != nil {
			return err
		}
	}
	return nil
}

// ValidateGenerator validates shell-sensitive fields in g.
func ValidateGenerator(g *Generator) error {
	if err := ValidateShellSafe(g.AppName, "AppName"); err != nil {
		return err
	}
	for _, da := range g.DynamicArgs {
		if err := ValidateShellSafe(da, "DynamicArgs"); err != nil {
			return err
		}
	}
	if err := ValidateSpecs(g.Specs); err != nil {
		return err
	}
	return ValidateSubs(g.Subs)
}

// WriteIndented writes each non-empty line of block to sb, prefixed with indent.
func WriteIndented(sb *strings.Builder, indent, block string) {
	for line := range strings.SplitSeq(block, "\n") {
		if line == "" {
			continue
		}
		fmt.Fprintf(sb, "%s%s\n", indent, line)
	}
}
