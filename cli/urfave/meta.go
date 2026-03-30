package urfave

import (
	"github.com/gechr/clib/complete"
	clilib "github.com/urfave/cli/v3"
)

// FlagMeta extracts completion metadata from a urfave command's flags.
// It reads flag properties via urfave interfaces and clib extras.
func FlagMeta(cmd *clilib.Command) []complete.FlagMeta {
	if cmd == nil {
		return nil
	}
	prepareFlagExtras(cmd)

	var flags []complete.FlagMeta
	for _, f := range cmd.Flags {
		flags = append(flags, flagToMeta(cmd, f))
	}
	return flags
}

func flagToMeta(cmd *clilib.Command, f clilib.Flag) complete.FlagMeta {
	meta := complete.FlagMeta{
		Persistent: isPersistentFlag(f),
	}

	// Name: first multi-char entry; Short: first single-char entry; Aliases: rest.
	if bif, ok := f.(*clilib.BoolWithInverseFlag); ok {
		populateNegatableNames(&meta, bif)
	} else {
		populateNames(&meta, f.Names())
	}

	// Help from DocGenerationFlag.
	if df, ok := f.(clilib.DocGenerationFlag); ok {
		meta.Help = df.GetUsage()
		meta.HasArg = df.TakesValue()
	}

	// Hidden from VisibleFlag.
	if vf, ok := f.(clilib.VisibleFlag); ok {
		meta.Hidden = !vf.IsVisible()
	}

	// Group: clib extra overrides urfave category.
	extra := getFlagExtra(cmd, f)
	if extra != nil && extra.Group != "" {
		meta.Group = extra.Group
	} else if cf, ok := f.(clilib.CategorizableFlag); ok {
		meta.Group = cf.GetCategory()
	}

	// IsSlice from DocGenerationMultiValueFlag.
	if mv, ok := f.(clilib.DocGenerationMultiValueFlag); ok {
		meta.IsSlice = mv.IsMultiValueFlag()
	}

	// IsCSV: check if GenericFlag with CSVFlag value.
	if gf, ok := f.(*clilib.GenericFlag); ok {
		if _, isCSV := gf.Value.(*CSVFlag); isCSV {
			meta.IsCSV = true
		}
	}

	// Clib extras.
	if extra != nil {
		meta.Complete = extra.Complete
		meta.Enum = extra.Enum
		meta.EnumDefault = extra.EnumDefault
		meta.EnumHighlight = extra.EnumHighlight
		meta.EnumTerse = extra.EnumTerse
		meta.Extension = extra.Extension
		meta.ValueHint = extra.Hint
		meta.NegativeDesc = extra.NegativeDesc
		meta.Placeholder = extra.Placeholder
		meta.PlaceholderOverride = extra.Placeholder != ""
		meta.PositiveDesc = extra.PositiveDesc
		meta.Terse = extra.Terse
	}

	// Fall back to urfave's default value for enum highlighting.
	if meta.EnumDefault == "" && len(meta.Enum) > 0 {
		if df, ok := f.(clilib.DocGenerationFlag); ok {
			if def := df.GetDefaultText(); def != "" {
				meta.EnumDefault = def
			}
		}
	}

	return meta
}

func isPersistentFlag(f clilib.Flag) bool {
	if lf, ok := f.(clilib.LocalFlag); ok {
		return !lf.IsLocal()
	}
	return true
}

func populateNegatableNames(meta *complete.FlagMeta, bif *clilib.BoolWithInverseFlag) {
	meta.Name = bif.Name
	meta.Negatable = true
	meta.InversePrefix = bif.InversePrefix
	if meta.InversePrefix == "" {
		meta.InversePrefix = clilib.DefaultInverseBoolPrefix
	}
	// Filter inverse names from aliases.
	for _, n := range bif.Aliases {
		if len(n) == 1 && meta.Short == "" {
			meta.Short = n
		} else {
			meta.Aliases = append(meta.Aliases, n)
		}
	}
}

func populateNames(meta *complete.FlagMeta, names []string) {
	for _, n := range names {
		switch {
		case meta.Name == "" && len(n) > 1:
			meta.Name = n
		case meta.Short == "" && len(n) == 1:
			meta.Short = n
		default:
			meta.Aliases = append(meta.Aliases, n)
		}
	}
	// If no multi-char name found, use first name.
	if meta.Name == "" && len(names) > 0 {
		meta.Name = names[0]
	}
}
