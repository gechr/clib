package cobra_test

import (
	"os/exec"
	"strings"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestCobraDependencyBoundary(t *testing.T) {
	for _, pkg := range []string{"./complete", "./cli/kong", "./cli/urfave"} {
		t.Run(pkg, func(t *testing.T) {
			cmd := exec.CommandContext(t.Context(), "go", "list", "-f", `{{join .Deps "\n"}}`, pkg)
			cmd.Dir = "../.."
			out, err := cmd.CombinedOutput()
			require.NoError(t, err, "%s", out)
			deps := strings.Fields(string(out))
			require.NotContains(t, deps, "github.com/spf13/cobra")
			require.NotContains(t, deps, "github.com/spf13/pflag")
		})
	}
}
