# Nginx Plugin

The **Nginx** plugin allows you to install, configure, and manage the Nginx web server on your VPS.

## Usage

All commands are run against a target server alias:

```bash
vps-init <target> nginx <command> [args]
```

## Commands

### Installation & Lifecycle

*   `install`: Installs Nginx using the system package manager (`apt-get`).
*   `status`: Checks the status of the Nginx service.
*   `start`, `stop`, `restart`, `reload`: Manages the systemd service.

### Site Management

*   `add-site <domain> [--proxy <port>] [--file <path>] [--ssl]`:
    *   **Proxy Mode (Default)**: Creates a reverse proxy config pointing to `localhost:<port>` (default 3000).
        ```bash
        vps-init prod nginx add-site api.example.com --proxy 8080
        ```
    *   **Custom Config**: Uploads a local configuration file.
        ```bash
        vps-init prod nginx add-site static.example.com --file ./mysite.conf
        ```
    *   **One-Step SSL**: Add `--ssl` to automatically install certificates after configuration.
        ```bash
        vps-init prod nginx add-site api.example.com --ssl
        ```
    *   *Safety*: Automatically validates config with `nginx -t` and rolls back if invalid.

*   `remove-site <domain>`:
    *   Removes the site configuration and correctly reloads Nginx.
    *   ```bash
        vps-init prod nginx remove-site api.example.com
        ```

### SSL / Security

*   `install-ssl [domain]`:
    *   Installs Certbot and requests a Let's Encrypt SSL certificate.
    *   **Interactive Mode**: If `[domain]` is omitted, lists available sites for selection.
    *   **Direct Mode**: `vps-init myalias nginx install-ssl mydomain.com`

### Observability

*   `logs`:
    *   Streams the Nginx logs (`access.log` / `error.log`) to your local terminal in real-time.
    *   Press `Ctrl+C` to stop streaming.

## Examples

**1. Setting up a Node.js/Go backend:**

```bash
# 1. Install Nginx
vps-init myserver nginx install

# 2. Add site proxying to your app running on port 3000
vps-init myserver nginx add-site myapp.com --proxy 3000

# 3. Secure it with SSL
vps-init myserver nginx install-ssl myapp.com
```

**2. Debugging issues:**

```bash
# Watch logs while you make requests
vps-init myserver nginx logs
```
