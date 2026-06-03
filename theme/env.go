package theme

import (
	"fmt"
	"os"
	"strings"
	"sync/atomic"
)

// DefaultEnvPrefix is the default environment variable prefix.
const DefaultEnvPrefix = "CLIB"

const (
	envTheme      = "THEME"
	envThemeDark  = "THEME_DARK"
	envThemeLight = "THEME_LIGHT"
)

var envPrefix atomic.Value // stores string; "" means no custom prefix

// SetEnvPrefix sets a custom environment variable prefix.
//
//	theme.SetEnvPrefix("MYAPP")
//	// Now checks MYAPP_THEME_LIGHT/MYAPP_THEME_DARK first, then CLIB_THEME_LIGHT/CLIB_THEME_DARK
func SetEnvPrefix(prefix string) {
	envPrefix.Store(strings.TrimRight(prefix, "_"))
}

// PairFromEnv builds a theme pair from <PREFIX>_THEME_LIGHT and <PREFIX>_THEME_DARK.
func PairFromEnv(opts ...PairOption) (*Pair, error) {
	lightName, lightEnv := lookupEnv(envThemeLight)
	darkName, darkEnv := lookupEnv(envThemeDark)
	lightName = strings.TrimSpace(lightName)
	darkName = strings.TrimSpace(darkName)
	if lightName == "" {
		return nil, fmt.Errorf("%s is required", lightEnv)
	}
	if darkName == "" {
		return nil, fmt.Errorf("%s is required", darkEnv)
	}

	var light Theme
	if err := light.UnmarshalText([]byte(lightName)); err != nil {
		return nil, fmt.Errorf("%s: %w", lightEnv, err)
	}
	var dark Theme
	if err := dark.UnmarshalText([]byte(darkName)); err != nil {
		return nil, fmt.Errorf("%s: %w", darkEnv, err)
	}
	return NewPair(&light, &dark, opts...)
}

func envOverride() *Theme {
	name, _ := lookupEnv(envTheme)
	name = strings.TrimSpace(name)
	if name == "" {
		return nil
	}
	var th Theme
	if err := th.UnmarshalText([]byte(name)); err != nil {
		return nil
	}
	return &th
}

func lookupEnv(suffix string) (string, string) {
	if p, ok := envPrefix.Load().(string); ok && p != "" {
		name := p + "_" + suffix
		if v := os.Getenv(name); v != "" {
			return v, name
		}
	}
	name := DefaultEnvPrefix + "_" + suffix
	return os.Getenv(name), name
}
