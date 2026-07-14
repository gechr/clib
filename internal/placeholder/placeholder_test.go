package placeholder_test

import (
	"testing"

	"github.com/gechr/clib/internal/placeholder"
	"github.com/stretchr/testify/require"
)

func TestForEnum(t *testing.T) {
	t.Parallel()

	for _, values := range [][]string{
		{"auto", "always", "never"},
		{"never", "auto", "always"},
		{"auto", "always", "never", "sometimes"},
		{"auto", "always", "never", "auto"},
	} {
		require.Equal(t, "when", placeholder.ForEnum(values))
	}
	for _, values := range [][]string{
		{"auto", "never"},
		{"auto", "auto", "never"},
		{"AUTO", "always", "never"},
	} {
		require.Empty(t, placeholder.ForEnum(values))
	}
}
