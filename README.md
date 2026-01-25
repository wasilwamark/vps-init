# VPS-Init

<div align="center">

<img src="./vps-init-logo.png" width="200" alt="VPS-Init Logo">

**A CLI tool for Easy Server Management**

**SSH all the way**

[![Go Version](https://img.shields.io/badge/Go-1.21+-blue.svg)](https://golang.org)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)
[![Build Status](https://img.shields.io/badge/Build-Passing-brightgreen.svg)](https://github.com/wasilwamark/vps-init)

</div>

## About

VPS-Init manages your servers over SSH. Quick, standardized server configuration without complex IaC tools.

## Installation

**Prerequisites**

- Go 1.21+
- SSH access to your server

**Install**

```bash
git clone https://github.com/wasilwamark/vps-init
cd vps-init
make install
```

**Add Server & Command**

```bash
vps-init alias add myserver user@1.2.3.4 --sudo-password 'password'
vps-init myserver system update
```

## How It Works

VPS-Init connects via SSH, executes commands, and disconnects. Simple as that.

## Plugins

**Core**

- [System](internal/services/system): OS package management

**Services**

- [Nginx](internal/services/nginx): Web server
- [MySQL/MariaDB](internal/services/mysql): Database
- [Redis](internal/services/redis): Cache
- [Fail2Ban](internal/services/fail2ban): Security
- [Wireguard](internal/services/wireguard): VPN
- [Restic](internal/services/restic): Backup
- [Firewall](internal/services/firewall): Firewall (UFW/Firewalld)
- [WordPress](internal/services/wordpress): CMS
- [Keycloak](internal/services/keycloak): Identity
- [Docker](internal/services/docker): Containers

## Example Usage

```bash
# System updates
vps-init myserver system update
vps-init myserver system upgrade

# Web server
vps-init myserver nginx install
vps-init myserver nginx install-ssl mydomain.com

# Database
vps-init myserver mysql install
vps-init myserver mysql create-db myapp

# Firewall
vps-init myserver firewall install
vps-init myserver firewall allow 80
```

## Contributing

Fork, branch, PR.
