package ssh

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// connection implements the plugin.Connection interface
type connection struct {
	config    *SSHConfig
	connected bool
	stats     *plugin.ConnectionStats
}

// New creates a new SSH connection with host and user (backward compatibility)
func New(host, user string) plugin.Connection {
	config := NewSSHConfig()
	config.Host = host
	config.User = user
	return newConnectionWithConfig(config)
}

// NewWithOptions creates a new SSH connection with options
func NewWithOptions(options ...SSHOption) plugin.Connection {
	config := NewSSHConfig()
	config.ApplySSHOptions(options...)
	return newConnectionWithConfig(config)
}

// newConnectionWithConfig creates a new SSH connection with a configuration
func newConnectionWithConfig(config *SSHConfig) plugin.Connection {
	if config.KeyPath == "" {
		config.KeyPath = findDefaultKey()
	}

	now := time.Now()
	return &connection{
		config: config,
		stats: &plugin.ConnectionStats{
			ConnectedAt:  now,
			LastActivity: now,
		},
	}
}

// Connect tests the SSH connection
func (c *connection) Connect() bool {
	ctx, cancel := context.WithTimeout(context.Background(), c.config.ConnectTimeout)
	defer cancel()

	result := c.runCommandWithContext(ctx, "echo 'connection-test'")
	c.connected = result.Success
	return result.Success
}

// Disconnect closes the connection
func (c *connection) Disconnect() {
	c.connected = false
}

// RunCommand executes a command over SSH (legacy interface)
func (c *connection) RunCommand(cmd string, sudo bool) plugin.Result {
	if sudo {
		// Use non-interactive sudo (this will fail if password is required)
		result := c.runCommandWithContext(context.Background(), fmt.Sprintf("sudo -n %s", cmd))
		return c.convertResult(result)
	}

	result := c.runCommandWithContext(context.Background(), cmd)
	return c.convertResult(result)
}

// RunCommandWithOutput executes a command and returns stdout/stderr
func (c *connection) RunCommandWithOutput(cmd string, sudo bool) (string, error) {
	result := c.RunCommand(cmd, sudo)
	if result.Success {
		return result.Output, nil
	}
	return "", fmt.Errorf("command failed: %s", result.Error)
}

// UploadFile uploads a file to the remote host
func (c *connection) UploadFile(localPath, remotePath string) error {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %v", err)
	}

	escapedContent := strings.ReplaceAll(string(content), "'", "'\"'\"'")
	cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", remotePath, escapedContent)

	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to upload file: %s", result.Stderr)
	}
	return nil
}

// DownloadFile downloads a file from the remote host
func (c *connection) DownloadFile(remotePath, localPath string) error {
	// Create temp file path
	tempFile := filepath.Join(os.TempDir(), filepath.Base(remotePath))

	// Use SCP to download the file
	scpArgs := []string{
		"-i", c.config.KeyPath,
		"-P", fmt.Sprintf("%d", c.config.Port),
	}

	if !c.config.StrictHostKeyChecking {
		scpArgs = append(scpArgs,
			"-o", "StrictHostKeyChecking=no",
			"-o", "UserKnownHostsFile=/dev/null",
			"-o", "LogLevel=ERROR",
		)
	}

	scpArgs = append(scpArgs,
		fmt.Sprintf("%s@%s:%s", c.config.User, c.config.Host, remotePath),
		tempFile,
	)

	command := exec.Command("scp", scpArgs...)
	if err := command.Run(); err != nil {
		return fmt.Errorf("failed to download file: %v", err)
	}

	// Move temp file to final location
	if err := os.Rename(tempFile, localPath); err != nil {
		os.Remove(tempFile) // Clean up temp file
		return fmt.Errorf("failed to save file: %v", err)
	}

	return nil
}

// Close closes the connection (alias for Disconnect)
func (c *connection) Close() error {
	c.Disconnect()
	return nil
}

// Helper methods for advanced features (from vps-init-ssh)

