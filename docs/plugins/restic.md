# Restic Backup Plugin

The Restic plugin provides secure, efficient backups to S3-compatible storage. It supports streaming database dumps directly to the backup repository without creating local temporary files.

## Usage

```bash
vps-init <target> restic <command> [args...]
```

## Commands

| Command | Description | Example |
| :--- | :--- | :--- |
| `install` | Install Restic | `vps-init prod restic install` |
| `init` | Configure S3 repo & credentials | `vps-init prod restic init` |
| `backup-db` | Stream database dump to S3 | `vps-init prod restic backup-db my_app_db` |
| `snapshots` | List stored backups | `vps-init prod restic snapshots` |
| `unlock` | Remove stale locks | `vps-init prod restic unlock` |

## Setup Guide

1.  **Install**:
    ```bash
    vps-init prod restic install
    ```
2.  **Initialize**:
    You will need your AWS/S3 Bucket URL, Access Key, and Secret Key.
    ```bash
    vps-init prod restic init
    ```
    *This saves credentials to `/etc/vps-init/restic.env` securely.*

3.  **Backup a Database**:
    ```bash
    vps-init prod restic backup-db wp_my_site
    ```

4.  **Check Backups**:
    ```bash
    vps-init prod restic snapshots
    ```
