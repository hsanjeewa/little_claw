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
	connPass, err := c.vault.GetPassword(task.HostIP)
	if err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to get connection password: %v", err))
	}

	config := &ssh.ClientConfig{
		User: task.User,
		Auth: []ssh.AuthMethod{
			ssh.Password(connPass),
		},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         5 * time.Second,
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
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to dial: %v", err))
	}
	defer conn.Close()

	session, err := conn.NewSession()
	if err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to create session: %v", err))
	}
	defer session.Close()

	modes := ssh.TerminalModes{
		ssh.ECHO:          0,
		ssh.TTY_OP_ISPEED: 14400,
		ssh.TTY_OP_OSPEED: 14400,
	}

	if err := session.RequestPty("vt100", 80, 40, modes); err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("request for pseudo terminal failed: %v", err))
	}

	stdin, err := session.StdinPipe()
	if err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to get stdin pipe: %v", err))
	}
	defer stdin.Close()

	stdout, err := session.StdoutPipe()
	if err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to get stdout pipe: %v", err))
	}

	stderr, err := session.StderrPipe()
	if err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to get stderr pipe: %v", err))
	}

	if err := session.Start(task.Command); err != nil {
		return "", fmt.Errorf("context: %w", fmt.Errorf("failed to start command: %v", err))
	}

	outputChan := make(chan string)
	errorChan := make(chan error)

	go func() {
		var outputBuf bytes.Buffer
		buf := make([]byte, 4096)

		for {
			n, err := stdout.Read(buf)
			if n > 0 {
				chunk := buf[:n]
				outputBuf.Write(chunk)

				chunkStr := strings.ToLower(string(chunk))
				if strings.Contains(chunkStr, "password") || strings.Contains(chunkStr, "[sudo]") {
					sudoPass, err := c.vault.GetPassword(task.HostIP)
					if err == nil {
						_, _ = stdin.Write([]byte(sudoPass + "\n"))
					}
				}
			}
			if err != nil {
				if err == io.EOF {
					break
				}
				errorChan <- fmt.Errorf("context: %w", fmt.Errorf("read error: %v", err))
				return
			}
		}

		stderrBuf := new(bytes.Buffer)
		_, _ = io.Copy(stderrBuf, stderr)
		if stderrBuf.Len() > 0 {
			outputBuf.Write(stderrBuf.Bytes())
		}

		err = session.Wait()
		if err != nil {
			errorChan <- fmt.Errorf("context: %w", fmt.Errorf("command execution failed: %v", err))
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
