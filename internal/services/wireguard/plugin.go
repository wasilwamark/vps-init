package wireguard

import (
	"context"
	"fmt"
	"strings"
	"time"

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
	res := conn.RunSudo("cat /etc/wireguard/wg0.conf", pass)
	if !res.Success {
		return fmt.Errorf("failed to read server config: %s", res.Stderr)
	}

	// Parse the config to extract the private key
	lines := strings.Split(res.Stdout, "\n")
	var sPriv string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "PrivateKey") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				sPriv = strings.TrimSpace(parts[1])
				break
			}
		}
	}

	if sPriv == "" {
		return fmt.Errorf("server private key not found in config")
	}

	// Derive public from private because getting it from wg show might require it running
	sPubRes := conn.RunCommand(fmt.Sprintf("echo '%s' | wg pubkey", sPriv), false)
	if !sPubRes.Success {
		return fmt.Errorf("failed to derive server public key: %s", sPubRes.Stderr)
	}
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
	if res := conn.RunSudo(fmt.Sprintf("bash -c 'cat %s >> /etc/wireguard/wg0.conf'", tmpPeer), pass); !res.Success {
		return fmt.Errorf("failed to update server config: %s", res.Stderr)
	}

	// Reload Server
	conn.RunSudo("bash -c 'wg syncconf wg0 <(wg-quick strip wg0)'", pass)

	// create Client Config
	// Remove /32 suffix from clientIP for Address field - just use the IP
	clientAddr := strings.Replace(clientIP, "/32", "", 1)
	clientConfig := fmt.Sprintf(`[Interface]
PrivateKey = %s
Address = %s
DNS = 1.1.1.1

[Peer]
PublicKey = %s
Endpoint = %s
AllowedIPs = 0.0.0.0/0
PersistentKeepalive = 25
`, cPriv, clientAddr, sPub, endpoint)

	// Display client information
	fmt.Printf("\n✅ Peer %s added successfully!\n\n", name)

	fmt.Printf("📱 Client Configuration:\n")
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n")
	fmt.Printf("🔑 Client Public Key: %s\n", cPub)
	fmt.Printf("🌐 Server Endpoint: %s\n", endpoint)
	fmt.Printf("📍 Client IP: %s\n", clientAddr)
	fmt.Printf("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━\n\n")

	fmt.Printf("📄 Complete Client Config for %s:\n", name)
	fmt.Println(clientConfig)

	// Generate QR Code
	fmt.Printf("\n📱 Scan this QR Code to add to your Wireguard client:\n")
	// Write to tmp file then qrencode
	tmpClient := fmt.Sprintf("/tmp/%s.conf", name)
	conn.WriteFile(clientConfig, tmpClient)
	conn.RunInteractive(fmt.Sprintf("qrencode -t ansiutf8 < %s", tmpClient))

	// Clean up
	conn.RunSudo(fmt.Sprintf("rm %s", tmpClient), pass)
	return nil
}

