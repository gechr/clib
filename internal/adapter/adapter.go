// Package adapter holds helpers shared by the cli/* framework adapters.
package adapter

import (
	"strings"

	"github.com/gechr/clib/help"
)

// FlagRefs flattens rendered flag sections into non-rendering reference
// metadata. Adapters use this when a dispatcher command intentionally hides
// its option sections but descriptions may still refer to those flags.
func FlagRefs(sections []help.Section) help.FlagRefs {
	var refs help.FlagRefs
	for _, section := range sections {
		for _, content := range section.Content {
			if flags, ok := content.(help.FlagGroup); ok {
				refs = append(refs, flags...)
			}
		}
	}
	return refs
}

// ApplyLongHelp applies opts to sections along with the default long-help
// policy: unless overridden via [help.WithAlwaysShowDescription] or
// [help.WithAlwaysShowExamples], the description blurb and examples are hidden
// on -h and shown on --help.
func ApplyLongHelp(sections []help.Section, args []string, opts ...help.Option) []help.Section {
	behavior := help.ResolvePolicy(opts...)
	sections = help.Apply(sections, opts...)
	if !behavior.AlwaysShowDescription {
		sections = help.Apply(sections, help.WithDescriptionOnLongHelp(args))
	}
	if !behavior.AlwaysShowExamples {
		sections = help.Apply(sections, help.WithExamplesOnLongHelp(args))
	}
	return sections
}

// NegatableLong returns the advertised long-name spelling of a negatable flag.
// A bare positive/negative tag advertises just that variant; the flag stays
// negatable either way, so the hidden spelling still parses and completes.
func NegatableLong(name, prefix string, positiveOnly, negativeOnly bool) string {
	switch {
	case positiveOnly:
		return name
	case negativeOnly:
		return prefix + name
	default:
		return "[" + prefix + "]" + name
	}
}

// ApplyFlagVisibility applies the clib hide-long/hide-short/no-indent extras
// to a mapped help flag.
func ApplyFlagVisibility(f *help.Flag, hideLong, hideShort, noIndent bool) {
	if hideLong {
		f.Long = ""
	}
	if hideShort {
		f.Short = ""
	}
	if noIndent {
		f.NoIndent = true
	}
}

// NormalizePlaceholder lowercases an explicit placeholder for consistency with
// clib's help style, unless the adapter was told to preserve it.
func NormalizePlaceholder(placeholder string, lowercase bool) string {
	if !lowercase {
		return placeholder
	}
	return strings.ToLower(placeholder)
}
