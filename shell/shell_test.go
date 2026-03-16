package shell_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gechr/clib/shell"
	"github.com/stretchr/testify/require"
)

func fakeProcess(name string) func() string {
	return func() string { return name }
}

func TestDetect_CompleteShellTakesPrecedence(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "zsh")
	t.Setenv(shell.EnvShell, "/bin/bash")
	t.Cleanup(shell.SetParentProcessName(fakeProcess("fish")))

	got := shell.Detect()
	if got != "zsh" {
		t.Errorf("Detect() = %q, want %q", got, "zsh")
	}
}

func TestDetect_CompleteShellPathReturnsBase(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "/usr/local/bin/fish")
	t.Setenv(shell.EnvShell, "/bin/bash")
	t.Cleanup(shell.SetParentProcessName(fakeProcess("")))

	got := shell.Detect()
	if got != "fish" {
		t.Errorf("Detect() = %q, want %q", got, "fish")
	}
}

func TestDetect_ProcessTakesPrecedenceOverShellEnv(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "")
	t.Setenv(shell.EnvShell, "/bin/bash")
	t.Cleanup(shell.SetParentProcessName(fakeProcess("zsh")))

	got := shell.Detect()
	if got != "zsh" {
		t.Errorf("Detect() = %q, want %q", got, "zsh")
	}
}

func TestDetect_FallsBackToShellEnv(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "")
	t.Setenv(shell.EnvShell, "/bin/zsh")
	t.Cleanup(shell.SetParentProcessName(fakeProcess("")))

	got := shell.Detect()
	if got != "zsh" {
		t.Errorf("Detect() = %q, want %q", got, "zsh")
	}
}

func TestDetect_EmptyWhenNothingSet(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "")
	t.Setenv(shell.EnvShell, "")
	t.Cleanup(shell.SetParentProcessName(fakeProcess("")))

	got := shell.Detect()
	if got != "" {
		t.Errorf("Detect() = %q, want empty", got)
	}
}

func TestDetect_UnknownProcessIgnored(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "")
	t.Setenv(shell.EnvShell, "/bin/fish")
	t.Cleanup(shell.SetParentProcessName(fakeProcess("node")))

	got := shell.Detect()
	if got != "fish" {
		t.Errorf("Detect() = %q, want %q", got, "fish")
	}
}

func TestDetectFromEnv(t *testing.T) {
	t.Setenv("TEST_SHELL", "/usr/bin/fish")

	got := shell.DetectFromEnv("TEST_SHELL")
	if got != "fish" {
		t.Errorf("DetectFromEnv() = %q, want %q", got, "fish")
	}
}

func TestDetectFromEnv_Empty(t *testing.T) {
	t.Setenv("TEST_SHELL", "")

	got := shell.DetectFromEnv("TEST_SHELL")
	if got != "" {
		t.Errorf("DetectFromEnv() = %q, want empty", got)
	}
}

func TestDetectFromEnv_Unset(t *testing.T) {
	got := shell.DetectFromEnv("NONEXISTENT_SHELL_VAR_FOR_TEST")
	if got != "" {
		t.Errorf("DetectFromEnv() = %q, want empty", got)
	}
}

func TestDetectFromEnv_UnknownIgnored(t *testing.T) {
	t.Setenv("TEST_SHELL", "/usr/bin/node")

	got := shell.DetectFromEnv("TEST_SHELL")
	if got != "" {
		t.Errorf("DetectFromEnv() = %q, want empty", got)
	}
}

func TestDetect_UnknownShellEnvIgnored(t *testing.T) {
	t.Setenv("COMPLETE_SHELL", "")
	t.Setenv(shell.EnvShell, "/usr/bin/node")
	t.Cleanup(shell.SetParentProcessName(fakeProcess("")))

	got := shell.Detect()
	if got != "" {
		t.Errorf("Detect() = %q, want empty", got)
	}
}

func TestDetectFromProcess(t *testing.T) {
	t.Cleanup(shell.SetParentProcessName(fakeProcess("fish")))

	got := shell.DetectFromProcess()
	if got != "fish" {
		t.Errorf("DetectFromProcess() = %q, want %q", got, "fish")
	}
}

func TestDetectFromProcess_UnknownIgnored(t *testing.T) {
	t.Cleanup(shell.SetParentProcessName(fakeProcess("node")))

	got := shell.DetectFromProcess()
	if got != "" {
		t.Errorf("DetectFromProcess() = %q, want empty", got)
	}
}

func TestDetectFromProcess_Empty(t *testing.T) {
	t.Cleanup(shell.SetParentProcessName(fakeProcess("")))

	got := shell.DetectFromProcess()
	if got != "" {
		t.Errorf("DetectFromProcess() = %q, want empty", got)
	}
}

func TestXDGDataHome_EnvSet(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/custom/data")

	got, err := shell.XDGDataHome()
	require.NoError(t, err)
	require.Equal(t, "/custom/data", got)
}

func TestXDGDataHome_EnvUnset(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	got, err := shell.XDGDataHome()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".local", "share"), got)
}

func TestXDGConfigHome_EnvSet(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "/custom/config")

	got, err := shell.XDGConfigHome()
	require.NoError(t, err)
	require.Equal(t, "/custom/config", got)
}

func TestXDGConfigHome_EnvUnset(t *testing.T) {
	t.Setenv("XDG_CONFIG_HOME", "")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	got, err := shell.XDGConfigHome()
	require.NoError(t, err)
	require.Equal(t, filepath.Join(home, ".config"), got)
}

func TestCompletionFile(t *testing.T) {
	t.Setenv("XDG_DATA_HOME", "/data")
	t.Setenv("XDG_CONFIG_HOME", "/config")

	tests := []struct {
		name    string
		command string
		shell   string
		want    string
	}{
		{
			name:    "bash",
			command: "myapp",
			shell:   shell.Bash,
			want:    "/data/bash-completion/completions/myapp",
		},
		{
			name:    "zsh",
			command: "myapp",
			shell:   shell.Zsh,
			want:    "/data/zsh/site-functions/_myapp",
		},
		{
			name:    "fish",
			command: "myapp",
			shell:   shell.Fish,
			want:    "/config/fish/completions/myapp.fish",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := shell.CompletionFile(tt.command, tt.shell)
			require.NoError(t, err)
			require.Equal(t, tt.want, got)
		})
	}
}

func TestCompletionFile_UnsupportedShell(t *testing.T) {
	_, err := shell.CompletionFile("myapp", "unsupported")
	require.Error(t, err)
	require.EqualError(t, err, `unsupported shell "unsupported"`)
}

func TestKnown(t *testing.T) {
	known := shell.Known()
	require.NotEmpty(t, known)
	require.Contains(t, known, "bash")
	require.Contains(t, known, "zsh")
	require.Contains(t, known, "fish")

	// Verify mutation safety: modifying the returned slice should not affect future calls.
	known[0] = "mutated"
	fresh := shell.Known()
	require.NotEqual(t, "mutated", fresh[0])
}

func TestIsKnown(t *testing.T) {
	tests := []struct {
		name string
		want bool
	}{
		{"bash", true},
		{"nonexistent", false},
		{"", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, shell.IsKnown(tt.name))
		})
	}
}
