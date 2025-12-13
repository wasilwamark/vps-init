# Docker Plugin

The **Docker** plugin provides a streamlined way to install and manage Docker and Docker Compose on your VPS.

## Usage

```bash
vps-init <target> docker <command> [args]
```

## Commands

### Installation

*   `install`:
    *   Installs Docker Engine using the official convenience script (`get.docker.com`).
    *   **Features**:
        *   Automatically adds your user to the `docker` group (no need for `sudo`).
        *   *Note*: You may need to log out and back in for group changes to take full effect.

*   `verify`: Runs `hello-world` to verify installation.

### Compose Shortcuts

Manage your multi-container applications easily.

*   `up [args...]`:
    *   Runs `docker compose up -d [args]`.
    *   Example: `vps-init prod docker up --build`
*   `down`: Stops and removes containers (`docker compose down`).
*   `pull`: Pulls latest images (`docker compose pull`).
*   `compose [args...]`:
    *   Pass raw commands to `docker compose`.
    *   Example: `vps-init prod docker compose restart app`

### Container Management

*   `ps`: Lists running containers.
    *   Equivalent to: `docker ps`
*   `logs [container]`:
    *   **Smart Logging**:
        *   If `container` is provided: Streams logs for that specific container.
        *   If NO argument provided: Streams logs for all compose services (`docker compose logs -f`).
*   `prune`: Cleans up unused system resources (`docker system prune -f`).

## Examples

**1. setting up a new server:**

```bash
# Install Docker
vps-init myserver docker install

# Verify it works
vps-init myserver docker verify
```

**2. Deploying a project:**

```bash
# SSH in to clone your repo first (or use git plugin later)
vps-init myserver ssh
git clone ...
exit

# Start it up
vps-init myserver docker up --build
```

**3. Monitoring:**

```bash
# Check running containers
vps-init myserver docker ps

# Watch logs
vps-init myserver docker logs
```
