//go:build linux

package shell

import (
	"os"
	"strconv"
	"strings"
)

// processName returns the name of the process with the given PID
// by reading /proc/<pid>/comm on Linux.
func processName(pid int) string {
	data, err := os.ReadFile("/proc/" + strconv.Itoa(pid) + "/comm")
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(data))
}
