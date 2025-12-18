package plugin

import (
	"context"
	"os"
	"time"
)

// Enhanced components copied from vps-init-core
// This file contains additional types and utilities that extend the base plugin system

// Enhanced system information structures copied from vps-init-core

type FileInfo struct {
	Name         string
	Size         int64
	Mode         os.FileMode
	ModTime      time.Time
	IsDir        bool
	Owner        string
	Group        string
	Permissions  string
}

type PlatformInfo struct {
	OS           string
	Version      string
	Architecture string
	Kernel       string
	IsDocker     bool
	IsVM         bool
}

type MemoryInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Available uint64
	Cached    uint64
	Buffers   uint64
}

type DiskInfo struct {
	Total     uint64
	Used      uint64
	Free      uint64
	Path      string
	MountPath string
	Filesystem string
}

type CPUInfo struct {
	Model     string
	Cores     int
	Frequency float64
	Usage     float64
}

type ConnectionStats struct {
	ConnectedAt    time.Time
	LastActivity   time.Time
	CommandsRun    int
	BytesSent      int64
	BytesReceived  int64
}

// Enhanced configuration options copied from vps-init-core

type EnhancedOption func(*EnhancedConfig)

type EnhancedConfig struct {
	// Basic command options
	HideOutput bool
	Timeout    int // seconds

	// Extended command options
	EnvVars     map[string]string
	WorkingDir  string
	Context     context.Context
	Handler     ErrorHandler
	RetryCount  int
	RetryDelay  time.Duration
}

// WithHideOutput configures command to hide output
func WithHideOutput() EnhancedOption {
	return func(c *EnhancedConfig) {
		c.HideOutput = true
	}
}

// WithTimeout sets command execution timeout
func WithTimeout(timeout int) EnhancedOption {
	return func(c *EnhancedConfig) {
		c.Timeout = timeout
	}
}

// WithWorkingDir sets the working directory for command execution
func WithWorkingDir(dir string) EnhancedOption {
	return func(c *EnhancedConfig) {
		c.WorkingDir = dir
	}
}

// WithEnvVar sets an environment variable for command execution
func WithEnvVar(key, value string) EnhancedOption {
	return func(c *EnhancedConfig) {
		if c.EnvVars == nil {
			c.EnvVars = make(map[string]string)
		}
		c.EnvVars[key] = value
	}
}

// WithContext sets the context for command execution
func WithContext(ctx context.Context) EnhancedOption {
	return func(c *EnhancedConfig) {
		c.Context = ctx
	}
}

// WithErrorHandler configures error handling
func WithErrorHandler(handler ErrorHandler) EnhancedOption {
	return func(c *EnhancedConfig) {
		c.Handler = handler
	}
}

// WithRetry configures retry logic
func WithRetry(attempts int, delay time.Duration) EnhancedOption {
	return func(c *EnhancedConfig) {
		c.RetryCount = attempts
		c.RetryDelay = delay
	}
}

// File operation options
type FileOption func(*FileConfig)

type FileConfig struct {
	Permissions string
	Owner       string
	Group       string
	Backup      bool
}

func WithFilePermissions(perm string) FileOption {
	return func(fc *FileConfig) {
		fc.Permissions = perm
	}
}

func WithFileOwner(user, group string) FileOption {
	return func(fc *FileConfig) {
		fc.Owner = user
		fc.Group = group
	}
}

func WithBackup(enabled bool) FileOption {
	return func(fc *FileConfig) {
		fc.Backup = enabled
	}
}

// User management options
type UserOption func(*UserConfig)

type UserConfig struct {
	Shell     string
	HomeDir   string
	Groups    []string
	Password  bool
}

func WithUserShell(shell string) UserOption {
	return func(uc *UserConfig) {
		uc.Shell = shell
	}
}

func WithUserHomeDir(homeDir string) UserOption {
	return func(uc *UserConfig) {
		uc.HomeDir = homeDir
	}
}

func WithUserGroups(groups []string) UserOption {
	return func(uc *UserConfig) {
		uc.Groups = groups
	}
}

func WithUserPasswordRequired(required bool) UserOption {
	return func(uc *UserConfig) {
		uc.Password = required
	}
}

// Error handling function type
type ErrorHandler func(*Result)

// Batch operation types
type BatchResult struct {
	Results []*Result
	Errors  []error
	Summary *BatchSummary
}

type BatchSummary struct {
	TotalCommands   int
	Successful      int
	Failed          int
	Duration        time.Duration
}

type ParallelResult struct {
	Results map[string]*Result
	Errors  map[string]error
	Summary *BatchSummary
}

// Version information structure
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit,omitempty"`
	BuildDate string `json:"build_date,omitempty"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// String returns the version string
func (v VersionInfo) String() string {
	version := "vps-init v" + v.Version
	if v.GitCommit != "" {
		shortCommit := v.GitCommit
		if len(shortCommit) > 8 {
			shortCommit = shortCommit[:8]
		}
		version += " (" + shortCommit + ")"
	}
	if v.BuildDate != "" {
		version += " built " + v.BuildDate
	}
	return version
}