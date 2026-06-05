package theme

import (
	"fmt"
	"strings"
)

const (
	themeNameDark                = "dark"
	themeNameLight               = "light"
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
	themeNameSolarizedDark       = "solarized-dark"
	themeNameSolarizedLight      = "solarized-light"
	themeNameTokyoNight          = "tokyo-night"
)

const (
	themeKeyDark                = "dark"
	themeKeyLight               = "light"
	themeKeyCatppuccinFrappe    = "catppuccinfrappe"
	themeKeyCatppuccinLatte     = "catppuccinlatte"
	themeKeyCatppuccinMacchiato = "catppuccinmacchiato"
	themeKeyCatppuccinMocha     = "catppuccinmocha"
	themeKeyGruvboxDark         = "gruvboxdark"
	themeKeyGruvboxLight        = "gruvboxlight"
	themeKeyMonochromeDark      = "monochromedark"
	themeKeyMonochromeLight     = "monochromelight"
	themeKeyOneDark             = "onedark"
	themeKeyPlainDark           = "plaindark"
	themeKeyPlainLight          = "plainlight"
	themeKeySolarizedDark       = "solarizeddark"
	themeKeySolarizedLight      = "solarizedlight"
	themeKeyTokyoNight          = "tokyonight"
)

var themeNames = []string{
	themeNameDark,
	themeNameLight,
	themeNameCatppuccinFrappe,
	themeNameCatppuccinLatte,
	themeNameCatppuccinMacchiato,
	themeNameCatppuccinMocha,
	themeNameDracula,
	themeNameGruvboxDark,
	themeNameGruvboxLight,
	themeNameMonokai,
	themeNameForBackground(themeNameMonochrome, BackgroundDark),
	themeNameForBackground(themeNameMonochrome, BackgroundLight),
	themeNameNord,
	themeNameOneDark,
	themeNameForBackground(themeNamePlain, BackgroundDark),
	themeNameForBackground(themeNamePlain, BackgroundLight),
	themeNameSynthwave,
	themeNameSolarizedDark,
	themeNameSolarizedLight,
	themeNameTokyoNight,
}

// Names returns the built-in theme names accepted by [Theme.UnmarshalText].
func Names() []string {
	return append([]string(nil), themeNames...)
}

func normalizePresetName(name string) string {
	name = strings.TrimSpace(strings.ToLower(name))
	replacer := strings.NewReplacer("-", "", "_", "", " ", "")
	return replacer.Replace(name)
}

func themeNameForBackground(name string, bg Background) string {
	return name + "-" + bg.String()
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
	case themeKeyDark:
		*t = *Dark()
	case themeKeyLight:
		*t = *Light()
	case themeKeyPlainDark:
		*t = *Plain(BackgroundDark)
	case themeKeyPlainLight:
		*t = *Plain(BackgroundLight)
	case themeKeyMonochromeDark:
		*t = *Monochrome(BackgroundDark)
	case themeKeyMonochromeLight:
		*t = *Monochrome(BackgroundLight)
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
	case themeKeySolarizedDark:
		*t = *SolarizedDark()
	case themeKeySolarizedLight:
		*t = *SolarizedLight()
	case themeKeyTokyoNight:
		*t = *TokyoNight()
	default:
		return fmt.Errorf("unknown theme %q (valid: %s)", text, strings.Join(themeNames, ", "))
	}
	return nil
}
