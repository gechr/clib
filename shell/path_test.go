package shell_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gechr/clib/shell"
	"github.com/stretchr/testify/require"
)

func TestExpandPath_Tilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, filepath.Join(home, "config.yaml"), shell.ExpandPath("~/config.yaml"))
}

func TestExpandPath_BareTilde(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, home, shell.ExpandPath("~"))
}

func TestExpandPath_EnvVar(t *testing.T) {
	t.Setenv("TEST_EXPAND_DIR", "/opt/data")
	require.Equal(t, "/opt/data/file.txt", shell.ExpandPath("$TEST_EXPAND_DIR/file.txt"))
}

func TestExpandPath_Empty(t *testing.T) {
	require.Empty(t, shell.ExpandPath(""))
}

func TestExpandPath_NoExpansion(t *testing.T) {
	require.Equal(t, "/absolute/path", shell.ExpandPath("/absolute/path"))
}