func (p *Plugin) removePeerHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	pass := getSudoPass(flags)

	// Get the config file to list peers
	configRes := conn.RunSudo("cat /etc/wireguard/wg0.conf", pass)
	if !configRes.Success {
		return fmt.Errorf("failed to read config file: %s", configRes.Stderr)
	}

	// Parse peers and build list
	lines := strings.Split(configRes.Stdout, "\n")
	var peers []struct {
		name    string
		pubKey  string
		allowed string
	}

	var currentName, currentPubKey, currentAllowed string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[Peer]") {
			// Save previous peer if exists
			if currentPubKey != "" {
				peers = append(peers, struct {
					name    string
					pubKey  string
					allowed string
				}{currentName, currentPubKey, currentAllowed})
			}
			currentName, currentPubKey, currentAllowed = "", "", ""
			continue
		}
		if strings.HasPrefix(line, "# Name =") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentName = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "PublicKey =") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentPubKey = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "AllowedIPs =") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentAllowed = strings.TrimSpace(parts[1])
			}
		}
	}
	// Save last peer if exists
	if currentPubKey != "" {
		peers = append(peers, struct {
			name    string
			pubKey  string
			allowed string
		}{currentName, currentPubKey, currentAllowed})
	}

	if len(peers) == 0 {
		fmt.Println("❌ No peers found in configuration")
		return nil
	}

	// Display peers for selection
	fmt.Println("📋 Available WireGuard Peers:")
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	for i, peer := range peers {
		displayName := peer.name
		if displayName == "" {
			displayName = "Unnamed"
		}
		fmt.Printf(" [%d] %s (%s)\n", i+1, displayName, peer.allowed)
	}
	fmt.Println("━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━")
	fmt.Printf("Select peer to remove (1-%d): ", len(peers))

	// Get user input
	var choice int
	fmt.Scanf("%d", &choice)

	if choice < 1 || choice > len(peers) {
		return fmt.Errorf("invalid selection: %d", choice)
	}

	selectedPeer := peers[choice-1]
	displayName := selectedPeer.name
	if displayName == "" {
		displayName = "Unnamed"
	}

	// Confirm removal
	fmt.Printf("\n⚠️  Are you sure you want to remove peer '%s' (%s)? [y/N]: ", displayName, selectedPeer.allowed)
	var confirm string
	fmt.Scanf("%s", &confirm)

	if strings.ToLower(confirm) != "y" && strings.ToLower(confirm) != "yes" {
		fmt.Println("❌ Operation cancelled")
		return nil
	}

	// Remove peer from runtime
	fmt.Printf("🗑️  Removing peer '%s' from WireGuard...\n", displayName)
	removeRes := conn.RunSudo(fmt.Sprintf("wg set wg0 peer %s remove", selectedPeer.pubKey), pass)
	if !removeRes.Success {
		return fmt.Errorf("failed to remove peer from runtime: %s", removeRes.Stderr)
	}

	// Remove from config file
	// Create new config without the selected peer
	var newConfig []string
	i := 0

	for i < len(lines) {
		line := lines[i]
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "[Peer]") {
			// Check if this is the peer to remove by looking for its public key
			isTargetPeer := false
			// Look ahead for the PublicKey
			for j := i + 1; j < len(lines) && j < i + 10; j++ {
				nextLine := strings.TrimSpace(lines[j])
				if strings.HasPrefix(nextLine, "PublicKey =") {
					parts := strings.SplitN(nextLine, "=", 2)
					if len(parts) == 2 {
						pubKey := strings.TrimSpace(parts[1])
						if pubKey == selectedPeer.pubKey {
							isTargetPeer = true
						}
					}
					break
				}
				// If we hit another [Peer] or [Interface], this peer doesn't have a public key
				if strings.HasPrefix(nextLine, "[") {
					break
				}
			}

			if isTargetPeer {
				// Skip this entire peer section
				// Advance i until we hit the next section or end
				i++
				for i < len(lines) {
					nextTrimmed := strings.TrimSpace(lines[i])
					if strings.HasPrefix(nextTrimmed, "[") {
						// We've reached the next section, don't advance i so it gets processed
						break
					}
					if nextTrimmed == "" {
						// Check if the next line starts a new section
						if i+1 < len(lines) && strings.HasPrefix(strings.TrimSpace(lines[i+1]), "[") {
							// Skip the empty line too
							i++
							break
						}
					}
					i++
				}
				continue
			} else {
				// This is not the peer to remove, add it
				newConfig = append(newConfig, line)
			}
		} else {
			// Not a peer section, add the line
			newConfig = append(newConfig, line)
		}
		i++
	}

	// Write new config
	newConfigStr := strings.Join(newConfig, "\n")
	tmpPath := "/tmp/wg0_new.conf"
	if !conn.WriteFile(newConfigStr, tmpPath) {
		return fmt.Errorf("failed to write new config")
	}

	// Backup and replace config
	backupPath := fmt.Sprintf("/etc/wireguard/wg0.conf.bak.%d", time.Now().Unix())
	conn.RunSudo(fmt.Sprintf("cp /etc/wireguard/wg0.conf %s", backupPath), pass)

	if res := conn.RunSudo(fmt.Sprintf("mv %s /etc/wireguard/wg0.conf", tmpPath), pass); !res.Success {
		return fmt.Errorf("failed to update config file: %s", res.Stderr)
	}

	// Reload configuration
	fmt.Println("🔄 Reloading WireGuard configuration...")
	reloadRes := conn.RunSudo("wg-quick down wg0 && wg-quick up wg0", pass)
	if !reloadRes.Success {
		// Try alternative reload method
		conn.RunSudo("wg syncconf wg0 <(wg-quick strip wg0)", pass)
	}

	fmt.Printf("✅ Peer '%s' removed successfully!\n", displayName)
	fmt.Printf("💾 Backup saved to: %s\n", backupPath)

	return nil
}

func (p *Plugin) listPeersHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	pass := getSudoPass(flags)

	// Get the current WireGuard status
	res := conn.RunSudo("wg show", pass)
	if !res.Success {
		return fmt.Errorf("failed to get wg status: %s", res.Stderr)
	}

	// Parse the config file to get peer names
	configRes := conn.RunSudo("cat /etc/wireguard/wg0.conf", pass)
	if !configRes.Success {
		return fmt.Errorf("failed to read config file: %s", configRes.Stderr)
	}

	// Build a map of public keys to device names
	peerNames := make(map[string]string)
	lines := strings.Split(configRes.Stdout, "\n")
	var currentName string

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "[Peer]") {
			currentName = ""
			continue
		}
		if strings.HasPrefix(line, "# Name =") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 {
				currentName = strings.TrimSpace(parts[1])
			}
		}
		if strings.HasPrefix(line, "PublicKey =") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) == 2 && currentName != "" {
				pubKey := strings.TrimSpace(parts[1])
				peerNames[pubKey] = currentName
				currentName = ""
			}
		}
	}

	// Parse the wg show output and replace public keys with names
	output := res.Stdout
	lines = strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "peer:") {
			parts := strings.Fields(line)
			if len(parts) >= 2 {
				pubKey := parts[1]
				if name, exists := peerNames[pubKey]; exists {
					fmt.Printf("peer: %s (%s)\n", pubKey, name)
				} else {
					fmt.Println(line)
				}
			} else {
				fmt.Println(line)
			}
		} else {
			fmt.Println(line)
		}
	}

	return nil
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
