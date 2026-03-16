package human_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gechr/clib/human"
	"github.com/stretchr/testify/require"
)

func TestContractHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, "~/config.yaml", human.ContractHome(filepath.Join(home, "config.yaml")))
}

func TestContractHome_NotUnderHome(t *testing.T) {
	require.Equal(t, "/tmp/file.txt", human.ContractHome("/tmp/file.txt"))
}

func TestContractHome_ExactHome(t *testing.T) {
	home, err := os.UserHomeDir()
	require.NoError(t, err)

	require.Equal(t, "~", human.ContractHome(home))
}
