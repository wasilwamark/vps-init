# Plugin Roadmap

This document outlines the planned plugins to be implemented for VPS-Init. We will implement these one by one to build a comprehensive VPS management suite.

## 0. Server Upgrades

*   [x] **System Upgrade**: Keep the server up-to-date.
    *   *Features*: Update, Upgrade, Full-Upgrade, Autoremove, Shell Access.

## 1. Web Servers

*   [x] **Nginx**: High-performance web server, reverse proxy, and load balancer.
    *   *Features*: Install, Configure Sites, SSL (Certbot integration).
*   [ ] **Apache**: Classic, robust web server.
    *   *Features*: Install, Virtual Hosts, Module management.
*   [ ] **Caddy**: Modern web server with automatic HTTPS.
    *   *Features*: Install, Caddyfile management.

## 2. Databases

*   [ ] **PostgreSQL**: Advanced open-source relational database.
    *   *Features*: Install, Create User/DB, Backup/Restore.
*   [ ] **MySQL / MariaDB**: Popular relational database.
    *   *Features*: Install, Secure Installation, User management.
*   [ ] **Redis**: In-memory data store.
    *   *Features*: Install, Config optimization.
*   [ ] **MongoDB**: NoSQL document database.
    *   *Features*: Install, Auth setup.

## 3. Container Management

*   [x] **Docker**: Standard container platform.
    *   *Features*: Install Engine, Run Containers, Manage Compose.
*   [ ] **Podman**: Daemonless container engine.
    *   *Features*: Install, Container management.

## 4. Language Runtimes

*   [ ] **Node.js**: JavaScript runtime.
    *   *Features*: Install (via NVM?), PM2 setup.
*   [ ] **Go**: Go programming language.
    *   *Features*: Install specific versions.
*   [ ] **Python**: Python environment.
    *   *Features*: Install, Pip, Virtualenv.

## 5. Security & Utilities

*   [x] **Fail2Ban**: Intrusion prevention.
    *   *Features*: Install, Jail configuration.
*   [ ] **Certbot**: SSL certificate management.
    *   *Features*: Install, Obtain certs (standalone or webroot). (Note: Integrated in Nginx plugin)
*   [] **UFW**: Firewall management.
    *   *Features*: Enable/Disable, Allow/Deny ports.
