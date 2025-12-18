package ssh

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// Connection interface defines the contract for SSH connections
type Connection interface {
	// Basic operations
	RunCommand(cmd string, sudo bool) plugin.Result
	RunCommandWithOutput(cmd string, sudo bool) (string, error)
	UploadFile(localPath, remotePath string) error
	DownloadFile(remotePath, localPath string) error
	Close() error

	// Enhanced SSH operations (from vps-init-ssh)
	Connect() bool
	Disconnect()
	Reconnect() error
	IsHealthy() bool
	GetConnectionStats() *plugin.ConnectionStats

	// Command execution
	RunSudo(cmd, password string) plugin.Result
	RunInteractive(cmd string) error
	Shell() error

	// File operations
	WriteFile(content, path string) error
	WriteFileFromLocal(localPath, remotePath string) error
	AppendFile(content, path string) error
	CopyFile(src, dst string) error
	MoveFile(src, dst string) error
	DeleteFile(path string) error
	CreateDirectory(path string) error
	RemoveDirectory(path string, recursive bool) error
	ListDirectory(path string) plugin.Result
	GetFileInfo(path string) plugin.FileInfo
	ChangePermissions(path, permissions string) error
	ChangeOwner(path, user, group string) error

	// System operations
	FileExists(path string) bool
	DirectoryExists(path string) bool
	Systemctl(action, service string) bool
	InstallPackage(packageName string) bool

	// Platform detection
	IsUbuntu() bool
	IsDebian() bool
	IsCentOS() bool
	IsRedHat() bool

	// Connection info
	User() string
	Host() string
	Port() int
}

// Config holds SSH connection configuration
type Config struct {
	Host         string
	User         string
	Port         int
	IdentityFile string
	SudoPass     string
	Timeout      time.Duration
}

// DefaultConfig returns default SSH configuration
func DefaultConfig() Config {
	return Config{
		Port:    22,
		Timeout: 30 * time.Second,
	}
}

// connection implements the Connection interface
type connection struct {
	config Config
}

// NewConnection creates a new SSH connection
func NewConnection(config Config) Connection {
	return &connection{config: config}
}

// NewConnectionFromAlias creates connection from alias string (user@host[:port])
func NewConnectionFromAlias(alias string) (Connection, error) {
	parts := strings.Split(alias, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid alias format: expected user@host[:port]")
	}

	user := parts[0]
	hostPort := parts[1]

	// Parse port if specified
	host := hostPort
	port := 22
	if strings.Contains(hostPort, ":") {
		hostParts := strings.Split(hostPort, ":")
		if len(hostParts) != 2 {
			return nil, fmt.Errorf("invalid host:port format")
		}
		host = hostParts[0]
		// Convert port string to int with error handling
		var err error
		_, err = fmt.Sscanf(hostParts[1], "%d", &port)
		if err != nil {
			return nil, fmt.Errorf("invalid port number: %v", err)
		}
	}

	config := Config{
		User: user,
		Host: host,
		Port: port,
	}

	return NewConnection(config), nil
}

// Connect establishes SSH connection and returns connection instance
func Connect(config Config) (Connection, error) {
	// Validate configuration
	if config.Host == "" {
		return nil, fmt.Errorf("host is required")
	}
	if config.User == "" {
		return nil, fmt.Errorf("user is required")
	}
	if config.Port == 0 {
		config.Port = 22
	}

	// Test connection
	conn := NewConnection(config)
	if !conn.Connect() {
		return nil, fmt.Errorf("failed to connect to %s@%s:%d", config.User, config.Host, config.Port)
	}

	return conn, nil
}

// ConnectFromAlias connects using alias string
func ConnectFromAlias(alias string) (Connection, error) {
	conn, err := NewConnectionFromAlias(alias)
	if err != nil {
		return nil, err
	}

	config := conn.(*connection).config
	return Connect(config)
}

