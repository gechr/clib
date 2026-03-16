package tag

import "strings"

// Clib tag keys.
const (
	Complete    = "complete"
	Default     = "default"
	Enum        = "enum"
	Ext         = "ext"
	Group       = "group"
	Inverse     = "inverse"
	Highlight   = "highlight"
	Hint        = "hint"
	Negatable   = "negatable"
	Negative    = "negative"
	Placeholder = "placeholder"
	Positive    = "positive"
	Terse       = "terse"
)

// Parse extracts the value for key from a clib tag string.
// Returns the unquoted value and true if found, or "" and false otherwise.
// Bare keys (no '=') return "" and true.
//
// Format: comma-separated entries, values optionally single-quoted:
//
//	"negatable,group='Filters',placeholder='repo'"
func Parse(s, key string) (string, bool) {
	for _, entry := range Split(s) {
		k, v, hasEq := strings.Cut(entry, "=")
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k != key {
			continue
		}
		if !hasEq {
			return "", true
		}
		v = strings.TrimPrefix(v, "'")
		v = strings.TrimSuffix(v, "'")
		return v, true
	}
	return "", false
}

// Split splits a clib tag on commas, respecting single-quoted values.
func Split(s string) []string {
	var parts []string
	var buf strings.Builder
	inQuote := false
	for _, r := range s {
		switch {
		case r == '\'':
			inQuote = !inQuote
			buf.WriteRune(r)
		case r == ',' && !inQuote:
			parts = append(parts, strings.TrimSpace(buf.String()))
			buf.Reset()
		default:
			buf.WriteRune(r)
		}
	}
	if buf.Len() > 0 {
		parts = append(parts, strings.TrimSpace(buf.String()))
	}
	return parts
}

// SplitCSV splits s on commas, trims whitespace from each element,
// and returns the resulting slice. Returns nil for empty input.
func SplitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	for i, p := range parts {
		parts[i] = strings.TrimSpace(p)
	}
	return parts
}
