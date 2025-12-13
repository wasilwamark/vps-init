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

## � Available Plugins

VPS-Init is built on a modular plugin architecture.

### [System Management](docs/plugins/system.md)
Update your OS packages, upgrade, and clean up.
- **Documentation**: [See docs/plugins/system.md](docs/plugins/system.md)
- **Commands**: `update`, `upgrade`, `full-upgrade`, `autoremove`
*   [**System Management**](docs/plugins/system.md): Update your OS packages, upgrade, and clean up.
*   [**Nginx**](docs/plugins/nginx.md): Web server and reverse proxy management.
*   [**Docker**](docs/plugins/docker.md): Container management.
*   [**Fail2Ban**](docs/plugins/fail2ban.md): Brute-force protection.
*   [**Firewall**](docs/plugins/system.md): (See System/UFW) - *Note: UFW is currently under System/Firewall*

## 🛠️ Example Usage

### Managing System Updates
```bash
# Update package lists
vps-init myserver system update

# Upgrade all packages
vps-init myserver system upgrade
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