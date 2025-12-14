package ssh

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Connection struct {
	Host      string
	User      string
	Port      int
	KeyPath   string
	connected bool
}

func New(host, user string) *Connection {
	return &Connection{
		Host:    host,
		User:    user,
		Port:    22,
		KeyPath: findDefaultKey(),
	}
}

func findDefaultKey() string {
	home, _ := os.UserHomeDir()
	possibleKeys := []string{
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
	}

	for _, key := range possibleKeys {
		if _, err := os.Stat(key); err == nil {
			return key
		}
	}

	return possibleKeys[0] // fallback
}

func (s *Connection) Connect() bool {
	result := s.RunCommand("echo 'connection-test'", true)
	s.connected = result.Success
	return result.Success
}

func (s *Connection) Disconnect() {
	s.connected = false
}

func (s *Connection) RunCommand(cmd string, check bool) *CommandResult {
	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-p", fmt.Sprintf("%d", s.Port),
		"-i", s.KeyPath,
		fmt.Sprintf("%s@%s", s.User, s.Host),
		cmd,
	}

	command := exec.Command("ssh", sshArgs...)

	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	err := command.Run()

	result := &CommandResult{
		Command: cmd,
		Stdout:  stdout.String(),
		Stderr:  stderr.String(),
		Success: err == nil,
	}

	if check && err != nil {
		result.Error = err
	}

	return result
}

func (s *Connection) RunSudo(cmd, password string) *CommandResult {
	if password == "" {
		// Try non-interactive sudo (this will fail if password is required)
		return s.RunCommand(fmt.Sprintf("sudo -n %s", cmd), false)
	}

	// Use sudo with password from stdin
	// -S: read password from stdin
	// -p '': empty prompt
	// Wrap command in sh -c to handle complex commands with &&, ||, etc.
	fullCmd := fmt.Sprintf("echo '%s' | sudo -S -p '' sh -c '%s'", password, cmd)

	// We mask the command in the result to avoid leaking password in logs if we had them
	// But RunCommand doesn't mask inputs.
	// For security, we should be careful.
	// However, since we are sending this over SSH, we construct the command string.

	result := s.RunCommand(fullCmd, false)

	// If failed, it might be wrong password or something else.
	// Clean up the command in the result so we don't accidentally print the password if result is printed
	result.Command = fmt.Sprintf("sudo %s", cmd)

	return result
}

func (s *Connection) WriteFile(content, path string) bool {
	// Use cat with heredoc to write file
	cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", path, content)
	result := s.RunCommand(cmd, false)
	return result.Success
}

func (s *Connection) AppendFile(content, path string) bool {
	cmd := fmt.Sprintf("cat >> %s << 'EOF'\n%s\nEOF", path, content)
	result := s.RunCommand(cmd, false)
	return result.Success
}

func (s *Connection) FileExists(path string) bool {
	cmd := fmt.Sprintf("test -f %s", path)
	result := s.RunCommand(cmd, false)
	return result.Success
}

func (s *Connection) DirectoryExists(path string) bool {
	cmd := fmt.Sprintf("test -d %s", path)
	result := s.RunCommand(cmd, false)
	return result.Success
}

func (s *Connection) InstallPackage(packageName string) bool {
	commands := []string{
		"apt-get update",
		fmt.Sprintf("apt-get install -y %s", packageName),
	}

	for _, cmd := range commands {
		result := s.RunCommand(cmd, false)
		if !result.Success {
			return false
		}
	}
	return true
}

func (s *Connection) Systemctl(action, service string) bool {
	cmd := fmt.Sprintf("systemctl %s %s", action, service)
	result := s.RunCommand(cmd, false)
	return result.Success
}

func (s *Connection) IsUbuntu() bool {
	result := s.RunCommand("cat /etc/os-release | grep -i ubuntu", false)
	return result.Success && strings.Contains(result.Stdout, "ubuntu")
}

type CommandResult struct {
	Command string
	Stdout  string
	Stderr  string
	Success bool
	Error   error
}

// RunInteractive runs a command and streams stdout/stderr to the current process
// It also connects stdin.
func (s *Connection) RunInteractive(cmd string) error {
	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ServerAliveInterval=60",
		"-o", "ServerAliveCountMax=120",
		"-p", fmt.Sprintf("%d", s.Port),
		"-t", // Force pseudo-terminal allocation for interactive feeling
		"-i", s.KeyPath,
		fmt.Sprintf("%s@%s", s.User, s.Host),
		cmd,
	}

	command := exec.Command("ssh", sshArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

// Shell opens an interactive shell session
func (s *Connection) Shell() error {
	sshArgs := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "ServerAliveInterval=60",
		"-o", "ServerAliveCountMax=120",
		"-p", fmt.Sprintf("%d", s.Port),
		"-t", // Force pseudo-terminal usage
		"-i", s.KeyPath,
		fmt.Sprintf("%s@%s", s.User, s.Host),
	}

	command := exec.Command("ssh", sshArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}
