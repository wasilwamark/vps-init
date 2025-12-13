# MySQL/MariaDB Plugin

The MySQL plugin allows you to install and manage a MariaDB database server on your VPS. It handles secure installation and provides helper commands for common database operations.

## Usage

```bash
vps-init <target> mysql <command> [args...]
```

## Commands

| Command | Description | Example |
| :--- | :--- | :--- |
| `install` | Installs MariaDB and runs security script | `vps-init prod mysql install` |
| `create-db` | Creates a new database | `vps-init prod mysql create-db my_app_db` |
| `create-user` | Creates a new user (localhost access) | `vps-init prod mysql create-user my_user "s3cr3t"` |
| `grant` | Grants all privileges on a DB to a user | `vps-init prod mysql grant my_user my_app_db` |
| `status` | Checks service status | `vps-init prod mysql status` |

## Quick Start

1.  **Install**:
    ```bash
    vps-init prod mysql install
    ```
2.  **Create Database & User**:
    ```bash
    vps-init prod mysql create-db my_blog
    vps-init prod mysql create-user blog_user "secure_password"
    vps-init prod mysql grant blog_user my_blog
    ```
