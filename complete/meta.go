package complete

import (
	"fmt"
	"strings"

	"github.com/gechr/clib/internal/tag"
)

// FlagMeta holds metadata extracted from a single struct field's tags.
type FlagMeta struct {
	Aliases             []string // flag aliases
	Complete            string   // completion directive
	Enum                []string // enum values
	EnumDefault         string   // default value annotation for enum display
	EnumHighlight       []string // highlight hint substrings (parallel to Enum)
	EnumTerse           []string // short descriptions for enum values (parallel to Enum)
	Extension           string   // file extension filter for completion (e.g. "yaml" or "yaml,yml")
	Group               string   // help section group
	HasArg              bool     // true if the flag takes a value (non-bool)
	Help                string   // help text for --help output
	Hidden              bool     // hidden flag
	HideLong            bool     // hide the long flag from help output
	HideShort           bool     // hide the short flag from help output
	IsArg               bool     // true if this is a positional argument
	IsCSV               bool     // true if the field type is CSVFlag or *CSVFlag
	IsSlice             bool     // true if the field type is a slice
	Name                string   // flag name
	NoIndent            bool     // suppress short-flag alignment indent in help
	InversePrefix       string   // prefix for negated flag (default "no-")
	Negatable           bool     // true if the flag supports --no- prefix
	NegativeDesc        string   // explicit description for --no- variant
	Optional            bool     // true if arg is optional
	Origin              string   // where this metadata came from (e.g. struct field name)
	Placeholder         string   // placeholder like <value>
	PlaceholderOverride bool     // true if placeholder was set via clib tag
	Persistent          bool     // true if the flag remains available on descendant subcommands
	PositiveDesc        string   // explicit description for positive variant (negatable flags)
	Short               string   // short flag letter
	Terse               string   // very short description for completions
	ValueHint           string   // value type hint for completion (file, dir, command, user, host, url, email)
}

// Desc returns the Terse description for use in completions.
// Falls back to Help if Terse is empty.
func (f *FlagMeta) Desc() string {
	if f.Terse != "" {
		return f.Terse
	}
	return f.Help
}

// ParseClibTag parses a clib:"..." struct tag value into meta.
// These are clib-specific annotations that supplement what the CLI
// framework provider (kong, cobra, etc.) already supplies.
//
// Format: comma-separated entries, values optionally single-quoted.
//
//	clib:"terse='Draft filter',complete='predictor=repo',group='filters'"
//
// Supported keys: complete, enum, group, inverse, negatable, negative, placeholder, positive, terse.
func (f *FlagMeta) ParseClibTag(t string) error {
	if t == "" {
		return nil
	}
	parts, err := tag.Split(t)
	if err != nil {
		return err
	}
	for _, entry := range parts {
		key, val, _ := strings.Cut(entry, "=")
		if key == "" {
			return fmt.Errorf("empty key in tag %q", t)
		}
		val = strings.TrimPrefix(val, "'")
		val = strings.TrimSuffix(val, "'")
		switch key {
		case tag.Complete:
			f.Complete = val
		case tag.Default:
			f.EnumDefault = val
		case tag.Ext:
			f.Extension = val
		case tag.Group:
			f.Group = val
		case tag.HideLong:
			f.HideLong = true
		case tag.HideShort:
			f.HideShort = true
		case tag.NoIndent:
			f.NoIndent = true
		case tag.Highlight:
			if val != "" {
				f.EnumHighlight = tag.SplitCSV(val)
			}
		case tag.Inverse:
			f.InversePrefix = val
		case tag.Negatable:
			f.Negatable = true
		case tag.Negative:
			f.NegativeDesc = val
		case tag.Placeholder:
			f.Placeholder = val
			f.PlaceholderOverride = true
		case tag.Positive:
			f.PositiveDesc = val
		case tag.Enum:
			if val != "" {
				f.Enum = tag.SplitCSV(val)
			}
		case tag.EnumTerse:
			if val != "" {
				f.EnumTerse = tag.SplitCSV(val)
			}
		case tag.Terse:
			f.Terse = val
		case tag.Hint:
			f.ValueHint = val
		default:
			return fmt.Errorf("unknown tag key %q in %q", key, t)
		}
	}

	return f.validateTagOnly()
}

// validateTagOnly checks constraints that are fully determined by clib tag keys alone.
func (f *FlagMeta) validateTagOnly() error {
	if f.HideLong && f.HideShort {
		return fmt.Errorf("%s: hide-long and hide-short are mutually exclusive", f.flagLabel())
	}
	return nil
}

// Validate checks constraints that depend on the fully populated FlagMeta,
// including fields set by the CLI framework (e.g. negatable from kong's tag).
// Callers should invoke this after all metadata is populated.
func (f *FlagMeta) Validate() error {
	p := f.flagLabel() + ": "
	if f.HideLong && f.Short == "" {
		return fmt.Errorf("%shide-long has no effect without a short flag", p)
	}
	if f.HideShort && f.Name == "" {
		return fmt.Errorf("%shide-short has no effect without a long flag", p)
	}
	if f.NoIndent && f.Short != "" {
		return fmt.Errorf("%sno-indent has no effect on flags with a short form", p)
	}
	if f.NegativeDesc != "" && !f.Negatable {
		return fmt.Errorf("%snegative requires negatable", p)
	}
	if f.PositiveDesc != "" && !f.Negatable {
		return fmt.Errorf("%spositive requires negatable", p)
	}
	if f.InversePrefix != "" && !f.Negatable {
		return fmt.Errorf("%sinverse requires negatable", p)
	}
	if len(f.EnumHighlight) > 0 && len(f.Enum) == 0 {
		return fmt.Errorf("%shighlight requires enum", p)
	}
	if len(f.EnumTerse) > 0 && len(f.EnumTerse) != len(f.Enum) {
		return fmt.Errorf(
			"%senum-terse length (%d) must match enum length (%d)",
			p,
			len(f.EnumTerse),
			len(f.Enum),
		)
	}
	if f.EnumDefault != "" && len(f.Enum) == 0 {
		return fmt.Errorf("%sdefault requires enum", p)
	}
	return nil
}

func (f *FlagMeta) flagLabel() string {
	flag := f.Name
	if flag == "" {
		flag = f.Short
	}
	if flag == "" {
		flag = f.Origin
	}
	if f.Name == "" {
		return "-" + flag
	}
	return "--" + flag
}
