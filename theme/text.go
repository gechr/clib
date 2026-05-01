package theme

import (
	"fmt"
	"strings"
)

const (
	themeNameDefault             = "default"
	themeNamePlain               = "plain"
	themeNameCatppuccinFrappe    = "catppuccin-frappe"
	themeNameCatppuccinLatte     = "catppuccin-latte"
	themeNameCatppuccinMacchiato = "catppuccin-macchiato"
	themeNameCatppuccinMocha     = "catppuccin-mocha"
	themeNameDracula             = "dracula"
	themeNameGruvboxDark         = "gruvbox-dark"
	themeNameGruvboxLight        = "gruvbox-light"
	themeNameMonochrome          = "monochrome"
	themeNameMonokai             = "monokai"
	themeNameNord                = "nord"
	themeNameOneDark             = "one-dark"
	themeNameSynthwave           = "synthwave"
	themeNameSolarized           = "solarized"
	themeNameTokyoNight          = "tokyo-night"
)

const (
	themeKeyCatppuccinFrappe    = "catppuccinfrappe"
	themeKeyCatppuccinLatte     = "catppuccinlatte"
	themeKeyCatppuccinMacchiato = "catppuccinmacchiato"
	themeKeyCatppuccinMocha     = "catppuccinmocha"
	themeKeyGruvboxDark         = "gruvboxdark"
	themeKeyGruvboxLight        = "gruvboxlight"
	themeKeyOneDark             = "onedark"
	themeKeyTokyoNight          = "tokyonight"
)

var validThemeNames = []string{
	themeNameDefault,

	themeNamePlain, // no styling

	themeNameCatppuccinFrappe,
	themeNameCatppuccinLatte,
	themeNameCatppuccinMacchiato,
	themeNameCatppuccinMocha,
	themeNameDracula,
	themeNameGruvboxDark,
	themeNameGruvboxLight,
	themeNameMonochrome,
	themeNameMonokai,
	themeNameNord,
	themeNameOneDark,
	themeNameSynthwave,
	themeNameSolarized,
	themeNameTokyoNight,
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
	case "", themeNameDefault:
		*t = *defaultTheme()
	case themeNamePlain:
		*t = *Plain()
	case themeNameMonochrome:
		*t = *Monochrome()
	case themeNameMonokai:
		*t = *Monokai()
	case themeKeyCatppuccinLatte:
		*t = *CatppuccinLatte()
	case themeKeyCatppuccinFrappe:
		*t = *CatppuccinFrappe()
	case themeKeyCatppuccinMacchiato:
		*t = *CatppuccinMacchiato()
	case themeKeyCatppuccinMocha:
		*t = *CatppuccinMocha()
	case themeNameDracula:
		*t = *Dracula()
	case themeKeyGruvboxDark:
		*t = *GruvboxDark()
	case themeKeyGruvboxLight:
		*t = *GruvboxLight()
	case themeNameNord:
		*t = *Nord()
	case themeKeyOneDark:
		*t = *OneDark()
	case themeNameSynthwave:
		*t = *Synthwave()
	case themeNameSolarized:
		*t = *Solarized()
	case themeKeyTokyoNight:
		*t = *TokyoNight()
	default:
		return fmt.Errorf("unknown theme %q (valid: %s)", text, strings.Join(validThemeNames, ", "))
	}
	return nil
}
