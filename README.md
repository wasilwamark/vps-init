# VPS-Init

<div align="center">

![VPS-Init Logo](https://via.placeholder.com/200x80/333/fff?text=VPS-Init)

**A Go-based CLI tool for quick and easy server configuration**

No Terraform required - just SSH and commands!

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/wasilwamark/vps-init)

[Quick Start](#-quick-start) • [Services](#-available-services) • [Examples](#-examples) • [Development](#-development)

</div>

## 🚀 Quick Start

### Prerequisites

- Go 1.21 or higher
- SSH access to your servers
- Ubuntu/Debian servers (currently supported)

### Installation

```bash
# Clone the repository
git clone https://github.com/wasilwamark/vps-init
cd vps-init

# Build and install
make install

# Or build manually
go build -o vps-init ./cmd/vps-init
sudo cp vps-init /usr/local/bin/
```

### First Use

```bash
# Add your server as an alias (optional but recommended)
vps-init alias add myserver mark@1.2.3.4

# Install services on your server
vps-init myserver nginx install
vps-init myserver docker install
vps-init myserver monitoring setup

# Install with SSL for your domain
vps-init mark@1.2.3.4 nginx install-ssl api.tiza.africa
```

## 📋 Available Services

### 🌐 Nginx Web Server
- `install` - Install and start Nginx
- `install-ssl <domain>` - Install Nginx with Let's Encrypt SSL
- `create-site <domain>` - Create a new site configuration

**What it does:**
- Installs Nginx from official repositories
- Configures proxy to localhost:3000 (perfect for Node.js apps)
- Sets up SSL with Let's Encrypt
- Adds security headers
- Auto-starts on boot

### 🐳 Docker Platform
- `install` - Install Docker Engine and Docker Compose

**What it does:**
- Installs latest Docker Engine
- Installs Docker Compose v2
- Adds user to docker group
- Creates standard directories (`/opt/docker/`)
- Auto-starts on boot

### 📊 Monitoring
- `setup` - Install monitoring and alerting

**What it does:**
- Installs monitoring tools (htop, iotop)
- Sets up disk usage alerts (80% threshold)
- Sets up memory usage alerts (85% threshold)
- Installs Uptime Kuma dashboard (port 3001)
- Creates web-based monitoring dashboard

### 🔧 Aliases Management
- `alias add <name> <user@host>` - Add server alias
- `alias list` - List all aliases
- `alias remove <name>` - Remove alias

## 🛠️ Examples

### Complete Web Server Setup
```bash
# Step 1: Add server alias
vps-init alias add production deploy@prod-server.com

# Step 2: Install Nginx with SSL for your API
vps-init production nginx install-ssl api.tiza.africa

# Step 3: Install Docker for your app
vps-init production docker install

# Step 4: Set up monitoring
vps-init production monitoring setup

# Step 5: Deploy your app (now you can SSH in and deploy)
ssh deploy@prod-server.com
cd /opt/docker
docker-compose up -d
```

### Development Server
```bash
# Quick dev setup
vps-init mark@dev-server.com docker install
vps-init mark@dev-server.com nginx create-site dev.myapp.com

# No SSL for dev
```

### Multiple Domains on One Server
```bash
# Main site with SSL
vps-init production nginx install-ssl myapp.com

# API subdomain with SSL
vps-init production nginx create-site api.myapp.com
# Then manually get SSL: ssh production "certbot --nginx -d api.myapp.com"
```

## 🏗️ How It Works

VPS-Init simplifies server management:

```
Your Command → SSH Connection → Direct Server Configuration
     ↓                ↓                     ↓
Simple CLI    → Secure Connection → Package Install + Config Write
```

1. **SSH Connection**: Uses your SSH keys (no passwords stored)
2. **Package Management**: Leverages apt-get/yum directly
3. **Configuration**: Writes proper config files to standard locations
4. **Service Control**: Uses systemctl for service management

**No layers of abstraction** - just the commands you'd run manually, automated!

## 📁 Project Structure

```
vps-init/
├── cmd/
│   └── vps-init/           # Main CLI application entry point
├── internal/
│   ├── cli/               # CLI command handlers
│   │   ├── root.go        # Root command and alias management
│   │   ├── nginx.go       # Nginx commands
│   │   ├── docker.go      # Docker commands
│   │   └── monitoring.go  # Monitoring commands
│   ├── config/            # Configuration and alias storage
│   ├── services/          # Service implementation modules
│   │   ├── nginx/         # Nginx installation and configuration
│   │   ├── docker/        # Docker setup
│   │   └── monitoring/    # Monitoring tools setup
│   └── ssh/               # SSH connection management
├── Makefile               # Build and installation instructions
├── go.mod                 # Go module dependencies
└── README.md              # This file
```

## 🔧 Development

### Build Commands

```bash
# Build for current platform
make build

# Build for all platforms (Linux, macOS, Windows)
make build-all

# Install to system PATH
make install

# Development build with debug info
make dev

# Run tests
make test

# Format code
make fmt

# Clean build artifacts
make clean
```

### Adding New Services

1. Create service module in `internal/services/yourservice/`
2. Implement CLI commands in `internal/cli/yourservice.go`
3. Add commands to root CLI in `internal/cli/root.go`
4. Update documentation

Example service structure:
```go
// internal/services/yourservice/service.go
type Service struct {
    ssh *ssh.Connection
}

func New(ssh *ssh.Connection) *Service {
    return &Service{ssh: ssh}
}

func (s *Service) Install() bool {
    // Your installation logic here
    return s.ssh.InstallPackage("your-package")
}
```

## 📦 What Gets Installed Where

### Nginx
- **Package**: `nginx`
- **Config**: `/etc/nginx/sites-available/`
- **SSL**: `/etc/letsencrypt/live/`
- **Logs**: `/var/log/nginx/`

### Docker
- **Packages**: `docker-ce`, `docker-ce-cli`, `containerd.io`
- **Compose**: `/usr/local/bin/docker-compose`
- **Data**: `/opt/docker/`
- **Sockets**: `/var/run/docker.sock`

### Monitoring
- **Tools**: `htop`, `iotop`
- **Scripts**: `/opt/monitoring/`
- **Logs**: `/var/log/monitoring/`
- **Uptime Kuma**: `/opt/uptime-kuma/`
- **Dashboard**: `/var/www/html/monitoring.html`

## 🔒 Security Considerations

- ✅ **SSH Key Authentication** - No passwords stored
- ✅ **Minimal Privileges** - Only installs packages, no system changes
- ✅ **Transparent Commands** - All commands visible in SSH logs
- ✅ **No Data Collection** - Tool runs locally, no telemetry
- ✅ **Open Source** - All code visible and auditable

## 🆘 Troubleshooting

### SSH Connection Issues
```bash
# Test SSH connection manually
ssh -v user@your-server

# Check if SSH key is loaded
ssh-add -l

# Add SSH key if needed
ssh-add ~/.ssh/id_rsa
```

### Permission Issues
```bash
# VPS-Init needs sudo access for package installation
# Test if you can install packages:
ssh user@server "sudo apt-get update"
```

### Service Not Starting
```bash
# Check service status
ssh user@server "systemctl status nginx"
ssh user@server "systemctl status docker"

# Check logs
ssh user@server "journalctl -u nginx -f"
```

### SSL Certificate Issues
```bash
# Check Let's Encrypt status
ssh user@server "certbot certificates"

# Manually obtain certificate
ssh user@server "certbot --nginx -d yourdomain.com"
```

## 🤝 Contributing

We love contributions! Here's how to help:

1. **Fork** the repository
2. **Create** a feature branch (`git checkout -b feature/amazing-feature`)
3. **Implement** your changes
4. **Add** tests if applicable
5. **Update** documentation
6. **Submit** a pull request

### Development Workflow
```bash
# Set up development environment
git clone https://github.com/wasilwamark/vps-init
cd vps-init
go mod download

# Make your changes
# ... edit files ...

# Test your changes
make test
make dev
./bin/vps-init-dev --help

# Submit PR
```

## 📚 Additional Resources

- [SSH Key Management](https://www.ssh.com/ssh/key/)
- [Nginx Configuration](https://nginx.org/en/docs/)
- [Docker Documentation](https://docs.docker.com/)
- [Let's Encrypt SSL](https://letsencrypt.org/)

## 🐛 Reporting Issues

Found a bug? Please open an issue with:

- **Description**: What happened and what you expected
- **Command**: The exact command you ran
- **Output**: Any error messages
- **Environment**: OS, Go version, server distro

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🙏 Acknowledgments

- [Cobra](https://github.com/spf13/cobra) - CLI framework
- [Terraform](https://www.terraform.io/) - Inspiration for declarative approach
- The Go community for excellent tooling

---

<div align="center">

**Built with ❤️ for developers who want simplicity**

[⭐ Star this repo](https://github.com/wasilwamark/vps-init) • [🐛 Report Issues](https://github.com/wasilwamark/vps-init/issues) • [💬 Discussions](https://github.com/wasilwamark/vps-init/discussions)

</div>