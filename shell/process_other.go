//go:build !darwin && !linux

package shell

// processName is a stub for unsupported platforms.
// It always returns an empty string.
func processName(_ int) string {
	return ""
}
