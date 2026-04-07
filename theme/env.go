package theme

import (
	"os"
	"strings"
	"sync/atomic"
)

// DefaultEnvPrefix is the default environment variable prefix.
const DefaultEnvPrefix = "CLIB"

const envTheme = "THEME"

var envPrefix atomic.Value // stores string; "" means no custom prefix

// SetEnvPrefix sets a custom environment variable prefix. Env vars are
// checked with the custom prefix first, then "CLIB" as fallback.
//
//	theme.SetEnvPrefix("MYAPP")
//	// Now checks MYAPP_THEME first, then CLIB_THEME
func SetEnvPrefix(prefix string) {
	envPrefix.Store(strings.TrimRight(prefix, "_"))
}

func getEnv(suffix string) string {
	if p, ok := envPrefix.Load().(string); ok && p != "" {
		if v := os.Getenv(p + "_" + suffix); v != "" {
			return v
		}
	}
	return os.Getenv(DefaultEnvPrefix + "_" + suffix)
}