// RunInteractive runs a command and streams stdout/stderr to the current process
func (c *connection) RunInteractive(cmd string) error {
	sshArgs := c.buildSSHArgs()
	sshArgs = append(sshArgs,
		"-t", // Force pseudo-terminal allocation for interactive feeling
		fmt.Sprintf("%s@%s", c.config.User, c.config.Host),
		cmd,
	)

	command := exec.Command("ssh", sshArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

// Shell opens an interactive shell session
func (c *connection) Shell() error {
	sshArgs := c.buildSSHArgs()
	sshArgs = append(sshArgs,
		"-t", // Force pseudo-terminal usage
		fmt.Sprintf("%s@%s", c.config.User, c.config.Host),
	)

	command := exec.Command("ssh", sshArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

// WriteFile writes content to a file on the remote host
func (c *connection) WriteFile(content, path string) error {
	// Use cat with heredoc to write file
	// Properly escape any single quotes in the content
	escapedContent := strings.ReplaceAll(content, "'", "'\"'\"'")
	cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", path, escapedContent)

	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to write file: %s", result.Stderr)
	}
	return nil
}

// AppendFile appends content to a file on the remote host
func (c *connection) AppendFile(content, path string) error {
	// Use cat with heredoc to append file
	escapedContent := strings.ReplaceAll(content, "'", "'\"'\"'")
	cmd := fmt.Sprintf("cat >> %s << 'EOF'\n%s\nEOF", path, escapedContent)

	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to append file: %s", result.Stderr)
	}
	return nil
}

// FileExists checks if a file exists on the remote host
func (c *connection) FileExists(path string) bool {
	cmd := fmt.Sprintf("test -f %s", path)
	result := c.runCommandWithContext(context.Background(), cmd)
	return result.Success
}

// DirectoryExists checks if a directory exists on the remote host
func (c *connection) DirectoryExists(path string) bool {
	cmd := fmt.Sprintf("test -d %s", path)
	result := c.runCommandWithContext(context.Background(), cmd)
	return result.Success
}

// Systemctl manages systemd services
func (c *connection) Systemctl(action, service string) bool {
	cmd := fmt.Sprintf("systemctl %s %s", action, service)
	result := c.runCommandWithContext(context.Background(), cmd)
	return result.Success
}

// InstallPackage installs a package using apt-get
func (c *connection) InstallPackage(packageName string) bool {
	commands := []string{
		"apt-get update",
		fmt.Sprintf("apt-get install -y %s", packageName),
	}

	for _, cmd := range commands {
		result := c.runCommandWithContext(context.Background(), cmd)
		if !result.Success {
			return false
		}
	}
	return true
}

// IsUbuntu checks if the remote system is Ubuntu
func (c *connection) IsUbuntu() bool {
	result := c.runCommandWithContext(context.Background(), "cat /etc/os-release | grep -i ubuntu")
	return result.Success && strings.Contains(result.Stdout, "ubuntu")
}

// IsDebian checks if the remote system is Debian
func (c *connection) IsDebian() bool {
	result := c.runCommandWithContext(context.Background(), "cat /etc/os-release | grep -i debian")
	return result.Success && strings.Contains(result.Stdout, "debian")
}

// IsCentOS checks if the remote system is CentOS
func (c *connection) IsCentOS() bool {
	result := c.runCommandWithContext(context.Background(), "cat /etc/os-release | grep -i centos")
	return result.Success && strings.Contains(result.Stdout, "centos")
}

// IsRedHat checks if the remote system is Red Hat Enterprise Linux
func (c *connection) IsRedHat() bool {
	result := c.runCommandWithContext(context.Background(), "cat /etc/os-release | grep -i 'red hat\\|rhel'")
	return result.Success && (strings.Contains(result.Stdout, "red hat") || strings.Contains(result.Stdout, "rhel"))
}

// User returns the SSH user
func (c *connection) User() string {
	return c.config.User
}

// Host returns the SSH host
func (c *connection) Host() string {
	return c.config.Host
}

// Port returns the SSH port
func (c *connection) Port() int {
	return c.config.Port
}

// Internal helper methods

// runCommandWithContext executes a command with context support
func (c *connection) runCommandWithContext(ctx context.Context, cmd string) *enhancedResult {
	// Update stats
	c.stats.LastActivity = time.Now()

	// Build command with environment variables and working directory from SSH config
	if len(c.config.EnvVars) > 0 {
		var envVars []string
		for k, v := range c.config.EnvVars {
			envVars = append(envVars, fmt.Sprintf("%s='%s'", k, strings.ReplaceAll(v, "'", "'\"'\"'")))
		}
		cmd = fmt.Sprintf("export %s; %s", strings.Join(envVars, " "), cmd)
	}

	if c.config.WorkingDir != "" {
		cmd = fmt.Sprintf("cd '%s' && %s", c.config.WorkingDir, cmd)
	}

	// Build SSH command arguments
	sshArgs := c.buildSSHArgs()
	sshArgs = append(sshArgs, fmt.Sprintf("%s@%s", c.config.User, c.config.Host), cmd)

	// Create the SSH command
	command := exec.CommandContext(ctx, "ssh", sshArgs...)

	var stdout, stderr strings.Builder
	command.Stdout = &stdout
	command.Stderr = &stderr

	// Run the command
	err := command.Run()

	// Create result
	result := &enhancedResult{
		Success:  err == nil,
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: 0,
		Command:  cmd,
		Duration: time.Since(c.stats.LastActivity),
		Timestamp: c.stats.LastActivity,
		Host:     c.config.Host,
		User:     c.config.User,
	}

	if err != nil {
		if exitError, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitError.ExitCode()
		} else {
			result.ExitCode = -1
		}
	}

	// Update command stats
	c.stats.CommandsRun++

	// Print output if not hidden
	if result.Success && result.Stdout != "" && !c.config.HideOutput {
		fmt.Print(result.Stdout)
	}

	return result
}

// buildSSHArgs builds the base SSH command arguments
func (c *connection) buildSSHArgs() []string {
	args := []string{
		"-p", fmt.Sprintf("%d", c.config.Port),
		"-i", c.config.KeyPath,
		"-o", fmt.Sprintf("ServerAliveInterval=%d", c.config.ServerAliveInterval),
		"-o", fmt.Sprintf("ServerAliveCountMax=%d", c.config.ServerAliveCountMax),
	}

	if c.config.StrictHostKeyChecking {
		args = append(args, "-o", "StrictHostKeyChecking=yes")
	} else {
		args = append(args, "-o", "StrictHostKeyChecking=no")
		args = append(args, "-o", "UserKnownHostsFile=/dev/null")
		args = append(args, "-o", "LogLevel=ERROR")
	}

	return args
}

// findDefaultKey finds the default SSH key
func findDefaultKey() string {
	home, _ := os.UserHomeDir()
	possibleKeys := []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_rsa"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
		filepath.Join(home, ".ssh", "id_dsa"),
	}

	for _, key := range possibleKeys {
		if _, err := os.Stat(key); err == nil {
			return key
		}
	}

	// Return default path even if it doesn't exist
	return possibleKeys[1] // id_rsa as fallback
}

