# Language Runtimes Plugin

The Language Runtimes plugin provides a unified interface for managing programming language runtimes on your VPS. It supports multiple popular programming languages and handles version management using standard version managers.

## Supported Languages

- **Node.js** - Managed via NVM (Node Version Manager)
- **Python** - Managed via pyenv
- **Go** - Direct installation
- **Java** - OpenJDK via apt (versions 8, 11, 17, 21)
- **Rust** - Managed via rustup
- **PHP** - Via apt and PPA (multiple versions)
- **Ruby** - Managed via rbenv
- **.NET** - Microsoft .NET SDK

## Commands

### Install a Language Runtime

```bash
# Install Node.js version 18
vps-init <server> runtimes install node 18

# Install Python 3.11
vps-init <server> runtimes install python 3.11

# Install Go 1.21
vps-init <server> runtimes install go 1.21

# Install Java 17
vps-init <server> runtimes install java 17

# Install Rust
vps-init <server> runtimes install rust latest

# Install PHP 8.2
vps-init <server> runtimes install php 8.2

# Install Ruby 3.2
vps-init <server> runtimes install ruby 3.2

# Install .NET 7
vps-init <server> runtimes install dotnet 7
```

### List Installed Runtimes

```bash
vps-init <server> runtimes list
```

### Switch Between Versions

```bash
# Switch to Node.js 16
vps-init <server> runtimes use node 16

# Switch to Python 3.10
vps-init <server> runtimes use python 3.10
```

### Show Current Status

```bash
vps-init <server> runtimes status
```

### Remove a Runtime Version

```bash
# Remove Node.js 14
vps-init <server> runtimes remove node 14

# Remove Python 3.9
vps-init <server> runtimes remove python 3.9
```

### Update Version Managers

```bash
vps-init <server> runtimes update
```

## Features

### Automatic Version Managers

The plugin automatically installs and configures version managers for supported languages:

- **NVM** for Node.js - Allows installing multiple Node.js versions
- **pyenv** for Python - Allows installing multiple Python versions
- **rbenv** for Ruby - Allows installing multiple Ruby versions
- **rustup** for Rust - Manages Rust toolchains

### Environment Setup

The plugin automatically:
- Updates shell profiles (`~/.bashrc`) with necessary environment variables
- Configures PATH for all installed runtimes
- Sets up version-specific aliases where applicable

### Version Detection

The plugin can:
- Detect existing installations
- Show current active versions
- List all installed versions for each language

## Examples

### Setting up a Node.js Development Environment

```bash
# Install Node.js 18 with NVM
vps-init myserver runtimes install node 18

# Switch to Node.js 16 for a legacy project
vps-init myserver runtimes use node 16

# Check current status
vps-init myserver runtimes status
```

### Setting up a Python Environment

```bash
# Install Python 3.11 with pyenv
vps-init myserver runtimes install python 3.11

# List available Python versions
vps-init myserver runtimes list

# Install additional packages if needed
vps-init myserver system cmd "pip install virtualenv"
```

### Multi-Language Setup

```bash
# Install multiple language runtimes
vps-init myserver runtimes install node 18
vps-init myserver runtimes install python 3.11
vps-init myserver runtimes install go 1.21
vps-init myserver runtimes install java 17

# Check all installed runtimes
vps-init myserver runtimes status
```

## Dependencies

The plugin requires the following system dependencies:
- `curl` - For downloading installers
- `wget` - For downloading packages
- `git` - For cloning version manager repositories

These dependencies are automatically checked and installed by the plugin.

## Notes

- For Go and Java, version switching is limited (requires reinstallation for Go, uses `update-alternatives` for Java)
- All installations are performed in the user's home directory when using version managers
- System-wide installations are performed for languages that support it (Java, PHP, .NET)
- The plugin respects existing installations and won't overwrite them unless explicitly requested