package cobra

import (
	"testing"

	cobralib "github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/stretchr/testify/require"
)

func TestExtend_NilFlag(_ *testing.T) {
	// Should not panic.
	Extend(nil, FlagExtra{Group: "test"})
}

func TestExtendCommand_NilCommand(_ *testing.T) {
	// Should not panic.
	ExtendCommand(nil, CommandExtra{Alias: "tool release"})
}

func TestExtendStoresExtraOnFlag(t *testing.T) {
	f := &pflag.Flag{Name: "repo"}
	Extend(f, FlagExtra{Group: "Filters"})
	require.NotNil(t, getExtra(f))
}

func TestExtendCommandStoresExtraOnCommand(t *testing.T) {
	cmd := &cobralib.Command{Use: "init"}
	ExtendCommand(cmd, CommandExtra{Alias: "tool release"})
	require.Equal(t, "tool release", getCommandExtra(cmd).Alias)
}
