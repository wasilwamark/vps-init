package plugin

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/Masterminds/semver/v3"
)

// GitInstaller handles plugin installation from git repositories
type GitInstaller struct {
	config      InstallerConfig
	vcs         VCSProvider
	validator   *Validator
	builders    map[string]Builder
}

// InstallerConfig contains configuration for plugin installation
type InstallerConfig struct {
	CacheDir      string
	InstallDir    string
	MaxDepth      int
	Timeout       time.Duration
	AllowInsecure bool
}

// VCSProvider defines interface for version control operations
type VCSProvider interface {
	Clone(ctx context.Context, repoURL, targetPath string, options CloneOptions) error
	Checkout(ctx context.Context, path, version string) error
	GetTags(ctx context.Context, path string) ([]string, error)
	GetBranches(ctx context.Context, path string) ([]string, error)
	GetCommit(ctx context.Context, path string) (string, error)
	GetRemoteURL(ctx context.Context, path string) (string, error)
	IsValidRepository(path string) bool
}

// CloneOptions contains options for cloning a repository
type CloneOptions struct {
	Depth       int
	Branch      string
	Recursive   bool
	SingleBranch bool
}

// Builder defines interface for building plugins
type Builder interface {
	Build(ctx context.Context, sourcePath, outputPath string, options BuildOptions) (*BuildResult, error)
	CanBuild(sourcePath string) bool
	GetLanguage() string
}

// BuildOptions contains options for building a plugin
type BuildOptions struct {
	GoVersion  string
	Flags      []string
	Tags       []string
	LDFlags    string
	Env        map[string]string
	OutputName string
}

// BuildResult contains the result of building a plugin
type BuildResult struct {
	BinaryPath string
	Checksum   string
	Metadata   PluginMetadata
	Warnings   []string
	BuildTime  time.Duration
}

// NewGitInstaller creates a new git-based plugin installer
func NewGitInstaller(config InstallerConfig) *GitInstaller {
	return &GitInstaller{
		config:    config,
		vcs:       NewGitProvider(),
		validator: NewValidator("1.0.0"), // Get from build info
		builders: make(map[string]Builder),
	}
}

// RepositoryInfo contains information about a git repository
type RepositoryInfo struct {
	URL       string
	Host      string // github.com, gitlab.com, etc.
	Owner     string
	Repo      string
	Protocol  string // https, ssh, git
	IsPrivate bool
}

// ParseRepositoryURL parses a git repository URL into components
func ParseRepositoryURL(repoURL string) (*RepositoryInfo, error) {
	// Remove .git suffix if present
	repoURL = strings.TrimSuffix(repoURL, ".git")

	// Handle different URL formats
	var parsed *url.URL
	var err error

	// Try parsing as URL first (https, git+https, etc.)
	if strings.Contains(repoURL, "://") {
		parsed, err = url.Parse(repoURL)
		if err != nil {
			return nil, fmt.Errorf("invalid repository URL: %w", err)
		}
	} else {
		// Handle scp-like URLs (git@github.com:user/repo.git)
		scpRegex := regexp.MustCompile(`^(?:(?P<user>[^@]+)@)?(?P<host>[^:]+):(?P<path>.+)$`)
		matches := scpRegex.FindStringSubmatch(repoURL)
		if len(matches) == 0 {
			return nil, fmt.Errorf("invalid repository URL format")
		}

		host := matches[scpRegex.SubexpIndex("host")]
		path := matches[scpRegex.SubexpIndex("path")]

		parsed = &url.URL{
			Scheme: "ssh",
			Host:   host,
			Path:   path,
		}
	}

	info := &RepositoryInfo{
		URL:      repoURL,
		Host:     parsed.Host,
		Protocol: parsed.Scheme,
	}

	// Extract owner and repo from path
	pathParts := strings.Split(strings.Trim(parsed.Path, "/"), "/")
	if len(pathParts) < 2 {
		return nil, fmt.Errorf("repository path must include owner and repository name")
	}

	info.Owner = pathParts[0]
	info.Repo = pathParts[1]

	// Determine if private based on protocol or known patterns
	info.IsPrivate = parsed.Scheme == "ssh" || parsed.Scheme == "git"

	return info, nil
}

