package plugin

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// GitProvider implements VCSProvider for git repositories
type GitProvider struct{}

// NewGitProvider creates a new git provider
func NewGitProvider() *GitProvider {
	return &GitProvider{}
}

// Clone clones a git repository
func (g *GitProvider) Clone(ctx context.Context, repoURL, targetPath string, options CloneOptions) error {
	args := []string{"clone"}

	if options.Recursive {
		args = append(args, "--recursive")
	}

	if options.Depth > 0 {
		args = append(args, "--depth", fmt.Sprintf("%d", options.Depth))
	}

	if options.SingleBranch {
		args = append(args, "--single-branch")
	}

	if options.Branch != "" {
		args = append(args, "--branch", options.Branch)
	}

	args = append(args, repoURL, targetPath)

	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// Checkout checks out a specific version, branch, or commit
func (g *GitProvider) Checkout(ctx context.Context, path, version string) error {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "checkout", version)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	return cmd.Run()
}

// GetTags gets all tags from a repository
func (g *GitProvider) GetTags(ctx context.Context, path string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "tag", "-l")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get tags: %w", err)
	}

	tags := strings.Split(strings.TrimSpace(string(output)), "\n")
	if len(tags) == 1 && tags[0] == "" {
		return []string{}, nil
	}

	return tags, nil
}

// GetBranches gets all branches from a repository
func (g *GitProvider) GetBranches(ctx context.Context, path string) ([]string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "branch", "-a")
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("failed to get branches: %w", err)
	}

	branches := strings.Split(strings.TrimSpace(string(output)), "\n")
	var result []string

	for _, branch := range branches {
		branch = strings.TrimSpace(branch)
		// Remove "* " prefix for current branch and "remotes/origin/" prefix for remote branches
		branch = strings.TrimPrefix(branch, "* ")
		branch = strings.TrimPrefix(branch, "remotes/origin/")
		if branch != "" && branch != "HEAD" {
			result = append(result, branch)
		}
	}

	return result, nil
}

// GetCommit gets the current commit hash
func (g *GitProvider) GetCommit(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "rev-parse", "HEAD")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get commit: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// GetRemoteURL gets the remote URL for a repository
func (g *GitProvider) GetRemoteURL(ctx context.Context, path string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", "-C", path, "config", "--get", "remote.origin.url")
	output, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("failed to get remote URL: %w", err)
	}

	return strings.TrimSpace(string(output)), nil
}

// IsValidRepository checks if a directory is a valid git repository
func (g *GitProvider) IsValidRepository(path string) bool {
	gitDir := filepath.Join(path, ".git")
	if stat, err := os.Stat(gitDir); err == nil {
		return stat.IsDir() || stat.Mode().IsRegular() // .git could be a file (git worktree)
	}
	return false
}