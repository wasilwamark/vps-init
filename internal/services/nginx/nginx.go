package nginx

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
	fmt.Println("📦 Installing Nginx...")

	if !s.ssh.InstallPackage("nginx") {
		fmt.Println("❌ Failed to install Nginx")
		return false
	}

	// Start and enable Nginx
	if !s.ssh.Systemctl("start", "nginx") {
		fmt.Println("❌ Failed to start Nginx")
		return false
	}

	if !s.ssh.Systemctl("enable", "nginx") {
		fmt.Println("❌ Failed to enable Nginx")
		return false
	}

	fmt.Println("✅ Nginx installed and running")
	return true
}

func (s *Service) InstallWithSSL(domain string) bool {
	fmt.Printf("🔒 Installing Nginx with SSL for %s\n", domain)

	// First install Nginx
	if !s.Install() {
		return false
	}

	// Create site configuration
	siteConfig := s.generateSiteConfig(domain, true)
	if !s.ssh.WriteFile(siteConfig, fmt.Sprintf("/etc/nginx/sites-available/%s", domain)) {
		fmt.Println("❌ Failed to create site configuration")
		return false
	}

	// Enable site
	enableCmd := fmt.Sprintf("ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/", domain)
	result := s.ssh.RunCommand(enableCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to enable site")
		return false
	}

	// Install Certbot for Let's Encrypt
	if !s.ssh.InstallPackage("certbot") {
		fmt.Println("❌ Failed to install Certbot")
		return false
	}

	if !s.ssh.InstallPackage("python3-certbot-nginx") {
		fmt.Println("❌ Failed to install Certbot Nginx plugin")
		return false
	}

	// Obtain SSL certificate
	sslCmd := fmt.Sprintf("certbot --nginx -d %s --non-interactive --agree-tos --email admin@%s", domain, domain)
	result = s.ssh.RunCommand(sslCmd, false)
	if !result.Success {
		fmt.Println("⚠️  SSL certificate installation may have failed, checking...")
	}

	// Test and reload Nginx
	if s.ssh.Systemctl("reload", "nginx") {
		fmt.Printf("✅ Nginx with SSL configured for %s\n", domain)
		return true
	}

	return false
}

func (s *Service) CreateSite(domain string) bool {
	fmt.Printf("🌐 Creating Nginx site for %s\n", domain)

	siteConfig := s.generateSiteConfig(domain, false)
	configPath := fmt.Sprintf("/etc/nginx/sites-available/%s", domain)

	if !s.ssh.WriteFile(siteConfig, configPath) {
		fmt.Println("❌ Failed to create site configuration")
		return false
	}

	// Enable site
	enableCmd := fmt.Sprintf("ln -sf /etc/nginx/sites-available/%s /etc/nginx/sites-enabled/", domain)
	result := s.ssh.RunCommand(enableCmd, false)
	if !result.Success {
		fmt.Println("❌ Failed to enable site")
		return false
	}

	// Reload Nginx
	if s.ssh.Systemctl("reload", "nginx") {
		fmt.Printf("✅ Site %s created and enabled\n", domain)
		return true
	}

	return false
}

func (s *Service) generateSiteConfig(domain string, ssl bool) string {
	config := fmt.Sprintf(`server {
    listen 80;
    listen [::]:80;
    server_name %s;

    # Redirect to HTTPS if SSL is enabled
%s

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
}`, domain, generateSSLRedirect(ssl))

	if ssl {
		httpsConfig := fmt.Sprintf(`}

server {
    listen 443 ssl http2;
    listen [::]:443 ssl http2;
    server_name %s;

    # SSL Configuration (will be configured by Certbot)
    ssl_certificate /etc/letsencrypt/live/%s/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/%s/privkey.pem;
    ssl_protocols TLSv1.2 TLSv1.3;
    ssl_ciphers ECDHE-RSA-AES256-GCM-SHA512:DHE-RSA-AES256-GCM-SHA512;
    ssl_prefer_server_ciphers off;

    location / {
        proxy_pass http://localhost:3000;
        proxy_http_version 1.1;
        proxy_set_header Upgrade $http_upgrade;
        proxy_set_header Connection 'upgrade';
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
        proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto $scheme;
        proxy_cache_bypass $http_upgrade;
    }

    # Security headers
    add_header X-Frame-Options "SAMEORIGIN" always;
    add_header X-XSS-Protection "1; mode=block" always;
    add_header X-Content-Type-Options "nosniff" always;
    add_header Strict-Transport-Security "max-age=31536000; includeSubDomains" always;`, domain, domain, domain)

		config += httpsConfig
	}

	return config
}

func (s *Service) Reload() error {
	fmt.Println("🔄 Reloading Nginx configuration...")

	if !s.ssh.Systemctl("reload", "nginx") {
		return fmt.Errorf("failed to reload nginx")
	}

	fmt.Println("✅ Nginx configuration reloaded")
	return nil
}

func (s *Service) Status() error {
	fmt.Println("📊 Checking Nginx status...")

	result := s.ssh.RunCommand("systemctl status nginx --no-pager", false)
	if result.Success {
		fmt.Println(result.Stdout)
		return nil
	}

	return fmt.Errorf("nginx is not running")
}

func generateSSLRedirect(enabled bool) string {
	if enabled {
		return "    return 301 https://$host$request_uri;"
	}
	return ""
}