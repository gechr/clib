package kong

import (
	"reflect"

	"github.com/gechr/clib/complete"
	"github.com/gechr/clib/internal/tag"
)

// fieldNameToFlag converts a Go CamelCase field name to a kebab-case flag name,
// matching kong's auto-derivation convention.
//
//	Config     → config
//	NoConfig   → no-config
//	DryRun     → dry-run
//	HTMLParser → html-parser
func fieldNameToFlag(name string) string {
	buf := make([]byte, 0, len(name)*2) //nolint:mnd // rough upper bound for kebab-case expansion
	for i := range len(name) {
		c := name[i]
		if c < 'A' || c > 'Z' {
			buf = append(buf, c)
			continue
		}
		if i > 0 {
			buf = appendHyphenIfNeeded(buf, name, i)
		}
		buf = append(buf, c+('a'-'A'))
	}
	return string(buf)
}

// appendHyphenIfNeeded inserts a '-' before an uppercase letter when the
// previous char is lowercase/digit, or when it's an uppercase-to-lowercase
// transition (e.g. "HTML" + "P" → "html-p").
func appendHyphenIfNeeded(buf []byte, name string, i int) []byte {
	prev := name[i-1]
	switch {
	case prev >= 'a' && prev <= 'z',
		prev >= '0' && prev <= '9':
		buf = append(buf, '-')
	case prev >= 'A' && prev <= 'Z' && i+1 < len(name) && name[i+1] >= 'a' && name[i+1] <= 'z':
		buf = append(buf, '-')
	}
	return buf
}

// Reflect extracts flag metadata from a CLI struct by reading Kong-style struct tags.
// It reads the bare-tag format: name:"foo" help:"..." short:"x" etc.
func Reflect(cli any) ([]complete.FlagMeta, error) {
	v := reflect.ValueOf(cli)
	if v.Kind() == reflect.Pointer {
		if v.IsNil() {
			return nil, nil
		}
		v = v.Elem()
	}
	t := v.Type()
	if t.Kind() != reflect.Struct {
		return nil, nil
	}

	return inspectStruct(t)
}

func inspectStruct(t reflect.Type) ([]complete.FlagMeta, error) {
	var flags []complete.FlagMeta

	csvType := reflect.TypeFor[CSVFlag]()
	completionFlagsType := reflect.TypeFor[CompletionFlags]()

	for field := range t.Fields() {
		if !field.IsExported() {
			continue
		}

		// Skip subcommand fields (cmd:"" tag).
		if _, ok := field.Tag.Lookup(tagCmd); ok {
			continue
		}

		// Handle embedded structs by recursing.
		if field.Anonymous {
			embedded, err := inspectEmbedded(field.Type, completionFlagsType)
			if err != nil {
				return nil, err
			}
			flags = append(flags, embedded...)
			continue
		}

		meta := complete.FlagMeta{
			Origin:     field.Name,
			Persistent: true,
		}

		meta.Name = field.Tag.Get(tagName)
		meta.Short = field.Tag.Get(tagShort)
		meta.Help = field.Tag.Get(tagHelp)
		meta.Placeholder = field.Tag.Get(tagPlaceholder)

		// clib-specific metadata: clib:"terse='...',complete='...',group='...'"
		if err := meta.ParseClibTag(field.Tag.Get(tagClib)); err != nil {
			return nil, err
		}

		// PlaceholderOverride: true if placeholder was set via either kong's
		// native placeholder:"" tag or the clib:"placeholder='...'" tag.
		// ParseClibTag already sets it for the clib path; cover the native path.
		if !meta.PlaceholderOverride && meta.Placeholder != "" {
			meta.PlaceholderOverride = true
		}

		if _, ok := field.Tag.Lookup(tagNegatable); ok {
			meta.Negatable = true
		}

		if _, ok := field.Tag.Lookup(tagHidden); ok {
			meta.Hidden = true
		}

		if _, ok := field.Tag.Lookup(tagArg); ok {
			meta.IsArg = true
			meta.Persistent = false
		}

		if _, ok := field.Tag.Lookup(tagOptional); ok {
			meta.Optional = true
		}

		if len(meta.Enum) == 0 {
			if enum := field.Tag.Get(tagEnum); enum != "" {
				meta.Enum = tag.SplitCSV(enum)
			}
		}

		// Fall back to kong's native default for enum highlighting.
		if meta.EnumDefault == "" && len(meta.Enum) > 0 {
			if def, ok := field.Tag.Lookup(tagDefault); ok && def != "" {
				meta.EnumDefault = def
			}
		}

		if aliases := field.Tag.Get(tagAliases); aliases != "" {
			meta.Aliases = tag.SplitCSV(aliases)
		}

		// Determine HasArg from field type.
		fieldType := field.Type
		if fieldType.Kind() == reflect.Pointer {
			fieldType = fieldType.Elem()
		}
		meta.HasArg = fieldType.Kind() != reflect.Bool
		meta.IsSlice = fieldType.Kind() == reflect.Slice

		// Determine IsCSV.
		if fieldType == csvType {
			meta.IsCSV = true
		}

		// Auto-derive flag name from field name when not explicitly tagged,
		// matching kong's convention (CamelCase → kebab-case).
		if meta.Name == "" && !meta.IsArg {
			meta.Name = fieldNameToFlag(field.Name)
		}

		if err := meta.Validate(); err != nil {
			return nil, err
		}
		flags = append(flags, meta)
	}

	return flags, nil
}

// inspectEmbedded recurses into an embedded struct type, skipping the
// CompletionFlags type which is internal infrastructure.
func inspectEmbedded(ft reflect.Type, skip reflect.Type) ([]complete.FlagMeta, error) {
	if ft.Kind() == reflect.Pointer {
		ft = ft.Elem()
	}
	if ft == skip || ft.Kind() != reflect.Struct {
		return nil, nil
	}
	return inspectStruct(ft)
}
