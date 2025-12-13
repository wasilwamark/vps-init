# WordPress Plugin

The WordPress plugin automates the deployment of a fully functional **LEMP Stack** (Linux, Nginx, MariaDB, PHP) and installs WordPress. It features an interactive wizard that handles everything from database creation to Nginx configuration.

## Usage

```bash
vps-init <target> wordpress <command> [args...]
```

## Commands

| Command | Description | Example |
| :--- | :--- | :--- |
| `install` | Installs system dependencies (PHP, WP-CLI) | `vps-init prod wordpress install` |
| `create-site` | **Wizard** to deploy a new site | `vps-init prod wordpress create-site example.com` |

## Deployment Flow (`create-site`)

The `create-site` command is an interactive wizard that performs the following steps automatically:

1.  **Collects Information**: Asks for database credentials, admin user details, etc.
2.  **Database Setup**: Creates the database and user (using MySQL plugin logic).
3.  **Web Root**: Creates `/var/www/<domain>`.
4.  **Download**: Downloads the latest WordPress core using WP-CLI.
5.  **Configure**: Generates `wp-config.php` with correct DB details.
6.  **Install**: Installs WordPress tables and sets up the admin user.
7.  **Permissions**: Sets correct ownership (`www-data`).
8.  **Nginx**: Generates and enables an Nginx server block, then reloads Nginx.

## Quick Start

1.  **Prerequisites**:
    Ensure Nginx and MySQL are installed.
    ```bash
    vps-init prod nginx install
    vps-init prod mysql install
    ```

2.  **Install WordPress Tools**:
    ```bash
    vps-init prod wordpress install
    ```

3.  **Deploy a Site**:
    ```bash
    vps-init prod wordpress create-site blog.example.com
    ```
    *Follow the interactive prompts.*

4.  **Secure It**:
    Use the Nginx plugin to add SSL.
    ```bash
    vps-init prod nginx install-ssl blog.example.com
    ```
