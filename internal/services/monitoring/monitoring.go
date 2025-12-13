package monitoring

import (
	"fmt"
	"github.com/wasilwamark/vps-init/internal/ssh"
)

type Service struct {
	ssh *ssh.Connection
}

func New(ssh *ssh.Connection) *Service {
	return &Service{ssh: ssh}
}

func (s *Service) Install() bool {
	return s.Setup()
}

func (s *Service) Setup() bool {
	fmt.Println("📊 Setting up monitoring...")

	// Install monitoring tools
	if !s.ssh.InstallPackage("htop") {
		fmt.Println("⚠️  Could not install htop")
	}

	if !s.ssh.InstallPackage("iotop") {
		fmt.Println("⚠️  Could not install iotop")
	}

	// Create monitoring directory
	mkdirCmd := "mkdir -p /opt/monitoring /var/log/monitoring"
	result := s.ssh.RunCommand(mkdirCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to create monitoring directories")
		return false
	}

	// Create disk monitoring script
	diskMonitorScript := `#!/bin/bash
THRESHOLD=80
LOG_FILE="/var/log/monitoring/disk_usage.log"
ALERT_FILE="/var/log/monitoring/disk_alert.log"

# Get current disk usage
USAGE=$(df / | awk 'NR==2 {print $5}' | sed 's/%//')
DATE=$(date '+%Y-%m-%d %H:%M:%S')

# Log current usage
echo "$DATE: Disk usage is ${USAGE}%" >> $LOG_FILE

# Check threshold
if [ $USAGE -gt $THRESHOLD ]; then
    echo "$DATE: ALERT - Disk usage exceeded ${THRESHOLD}%. Current: ${USAGE}%" >> $ALERT_FILE
    # Send alert (you can add email/slack webhook here)
    echo "Disk usage critical: ${USAGE}% on $(hostname)"
fi`

	if !s.ssh.WriteFile(diskMonitorScript, "/opt/monitoring/disk_monitor.sh") {
		fmt.Println("❌ Failed to create disk monitoring script")
		return false
	}

	// Make script executable
	chmodCmd := "chmod +x /opt/monitoring/disk_monitor.sh"
	result = s.ssh.RunCommand(chmodCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to make disk monitor executable")
		return false
	}

	// Set up cron job for disk monitoring (every 5 minutes)
	cronCmd := "(crontab -l 2>/dev/null; echo '*/5 * * * * /opt/monitoring/disk_monitor.sh') | crontab -"
	result = s.ssh.RunCommand(cronCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to set up disk monitoring cron job")
		return false
	}

	// Create memory monitoring script
	memoryMonitorScript := `#!/bin/bash
THRESHOLD=85
LOG_FILE="/var/log/monitoring/memory_usage.log"
ALERT_FILE="/var/log/monitoring/memory_alert.log"

# Get current memory usage
USAGE=$(free | grep Mem | awk '{printf("%.0f"), $3/$2 * 100.0}')
DATE=$(date '+%Y-%m-%d %H:%M:%S')

# Log current usage
echo "$DATE: Memory usage is ${USAGE}%" >> $LOG_FILE

# Check threshold
if [ $USAGE -gt $THRESHOLD ]; then
    echo "$DATE: ALERT - Memory usage exceeded ${THRESHOLD}%. Current: ${USAGE}%" >> $ALERT_FILE
    echo "Memory usage critical: ${USAGE}% on $(hostname)"
fi`

	if !s.ssh.WriteFile(memoryMonitorScript, "/opt/monitoring/memory_monitor.sh") {
		fmt.Println("❌ Failed to create memory monitoring script")
		return false
	}

	// Make script executable
	chmodCmd = "chmod +x /opt/monitoring/memory_monitor.sh"
	result = s.ssh.RunCommand(chmodCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to make memory monitor executable")
		return false
	}

	// Set up cron job for memory monitoring (every 2 minutes)
	cronCmd = "(crontab -l 2>/dev/null; echo '*/2 * * * * /opt/monitoring/memory_monitor.sh') | crontab -"
	result = s.ssh.RunCommand(cronCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to set up memory monitoring cron job")
		return false
	}

	// Install and set up Uptime Kuma
	fmt.Println("📈 Installing Uptime Kuma...")
	uptimeKumaScript := `
# Install Node.js if not present
curl -fsSL https://deb.nodesource.com/setup_18.x | bash -
apt-get install -y nodejs

# Install PM2 for process management
npm install -g pm2

# Create directory for Uptime Kuma
mkdir -p /opt/uptime-kuma
cd /opt/uptime-kuma

# Clone and setup Uptime Kuma
git clone https://github.com/louislam/uptime-kuma.git .
npm run setup

# Start with PM2 on port 3001
pm2 start server/server.js --name uptime-kuma -- --port=3001
pm2 save
pm2 startup
`

	result = s.ssh.RunCommand(uptimeKumaScript, false)
	if !result.Success {
		fmt.Println("⚠️  Uptime Kuma installation may have failed")
	}

	// Create monitoring dashboard
	dashboardHTML := `<!DOCTYPE html>
<html>
<head>
    <title>Server Monitoring Dashboard</title>
    <style>
        body { font-family: Arial, sans-serif; margin: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; }
        .card { background: white; padding: 20px; margin: 10px 0; border-radius: 8px; box-shadow: 0 2px 4px rgba(0,0,0,0.1); }
        .metric { display: flex; justify-content: space-between; margin: 10px 0; }
        .status-good { color: green; }
        .status-warning { color: orange; }
        .status-critical { color: red; }
    </style>
</head>
<body>
    <div class="container">
        <h1>🖥️ Server Monitoring Dashboard</h1>

        <div class="card">
            <h2>System Information</h2>
            <div class="metric">
                <span>Hostname:</span>
                <span id="hostname">Loading...</span>
            </div>
            <div class="metric">
                <span>Uptime:</span>
                <span id="uptime">Loading...</span>
            </div>
        </div>

        <div class="card">
            <h2>Resource Usage</h2>
            <div class="metric">
                <span>Disk Usage:</span>
                <span id="disk" class="status-good">Loading...</span>
            </div>
            <div class="metric">
                <span>Memory Usage:</span>
                <span id="memory" class="status-good">Loading...</span>
            </div>
        </div>

        <div class="card">
            <h2>Services</h2>
            <div class="metric">
                <span>Uptime Kuma:</span>
                <a href="http://localhost:3001" target="_blank">Open Dashboard</a>
            </div>
        </div>

        <div class="card">
            <h2>Recent Logs</h2>
            <pre style="max-height: 300px; overflow-y: auto; background: #f9f9f9; padding: 10px;">
Check logs: /var/log/monitoring/
            </pre>
        </div>
    </div>
</body>
</html>`

	// Install nginx for dashboard if not present
	s.ssh.InstallPackage("nginx")
	s.ssh.Systemctl("start", "nginx")

	// Write dashboard
	if s.ssh.WriteFile(dashboardHTML, "/var/www/html/monitoring.html") {
		fmt.Println("✅ Monitoring dashboard available at http://your-server-ip/monitoring.html")
	}

	fmt.Println("✅ Monitoring setup completed!")
	fmt.Println("📊 Uptime Kuma available at http://your-server-ip:3001")
	return true
}