package helpers

import (
	"fmt"
	"time"

	"golang.org/x/crypto/ssh"
)

type SSHClient struct {
	client *ssh.Client
	config *SSHConfig
}

type SSHConfig struct {
	Host     string
	Port     int
	User     string
	Password string
	KeyPath  string
}

func NewSSHClient(config *SSHConfig) (*SSHClient, error) {
	sshConfig := &ssh.ClientConfig{
		User:            config.User,
		Auth:            []ssh.AuthMethod{},
		HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		Timeout:         30 * time.Second,
	}

	if config.Password != "" {
		sshConfig.Auth = append(sshConfig.Auth, ssh.Password(config.Password))
	}

	if config.KeyPath != "" {
		key, err := ssh.ParsePrivateKey([]byte(config.KeyPath))
		if err != nil {
			return nil, fmt.Errorf("failed to parse private key: %w", err)
		}
		sshConfig.Auth = append(sshConfig.Auth, ssh.PublicKeys(key))
	}

	addr := fmt.Sprintf("%s:%d", config.Host, config.Port)
	client, err := ssh.Dial("tcp", addr, sshConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to dial: %w", err)
	}

	return &SSHClient{
		client: client,
		config: config,
	}, nil
}

func (c *SSHClient) RunCommand(cmd string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	output, err := session.CombinedOutput(cmd)
	if err != nil {
		return string(output), fmt.Errorf("command failed: %w", err)
	}

	return string(output), nil
}

func (c *SSHClient) RunCommandWithOutput(cmd string) (stdout, stderr string, err error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", "", fmt.Errorf("failed to create session: %w", err)
	}
	defer session.Close()

	stdoutBytes, err := session.Output(cmd)
	if err != nil {
		return string(stdoutBytes), "", fmt.Errorf("command failed: %w", err)
	}

	return string(stdoutBytes), "", nil
}

func (c *SSHClient) Close() error {
	return c.client.Close()
}

func (c *SSHClient) TestConnection() error {
	_, err := c.RunCommand("echo 'connection test'")
	return err
}
