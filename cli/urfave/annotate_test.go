package urfave

import (
	"testing"

	"github.com/stretchr/testify/require"
	clilib "github.com/urfave/cli/v3"
)

func TestExtend_NilFlag(t *testing.T) {
	resetExtras()
	t.Cleanup(resetExtras)

	// Should not panic.
	Extend(nil, FlagExtra{Group: "test"})
}

func TestExtendCommand_NilCommand(t *testing.T) {
	resetExtras()
	t.Cleanup(resetExtras)

	// Should not panic.
	ExtendCommand(nil, CommandExtra{PathArgs: true})
}

func TestExtendStoresExtras(t *testing.T) {
	resetExtras()
	t.Cleanup(resetExtras)

	f := &clilib.StringFlag{Name: "repo"}
	Extend(f, FlagExtra{Group: "Filters"})
	cmd := &clilib.Command{Name: "test", Flags: []clilib.Flag{f}}
	prepareFlagExtras(cmd)
	require.NotNil(t, getFlagExtra(cmd, f))

	ExtendCommand(cmd, CommandExtra{PathArgs: true})
	require.NotNil(t, getCommandExtra(cmd))
}

type valueFlag struct {
	name  string
	names []string
}

func (f valueFlag) String() string           { return f.name }
func (f valueFlag) Get() any                 { return nil }
func (f valueFlag) PreParse() error          { return nil }
func (f valueFlag) PostParse() error         { return nil }
func (f valueFlag) Set(string, string) error { return nil }
func (f valueFlag) Names() []string          { return append([]string{f.name}, f.names...) }
func (f valueFlag) IsSet() bool              { return false }

func TestExtendStoresExtrasForValueFlag(t *testing.T) {
	resetExtras()
	t.Cleanup(resetExtras)

	flag := valueFlag{name: "repo", names: []string{"r"}}
	Extend(flag, FlagExtra{Group: "Filters", Placeholder: "owner/repo"})

	cmd := &clilib.Command{Name: "test", Flags: []clilib.Flag{flag}}
	prepareFlagExtras(cmd)

	extra := getFlagExtra(cmd, flag)
	require.NotNil(t, extra)
	require.Equal(t, "Filters", extra.Group)
	require.Equal(t, "owner/repo", extra.Placeholder)
}
