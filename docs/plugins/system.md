# System Plugin

The **System** plugin manages core system updates and package maintenance on your VPS.

## Usage

All commands should be run against a target server alias (e.g., `ovh`) or connection string (`user@host`).

### Commands

| Command | Description | Example |
| :--- | :--- | :--- |
| `update` | Updates package lists (`apt-get update`) | `vps-init prod system update` |
| `upgrade` | Upgrades installed packages | `vps-init prod system upgrade` |
| `full-upgrade` | Performs dist-upgrade | `vps-init prod system full-upgrade` |
| `autoremove` | Removes unused packages | `vps-init prod system autoremove` |
| `install` | Installs specific packages | `vps-init prod system install git curl` |
| `uninstall` | Removes specific packages | `vps-init prod system uninstall apache2` |
| `shell` | Opens interactive SSH shell | `vps-init prod system shell` |

### Sudo Privileges

These commands generally require root privileges. If your user is not `root`, you must provide the user's password so `sudo` can operate.

This is done via an environment variable specific to your server alias:

`SSH_SUDO_PWD_<ALIAS>`

**Example:**

If your alias is `ovh`:

```bash
export SSH_SUDO_PWD_OVH='your-secret-password'
vps-init ovh system update
```

The tool will automatically detect the alias `ovh`, look for `SSH_SUDO_PWD_OVH`, and inject the password when running sudo commands.

**Method 2: Stored Secret (Recommended)**

You can save the password securely when adding the alias:

```bash
vps-init alias add ovh user@host --sudo-password 'your-secret-password'
```

This saves the password to `~/.vps-init/secrets.json` with restricted permissions. The tool will check this file if the environment variable is not set.
