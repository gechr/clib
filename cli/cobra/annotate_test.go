package cobra

import (
	"testing"

	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestExtend_NilFlag(t *testing.T) {
	resetExtras()
	t.Cleanup(resetExtras)

	// Should not panic.
	Extend(nil, FlagExtra{Group: "test"})
}

func TestExtendStoresExtraOnFlag(t *testing.T) {
	resetExtras()
	t.Cleanup(resetExtras)

	f := &pflag.Flag{Name: "repo"}
	Extend(f, FlagExtra{Group: "Filters"})
	require.NotNil(t, getExtra(f))
}
