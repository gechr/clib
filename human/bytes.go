package human

import (
	"fmt"
	"strconv"
	"strings"
)

// SI byte size constants (powers of 1000).
const (
	KB = 1000
	MB = 1000 * KB
	GB = 1000 * MB
	TB = 1000 * GB
	PB = 1000 * TB
	EB = 1000 * PB
)

// IEC byte size constants (powers of 1024).
const (
	KiB = 1024
	MiB = 1024 * KiB
	GiB = 1024 * MiB
	TiB = 1024 * GiB
	PiB = 1024 * TiB
	EiB = 1024 * PiB
)

// Unit label constants.
const (
	UnitB   = "B"
	UnitKB  = "KB"
	UnitMB  = "MB"
	UnitGB  = "GB"
	UnitTB  = "TB"
	UnitPB  = "PB"
	UnitEB  = "EB"
	UnitKiB = "KiB"
	UnitMiB = "MiB"
	UnitGiB = "GiB"
	UnitTiB = "TiB"
	UnitPiB = "PiB"
	UnitEiB = "EiB"
)

// ParseByteSize parses a human-readable byte size string like "27.61 MiB" or
// "1.5 GB" into a byte count. Supports both IEC (KiB, MiB, GiB, TiB, PiB, EiB)
// and SI (KB, MB, GB, TB, PB, EB) units. Returns 0 for empty or unparseable input.
func ParseByteSize(s string) float64 {
	s = strings.TrimSpace(s)
	if s == "" {
		return 0
	}

	i := 0
	for i < len(s) && (s[i] == '.' || (s[i] >= '0' && s[i] <= '9')) {
		i++
	}
	if i == 0 {
		return 0
	}

	num, err := strconv.ParseFloat(s[:i], 64)
	if err != nil {
		return 0
	}

	unit := strings.TrimSpace(s[i:])
	switch unit {
	case UnitEiB:
		return num * EiB
	case UnitPiB:
		return num * PiB
	case UnitTiB:
		return num * TiB
	case UnitGiB:
		return num * GiB
	case UnitMiB:
		return num * MiB
	case UnitKiB:
		return num * KiB
	case UnitEB:
		return num * EB
	case UnitPB:
		return num * PB
	case UnitTB:
		return num * TB
	case UnitGB:
		return num * GB
	case UnitMB:
		return num * MB
	case UnitKB, "kB":
		return num * KB
	case "bytes", "byte", UnitB, "":
		return num
	default:
		return num
	}
}

// FormatSIBytes formats a byte count using SI decimal units (KB, MB, GB, TB, PB, EB).
func FormatSIBytes(b float64) string {
	switch {
	case b >= EB:
		return fmt.Sprintf("%.2f %s", b/EB, UnitEB)
	case b >= PB:
		return fmt.Sprintf("%.2f %s", b/PB, UnitPB)
	case b >= TB:
		return fmt.Sprintf("%.2f %s", b/TB, UnitTB)
	case b >= GB:
		return fmt.Sprintf("%.2f %s", b/GB, UnitGB)
	case b >= MB:
		return fmt.Sprintf("%.2f %s", b/MB, UnitMB)
	case b >= KB:
		return fmt.Sprintf("%.2f %s", b/KB, UnitKB)
	default:
		return fmt.Sprintf("%.0f %s", b, UnitB)
	}
}

// FormatIECBytes formats a byte count using IEC binary units (KiB, MiB, GiB, TiB, PiB, EiB).
func FormatIECBytes(b float64) string {
	switch {
	case b >= EiB:
		return fmt.Sprintf("%.2f %s", b/EiB, UnitEiB)
	case b >= PiB:
		return fmt.Sprintf("%.2f %s", b/PiB, UnitPiB)
	case b >= TiB:
		return fmt.Sprintf("%.2f %s", b/TiB, UnitTiB)
	case b >= GiB:
		return fmt.Sprintf("%.2f %s", b/GiB, UnitGiB)
	case b >= MiB:
		return fmt.Sprintf("%.2f %s", b/MiB, UnitMiB)
	case b >= KiB:
		return fmt.Sprintf("%.2f %s", b/KiB, UnitKiB)
	default:
		return fmt.Sprintf("%.0f %s", b, UnitB)
	}
}