// RunCommand executes a command on the remote host
func (c *connection) RunCommand(cmd string, sudo bool) plugin.Result {
	if sudo {
		sudoCmd := fmt.Sprintf("sudo -S %s", cmd)
		return c.runCommandWithContext(context.Background(), sudoCmd)
	}
	return c.runCommandWithContext(context.Background(), cmd)
}

// RunCommandWithOutput executes a command and returns output as string
func (c *connection) RunCommandWithOutput(cmd string, sudo bool) (string, error) {
	result := c.RunCommand(cmd, sudo)
	if result.Success {
		return result.Stdout, nil
	}
	return "", fmt.Errorf("command failed: %s", result.Stderr)
}

// RunSudo executes a command with sudo privileges
func (c *connection) RunSudo(cmd, password string) plugin.Result {
	if password == "" {
		return plugin.Result{
			Success: false,
			Error:   "sudo password is required for sudo commands",
		}
	}

	sudoCmd := fmt.Sprintf("echo '%s' | sudo -S %s", password, cmd)
	return c.runCommandWithContext(context.Background(), sudoCmd)
}

// Host returns the remote host
func (c *connection) Host() string {
	return c.config.Host
}

// User returns the username
func (c *connection) User() string {
	return c.config.User
}

// Port returns the port number
func (c *connection) Port() int {
	return c.config.Port
}

// buildSSHArgs builds SSH command arguments
func (c *connection) buildSSHArgs() []string {
	args := []string{
		"-o", "StrictHostKeyChecking=no",
		"-o", "UserKnownHostsFile=/dev/null",
		"-o", "LogLevel=ERROR",
		"-o", "ConnectTimeout=10",
		"-o", "ServerAliveInterval=30",
		"-o", "ServerAliveCountMax=3",
	}

	// Add port if not default
	if c.config.Port != 22 {
		args = append(args, "-p", fmt.Sprintf("%d", c.config.Port))
	}

	// Add identity file if specified
	if c.config.IdentityFile != "" {
		// Expand tilde to home directory
		identityFile := c.config.IdentityFile
		if strings.HasPrefix(identityFile, "~/") {
			home, err := os.UserHomeDir()
			if err == nil {
				identityFile = filepath.Join(home, identityFile[2:])
			}
		}
		args = append(args, "-i", identityFile)
	}

	return args
}

// runCommandWithContext executes a command with context
func (c *connection) runCommandWithContext(ctx context.Context, cmd string) plugin.Result {
	startTime := time.Now()

	// Build SSH command
	sshArgs := c.buildSSHArgs()
	sshArgs = append(sshArgs,
		fmt.Sprintf("%s@%s", c.config.User, c.config.Host),
		cmd,
	)

	// Create command with context if timeout is specified
	command := exec.CommandContext(ctx, "ssh", sshArgs...)

	// Set up buffers
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr

	// Run command
	err := command.Run()

	// Create result
	result := plugin.Result{
		Success:   err == nil,
		Stdout:    stdout.String(),
		Stderr:    stderr.String(),
		ExitCode:  getExitCode(err),
		Duration:  time.Since(startTime).String(),
		Timestamp: startTime.Format(time.RFC3339),
	}

	// Set output/error based on success
	if result.Success {
		result.Output = result.Stdout
	} else {
		result.Error = result.Stderr
	}

	return result
}

// testConnection tests if SSH connection is working
func (c *connection) testConnection() error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := "echo 'connection_test'"
	result := c.runCommandWithContext(ctx, cmd)

	if !result.Success {
		return fmt.Errorf("connection test failed: %s", result.Stderr)
	}

	return nil
}

// getExitCode extracts exit code from error
func getExitCode(err error) int {
	if err == nil {
		return 0
	}

	if exitError, ok := err.(*exec.ExitError); ok {
		return exitError.ExitCode()
	}

	return 1
}

