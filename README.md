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
*   [**Firewall**](docs/plugins/firewall-plugin.md): UFW firewall management with comprehensive rule management.
*   [**Wireguard**](docs/plugins/wireguard.md): Personal VPN with QR code setup.
*   [**MySQL/MariaDB**](docs/plugins/mysql.md): Database management with user and database operations.
*   [**WordPress**](docs/plugins/wordpress.md): Automated LEMP stack & site deployment.
*   [**Redis**](docs/plugins/redis.md): Redis database server management with backup capabilities.
*   [**Restic**](docs/plugins/restic.md): S3 backup manager with database support.
*   [**Language Runtime**](internal/services/runtimes/README.md): Multiple language runtime management (Node.js, Python, Go, Java, Rust, PHP, Ruby, .NET).


### Plugin Features

#### **Database & Storage**
- **MySQL/MariaDB**: Complete database server with user management
- **Redis**: High-performance in-memory data store with persistence
- **Restic**: S3-compatible backup system

#### **Security & Networking**
- **Firewall**: UFW-based firewall with rule management
- **Fail2Ban**: Intrusion prevention and brute-force protection
- **Wireguard**: VPN server with QR code setup

#### **Web & Applications**
- **WordPress**: Automated LEMP stack installation and site management
- **Nginx**: Web server configuration and SSL management

#### **Containerization**
- **Docker**: Container management with Compose support

#### **Development Tools**
- **Language Runtime**: Multi-language runtime management

#### **Server Management**
- **System**: OS package management and system administration
- **Alias**: Server connection management

## 🛠️ Example Usage

### Managing System Updates
```bash
# Update package lists
vps-init myserver system update

# Upgrade all packages
vps-init myserver system upgrade
```

### Database & Storage Management
```bash
# Install and secure MySQL
vps-init myserver mysql install

# Create a database
vps-init myserver mysql create-db myapp

# Create a database user
vps-init myserver mysql create-user app_user 'strong_password'

# Grant privileges
vps-init myserver mysql grant app_user myapp ALL PRIVILEGES

# Install Redis with backup support
vps-init myserver redis install

# Create Redis backup
vps-init myserver redis backup

# Restic backup management
vps-init myserver restic init s3:mybucket
vps-init myserver restic backup-db mydatabase
```

### Web Server Management
```bash
# Install Nginx with SSL
vps-init myserver nginx install
vps-init myserver nginx install-ssl mydomain.com

# Install WordPress with LEMP stack
vps-init myserver wordpress install

# Install Redis cache for WordPress
vps-init myserver redis install
vps-init myserver firewall allow 6379
vps-init myserver firewall allow 6379
```

### Security & Hardening
```bash
# Install and configure firewall
vps-init myserver firewall install
vps-init myserver firewall enable

# Install fail2ban for brute-force protection
vps-init myserver fail2ban install

# Set up personal VPN
vps-init myserver wireguard install
vps-init myserver wireguard add-client my-device

# Check firewall status
vps-init myserver firewall status

# Review fail2ban status
vps-init myserver fail2ban status
```

### Development Environment
```bash
# Install multiple runtimes for development
vps-init myserver runtime install node 18
vps-init myserver runtime install python 3.11
vps-init myserver runtime install go 1.21
vps-init myserver runtime install java 17
vps-init myserver runtime install rust latest

# Switch between versions
vps-init myserver runtime use node 16
vps-init myserver runtime use python 3.9
```

### Container Management
```bash
# Install Docker and Docker Compose
vps-init myserver docker install

# Deploy a multi-container application
vps-init myserver docker deploy ./docker-compose.yml

# Install Portainer for web UI management
vps-init myserver docker install-portainer
```

### Managing Language Runtime
```bash
# Install Node.js 18 with NVM
vps-init myserver runtime install node 18

# Install Python 3.11 with pyenv
vps-init myserver runtime install python 3.11

# Install Go 1.21
vps-init myserver runtime install go 1.21

# Install Java 17
vps-init myserver runtime install java 17

# List all installed runtime
vps-init myserver runtime list

# Check current active versions
vps-init myserver runtime status

# Switch Node.js versions
vps-init myserver runtime use node 16

# Install multiple languages for a development environment
vps-init myserver runtime install node 18
vps-init myserver runtime install python 3.11
vps-init myserver runtime install go 1.21
vps-init myserver runtime install rust latest
```

### Firewall Management
```bash
# Install and configure UFW firewall with secure defaults
vps-init myserver firewall install

# Allow web traffic
vps-init myserver firewall allow 80
vps-init myserver firewall allow 443

# Allow SSH from specific IP
vps-init myserver firewall allow 22 tcp 192.168.1.100

# Enable firewall
vps-init myserver firewall enable

# Check firewall status
vps-init myserver firewall status

# Deny specific port
vps-init myserver firewall deny 23

# Delete rule by number
vps-init myserver firewall delete 3

# Reset firewall to defaults
vps-init myserver firewall reset

# Configure logging
vps-init myserver firewall logging high
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