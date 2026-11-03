package kong_test

import (
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/gechr/clib/cli/kong"
	"github.com/stretchr/testify/require"
)

type customBool int

func (b *customBool) Decode(ctx *konglib.DecodeContext) error {
	value := true
	if ctx.Scan.Peek().Type == konglib.FlagValueToken {
		if err := ctx.Scan.PopValueInto("bool", &value); err != nil {
			return err
		}
	}
	*b = 0
	if value {
		*b = 1
	}
	return nil
}

func (*customBool) IsBool() bool { return true }

func TestReflect_CustomNegatableBool(t *testing.T) {
	var cli struct {
		Enabled  customBool  `help:"Enable feature"          negatable:""`
		Optional *customBool `help:"Enable optional feature" negatable:""`
	}
	parser, err := konglib.New(&cli)
	require.NoError(t, err)

	flags, err := kong.Reflect(&cli)
	require.NoError(t, err)
	require.Len(t, flags, 2)
	for _, meta := range flags {
		t.Run(meta.Name, func(t *testing.T) {
			require.True(t, meta.Negatable)
			require.False(t, meta.HasArg, "custom boolean flags accept a bare switch")
			require.Empty(
				t,
				meta.Placeholder,
				"custom boolean flags have no inferred numeric placeholder",
			)
		})
	}

	_, err = parser.Parse([]string{"--enabled", "--no-optional"})
	require.NoError(t, err)
	require.Equal(t, customBool(1), cli.Enabled)
	require.NotNil(t, cli.Optional)
	require.Zero(t, *cli.Optional)
}
