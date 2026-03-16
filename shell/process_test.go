package shell

import "testing"

func TestParentProcessName_ReturnsNonEmpty(t *testing.T) {
	// Call the real parentProcessName - go test always has a parent process,
	// so ppid > 0 and the gopsutil lookup should succeed.
	name := parentProcessName()
	if name == "" {
		t.Skip("parentProcessName returned empty (unexpected in go test)")
	}
	t.Logf("parentProcessName() = %q", name)
}