// convertResult converts enhanced result to plugin.Result
func (c *connection) convertResult(r *enhancedResult) plugin.Result {
	return plugin.Result{
		Success:   r.Success,
		Output:    r.Stdout,
		Error:     r.Stderr,
		Stdout:    r.Stdout,
		Stderr:    r.Stderr,
		ExitCode:  r.ExitCode,
		Duration:  r.Duration.String(),
		Timestamp: r.Timestamp.Format(time.RFC3339),
	}
}

// enhancedResult is our internal result structure that matches vps-init-core
type enhancedResult struct {
	Success   bool
	Stdout    string
	Stderr    string
	ExitCode  int
	Command   string
	Duration  time.Duration
	Timestamp time.Time
	Host      string
	User      string
}

// Implement the remaining Connection interface methods

// RunSudo executes a command with sudo
func (c *connection) RunSudo(cmd, password string) plugin.Result {
	if password == "" {
		// Try non-interactive sudo (this will fail if password is required)
		result := c.runCommandWithContext(context.Background(), fmt.Sprintf("sudo -n %s", cmd))
		return c.convertResult(result)
	}

	// Use sudo with password from stdin
	// -S: read password from stdin
	// -p '': empty prompt
	// Wrap command in sh -c to handle complex commands with &&, ||, etc.
	fullCmd := fmt.Sprintf("echo '%s' | sudo -S -p '' sh -c '%s'", password, cmd)

	// Execute with the command masked in the result to avoid leaking password
	result := c.runCommandWithContext(context.Background(), fullCmd)
	// Hide the password in the command field
	result.Command = fmt.Sprintf("sudo -S %s", cmd)
	return c.convertResult(result)
}

