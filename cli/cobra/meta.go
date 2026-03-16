package cobra

import (
	"strings"

	"github.com/gechr/clib/complete"
	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// FlagMeta extracts completion metadata from a cobra command's flags.
// It reads pflag properties and clib extras from all flags on the command.
func FlagMeta(cmd *cobralib.Command) []complete.FlagMeta {
	if cmd == nil {
		return nil
	}
	var flags []complete.FlagMeta
	seen := map[string]struct{}{}

	appendFlags := func(fs *pflag.FlagSet, persistent bool) {
		fs.VisitAll(func(f *pflag.Flag) {
			if _, ok := seen[f.Name]; ok {
				return
			}
			meta := pflagToMeta(f)
			meta.Persistent = persistent
			flags = append(flags, meta)
			seen[f.Name] = struct{}{}
		})
	}

	appendFlags(cmd.LocalNonPersistentFlags(), false)
	appendFlags(cmd.PersistentFlags(), true)

	return flags
}

func pflagToMeta(f *pflag.Flag) complete.FlagMeta {
	meta := complete.FlagMeta{
		Name:   f.Name,
		Short:  f.Shorthand,
		Help:   f.Usage,
		Hidden: f.Hidden,
		HasArg: f.Value.Type() != pflagTypeBool,
	}

	typeName := f.Value.Type()
	meta.IsSlice = strings.Contains(typeName, "Slice") || strings.Contains(typeName, "Array")

	if _, ok := f.Value.(*CSVFlag); ok {
		meta.IsCSV = true
	}

	if extra := getExtra(f); extra != nil {
		meta.Aliases = extra.Aliases
		meta.Complete = extra.Complete
		meta.Enum = extra.Enum
		meta.EnumDefault = extra.EnumDefault
		meta.EnumHighlight = extra.EnumHighlight
		meta.Extension = extra.Extension
		meta.Group = extra.Group
		meta.ValueHint = extra.Hint
		meta.Negatable = extra.Negatable
		meta.NegativeDesc = extra.NegativeDesc
		meta.Placeholder = extra.Placeholder
		meta.PlaceholderOverride = extra.Placeholder != ""
		meta.PositiveDesc = extra.PositiveDesc
		meta.Terse = extra.Terse
	}

	// Fall back to pflag's default value for enum highlighting.
	if meta.EnumDefault == "" && len(meta.Enum) > 0 && f.DefValue != "" {
		meta.EnumDefault = f.DefValue
	}

	return meta
}
