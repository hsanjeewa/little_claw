package ssh

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

// fakeExitError mimics *ssh.ExitError so we don't need a real SSH session.
type fakeExitError struct{ code int }

func (e fakeExitError) Error() string  { return fmt.Sprintf("Process exited with status %d", e.code) }
func (e fakeExitError) ExitStatus() int { return e.code }

func TestFormatExecutionError_IncludesCommandAndOutput(t *testing.T) {
	cmd := "sudo apt-get update"
	execErr := fakeExitError{code: 1}
	// On a PTY the real diagnostic arrives on the combined stdout stream.
	combined := "E: Failed to fetch http://archive.ubuntu.com/ 404 Not Found\nE: Some index files failed to download."

	err := formatExecutionError(cmd, execErr, "", combined)

	msg := err.Error()
	if !strings.Contains(msg, cmd) {
		t.Fatalf("error should include the failing command, got:\n%s", msg)
	}
	if !strings.Contains(msg, "Failed to fetch") {
		t.Fatalf("error should include the captured output/stderr so the cause is visible, got:\n%s", msg)
	}
	if !strings.Contains(msg, "exit status 1") && !strings.Contains(msg, "status 1") {
		t.Fatalf("error should preserve the exit status, got:\n%s", msg)
	}
}

func TestFormatExecutionError_EmptyCombinedStillReadable(t *testing.T) {
	cmd := "sudo apt-get update"
	execErr := fakeExitError{code: 1}

	err := formatExecutionError(cmd, execErr, "", "")
	msg := err.Error()
	if !strings.Contains(msg, cmd) {
		t.Fatalf("error should include the command even with no captured output, got:\n%s", msg)
	}
	if !errors.Is(err, execErr) {
		t.Fatal("error should wrap the original exec error for %w unwrapping")
	}
}

// --- passwordless-first sudo handling (Ansible-style) ---

func TestPrepareSudoCommand_PasswordlessWhenNoVaultPassword(t *testing.T) {
	// No per-host sudo password configured: prefer passwordless sudo and run
	// non-interactively (-n) so a missing password fails fast instead of hanging.
	cmd, feedPassword := prepareSudoCommand("sudo apt-get update", "", false)
	if cmd != "sudo -n apt-get update" {
		t.Fatalf("expected non-interactive sudo, got %q", cmd)
	}
	if feedPassword {
		t.Fatal("should not feed a password when none is configured")
	}
}

func TestPrepareSudoCommand_UsesVaultPasswordWhenConfigured(t *testing.T) {
	// Per-host sudo password present in the vault: run normally and feed it on prompt.
	cmd, feedPassword := prepareSudoCommand("sudo apt-get update", "s3cr3t", true)
	if cmd != "sudo apt-get update" {
		t.Fatalf("expected unchanged command, got %q", cmd)
	}
	if !feedPassword {
		t.Fatal("should feed the configured vault password on prompt")
	}
}

func TestPrepareSudoCommand_NonSudoCommandUnchanged(t *testing.T) {
	cmd, feedPassword := prepareSudoCommand("uptime", "", false)
	if cmd != "uptime" {
		t.Fatalf("non-sudo command should be unchanged, got %q", cmd)
	}
	if feedPassword {
		t.Fatal("no password feeding for non-sudo command")
	}
}
