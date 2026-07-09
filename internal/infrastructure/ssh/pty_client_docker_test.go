package ssh

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/devops/agent/internal/domain/agent"
	"github.com/devops/agent/internal/infrastructure/security"
)

// requireDockerHost skips the test if the given SSH target is not reachable,
// so the Docker integration tests degrade gracefully when containers are down.
func requireDockerHost(t *testing.T, host string, port int) {
	t.Helper()
	addr := fmt.Sprintf("%s:%d", host, port)
	conn, err := net.DialTimeout("tcp", addr, 3*time.Second)
	if err != nil {
		t.Skipf("Docker test host %s not reachable (start it with `just up` / the password-enforced container): %v", addr, err)
	}
	conn.Close()
}

// loadTestKey reads the repo's test SSH private key, searching upward from the
// working directory so the test works regardless of the package it runs from.
func loadTestKey(t *testing.T) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("failed to get working dir: %v", err)
	}
	for {
		p := filepath.Join(dir, "test_keys", "id_ed25519")
		if data, err := os.ReadFile(p); err == nil {
			return string(data)
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	t.Fatalf("test_keys/id_ed25519 not found (searched upward from cwd)")
	return ""
}

// newTestVault builds a real encrypted vault seeded with the test key for the
// given alias. sudoPassword, when non-empty, is stored as the per-host sudo
// password (Ansible-style) to exercise the password-fed escalation path.
func newTestVault(t *testing.T, alias, sudoPassword string) *security.LocalEncryptedVault {
	t.Helper()
	v := security.NewLocalEncryptedVault([]byte("a-very-secret-key-32-bytes-long!"))
	if err := v.EncryptAndStore(alias, loadTestKey(t)); err != nil {
		t.Fatalf("failed to store test key: %v", err)
	}
	if sudoPassword != "" {
		if err := v.EncryptAndStoreSudoPassword(alias, sudoPassword); err != nil {
			t.Fatalf("failed to store sudo password: %v", err)
		}
	}
	return v
}

// TestExecute_AgainstDocker_PasswordlessSudo verifies the agent executes a
// privileged command on the live test container using passwordless sudo
// (sudo -n). Requires `just up` (or equivalent) to have started the containers.
func TestExecute_AgainstDocker_PasswordlessSudo(t *testing.T) {
	requireDockerHost(t, "127.0.0.1", 2222)
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}
	client := NewSSHClient(newTestVault(t, "web-prod-01", ""))

	task, err := agent.NewTask("t1", "web-prod-01", "127.0.0.1", 2222, "deployer", "sudo -n id -un", true)
	if err != nil {
		t.Fatalf("failed to build task: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := client.Execute(ctx, task)
	if err != nil {
		t.Fatalf("Execute failed: %v", err)
	}
	// NOPASSWD sudo as deployer should report root.
	if !strings.Contains(out, "root") {
		t.Fatalf("expected 'sudo -n id -un' to report root, got: %q", out)
	}
}

// TestExecute_AgainstDocker_FailureShowsDiagnostics verifies that a failing
// privileged command surfaces the command and captured output, not a bare
// exit status.
func TestExecute_AgainstDocker_FailureShowsDiagnostics(t *testing.T) {
	requireDockerHost(t, "127.0.0.1", 2222)
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}
	client := NewSSHClient(newTestVault(t, "web-prod-01", ""))

	task, err := agent.NewTask("t2", "web-prod-01", "127.0.0.1", 2222, "deployer", "sudo -n apt-get update", true)
	if err != nil {
		t.Fatalf("failed to build task: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	_, err = client.Execute(ctx, task)
	if err == nil {
		t.Fatal("expected failure for apt-get on Alpine (apk is the package manager)")
	}
	msg := err.Error()
	if !strings.Contains(msg, "sudo -n apt-get update") {
		t.Fatalf("error should name the failing command, got:\n%s", msg)
	}
	// Alpine has no apt-get, so the diagnostic must be present (not truncated).
	if !strings.Contains(msg, "apt-get") {
		t.Fatalf("error should include the underlying cause, got:\n%s", msg)
	}
}

// TestExecute_AgainstDocker_WithVaultSudoPassword verifies the executor does
// not break when a per-host sudo password is configured in the vault (the
// container still uses NOPASSWD, so the password is simply never needed).
func TestExecute_AgainstDocker_WithVaultSudoPassword(t *testing.T) {
	requireDockerHost(t, "127.0.0.1", 2222)
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}
	client := NewSSHClient(newTestVault(t, "web-prod-01", "dummy-sudo-password"))

	task, err := agent.NewTask("t3", "web-prod-01", "127.0.0.1", 2222, "deployer", "sudo id -un", true)
	if err != nil {
		t.Fatalf("failed to build task: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := client.Execute(ctx, task)
	if err != nil {
		t.Fatalf("Execute failed with configured sudo password: %v", err)
	}
	if !strings.Contains(out, "root") {
		t.Fatalf("expected sudo id -un to report root, got: %q", out)
	}
}

// TestExecute_AgainstDocker_SudoPasswordRequired verifies the real escalation
// path: the host REQUIRES a sudo password (no NOPASSWD) and the executor must
// detect the sudo prompt and feed the per-host password from the vault. Start
// the container with `just up-sudopw` (alias "sudopw-prod-01", password
// "Sup3rSecret!", port 2227) and set SUDO_PASS_SUDOPW_PROD_01=Sup3rSecret! in
// .env so the vault supplies the correct password. The test skips if the
// container is not running.
func TestExecute_AgainstDocker_SudoPasswordRequired(t *testing.T) {
	requireDockerHost(t, "127.0.0.1", 2227)
	if testing.Short() {
		t.Skip("skipping Docker integration test in short mode")
	}
	client := NewSSHClient(newTestVault(t, "sudopw-prod-01", "Sup3rSecret!"))

	task, err := agent.NewTask("t4", "sudopw-prod-01", "127.0.0.1", 2227, "ubuntu", "sudo id -un", true)
	if err != nil {
		t.Fatalf("failed to build task: %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	out, err := client.Execute(ctx, task)
	if err != nil {
		t.Fatalf("Execute failed despite correct vault sudo password: %v", err)
	}
	if !strings.Contains(out, "root") {
		t.Fatalf("expected sudo id -un to report root after password feed, got: %q", out)
	}
}
