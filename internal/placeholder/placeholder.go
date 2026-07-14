package placeholder

import "github.com/gechr/x/set"

const (
	placeholderWhen = "when"

	valueAlways = "always"
	valueAuto   = "auto"
	valueNever  = "never"
)

// ForEnum returns the conventional placeholder for enums containing the
// standard automatic, forced-on, and forced-off choices.
func ForEnum(values []string) string {
	required := set.New(valueAuto, valueAlways, valueNever)
	if !required.SubsetOf(set.New(values...)) {
		return ""
	}
	return placeholderWhen
}
