package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type Config struct {
	configDir string
	aliases   map[string]string
	secrets   map[string]string
}

func New() *Config {
	homeDir, _ := os.UserHomeDir()
	configDir := filepath.Join(homeDir, ".vps-init")

	cfg := &Config{
		configDir: configDir,
		aliases:   make(map[string]string),
		secrets:   make(map[string]string),
	}

	cfg.loadAliases()
	cfg.loadSecrets()
	return cfg
}

func (c *Config) loadAliases() {
	aliasesFile := filepath.Join(c.configDir, "aliases.json")

	if _, err := os.Stat(aliasesFile); os.IsNotExist(err) {
		// Create config directory
		os.MkdirAll(c.configDir, 0755)
		return
	}

	data, err := os.ReadFile(aliasesFile)
	if err != nil {
		return
	}

	json.Unmarshal(data, &c.aliases)
}

func (c *Config) saveAliases() error {
	aliasesFile := filepath.Join(c.configDir, "aliases.json")

	data, err := json.MarshalIndent(c.aliases, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(aliasesFile, data, 0644)
}

func (c *Config) SetAlias(alias, connection string) error {
	c.aliases[alias] = connection
	return c.saveAliases()
}

func (c *Config) GetAlias(alias string) (string, bool) {
	conn, exists := c.aliases[alias]
	return conn, exists
}

func (c *Config) GetAliases() map[string]string {
	return c.aliases
}

func (c *Config) RemoveAlias(alias string) error {
	if _, exists := c.aliases[alias]; !exists {
		return fmt.Errorf("alias '%s' does not exist", alias)
	}

	delete(c.aliases, alias)
	return c.saveAliases()
}

func (c *Config) ResolveTarget(target string) string {
	// If it contains @, it's already a connection string
	if strings.Contains(target, "@") {
		return target
	}

	// Try to resolve as alias
	if conn, exists := c.GetAlias(target); exists {
		return conn
	}

	// Return as-is
	return target
}

// Secrets Management

func (c *Config) loadSecrets() {
	if c.secrets == nil {
		c.secrets = make(map[string]string)
	}

	secretsFile := filepath.Join(c.configDir, "secrets.json")

	if _, err := os.Stat(secretsFile); os.IsNotExist(err) {
		return
	}

	data, err := os.ReadFile(secretsFile)
	if err != nil {
		return
	}

	json.Unmarshal(data, &c.secrets)
}

func (c *Config) saveSecrets() error {
	secretsFile := filepath.Join(c.configDir, "secrets.json")

	data, err := json.MarshalIndent(c.secrets, "", "  ")
	if err != nil {
		return err
	}

	// Use 0600 for secrets
	return os.WriteFile(secretsFile, data, 0600)
}

func (c *Config) SetSecret(alias, password string) error {
	if c.secrets == nil {
		c.secrets = make(map[string]string)
	}
	c.secrets[alias] = password
	return c.saveSecrets()
}

func (c *Config) GetSecret(alias string) (string, bool) {
	if c.secrets == nil {
		return "", false
	}
	pass, exists := c.secrets[alias]
	return pass, exists
}

type Connection struct {
	User string
	Host string
	Port int
}

func ParseConnection(connStr string) (*Connection, error) {
	parts := strings.Split(connStr, "@")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid connection format, expected user@host")
	}

	user := parts[0]
	host := parts[1]

	return &Connection{
		User: user,
		Host: host,
		Port: 22,
	}, nil
}
