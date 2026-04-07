package theme

import (
	"fmt"
	"strings"
)

var validThemeNames = []string{
	"default",
	"monochrome",
	"monokai",
	"catppuccin-latte",
	"catppuccin-frappe",
	"catppuccin-macchiato",
	"catppuccin-mocha",
	"dracula",
}

func normalizePresetName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	replacer := strings.NewReplacer("-", "", "_", "", " ", "")
	return replacer.Replace(name)
}

// String returns the preset name for built-in themes, or "custom" for themes
// that were modified programmatically.
func (t *Theme) String() string {
	if t == nil || t.name == "" {
		return "custom"
	}
	return t.name
}

// MarshalText implements [encoding.TextMarshaler].
func (t *Theme) MarshalText() ([]byte, error) {
	if t == nil {
		return nil, fmt.Errorf("cannot marshal nil theme")
	}
	if t.name == "" {
		return nil, fmt.Errorf("cannot marshal custom theme")
	}
	return []byte(t.name), nil
}

// UnmarshalText implements [encoding.TextUnmarshaler].
func (t *Theme) UnmarshalText(text []byte) error {
	if t == nil {
		return fmt.Errorf("cannot unmarshal theme into nil receiver")
	}

	switch normalizePresetName(string(text)) {
	case "", "default":
		*t = *defaultTheme()
	case "monochrome":
		*t = *Monochrome()
	case "monokai":
		*t = *Monokai()
	case "catppuccinlatte":
		*t = *CatppuccinLatte()
	case "catppuccinfrappe":
		*t = *CatppuccinFrappe()
	case "catppuccinmacchiato":
		*t = *CatppuccinMacchiato()
	case "catppuccinmocha":
		*t = *CatppuccinMocha()
	case "dracula":
		*t = *Dracula()
	default:
		return fmt.Errorf("unknown theme %q (valid: %s)", text, strings.Join(validThemeNames, ", "))
	}
	return nil
}
