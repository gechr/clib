package cobra

import (
	"strings"

	"github.com/gechr/clib/cli/internal/csvutil"
)

// CSVFlag implements pflag.Value for comma-separated values.
type CSVFlag struct {
	Values []string
}

// String returns the comma-separated string representation.
func (c *CSVFlag) String() string {
	return strings.Join(c.Values, ",")
}

// Set appends comma-separated values, filtering empty entries.
func (c *CSVFlag) Set(val string) error {
	c.Values = csvutil.Append(c.Values, val)
	return nil
}

// Type returns the pflag type name.
func (c *CSVFlag) Type() string {
	return "csv"
}
