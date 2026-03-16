package cobra_test

import (
	"testing"

	"github.com/gechr/clib/cli/cobra"
	"github.com/stretchr/testify/require"
)

func TestCSVFlag_String_Empty(t *testing.T) {
	f := &cobra.CSVFlag{}
	require.Empty(t, f.String())
}

func TestCSVFlag_String_Values(t *testing.T) {
	f := &cobra.CSVFlag{Values: []string{"a", "b", "c"}}
	require.Equal(t, "a,b,c", f.String())
}

func TestCSVFlag_Set_SingleValue(t *testing.T) {
	f := &cobra.CSVFlag{}
	err := f.Set("foo")
	require.NoError(t, err)
	require.Equal(t, []string{"foo"}, f.Values)
}

func TestCSVFlag_Set_CommaSeparated(t *testing.T) {
	f := &cobra.CSVFlag{}
	err := f.Set("a, b, c")
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, f.Values)
}

func TestCSVFlag_Set_Accumulates(t *testing.T) {
	f := &cobra.CSVFlag{}
	require.NoError(t, f.Set("a,b"))
	require.NoError(t, f.Set("c"))
	require.Equal(t, []string{"a", "b", "c"}, f.Values)
}

func TestCSVFlag_Type(t *testing.T) {
	f := &cobra.CSVFlag{}
	require.Equal(t, "csv", f.Type())
}
