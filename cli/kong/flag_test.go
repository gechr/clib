package kong_test

import (
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/cli/kong"
	"github.com/stretchr/testify/require"
)

func TestCSVFlag_String(t *testing.T) {
	f := kong.CSVFlag{Values: []string{"a", "b", "c"}}
	require.Equal(t, "a,b,c", f.String())
}

func TestCSVFlag_String_Empty(t *testing.T) {
	f := kong.CSVFlag{}
	require.Empty(t, f.String())
}

func TestCSVFlag_String_Single(t *testing.T) {
	f := kong.CSVFlag{Values: []string{"only"}}
	require.Equal(t, "only", f.String())
}

func TestCSVFlag_Decode(t *testing.T) {
	type CLI struct {
		Tags kong.CSVFlag `name:"tags"`
	}
	var cli CLI
	k, err := konglib.New(&cli)
	require.NoError(t, err)
	_, err = k.Parse([]string{"--tags", "a,b,c"})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, cli.Tags.Values)
}

func TestCSVFlag_Decode_Cumulative(t *testing.T) {
	type CLI struct {
		Tags kong.CSVFlag `name:"tags"`
	}
	var cli CLI
	k, err := konglib.New(&cli)
	require.NoError(t, err)
	_, err = k.Parse([]string{"--tags", "a,b", "--tags", "c"})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, cli.Tags.Values)
}

func TestCSVFlag_Decode_TrimsSpaces(t *testing.T) {
	type CLI struct {
		Tags kong.CSVFlag `name:"tags"`
	}
	var cli CLI
	k, err := konglib.New(&cli)
	require.NoError(t, err)
	_, err = k.Parse([]string{"--tags", " a , b , c "})
	require.NoError(t, err)
	require.Equal(t, []string{"a", "b", "c"}, cli.Tags.Values)
}
