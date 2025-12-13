package wireguard

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

func (p *Plugin) Name() string                                   { return "wireguard" }
func (p *Plugin) Description() string                            { return "Wireguard VPN Server" }
func (p *Plugin) Author() string                                 { return "VPS-Init" }
func (p *Plugin) Version() string                                { return "0.0.1" }
func (p *Plugin) Dependencies() []string                         { return []string{} }
func (p *Plugin) Initialize(config map[string]interface{}) error { return nil }
func (p *Plugin) Start(ctx context.Context) error                { return nil }
func (p *Plugin) Stop(ctx context.Context) error                 { return nil }
func (p *Plugin) GetRootCommand() *cobra.Command                 { return nil }

func (p *Plugin) GetCommands() []plugin.Command {
	return []plugin.Command{
		{
			Name:        "install",
			Description: "Install Wireguard and tools",
			Handler:     p.installHandler,
		},
		{
			Name:        "setup",
			Description: "Configure Wireguard Server (Interactive)",
			Handler:     p.setupHandler,
		},
		{
			Name:        "add-peer",
			Description: "Add a new client/peer",
			Handler:     p.addPeerHandler,
		},
		{
			Name:        "remove-peer",
			Description: "Remove a peer",
			Handler:     p.removePeerHandler,
		},
		{
			Name:        "list-peers",
			Description: "List configured peers",
			Handler:     p.listPeersHandler,
		},
		{
			Name:        "status",
			Description: "Show Wireguard status",
			Handler:     p.statusHandler,
		},
	}
}

// Handlers

func (p *Plugin) installHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🛡️  Installing Wireguard & Tools...")
	pass := getSudoPass(flags)

	// Update first
	if res := conn.RunSudo("apt-get update", pass); !res.Success {
		return fmt.Errorf("apt update failed: %s", res.Stderr)
	}

	// Install packages: wireguard, wireguard-tools, qrencode (for QR display)
	pkgs := "wireguard wireguard-tools qrencode iptables"
	if res := conn.RunSudo(fmt.Sprintf("apt-get install -y %s", pkgs), pass); !res.Success {
		return fmt.Errorf("installation failed: %s", res.Stderr)
	}

	fmt.Println("✅ Wireguard installed.")
	return nil
}

func (p *Plugin) setupHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("⚙️  Setting up Wireguard Server...")
	pass := getSudoPass(flags)

	// 1. Generate Server Keys
	privKey, pubKey, err := generateKeys(conn)
	if err != nil {
		return err
	}

	// 2. Interactive Config defaults
	port := "51820"
	cidr := "10.100.0.1/24"
	iface := getMainInterface(conn)

	fmt.Printf("Using Interface: %s\n", iface)
	fmt.Printf("Using Port: %s\n", port)
	fmt.Printf("Using Internal IP: %s\n", cidr)

	// 3. Create Config
	// IP forwarding rule
	postUp := fmt.Sprintf("iptables -A FORWARD -i wg0 -j ACCEPT; iptables -t nat -A POSTROUTING -o %s -j MASQUERADE", iface)
	postDown := fmt.Sprintf("iptables -D FORWARD -i wg0 -j ACCEPT; iptables -t nat -D POSTROUTING -o %s -j MASQUERADE", iface)

	config := fmt.Sprintf(`[Interface]
Address = %s
SaveConfig = true
PostUp = %s
PostDown = %s
ListenPort = %s
PrivateKey = %s
`, cidr, postUp, postDown, port, privKey)

	// Write Config
	tmpPath := "/tmp/wg0.conf"
	if !conn.WriteFile(config, tmpPath) {
		return fmt.Errorf("failed to write tmp config")
	}

	// Move to /etc/wireguard/
	conn.RunSudo("mkdir -p /etc/wireguard", pass)
	if res := conn.RunSudo(fmt.Sprintf("mv %s /etc/wireguard/wg0.conf", tmpPath), pass); !res.Success {
		return fmt.Errorf("failed to move config: %s", res.Stderr)
	}
	conn.RunSudo("chmod 600 /etc/wireguard/wg0.conf", pass)

	// 4. Enable IP Forwarding
	conn.RunSudo("sysctl -w net.ipv4.ip_forward=1", pass)
	// Make persistent
	conn.RunSudo("echo 'net.ipv4.ip_forward=1' > /etc/sysctl.d/99-wireguard.conf", pass)

	// 5. Firewall Rules (UFW) if installed
	// Try to allow 51820/udp
	conn.RunSudo(fmt.Sprintf("ufw allow %s/udp", port), pass)

	// 6. Start Service
	if res := conn.RunSudo("systemctl enable wg-quick@wg0", pass); !res.Success {
		return fmt.Errorf("failed to enable service: %s", res.Stderr)
	}
	if res := conn.RunSudo("systemctl start wg-quick@wg0", pass); !res.Success {
		return fmt.Errorf("failed to start service: %s", res.Stderr)
	}

	fmt.Printf("✅ Wireguard Server configured and running!\nPublic Key: %s\n", pubKey)
	return nil
}

