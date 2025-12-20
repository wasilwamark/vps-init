# Changelog

All notable changes to this project will be documented in this file.

## [Unreleased] - 2025-12-20

### 🐛 Bug Fixes

*   **SSH Module**: Improve WriteFile and AppendFile escaping to prevent configuration errors
    *   Enhanced escaping mechanisms for single quotes in file content
    *   More robust file writing operations over SSH connections

### ✨ Improvements

*   **Documentation**: Added project logo to README.md for better branding
*   **Developer Guide**: Updated CLAUDE.md with git operations policy and enhanced development guidelines

---

## [v0.0.2] - 2025-12-14

### 🚀 New Plugins

*   **WireGuard Plugin**:
    *   `install`: Install WireGuard and tools (wireguard, wireguard-tools, qrencode)
    *   `setup`: Configure WireGuard Server with interactive setup
    *   `add-peer <name>`: Add new client/peer with QR code generation
    *   `remove-peer`: Interactive peer removal with confirmation and backup
    *   `list-peers`: List configured peers with device names
    *   `status`: Show WireGuard service and interface status
*   **Docker Plugin**:
    *   `install`: Install Docker and Docker Compose
    *   `status`: Check Docker service status
*   **Nginx Plugin**:
    *   `install`: Install Nginx web server
    *   `list-sites`: List configured virtual hosts
    *   `status`: Show Nginx service status
*   **Fail2Ban Plugin**:
    *   `install`: Install and configure Fail2Ban
    *   `status`: Show Fail2Ban service status
*   **MySQL Plugin**:
    *   `install`: Install MySQL database server
    *   `create-db <name>`: Create new database
    *   `create-user <user> <password>`: Create new MySQL user
    *   `status`: Show MySQL service status
*   **WordPress Plugin**:
    *   `install <domain>`: Install WordPress with Nginx configuration
    *   `site-list`: List WordPress installations
*   **Restic Plugin**:
    *   `install`: Install Restic backup tool
    *   `init <repo>`: Initialize new backup repository
    *   `backup <path> <repo>`: Create backup with support for MySQL, PostgreSQL, MongoDB
    *   `snapshots <repo>`: List backup snapshots
    *   `restore <snapshot> <target> <repo>`: Restore from backup

### ✨ Improvements

*   **System Plugin Enhancements**:
    *   Added `shell` command to open interactive shell on remote server
    *   Added `install` command to install vps-init on remote server
    *   Added `uninstall` command to remove vps-init from remote server
*   **WireGuard Enhancements**:
    *   Enhanced `list-peers` to display device names alongside public keys for better readability
    *   Interactive peer removal with numbered selection menu
    *   Automatic backup creation before peer removal with timestamp
    *   QR code generation for easy mobile client configuration
*   **External Plugin Loading**: Support for loading external plugins from custom paths

### 🐛 Bug Fixes

*   Fixed WireGuard peer addition regression
*   Fixed Nginx plugin installation issue
*   Improved configuration file parsing for WireGuard peer management

### ⚙️ Infrastructure

*   Enhanced plugin architecture for better extensibility
*   Improved error handling and user feedback across all plugins
*   Better SSH connection management and timeout handling

---

## [v0.0.1] - 2025-12-13

Initial release of VPS-Init, a CLI tool for simple server management over SSH.

### 🚀 Features

*   **Plugin Architecture**: Modular design to support various services.
*   **System Plugin**:
    *   `system update`: Update package lists.
    *   `system upgrade`: Upgrade installed packages.
    *   `system full-upgrade`: Perform distribution upgrades.
    *   `system autoremove`: Clean up unused packages.
*   **Alias Management**:
    *   `alias add`: Save server connection details.
    *   `alias list`: View configured servers.
    *   `alias remove`: Delete aliases.
*   **Secure Sudo Handling**:
    *   Support for running root-level commands safely.
    *   **Secrets Store**: Automatically save passwords securely to `~/.vps-init/secrets.json` when adding aliases with `--sudo-password`.
    *   **Environment Variables**: Support for `SSH_SUDO_PWD_<ALIAS>` for CI/CD or session-based overrides.
*   **Direct Execution**: Run commands directly against servers using aliases or raw `user@host` strings.

### 📚 Documentation

*   Comprehensive `README.md` with quick start and architecture overview.
*   Detailed `docs/plugins/system.md` usage guide.
*   Developer guide for creating new plugins.

### ⚙️ Infrastructure

*   GitHub Actions workflow for automated releases.
*   GoReleaser configuration for multi-platform builds (Linux/macOS).
