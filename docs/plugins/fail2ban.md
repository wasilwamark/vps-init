# Fail2Ban Plugin

The Fail2Ban plugin scans log files (e.g., `/var/log/auth.log`) and bans IPs that show the malicious signs -- too many password failures, seeking for exploits, etc.

## Usage

```bash
vps-init <target> fail2ban <command> [args...]
```

## Commands

| Command | Description | Example |
| :--- | :--- | :--- |
| `install` | Installs Fail2Ban and ensures `sshd` jail is active | `vps-init prod fail2ban install` |
| `status` | Shows status of the Fail2Ban server and active jails | `vps-init prod fail2ban status` |
| `banned` | Lists currently banned IPs (defaults to `sshd` jail) | `vps-init prod fail2ban banned`<br>`vps-init prod fail2ban banned nginx-http-auth` |
| `unban` | Unbans a specific IP | `vps-init prod fail2ban unban 192.168.1.1` |