func (p *Plugin) addPeerHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	if len(args) < 1 {
		return fmt.Errorf("usage: add-peer <name>")
	}
	name := args[0]
	pass := getSudoPass(flags)

	// Generate Client Keys
	cPriv, cPub, err := generateKeys(conn)
	if err != nil {
		return err
	}

	// Get Server Public Key
	res := conn.RunSudo("cat /etc/wireguard/wg0.conf | grep PrivateKey | cut -d = -f 2", pass)
	sPriv := strings.TrimSpace(res.Stdout)
	// Derive public from private because getting it from wg show might require it running
	sPubRes := conn.RunCommand(fmt.Sprintf("echo '%s' | wg pubkey", sPriv), false)
	sPub := strings.TrimSpace(sPubRes.Stdout)

	// Find available IP
	// For simplicity, we just count peers + 2 (since .1 is server)
	// A real implementation would parse the file properly.
	// MVP: Check file for AllowedIPs, find max.
	// Hack: Just randomly pick or increment?
	// Let's grab the last IP octet from the config file if possible, or just fail if exists.
	// Simple strategy: 10.100.0.X
	res = conn.RunSudo("grep AllowedIPs /etc/wireguard/wg0.conf | wc -l", pass)
	countStr := strings.TrimSpace(res.Stdout)
	var count int
	fmt.Sscanf(countStr, "%d", &count)
	ipSuffix := count + 2 // Start at .2
	clientIP := fmt.Sprintf("10.100.0.%d/32", ipSuffix)

	// Get Server Endpoint (Public IP)
	// Try to guess or use host
	endpoint := fmt.Sprintf("%s:51820", conn.Host)

	// Add Peer to Server Config
	peerBlock := fmt.Sprintf(`
[Peer]
# Name = %s
PublicKey = %s
AllowedIPs = %s
`, name, cPub, clientIP)

	// Append to server config
	tmpPeer := "/tmp/wg_peer_add"
	conn.WriteFile(peerBlock, tmpPeer)
	if res := conn.RunSudo(fmt.Sprintf("cat %s >> /etc/wireguard/wg0.conf", tmpPeer), pass); !res.Success {
		return fmt.Errorf("failed to update server config")
	}

	// Reload Server
	conn.RunSudo("wg syncconf wg0 <(wg-quick strip wg0)", pass)

	// create Client Config
	clientConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = 1.1.1.1

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
`, cPriv, strings.Replace(clientIP, "/32", "/24", 1), sPub, endpoint)

	fmt.Printf("\n📋 Client Config for %s:\n", name)
	fmt.Println("-------------------------------------------")
	fmt.Println(clientConfig)
	fmt.Println("-------------------------------------------")

	// Generate QR Code
	fmt.Println("\n📱 Scan this QR Code to connect:")
	// Write to tmp file then qrencode
	tmpClient := fmt.Sprintf("/tmp/%s.conf", name)
	conn.WriteFile(clientConfig, tmpClient)
	conn.RunInteractive(fmt.Sprintf("qrencode -t ansiutf8 < %s", tmpClient))

	// Clean up
	conn.RunSudo(fmt.Sprintf("rm %s", tmpClient), pass)
	fmt.Printf("\n✅ Peer %s added.\n", name)
	return nil
}

func (p *Plugin) removePeerHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	// Removing logic is harder with text file parsing without a tool.
	// wg set wg0 peer <pubkey> remove works for runtime, but config file persistence is tricky without 'wg-quick save' overwriting formats.
	// For MVP, implementing remove is risky without parsing.
	return fmt.Errorf("remove-peer not implemented in this version (edit /etc/wireguard/wg0.conf manually)")
}

func (p *Plugin) listPeersHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	return conn.RunInteractive("sudo wg show")
}

func (p *Plugin) statusHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("🔌 Wireguard Service Status:")
	conn.RunInteractive("systemctl status wg-quick@wg0")
	fmt.Println("\n📊 Interface Status:")
	return conn.RunInteractive("sudo wg show")
}

// Helpers

func generateKeys(conn *ssh.Connection) (string, string, error) {
	// Returns private, public
	res := conn.RunCommand("wg genkey", false)
	if !res.Success {
		return "", "", fmt.Errorf("failed to gen key: %s", res.Stderr)
	}
	priv := strings.TrimSpace(res.Stdout)

	res = conn.RunCommand(fmt.Sprintf("echo '%s' | wg pubkey", priv), false)
	if !res.Success {
		return "", "", fmt.Errorf("failed to gen pub key")
	}
	pub := strings.TrimSpace(res.Stdout)
	return priv, pub, nil
}

func getMainInterface(conn *ssh.Connection) string {
	// Try to guess default interface
	res := conn.RunCommand("ip route | grep default | awk '{print $5}'", false)
	if res.Success {
		return strings.TrimSpace(res.Stdout)
	}
	return "eth0" // Fallback
}

func getSudoPass(flags map[string]interface{}) string {
	if v, ok := flags["sudo-password"]; ok {
		return v.(string)
	}
	return ""
}