// InstallOptions contains options for installing a plugin
type InstallOptions struct {
	Version      string
	Branch       string
	Commit       string
	Name         string
	Force        bool
	NoVerify     bool
	BuildOptions BuildOptions
}

// Install installs a plugin from a git repository
func (gi *GitInstaller) Install(ctx context.Context, repoURL string, options InstallOptions) (*PluginMetadata, error) {
	// Parse repository URL
	repoInfo, err := ParseRepositoryURL(repoURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse repository URL: %w", err)
	}

	// Create cache directory if it doesn't exist
	if err := os.MkdirAll(gi.config.CacheDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create cache directory: %w", err)
	}

	// Clone repository
	cachePath := filepath.Join(gi.config.CacheDir, fmt.Sprintf("%s-%s", repoInfo.Owner, repoInfo.Repo))
	if err := gi.cloneRepository(ctx, repoURL, cachePath, options); err != nil {
		return nil, fmt.Errorf("failed to clone repository: %w", err)
	}

	// Checkout specific version if requested
	if options.Version != "" || options.Branch != "" || options.Commit != "" {
		if err := gi.checkoutVersion(ctx, cachePath, options); err != nil {
			return nil, fmt.Errorf("failed to checkout version: %w", err)
		}
	}

	// Discover plugin in repository
	pluginPath, metadata, err := gi.discoverPlugin(ctx, cachePath)
	if err != nil {
		return nil, fmt.Errorf("failed to discover plugin: %w", err)
	}

	// Validate plugin
	if !options.NoVerify {
		if validationErrors := gi.validator.ValidatePluginMetadata(metadata); len(validationErrors) > 0 {
			return nil, fmt.Errorf("plugin validation failed: %v", validationErrors)
		}
	}

	// Build plugin
	buildResult, err := gi.buildPlugin(ctx, pluginPath, options.BuildOptions)
	if err != nil {
		return nil, fmt.Errorf("failed to build plugin: %w", err)
	}

	// Install plugin
	installPath, err := gi.installPlugin(ctx, buildResult, options.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to install plugin: %w", err)
	}

	// Update metadata with installation info
	metadata.InstallPath = installPath
	metadata.InstalledAt = time.Now().Format(time.RFC3339)
	metadata.Checksum = buildResult.Checksum
	metadata.Source = repoURL
	metadata.Validated = true
	metadata.TrustLevel = GetTrustLevel(metadata)

	return &metadata, nil
}

// cloneRepository clones a git repository to the cache directory
func (gi *GitInstaller) cloneRepository(ctx context.Context, repoURL, cachePath string, options InstallOptions) error {
	// Check if already exists
	if _, err := os.Stat(cachePath); err == nil {
		if !options.Force {
			// Pull latest changes
			return gi.vcs.Checkout(ctx, cachePath, "HEAD")
		}
		// Remove existing directory
		os.RemoveAll(cachePath)
	}

	cloneOptions := CloneOptions{
		Depth:       gi.config.MaxDepth,
		Branch:      options.Branch,
		Recursive:   true,
		SingleBranch: options.Branch != "",
	}

	return gi.vcs.Clone(ctx, repoURL, cachePath, cloneOptions)
}

// checkoutVersion checks out a specific version, branch, or commit
func (gi *GitInstaller) checkoutVersion(ctx context.Context, cachePath string, options InstallOptions) error {
	version := options.Version
	if version == "" {
		version = options.Branch
	}
	if version == "" {
		version = options.Commit
	}

	if version == "" {
		return nil // No specific version requested
	}

	// Try to resolve version to tag, branch, or commit
	if err := gi.vcs.Checkout(ctx, cachePath, version); err != nil {
		// If version looks like a semantic version, try to find matching tag
		if _, err := semver.NewVersion(version); err == nil {
			tags, err := gi.vcs.GetTags(ctx, cachePath)
			if err != nil {
				return fmt.Errorf("failed to get tags: %w", err)
			}

			for _, tag := range tags {
				if strings.TrimPrefix(tag, "v") == strings.TrimPrefix(version, "v") {
					return gi.vcs.Checkout(ctx, cachePath, tag)
				}
			}
		}

		return fmt.Errorf("version %s not found", version)
	}

	return nil
}

