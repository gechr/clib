//go:build darwin

package shell

import "golang.org/x/sys/unix"

// processName returns the name of the process with the given PID
// using the kern.proc.pid sysctl on macOS.
func processName(pid int) string {
	kinfo, err := unix.SysctlKinfoProc("kern.proc.pid", pid)
	if err != nil {
		return ""
	}
	comm := kinfo.Proc.P_comm
	// Find the null terminator.
	n := 0
	for n < len(comm) && comm[n] != 0 {
		n++
	}
	return string(comm[:n])
}
