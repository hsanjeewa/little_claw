package ssh

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/devops/agent/internal/domain/agent"
)

type SSHClient struct {
	vault agent.SecretVault
}

func NewSSHClient(vault agent.SecretVault) *SSHClient {
	return &SSHClient{
		vault: vault,
	}
}

func (c *SSHClient) Execute(ctx context.Context, task agent.Task) (string, error) {
	keyData, err := c.vault.GetPrivateKey(task.HostAlias)
	if err != nil {
		return "", fmt.Errorf("context: failed to get private key: %w", err)
	}

	signer, err := ssh.ParsePrivateKey([]byte(keyData))
	if err != nil {
		return "", fmt.Errorf("context: failed to parse private key: %w", err)
	}

	// Ansible-style sudo handling: prefer a per-host password from the encrypted
	// vault. When none is configured, fall back to passwordless sudo and run it
	// non-interactively (-n) so a missing password fails fast instead of hanging
	// on a prompt.
	sudoPass, sudoErr := c.vault.GetSudoPassword(task.HostAlias)
	haveSudoPass := sudoErr == nil
	runCmd, feedPassword := prepareSudoCommand(task.Command, sudoPass, haveSudoPass)

	config := &ssh.ClientConfig{
		User: task.User,
		Auth: []ssh.AuthMethod{
			ssh.PublicKeys(signer),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         10 * time.Second,
	}

	addr := fmt.Sprintf("%s:%d", task.HostIP, task.Port)

	dialCtx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	select {
	case <-dialCtx.Done():
		return "", fmt.Errorf("context: %w", dialCtx.Err())
	default:
	}

	conn, err := ssh.Dial("tcp", addr, config)
	if err != nil {
		return "", fmt.Errorf("context: failed to dial: %w", err)
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("context: failed to create session: %w", err)
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		return "", fmt.Errorf("context: request for pseudo terminal failed: %w", err)
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("context: failed to get stdin pipe: %w", err)
	}
	defer stdin.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("context: failed to get stdout pipe: %w", err)
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("context: failed to get stderr pipe: %w", err)
	}

	if err := session.Start(runCmd); err != nil {
		return "", fmt.Errorf("context: failed to start command: %w", err)
	}

	outputChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		var outputBuf bytes.Buffer
		stderrBuf := new(bytes.Buffer)
		buf := make([]byte, 4096)

		// Drain stderr concurrently so the diagnostic is never lost, regardless
		// of whether the remote PTY merges stdout/stderr into a single stream.
		stderrDone := make(chan struct{})
		go func() {
			_, _ = io.Copy(stderrBuf, stderr)
			close(stderrDone)
		}()

		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				outputBuf.Write(chunk)

				if feedPassword {
					chunkStr := strings.ToLower(string(chunk))
					if strings.Contains(chunkStr, "password") || strings.Contains(chunkStr, "[sudo]") {
						_, _ = stdin.Write([]byte(sudoPass + "\n"))
					}
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				errorChan <- fmt.Errorf("context: read error: %w", err)
				return
			}
		}

		<-stderrDone
		if stderrBuf.Len() > 0 {
			outputBuf.Write(stderrBuf.Bytes())
		}

		err = session.Wait()
		if err != nil {
			// On a PTY the remote shell merges stdout and stderr into the
			// single tty stream, so the meaningful diagnostic usually lives in
			// outputBuf rather than the (often empty) stderr pipe. Capture both
			// so the operator can see exactly why the command failed.
			stderrMsg := strings.TrimSpace(stderrBuf.String())
			combined := strings.TrimSpace(outputBuf.String())
			errorChan <- formatExecutionError(task.Command, err, stderrMsg, combined)
		}

		outputChan <- outputBuf.String()
	}()

	select {
	case <-ctx.Done():
		return "", fmt.Errorf("context: %w", ctx.Err())
	case err := <-errorChan:
		return "", err
	case output := <-outputChan:
		return output, nil
	}
}

// formatExecutionError builds a diagnosable error for a failed command. It
// always names the command that failed and appends any captured output so the
// operator can see the underlying cause (e.g. an apt 404 or a sudo prompt)
// instead of a bare "Process exited with status 1". The original error is
// wrapped so callers can still use errors.Is/As.
func formatExecutionError(command string, execErr error, stderrMsg, combined string) error {
	detail := stderrMsg
	if detail == "" {
		detail = combined
	}
	if detail == "" {
		return fmt.Errorf("context: command %q execution failed: %w", command, execErr)
	}
	return fmt.Errorf("context: command %q execution failed: %w\n%s", command, execErr, detail)
}

// prepareSudoCommand implements Ansible-style privilege escalation:
//   - If a per-host sudo password is configured in the vault, run the command
//     as-is and return feedPassword=true so the caller supplies it on prompt.
//   - If no password is configured, fall back to passwordless sudo and force
//     non-interactive mode (-n) so a missing password fails fast rather than
//     hanging on a prompt (feedPassword=false).
//   - Non-sudo commands are returned unchanged.
func prepareSudoCommand(command, sudoPass string, haveSudoPass bool) (string, bool) {
	if !strings.HasPrefix(strings.TrimSpace(command), "sudo") {
		return command, false
	}
	if haveSudoPass {
		return command, true
	}
	// Insert -n right after "sudo" (and any leading spaces).
	trimmed := strings.TrimSpace(command)
	return "sudo -n" + trimmed[len("sudo"):], false
}
