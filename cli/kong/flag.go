package kong

import (
	"reflect"
	"strings"

	konglib "github.com/alecthomas/kong"
	xstrings "github.com/gechr/x/strings"
)

// CSVFlag implements kong.MapperValue that splits comma-separated values.
type CSVFlag struct {
	Values []string
}

// CSVFlagPlaceholders returns a Kong option that makes Kong describe CSVFlag
// values using each flag's name (for example, --route=ROUTE) instead
// of the Go type name (--route=CSV-FLAG). Explicit placeholder tags are
// preserved.
//
// Pass this option to [konglib.New] or [konglib.Must].
func CSVFlagPlaceholders() konglib.Option {
	return konglib.PostBuild(func(k *konglib.Kong) error {
		normalizeCSVFlagPlaceholders(k.Model.Node)
		return nil
	})
}

func normalizeCSVFlagPlaceholders(node *konglib.Node) {
	for _, flag := range node.Flags {
		if flag.PlaceHolder == "" && isCSVFlag(flag) {
			// Kong prefers the named Go type when deriving a placeholder. Clear
			// only that inference so FormatPlaceHolder falls through to the
			// individual flag name.
			flag.Tag.TypeName = ""
		}
	}
	for _, child := range node.Children {
		normalizeCSVFlagPlaceholders(child)
	}
}

// isCSVFlag reports whether the kong flag's target type is CSVFlag or *CSVFlag.
func isCSVFlag(flag *konglib.Flag) bool {
	csvType := reflect.TypeFor[CSVFlag]()
	t := flag.Target.Type()
	if t.Kind() == reflect.Pointer {
		t = t.Elem()
	}
	return t == csvType
}

// String returns the comma-separated string representation.
func (c *CSVFlag) String() string {
	return strings.Join(c.Values, ",")
}

// Decode implements kong.MapperValue by splitting comma-separated values.
func (c *CSVFlag) Decode(ctx *konglib.DecodeContext) error {
	var value string
	if err := ctx.Scan.PopValueInto("value", &value); err != nil {
		return err
	}
	c.Values = xstrings.AppendCSV(c.Values, value)
	return nil
}
