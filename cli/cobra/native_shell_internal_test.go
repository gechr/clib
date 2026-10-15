package cobra

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestPatchCompletionScript(t *testing.T) {
	for _, script := range []string{"changed", "old middle old"} {
		_, err := patchCompletionScript(script, "old", "new")
		require.EqualError(
			t,
			err,
			fmt.Sprintf(
				"cobra completion template changed: expected one patch match, got %d",
				strings.Count(script, "old"),
			),
		)
	}
	got, err := patchCompletionScript("old", "old", "new")
	require.NoError(t, err)
	require.Equal(t, "new", got)
}
