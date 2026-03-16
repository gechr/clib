package kong

import (
	"strings"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/cli/internal/csvutil"
)

// CSVFlag implements kong.MapperValue that splits comma-separated values.
type CSVFlag struct {
	Values []string
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
	c.Values = csvutil.Append(c.Values, value)
	return nil
}
