package urfave_test

import (
	"testing"

	"github.com/gechr/clib/cli/urfave"
	"github.com/stretchr/testify/require"
)

func TestCSVFlag_Set(t *testing.T) {
	c := &urfave.CSVFlag{}
	err := c.Set("a, b, c")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, c.Values)

	// Append more values.
	err = c.Set("d,e")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c", "d", "e"}, c.Values)
}

func TestCSVFlag_String(t *testing.T) {
	c := &urfave.CSVFlag{Values: []string{"x", "y", "z"}}
	require.Equal(t, "x,y,z", c.String())
}

func TestCSVFlag_String_Empty(t *testing.T) {
	c := &urfave.CSVFlag{}
	require.Empty(t, c.String())
}

func TestCSVFlag_Get(t *testing.T) {
	c := &urfave.CSVFlag{Values: []string{"a", "b"}}
	got, ok := c.Get().([]string)
	require.True(t, ok)
	require.Equal(t, []string{"a", "b"}, got)
}

func TestCSVFlag_Type(t *testing.T) {
	c := &urfave.CSVFlag{}
	require.Equal(t, "csv", c.Type())
}
