package kong

import (
	"testing"

	konglib "github.com/alecthomas/kong"
	"github.com/stretchr/testify/require"
)

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