// discoverPlugin finds the plugin in a cloned repository
func (gi *GitInstaller) discoverPlugin(ctx context.Context, repoPath string) (string, PluginMetadata, error) {
	// Look for plugin.yaml in root
	metadataPath := filepath.Join(repoPath, "plugin.yaml")
	if _, err := os.Stat(metadataPath); err != nil {
		// Look in subdirectories
		entries, err := os.ReadDir(repoPath)
		if err != nil {
			return "", PluginMetadata{}, fmt.Errorf("failed to read repository: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				testPath := filepath.Join(repoPath, entry.Name(), "plugin.yaml")
				if _, err := os.Stat(testPath); err == nil {
					metadataPath = testPath
					repoPath = filepath.Join(repoPath, entry.Name())
					break
				}
			}
		}

		if _, err := os.Stat(metadataPath); err != nil {
			return "", PluginMetadata{}, fmt.Errorf("plugin.yaml not found in repository")
		}
	}

	// Read and parse metadata
	data, err := os.ReadFile(metadataPath)
	if err != nil {
		return "", PluginMetadata{}, fmt.Errorf("failed to read plugin metadata: %w", err)
	}

	var metadata PluginMetadata
	if err := parseYAML(data, &metadata); err != nil {
		return "", PluginMetadata{}, fmt.Errorf("failed to parse plugin metadata: %w", err)
	}

	// Validate metadata
	if validationErrors := gi.validator.ValidatePluginMetadata(metadata); len(validationErrors) > 0 {
		return "", PluginMetadata{}, fmt.Errorf("plugin metadata validation failed: %v", validationErrors)
	}

	return repoPath, metadata, nil
}

// buildPlugin builds a plugin from source
func (gi *GitInstaller) buildPlugin(ctx context.Context, pluginPath string, options BuildOptions) (*BuildResult, error) {
	// Find appropriate builder
	var builder Builder
	for _, b := range gi.builders {
		if b.CanBuild(pluginPath) {
			builder = b
			break
		}
	}

	if builder == nil {
		return nil, fmt.Errorf("no suitable builder found for plugin at %s", pluginPath)
	}

	// Create temporary output directory
	tempDir, err := os.MkdirTemp("", "vps-init-build-*")
	if err != nil {
		return nil, fmt.Errorf("failed to create temporary directory: %w", err)
	}
	defer os.RemoveAll(tempDir)

	// Build plugin
	result, err := builder.Build(ctx, pluginPath, tempDir, options)
	if err != nil {
		return nil, fmt.Errorf("build failed: %w", err)
	}

	// Calculate checksum
	checksum, err := calculateChecksum(result.BinaryPath)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate checksum: %w", err)
	}
	result.Checksum = checksum

	return result, nil
}

// installPlugin installs a built plugin to the plugins directory
func (gi *GitInstaller) installPlugin(ctx context.Context, result *BuildResult, name string) (string, error) {
	if name == "" {
		name = result.Metadata.Name
	}

	installDir := filepath.Join(gi.config.InstallDir, name)
	if err := os.MkdirAll(installDir, 0755); err != nil {
		return "", fmt.Errorf("failed to create install directory: %w", err)
	}

	// Copy binary
	binaryName := name + ".so"
	installPath := filepath.Join(installDir, binaryName)

	if err := copyFile(result.BinaryPath, installPath); err != nil {
		return "", fmt.Errorf("failed to copy plugin binary: %w", err)
	}

	// Write metadata
	metadataPath := filepath.Join(installDir, "plugin.yaml")
	if err := writeYAML(metadataPath, result.Metadata); err != nil {
		return "", fmt.Errorf("failed to write plugin metadata: %w", err)
	}

	return installPath, nil
}

// calculateChecksum calculates SHA256 checksum of a file
func calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// copyFile copies a file from src to dst
func copyFile(src, dst string) error {
	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()

	_, err = io.Copy(destination, source)
	return err
}

// parseYAML parses YAML data (placeholder implementation)
func parseYAML(data []byte, v interface{}) error {
	// Import and use yaml.Unmarshal here
	return nil
}

// writeYAML writes data as YAML (placeholder implementation)
func writeYAML(path string, data interface{}) error {
	// Import and use yaml.Marshal here
	return nil
}