// convertResult converts internal result to plugin result
func (c *connection) convertResult(result result) plugin.Result {
	return plugin.Result{
		Success:   result.exitCode == 0,
		Stdout:    result.stdout,
		Stderr:    result.stderr,
		ExitCode:  result.exitCode,
		Duration:  result.duration.String(),
		Timestamp: result.timestamp.Format(time.RFC3339),
	}
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

// WriteFileFromLocal copies a local file to remote
func (c *connection) WriteFileFromLocal(localPath, remotePath string) error {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %v", err)
	}
	return c.WriteFile(string(content), remotePath)
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

// result represents the result of a command execution
type result struct {
	stdout    string
	stderr    string
	exitCode  int
	duration  time.Duration
	timestamp time.Time
}

// Required interface methods

// UploadFile uploads a local file to remote host
func (c *connection) UploadFile(localPath, remotePath string) error {
	content, err := os.ReadFile(localPath)
	if err != nil {
		return fmt.Errorf("failed to read local file: %v", err)
	}
	return c.WriteFile(string(content), remotePath)
}

// DownloadFile downloads a remote file to local host
func (c *connection) DownloadFile(remotePath, localPath string) error {
	result := c.RunCommand(fmt.Sprintf("cat %s", remotePath), false)
	if !result.Success {
		return fmt.Errorf("failed to read remote file: %s", result.Stderr)
	}

	return os.WriteFile(localPath, []byte(result.Stdout), 0644)
}

// Close closes the SSH connection
func (c *connection) Close() error {
	// For SSH connections, there's no persistent connection to close
	return nil
}

// Connect establishes and tests connection
func (c *connection) Connect() bool {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	cmd := "echo 'connection_test'"
	result := c.runCommandWithContext(ctx, cmd)
	return result.Success
}

// Disconnect disconnects from SSH
func (c *connection) Disconnect() {
	// No-op for SSH
}

// Reconnect reconnects to SSH
func (c *connection) Reconnect() error {
	if c.Connect() {
		return nil
	}
	return fmt.Errorf("reconnection failed")
}

// IsHealthy checks if connection is healthy
func (c *connection) IsHealthy() bool {
	return c.Connect()
}

// GetConnectionStats returns connection statistics
func (c *connection) GetConnectionStats() *plugin.ConnectionStats {
	return &plugin.ConnectionStats{
		ConnectedAt:  time.Now(),
		LastActivity: time.Now(),
		CommandsRun:  0,
		BytesSent:    0,
		BytesReceived: 0,
	}
}

// CreateDirectory creates a directory
func (c *connection) CreateDirectory(path string) error {
	result := c.RunCommand(fmt.Sprintf("mkdir -p %s", path), false)
	if !result.Success {
		return fmt.Errorf("failed to create directory: %s", result.Stderr)
	}
	return nil
}

// RemoveDirectory removes a directory
func (c *connection) RemoveDirectory(path string, recursive bool) error {
	cmd := "rmdir"
	if recursive {
		cmd = "rm -rf"
	}
	result := c.RunCommand(fmt.Sprintf("%s %s", cmd, path), false)
	if !result.Success {
		return fmt.Errorf("failed to remove directory: %s", result.Stderr)
	}
	return nil
}

// ListDirectory lists directory contents
func (c *connection) ListDirectory(path string) plugin.Result {
	return c.RunCommand(fmt.Sprintf("ls -la %s", path), false)
}

// GetFileInfo gets file information
func (c *connection) GetFileInfo(path string) plugin.FileInfo {
	result := c.RunCommand(fmt.Sprintf("stat -c '%%n|%%s|%%Y|%%f' %s", path), false)
	if !result.Success {
		return plugin.FileInfo{
			Name: filepath.Base(path),
		}
	}

	parts := strings.Split(strings.TrimSpace(result.Stdout), "|")
	if len(parts) < 4 {
		return plugin.FileInfo{
			Name: filepath.Base(path),
		}
	}

	// Parse size, modtime, and mode
	var size int64
	mode := os.FileMode(0644)

	if len(parts) > 1 {
		if s, err := fmt.Sscanf(parts[1], "%d", &size); err != nil || s != 1 {
			size = 0
		}
	}

	modTime := time.Now()
	if len(parts) > 2 {
		var timestamp int64
		if n, err := fmt.Sscanf(parts[2], "%d", &timestamp); err == nil && n == 1 {
			modTime = time.Unix(timestamp, 0)
		}
	}

	isDir := false
	if len(parts) > 3 && len(parts[3]) > 0 {
		// Check if it's a directory by checking the first character of file mode
		isDir = strings.HasPrefix(parts[3], "4")
	}

	return plugin.FileInfo{
		Name:        filepath.Base(path),
		Size:        size,
		Mode:        mode,
		ModTime:     modTime,
		IsDir:       isDir,
		Owner:       "unknown",
		Group:       "unknown",
		Permissions: fmt.Sprintf("%o", mode),
	}
}

// ChangePermissions changes file permissions
func (c *connection) ChangePermissions(path, permissions string) error {
	result := c.RunCommand(fmt.Sprintf("chmod %s %s", permissions, path), false)
	if !result.Success {
		return fmt.Errorf("failed to change permissions: %s", result.Stderr)
	}
	return nil
}

// ChangeOwner changes file ownership
func (c *connection) ChangeOwner(path, user, group string) error {
	owner := user
	if group != "" {
		owner = fmt.Sprintf("%s:%s", user, group)
	}
	result := c.RunCommand(fmt.Sprintf("chown %s %s", owner, path), true)
	if !result.Success {
		return fmt.Errorf("failed to change ownership: %s", result.Stderr)
	}
	return nil
}

// FileExists checks if file exists
func (c *connection) FileExists(path string) bool {
	result := c.RunCommand(fmt.Sprintf("test -f %s", path), false)
	return result.Success
}

// DirectoryExists checks if directory exists
func (c *connection) DirectoryExists(path string) bool {
	result := c.RunCommand(fmt.Sprintf("test -d %s", path), false)
	return result.Success
}

// Systemctl manages systemd services
func (c *connection) Systemctl(action, service string) bool {
	result := c.RunCommand(fmt.Sprintf("systemctl %s %s", action, service), true)
	return result.Success
}

// InstallPackage installs a package
func (c *connection) InstallPackage(packageName string) bool {
	result := c.RunCommand(fmt.Sprintf("apt-get install -y %s", packageName), true)
	return result.Success
}

// IsUbuntu checks if system is Ubuntu
func (c *connection) IsUbuntu() bool {
	result := c.RunCommand("lsb_release -si", false)
	return result.Success && strings.Contains(strings.ToUpper(result.Stdout), "UBUNTU")
}

// IsDebian checks if system is Debian
func (c *connection) IsDebian() bool {
	result := c.RunCommand("lsb_release -si", false)
	return result.Success && strings.Contains(strings.ToUpper(result.Stdout), "DEBIAN")
}

// IsCentOS checks if system is CentOS
func (c *connection) IsCentOS() bool {
	result := c.RunCommand("lsb_release -si", false)
	return result.Success && strings.Contains(strings.ToUpper(result.Stdout), "CENTOS")
}

// IsRedHat checks if system is RedHat
func (c *connection) IsRedHat() bool {
	result := c.RunCommand("lsb_release -si", false)
	return result.Success && strings.Contains(strings.ToUpper(result.Stdout), "RED HAT")
}

// Helper function to create connection from SSHConfig (used by config.go)
func newConnectionWithConfig(config *SSHConfig) plugin.Connection {
	connConfig := Config{
		Host:         config.Host,
		User:         config.User,
		Port:         config.Port,
		IdentityFile: config.KeyPath,
		Timeout:      30 * time.Second,
	}
	return NewConnection(connConfig)
}