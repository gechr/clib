package urfave

import (
	"strings"

	"github.com/gechr/clib/cli/internal/csvutil"
)

// CSVFlag implements urfave/cli's Value interface for comma-separated values.
// Use with *clilib.GenericFlag{Name: "x", Value: &CSVFlag{}}.
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

// Get returns the underlying string slice.
func (c *CSVFlag) Get() any {
	return c.Values
}

// Type returns the flag type name.
func (c *CSVFlag) Type() string {
	return "csv"
}
