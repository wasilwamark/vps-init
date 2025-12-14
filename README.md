# VPS-Init

<div align="center">

![VPS-Init Logo](https://via.placeholder.com/200x80/333/fff?text=VPS-Init)

**A Go-based CLI tool for quick and easy server configuration**

**No Agents • No Terraform • Just SSH**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/wasilwamark/vps-init)

</div>

## About

VPS-Init is a lightweight CLI tool designed to simplify server management. Unlike complex infrastructure-as-code tools like Terraform or Ansible, VPS-Init works directly over SSH to run common configuration tasks. It's perfect for developers who manage a few VPS instances and want a quick, standardized way to update systems and install services without setting up a control node.

## 🚀 Installation & Quick Start

### Prerequisites
- Go 1.21+
- SSH access to your servers (Ubuntu/Debian supported)

### Installation

```bash
# Clone and Install
git clone https://github.com/wasilwamark/vps-init
cd vps-init
make install
```

### Quick Start

1.  **Add your server alias** (and optionally store sudo password):
    ```bash
    vps-init alias add myserver user@1.2.3.4 --sudo-password 'my-secret'
    ```

2.  **Run a command**:
    ```bash
    vps-init myserver system update
    ```

## 🏗️ How It Works

VPS-Init is agentless. It simply:
1.  **Connects** to your server via standard SSH (using your local SSH keys).
2.  **Executes** commands or scripts remotely.
3.  **Injects** sudo passwords securely only when needed (never stored in plain text history).
4.  **Disconnects** immediately after the task is done.

## 📦 Available Plugins

VPS-Init is built on a modular plugin architecture.

### Core Plugins

*   [**System Management**](docs/plugins/system.md): Update your OS packages, upgrade, and clean up.
*   [**Alias Manager**](docs/plugins/alias.md): Manage server aliases for quick access.

### Service Plugins

*   [**Nginx**](docs/plugins/nginx.md): Web server and reverse proxy management.
*   [**Docker**](docs/plugins/docker.md): Container management with Docker Engine and Compose.
*   [**Fail2Ban**](docs/plugins/fail2ban.md): Brute-force protection.
*   [**Wireguard**](docs/plugins/wireguard.md): Personal VPN with QR code setup.
*   [**MySQL/MariaDB**](docs/plugins/mysql.md): Database management.
*   [**WordPress**](docs/plugins/wordpress.md): Automated LEMP stack & site deployment.
*   [**Restic**](docs/plugins/restic.md): S3 Backups for Files and Databases.
*   [**Language Runtimes**](internal/services/runtimes/README.md): Manage programming language runtimes (Node.js, Python, Go, Java, Rust, PHP, Ruby, .NET).

### System Utilities

*   [**Firewall**](docs/plugins/system.md): (See System/UFW) - *Note: UFW is currently under System/Firewall*

## 🛠️ Example Usage

### Managing System Updates
```bash
# Update package lists
vps-init myserver system update

# Upgrade all packages
vps-init myserver system upgrade
```

### Managing Language Runtimes
```bash
# Install Node.js 18 with NVM
vps-init myserver runtimes install node 18

# Install Python 3.11 with pyenv
vps-init myserver runtimes install python 3.11

# Install Go 1.21
vps-init myserver runtimes install go 1.21

# Install Java 17
vps-init myserver runtimes install java 17

# List all installed runtimes
vps-init myserver runtimes list

# Check current active versions
vps-init myserver runtimes status

# Switch Node.js versions
vps-init myserver runtimes use node 16

# Install multiple languages for a development environment
vps-init myserver runtimes install node 18
vps-init myserver runtimes install python 3.11
vps-init myserver runtimes install go 1.21
vps-init myserver runtimes install rust latest
```

### Managing Aliases
```bash
# List all configured servers
vps-init alias list

# Add a new one
vps-init alias add dev ubuntu@dev.example.com
```

### Sudo Password Handling
If you didn't save the password during alias add, you can use an environment variable:
```bash
export SSH_SUDO_PWD_DEV='secretpass'
vps-init dev system update
```

## 🤝 Contributing

We welcome contributions! Whether it's adding new plugins or fixing bugs.

1.  **Fork** the repository.
2.  **Create** a branch (`git checkout -b feature/cool-plugin`).
3.  **Implement** your changes (See `docs/PLUGIN_DEVELOPMENT.md`).
4.  **Submit** a Pull Request.

---

<div align="center">

[⭐ Star on GitHub](https://github.com/wasilwamark/vps-init) • [🐛 Report Issues](https://github.com/wasilwamark/vps-init/issues)

</div>