// RunInteractive runs a command and streams stdout/stderr to the current process
func (c *connection) RunInteractive(cmd string) error {
	sshArgs := c.buildSSHArgs()
	sshArgs = append(sshArgs,
		"-t", // Force pseudo-terminal allocation for interactive feeling
		fmt.Sprintf("%s@%s", c.config.User, c.config.Host),
		cmd,
	)

	command := exec.Command("ssh", sshArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

// Shell opens an interactive shell session
func (c *connection) Shell() error {
	sshArgs := c.buildSSHArgs()
	sshArgs = append(sshArgs,
		"-t", // Force pseudo-terminal usage
		fmt.Sprintf("%s@%s", c.config.User, c.config.Host),
	)

	command := exec.Command("ssh", sshArgs...)
	command.Stdout = os.Stdout
	command.Stderr = os.Stderr
	command.Stdin = os.Stdin

	return command.Run()
}

// WriteFile writes content to a file on the remote host
func (c *connection) WriteFile(content, path string) error {
	// Use cat with heredoc to write file
	// Properly escape any single quotes in the content
	escapedContent := strings.ReplaceAll(content, "'", "'\"'\"'")
	cmd := fmt.Sprintf("cat > %s << 'EOF'\n%s\nEOF", path, escapedContent)

	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to write file: %s", result.Stderr)
	}
	return nil
}

// WriteFileFromLocal copies a local file to remote
func (c *connection) WriteFileFromLocal(localPath, remotePath string) error {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %v", err)
	}
	return c.WriteFile(string(content), remotePath)
}

// AppendFile appends content to a file on the remote host
func (c *connection) AppendFile(content, path string) error {
	// Use cat with heredoc to append file
	escapedContent := strings.ReplaceAll(content, "'", "'\"'\"'")
	cmd := fmt.Sprintf("cat >> %s << 'EOF'\n%s\nEOF", path, escapedContent)

	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to append file: %s", result.Stderr)
	}
	return nil
}

// CopyFile copies a file on the remote host
func (c *connection) CopyFile(src, dst string) error {
	cmd := fmt.Sprintf("cp '%s' '%s'", src, dst)
	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to copy file: %s", result.Stderr)
	}
	return nil
}

// MoveFile moves/renames a file on the remote host
func (c *connection) MoveFile(src, dst string) error {
	cmd := fmt.Sprintf("mv '%s' '%s'", src, dst)
	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to move file: %s", result.Stderr)
	}
	return nil
}

// DeleteFile deletes a file on the remote host
func (c *connection) DeleteFile(path string) error {
	cmd := fmt.Sprintf("rm -f '%s'", path)
	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to delete file: %s", result.Stderr)
	}
	return nil
}

// CreateDirectory creates a directory on the remote host
func (c *connection) CreateDirectory(path string) error {
	// Create with -p flag to handle parent directories
	cmd := fmt.Sprintf("mkdir -p '%s'", path)
	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to create directory: %s", result.Stderr)
	}
	return nil
}

// RemoveDirectory removes a directory on the remote host
func (c *connection) RemoveDirectory(path string, recursive bool) error {
	cmd := "rmdir"
	if recursive {
		cmd = "rm -rf"
	}
	cmd += fmt.Sprintf(" '%s'", path)

	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to remove directory: %s", result.Stderr)
	}
	return nil
}

// ListDirectory lists the contents of a directory
func (c *connection) ListDirectory(path string) plugin.Result {
	cmd := fmt.Sprintf("ls -la '%s'", path)
	result := c.runCommandWithContext(context.Background(), cmd)
	return c.convertResult(result)
}

// GetFileInfo gets file information
func (c *connection) GetFileInfo(path string) plugin.FileInfo {
	// Simplified implementation - return basic info
	return plugin.FileInfo{
		Name: filepath.Base(path),
		// In a full implementation, would parse stat output to get real file info
	}
}

// ChangePermissions changes file permissions
func (c *connection) ChangePermissions(path, permissions string) error {
	cmd := fmt.Sprintf("chmod %s '%s'", permissions, path)
	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to change permissions: %s", result.Stderr)
	}
	return nil
}

// ChangeOwner changes file owner
func (c *connection) ChangeOwner(path, user, group string) error {
	owner := user
	if group != "" {
		owner += ":" + group
	}
	cmd := fmt.Sprintf("chown %s '%s'", owner, path)
	result := c.runCommandWithContext(context.Background(), cmd)
	if !result.Success {
		return fmt.Errorf("failed to change owner: %s", result.Stderr)
	}
	return nil
}

// Reconnect attempts to reconnect to the SSH server
func (c *connection) Reconnect() error {
	c.Disconnect()
	if c.Connect() {
		return nil
	}
	return fmt.Errorf("failed to reconnect")
}

// IsHealthy checks if the connection is healthy and responsive
func (c *connection) IsHealthy() bool {
	if !c.connected {
		return false
	}
	// Try to run a simple command to test responsiveness
	result := c.runCommandWithContext(context.Background(), "echo 'health-check'")
	return result.Success
}

// GetConnectionStats returns basic connection statistics
func (c *connection) GetConnectionStats() *plugin.ConnectionStats {
	c.stats.LastActivity = time.Now()
	return c.stats
}