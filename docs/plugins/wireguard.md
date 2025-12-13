# Wireguard Plugin

The Wireguard plugin turns your VPS into a secure, personal VPN server. It handles installation, configuration, key generation, and client management with QR codes for easy mobile setup.

## Usage

```bash
vps-init <target> wireguard <command> [args...]
```

## Commands

| Command | Description | Example |
| :--- | :--- | :--- |
| `install` | Installs Wireguard, tools, and QREncode | `vps-init prod wireguard install` |
| `setup` | Configures the server, keys, and firewall | `vps-init prod wireguard setup` |
| `add-peer` | Adds a client and **shows QR code** | `vps-init prod wireguard add-peer my-phone` |
| `status` | Shows VPN interface status and connected peers | `vps-init prod wireguard status` |
| `list-peers` | Lists configured peers (alias for status) | `vps-init prod wireguard list-peers` |

## Quick Start

1.  **Install**:
    ```bash
    vps-init prod wireguard install
    ```
2.  **Setup Server**:
    ```bash
    vps-init prod wireguard setup
    ```
3.  **Connect a Phone**:
    *   Install Wireguard App on your phone.
    *   Run:
        ```bash
        vps-init prod wireguard add-peer my-phone
        ```
    *   Scan the QR code that appears in your terminal.
