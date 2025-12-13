# Changelog

All notable changes to this project will be documented in this file.

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
