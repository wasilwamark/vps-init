package ssh

import (
	"time"

	"github.com/wasilwamark/vps-init/pkg/plugin"
)

// SSHConfig holds SSH-specific configuration that extends the VPS-Init plugin config
type SSHConfig struct {
	// Basic connection info
	Host     string
	User     string
	Port     int
	KeyPath  string

	// SSH-specific connection options
	StrictHostKeyChecking bool
	ServerAliveInterval   int
	ServerAliveCountMax   int
	ConnectTimeout        time.Duration

	// Enhanced options from vps-init-core
	HideOutput bool
	Timeout    int // seconds
	EnvVars    map[string]string
	WorkingDir string
}

// SSHOption represents a configuration option for SSH connections
type SSHOption func(*SSHConfig)

// SSH-specific option functions
func WithKeyPath(keyPath string) SSHOption {
	return func(c *SSHConfig) {
		c.KeyPath = keyPath
	}
}

func WithStrictHostKeyChecking(enable bool) SSHOption {
	return func(c *SSHConfig) {
		c.StrictHostKeyChecking = enable
	}
}

func WithServerAliveInterval(seconds int) SSHOption {
	return func(c *SSHConfig) {
		c.ServerAliveInterval = seconds
	}
}

func WithServerAliveCountMax(count int) SSHOption {
	return func(c *SSHConfig) {
		c.ServerAliveCountMax = count
	}
}

func WithConnectTimeout(timeout time.Duration) SSHOption {
	return func(c *SSHConfig) {
		c.ConnectTimeout = timeout
	}
}

// Core configuration option functions
func WithHost(host string) SSHOption {
	return func(c *SSHConfig) {
		c.Host = host
	}
}

func WithUser(user string) SSHOption {
	return func(c *SSHConfig) {
		c.User = user
	}
}

func WithPort(port int) SSHOption {
	return func(c *SSHConfig) {
		c.Port = port
	}
}

// NewSSHConfig creates a default SSH configuration
func NewSSHConfig() *SSHConfig {
	return &SSHConfig{
		Port:                   22,
		StrictHostKeyChecking:  false,
		ServerAliveInterval:    60,
		ServerAliveCountMax:    3,
		ConnectTimeout:         30 * time.Second,
		EnvVars:                make(map[string]string),
	}
}

// ApplySSHOptions applies SSH-specific options to a configuration
func (c *SSHConfig) ApplySSHOptions(options ...SSHOption) {
	for _, opt := range options {
		opt(c)
	}
}

// Clone creates a copy of the configuration
func (c *SSHConfig) Clone() *SSHConfig {
	clone := &SSHConfig{
		Host:                  c.Host,
		User:                  c.User,
		Port:                  c.Port,
		KeyPath:               c.KeyPath,
		StrictHostKeyChecking:  c.StrictHostKeyChecking,
		ServerAliveInterval:    c.ServerAliveInterval,
		ServerAliveCountMax:    c.ServerAliveCountMax,
		ConnectTimeout:         c.ConnectTimeout,
		HideOutput:            c.HideOutput,
		Timeout:               c.Timeout,
		WorkingDir:            c.WorkingDir,
		EnvVars:               make(map[string]string),
	}

	// Copy environment variables
	for k, v := range c.EnvVars {
		clone.EnvVars[k] = v
	}

	return clone
}

// SSHConnector interface for creating connections
type SSHConnector interface {
	NewConnection(options ...SSHOption) plugin.Connection
}

// DefaultConnector is the default implementation of SSHConnector
type DefaultConnector struct{}

// NewConnector creates a new SSH connector
func NewConnector() SSHConnector {
	return &DefaultConnector{}
}

// NewConnection creates a new SSH connection with the given options
func (dc *DefaultConnector) NewConnection(options ...SSHOption) plugin.Connection {
	config := NewSSHConfig()
	config.ApplySSHOptions(options...)
	return newConnectionWithConfig(config)
}