package kong

import (
	"reflect"
	"testing"
	"time"

	konglib "github.com/alecthomas/kong"
	"github.com/stretchr/testify/require"
)

type stringerInt int

func (stringerInt) String() string { return "value" }

type textUnmarshalerInt int

func (*textUnmarshalerInt) UnmarshalText([]byte) error { return nil }

func TestClibGroup_NilTag(t *testing.T) {
	// clibGroup is the only helper that can be reached with a nil Tag
	// without first hitting kong's own nil-Tag panics (e.g. IsCounter).
	f := &konglib.Flag{
		Value: &konglib.Value{
			Name: "test",
			Tag:  nil,
		},
	}
	group, err := clibGroup(f)
	require.NoError(t, err)
	require.Empty(t, group)
}

func TestKongIntegerType(t *testing.T) {
	require.True(t, kongIntegerType(reflect.TypeFor[int]()))
	require.True(t, kongIntegerType(reflect.TypeFor[[]uint64]()))
	require.False(t, kongIntegerType(reflect.TypeFor[time.Duration]()))
	require.False(t, kongIntegerType(reflect.TypeFor[stringerInt]()))
	require.False(t, kongIntegerType(reflect.TypeFor[textUnmarshalerInt]()))
}
