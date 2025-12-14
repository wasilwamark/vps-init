package restic

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/wasilwamark/vps-init/internal/ssh"
	"github.com/wasilwamark/vps-init/pkg/plugin"
)

type Plugin struct{}

func (p *Plugin) Name() string                                   { return "restic" }
func (p *Plugin) Description() string                            { return "Restic Backup Manager (S3)" }
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
			Description: "Install Restic",
			Handler:     p.installHandler,
		},
		{
			Name:        "init",
			Description: "Initialize S3 Repository",
			Handler:     p.initHandler,
		},
		{
			Name:        "backup-db",
			Description: "Stream Database Backup to Repo",
			Handler:     p.backupDbHandler,
		},
		{
			Name:        "snapshots",
			Description: "List Snapshots",
			Handler:     p.snapshotsHandler,
		},
		{
			Name:        "restore-db",
			Description: "Restore Database from Backup",
			Handler:     p.restoreDbHandler,
		},
		{
			Name:        "unlock",
			Description: "Unlock Repository",
			Handler:     p.unlockHandler,
		},
	}
}

// Handlers

func (p *Plugin) installHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("💾 Installing Restic...")
	pass := getSudoPass(flags)

	// Update
	if res := conn.RunSudo("apt-get update", pass); !res.Success {
		return fmt.Errorf("apt update failed: %s", res.Stderr)
	}

	// Install
	if res := conn.RunSudo("apt-get install -y restic", pass); !res.Success {
		return fmt.Errorf("installation failed: %s", res.Stderr)
	}

	fmt.Println("✅ Restic installed.")
	return nil
}

func (p *Plugin) initHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	fmt.Println("⚙️  Initializing Repository Configuration...")
	pass := getSudoPass(flags)

	// Interactive Input
	var repo, id, key, password string

	fmt.Print("S3 Repository URL (e.g., s3:s3.amazonaws.com/my-bucket): ")
	fmt.Scanln(&repo)
	if repo == "" {
		return fmt.Errorf("repo url required")
	}

	fmt.Print("AWS Access Key ID: ")
	fmt.Scanln(&id)
	if id == "" {
		return fmt.Errorf("access key required")
	}

	fmt.Print("AWS Secret Access Key: ")
	fmt.Scanln(&key)
	if key == "" {
		return fmt.Errorf("secret key required")
	}

	fmt.Print("Repository Password: ")
	fmt.Scanln(&password)
	if password == "" {
		return fmt.Errorf("password required")
	}

	// Format S3 URL properly if needed
	// If user just entered bucket name, convert to full S3 URL
	if !strings.HasPrefix(repo, "s3:") && !strings.HasPrefix(repo, "/") {
		// Assume it's an S3 bucket name, format it properly
		repo = fmt.Sprintf("s3:s3.amazonaws.com/%s", repo)
		fmt.Printf("📝 Formatted repository URL: %s\n", repo)
	}

	// Save to config file
	envContent := fmt.Sprintf(`export RESTIC_REPOSITORY="%s"
export AWS_ACCESS_KEY_ID="%s"
export AWS_SECRET_ACCESS_KEY="%s"
export RESTIC_PASSWORD="%s"
`, repo, id, key, password)

	conn.RunSudo("mkdir -p /etc/vps-init", pass)
	conn.WriteFile(envContent, "/tmp/restic.env")
	conn.RunSudo("mv /tmp/restic.env /etc/vps-init/restic.env", pass)
	conn.RunSudo("chmod 600 /etc/vps-init/restic.env", pass)

	fmt.Println("🔒 Credentials saved to /etc/vps-init/restic.env")

	// Initialize Repo
	cmd := "bash -c 'source /etc/vps-init/restic.env && restic init'"
	// We run directly as root? or standard user? standard user might not read /etc/vps-init/restic.env if 600 root
	// Let's run as root for now since backups usually need root to read all files
	fmt.Println("🚀 Initializing backend...")
	if res := conn.RunSudo(cmd, pass); !res.Success {
		if strings.Contains(res.Stderr, "config file already exists") || strings.Contains(res.Stdout, "already initialized") {
			fmt.Println("⚠️  Repository already initialized.")
		} else {
			return fmt.Errorf("restic init failed: %s", res.Stderr)
		}
	} else {
		fmt.Println("✅ Repository initialized successfully.")
	}

	return nil
}

// Database Discovery Logic

func (p *Plugin) backupDbHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	pass := getSudoPass(flags)

	// 1. Discover Database Instances (Host Services & Docker Containers)
	fmt.Println("🔍 Scanning for database instances...")
	instances, err := discoverInstances(conn, pass)
	if err != nil {
		fmt.Printf("Warning during scan: %v\n", err)
	}

	if len(instances) == 0 {
		return fmt.Errorf("no database instances found")
	}

	// 2. Select Instance
	fmt.Println("\nFound Database Instances:")
	for i, inst := range instances {
		source := "Host"
		if inst.Type == "docker" {
			source = fmt.Sprintf("Docker Container (%s)", inst.ContainerName)
		}
		fmt.Printf("  [%d] %-10s %s\n", i+1, strings.ToUpper(inst.Engine), source)
	}

	fmt.Print("\nSelect instance (enter number): ")
	var instIdx int
	_, err = fmt.Scanln(&instIdx)
	if err != nil || instIdx < 1 || instIdx > len(instances) {
		return fmt.Errorf("invalid selection")
	}
	targetInst := instances[instIdx-1]

	// 3. Configure Credentials (Interactive)
	// We try to detect defaults to offer them, but ALWAYS ask.
	detectedUser, detectedPass := detectCredentials(conn, targetInst, pass)

	var user, dbPass string
	fmt.Printf("Database User [%s]: ", detectedUser)
	fmt.Scanln(&user)
	if user == "" {
		user = detectedUser
	}

	cleanedPass := ""
	if detectedPass != "" {
		cleanedPass = "*****"
	}
	fmt.Printf("Database Password [%s]: ", cleanedPass)
	// We can't easily read hidden input from remote execution context if we were running remote binary,
	// but here we are local CLI.
	// fmt.Scanln reads space-delimited. We need line.
	// Actually for simplicity in this prototype, explicit input is fine.
	var inputPass string
	fmt.Scanln(&inputPass)
	if inputPass != "" {
		dbPass = inputPass
	} else {
		dbPass = detectedPass
	}

	// 4. List Databases in Instance
	fmt.Println("🔍 Listing databases...")
	dbs, err := listDatabases(conn, targetInst, user, dbPass, pass)
	if err != nil {
		return fmt.Errorf("failed to list databases: %v", err)
	}
	if len(dbs) == 0 {
		return fmt.Errorf("no databases found in this instance")
	}

	fmt.Println("\nAvailable Databases:")
	for i, db := range dbs {
		fmt.Printf("  [%d] %s\n", i+1, db)
	}

	fmt.Print("\nSelect database to backup (enter number): ")
	var dbIdx int
	_, err = fmt.Scanln(&dbIdx)
	if err != nil || dbIdx < 1 || dbIdx > len(dbs) {
		return fmt.Errorf("invalid selection")
	}
	targetDBName := dbs[dbIdx-1]

	// 5. Perform Backup
	targetInfo := DatabaseInfo{
		Name:        targetDBName,
		Engine:      targetInst.Engine,
		Type:        targetInst.Type,
		ContainerID: targetInst.ContainerID,
		User:        user,
		Password:    dbPass,
	}

	return p.performBackup(conn, targetInfo, pass)
}

func (p *Plugin) performBackup(conn *ssh.Connection, targetDB DatabaseInfo, sudoPass string) error {
	fmt.Printf("📦 Streaming backup of %s (%s)...\n", targetDB.Name, targetDB.Engine)

	var dumpCmd string
	var ext string

	switch targetDB.Engine {
	case "mysql":
		ext = "sql"
		if targetDB.Type == "docker" {
			passFlag := ""
			if targetDB.Password != "" {
				passFlag = fmt.Sprintf("-p'%s'", targetDB.Password)
			}
			dumpCmd = fmt.Sprintf("docker exec -i %s mysqldump -u %s %s %s", targetDB.ContainerID, targetDB.User, passFlag, targetDB.Name)
		} else {
			passFlag := ""
			if targetDB.Password != "" {
				passFlag = fmt.Sprintf("-p'%s'", targetDB.Password)
			} // Caution with ps visibility, but standard practice in scripts often requires .my.cnf or env.
			dumpCmd = fmt.Sprintf("mysqldump -u %s %s --single-transaction --quick --lock-tables=false %s", targetDB.User, passFlag, targetDB.Name)
		}

	case "postgres":
		ext = "sql"
		envPrefix := ""
		if targetDB.Password != "" {
			envPrefix = fmt.Sprintf("PGPASSWORD='%s' ", targetDB.Password)
		}

		if targetDB.Type == "docker" {
			dumpCmd = fmt.Sprintf("docker exec -i -e PGPASSWORD='%s' %s pg_dump -U %s %s", targetDB.Password, targetDB.ContainerID, targetDB.User, targetDB.Name)
		} else {
			dumpCmd = fmt.Sprintf("%spg_dump -U %s %s", envPrefix, targetDB.User, targetDB.Name)
		}

	case "mongo":
		ext = "archive"
		authFlags := ""
		if targetDB.User != "" && targetDB.Password != "" {
			authFlags = fmt.Sprintf("--username %s --password '%s' --authenticationDatabase admin", targetDB.User, targetDB.Password)
		}

		if targetDB.Type == "docker" {
			dumpCmd = fmt.Sprintf("docker exec -i %s mongodump %s --db %s --archive", targetDB.ContainerID, authFlags, targetDB.Name)
		} else {
			dumpCmd = fmt.Sprintf("mongodump %s --db %s --archive", authFlags, targetDB.Name)
		}
	}

	// Pipe to Restic
	fullCmd := fmt.Sprintf("bash -c 'source /etc/vps-init/restic.env && %s | restic backup --stdin --stdin-filename %s.%s'", dumpCmd, targetDB.Name, ext)

	if res := conn.RunSudo(fullCmd, sudoPass); !res.Success {
		return fmt.Errorf("backup failed: %s", res.Stderr)
	}

	fmt.Println("✅ Database backup completed.")
	return nil
}

// Structures

type DatabaseInstance struct {
	Engine        string // "mysql", "postgres", "mongo"
	Type          string // "host", "docker"
	ContainerID   string
	ContainerName string
}

type DatabaseInfo struct {
	Name        string
	Engine      string
	Type        string
	ContainerID string
	User        string
	Password    string
}

// Discovery Logic

func discoverInstances(conn *ssh.Connection, sudoPass string) ([]DatabaseInstance, error) {
	var inst []DatabaseInstance

	// 1. Host Services
	if conn.RunCommand("which mysql", false).Success {
		inst = append(inst, DatabaseInstance{Engine: "mysql", Type: "host"})
	}
	if conn.RunCommand("which psql", false).Success {
		inst = append(inst, DatabaseInstance{Engine: "postgres", Type: "host"})
	}
	if conn.RunCommand("which mongosh", false).Success || conn.RunCommand("which mongo", false).Success {
		inst = append(inst, DatabaseInstance{Engine: "mongo", Type: "host"})
	}

	// 2. Docker Services
	if conn.RunCommand("which docker", false).Success {
		res := conn.RunSudo("docker ps --format '{{.ID}}|{{.Names}}|{{.Image}}'", sudoPass)
		if res.Success {
			lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
			for _, line := range lines {
				parts := strings.Split(line, "|")
				if len(parts) < 3 {
					continue
				}
				id, name, image := parts[0], parts[1], parts[2]

				if strings.Contains(image, "mysql") || strings.Contains(image, "mariadb") {
					inst = append(inst, DatabaseInstance{Engine: "mysql", Type: "docker", ContainerID: id, ContainerName: name})
				}
				if strings.Contains(image, "postgres") {
					inst = append(inst, DatabaseInstance{Engine: "postgres", Type: "docker", ContainerID: id, ContainerName: name})
				}
				if strings.Contains(image, "mongo") {
					inst = append(inst, DatabaseInstance{Engine: "mongo", Type: "docker", ContainerID: id, ContainerName: name})
				}
			}
		}
	}

	return inst, nil
}

func detectCredentials(conn *ssh.Connection, inst DatabaseInstance, sudoPass string) (string, string) {
	user := "root"
	pass := ""

	if inst.Engine == "postgres" {
		user = "postgres"
	}
	if inst.Engine == "mongo" {
		user = ""
	} // Often noauth by default or complex

	if inst.Type == "docker" {
		// Try to extract from Env
		if inst.Engine == "mysql" {
			pass, _ = getDockerEnv(conn, inst.ContainerID, sudoPass, []string{"MYSQL_ROOT_PASSWORD", "MARIADB_ROOT_PASSWORD"})
		} else if inst.Engine == "postgres" {
			pass, _ = getDockerEnv(conn, inst.ContainerID, sudoPass, []string{"POSTGRES_PASSWORD"})
			u, _ := getDockerEnv(conn, inst.ContainerID, sudoPass, []string{"POSTGRES_USER"})
			if u != "" {
				user = u
			}
		} else if inst.Engine == "mongo" {
			pass, _ = getDockerEnv(conn, inst.ContainerID, sudoPass, []string{"MONGO_INITDB_ROOT_PASSWORD"})
			u, _ := getDockerEnv(conn, inst.ContainerID, sudoPass, []string{"MONGO_INITDB_ROOT_USERNAME"})
			if u != "" {
				user = u
			}
		}
	}

	return user, pass
}

func listDatabases(conn *ssh.Connection, inst DatabaseInstance, user, pass, sudoPass string) ([]string, error) {
	var dbs []string
	var cmd string

	if inst.Type == "host" {
		switch inst.Engine {
		case "mysql":
			passFlag := ""
			if pass != "" {
				passFlag = fmt.Sprintf("-p'%s'", pass)
			}
			cmd = fmt.Sprintf("mysql -u %s %s -N -e 'SHOW DATABASES'", user, passFlag)
		case "postgres":
			// Postgres usually requires peer auth or password env
			// Sudo as postgres user is common fallback if no pass
			if pass == "" && user == "postgres" {
				cmd = "sudo -u postgres psql -l -t -A -F '|' | cut -d'|' -f1"
			} else {
				env := ""
				if pass != "" {
					env = fmt.Sprintf("PGPASSWORD='%s' ", pass)
				}
				cmd = fmt.Sprintf("%spsql -U %s -l -t -A -F '|' | cut -d'|' -f1", env, user)
			}
		case "mongo":
			auth := ""
			if user != "" && pass != "" {
				auth = fmt.Sprintf("--username %s --password '%s' --authenticationDatabase admin", user, pass)
			}
			// Try mongosh
			cmd = fmt.Sprintf("mongosh %s --quiet --eval 'db.adminCommand( { listDatabases: 1 } ).databases.forEach(db => print(db.name))'", auth)
		}
	} else {
		// Docker
		switch inst.Engine {
		case "mysql":
			passFlag := ""
			if pass != "" {
				passFlag = fmt.Sprintf("-p'%s'", pass)
			}
			cmd = fmt.Sprintf("docker exec -i %s mysql -u %s %s -N -e 'SHOW DATABASES'", inst.ContainerID, user, passFlag)
		case "postgres":
			cmd = fmt.Sprintf("docker exec -i -e PGPASSWORD='%s' %s psql -U %s -l -t -A -F '|' | cut -d'|' -f1", pass, inst.ContainerID, user)
		case "mongo":
			auth := ""
			if user != "" && pass != "" {
				auth = fmt.Sprintf("--username %s --password '%s' --authenticationDatabase admin", user, pass)
			}
			cmd = fmt.Sprintf("docker exec -i %s mongosh %s --quiet --eval 'db.adminCommand( { listDatabases: 1 } ).databases.forEach(db => print(db.name))'", inst.ContainerID, auth)
		}
	}

	res := conn.RunSudo(cmd, sudoPass)

	// Fallback for mongo if mongosh fails
	if inst.Engine == "mongo" && !res.Success {
		if strings.Contains(cmd, "mongosh") {
			cmd = strings.ReplaceAll(cmd, "mongosh", "mongo")
			res = conn.RunSudo(cmd, sudoPass)
		}
	}

	if !res.Success {
		return nil, fmt.Errorf("%s", res.Stderr)
	}

	lines := strings.Split(strings.TrimSpace(res.Stdout), "\n")
	for _, line := range lines {
		name := strings.TrimSpace(line)
		if name != "" && !isSystemDB(name) && !strings.Contains(name, "template") {
			dbs = append(dbs, name)
		}
	}
	return dbs, nil
}

// Helpers

func getDockerEnv(conn *ssh.Connection, id, sudoPass string, keys []string) (string, error) {
	inspectCmd := fmt.Sprintf("docker inspect %s --format '{{range .Config.Env}}{{println .}}{{end}}'", id)
	res := conn.RunSudo(inspectCmd, sudoPass)
	if !res.Success {
		return "", fmt.Errorf("inspect failed")
	}

	lines := strings.Split(res.Stdout, "\n")
	for _, line := range lines {
		for _, key := range keys {
			if strings.HasPrefix(line, key+"=") {
				return strings.TrimPrefix(line, key+"="), nil
			}
		}
	}
	return "", nil
}

func isSystemDB(name string) bool {
	switch name {
	case "information_schema", "performance_schema", "mysql", "sys", "admin", "local", "config":
		return true
	}
	return false
}

func (p *Plugin) restoreDbHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	pass := getSudoPass(flags)

	// 1. List Snapshots
	fmt.Println("📋 Fetching available snapshots...")
	cmd := "bash -c 'source /etc/vps-init/restic.env && restic snapshots --json'"
	res := conn.RunSudo(cmd, pass)
	if !res.Success {
		return fmt.Errorf("failed to list snapshots: %s", res.Stderr)
	}

	// Parse JSON to extract snapshots
	type Snapshot struct {
		ID    string   `json:"short_id"`
		Time  string   `json:"time"`
		Paths []string `json:"paths"`
	}
	var snapshots []Snapshot
	if err := json.Unmarshal([]byte(res.Stdout), &snapshots); err != nil {
		return fmt.Errorf("failed to parse snapshots: %v", err)
	}

	if len(snapshots) == 0 {
		return fmt.Errorf("no snapshots found")
	}

	// 2. Display and Select Snapshot
	fmt.Println("\nAvailable Backups:")
	for i, snap := range snapshots {
		// Extract filename from path
		filename := "unknown"
		if len(snap.Paths) > 0 {
			filename = snap.Paths[0]
		}
		fmt.Printf("  [%d] %s - %s\n", i+1, snap.Time[:19], filename)
	}

	fmt.Print("\nSelect backup to restore (enter number): ")
	var snapIdx int
	_, err := fmt.Scanln(&snapIdx)
	if err != nil || snapIdx < 1 || snapIdx > len(snapshots) {
		return fmt.Errorf("invalid selection")
	}
	selectedSnap := snapshots[snapIdx-1]

	// Extract database name and engine from filename
	filename := selectedSnap.Paths[0]
	dbName := strings.TrimSuffix(strings.TrimPrefix(filename, "/"), ".sql")
	dbName = strings.TrimSuffix(dbName, ".archive")

	// Determine engine from extension
	engine := "postgres"
	if strings.HasSuffix(filename, ".sql") {
		// Could be mysql or postgres, we'll ask
		fmt.Print("Database engine (mysql/postgres): ")
		fmt.Scanln(&engine)
	} else if strings.HasSuffix(filename, ".archive") {
		engine = "mongo"
	}

	// 3. Discover Target Instances
	fmt.Println("\n🔍 Scanning for database instances...")
	instances, err := discoverInstances(conn, pass)
	if err != nil {
		return fmt.Errorf("failed to discover instances: %v", err)
	}

	// Filter by engine
	var matchingInst []DatabaseInstance
	for _, inst := range instances {
		if inst.Engine == engine {
			matchingInst = append(matchingInst, inst)
		}
	}

	if len(matchingInst) == 0 {
		return fmt.Errorf("no %s instances found", engine)
	}

	// 4. Select Target Instance
	fmt.Printf("\nFound %s Instances:\n", strings.ToUpper(engine))
	for i, inst := range matchingInst {
		source := "Host"
		if inst.Type == "docker" {
			source = fmt.Sprintf("Docker Container (%s)", inst.ContainerName)
		}
		fmt.Printf("  [%d] %s\n", i+1, source)
	}

	fmt.Print("\nSelect target instance (enter number): ")
	var instIdx int
	_, err = fmt.Scanln(&instIdx)
	if err != nil || instIdx < 1 || instIdx > len(matchingInst) {
		return fmt.Errorf("invalid selection")
	}
	targetInst := matchingInst[instIdx-1]

	// 5. Get Credentials
	detectedUser, detectedPass := detectCredentials(conn, targetInst, pass)
	var user, dbPass string
	fmt.Printf("Database User [%s]: ", detectedUser)
	fmt.Scanln(&user)
	if user == "" {
		user = detectedUser
	}

	cleanedPass := ""
	if detectedPass != "" {
		cleanedPass = "*****"
	}
	fmt.Printf("Database Password [%s]: ", cleanedPass)
	var inputPass string
	fmt.Scanln(&inputPass)
	if inputPass != "" {
		dbPass = inputPass
	} else {
		dbPass = detectedPass
	}

	// 6. Confirm Restore (DESTRUCTIVE!)
	fmt.Printf("\n⚠️  WARNING: This will OVERWRITE the '%s' database!\n", dbName)
	fmt.Print("Type 'yes' to confirm: ")
	var confirm string
	fmt.Scanln(&confirm)
	if confirm != "yes" {
		return fmt.Errorf("restore cancelled")
	}

	// 7. Perform Restore
	fmt.Printf("🔄 Restoring %s to %s instance...\n", filename, engine)

	var restoreCmd string
	switch engine {
	case "mysql":
		passFlag := ""
		if dbPass != "" {
			passFlag = fmt.Sprintf("-p'%s'", dbPass)
		}
		if targetInst.Type == "docker" {
			restoreCmd = fmt.Sprintf("bash -c 'source /etc/vps-init/restic.env && restic dump %s %s | docker exec -i %s mysql -u %s %s %s'",
				selectedSnap.ID, filename, targetInst.ContainerID, user, passFlag, dbName)
		} else {
			restoreCmd = fmt.Sprintf("bash -c 'source /etc/vps-init/restic.env && restic dump %s %s | mysql -u %s %s %s'",
				selectedSnap.ID, filename, user, passFlag, dbName)
		}

	case "postgres":
		if targetInst.Type == "docker" {
			restoreCmd = fmt.Sprintf("bash -c 'source /etc/vps-init/restic.env && restic dump %s %s | docker exec -i -e PGPASSWORD='%s' %s psql -U %s %s'",
				selectedSnap.ID, filename, dbPass, targetInst.ContainerID, user, dbName)
		} else {
			env := ""
			if dbPass != "" {
				env = fmt.Sprintf("PGPASSWORD='%s' ", dbPass)
			}
			restoreCmd = fmt.Sprintf("bash -c 'source /etc/vps-init/restic.env && %srestic dump %s %s | psql -U %s %s'",
				env, selectedSnap.ID, filename, user, dbName)
		}

	case "mongo":
		auth := ""
		if user != "" && dbPass != "" {
			auth = fmt.Sprintf("--username %s --password '%s' --authenticationDatabase admin", user, dbPass)
		}
		if targetInst.Type == "docker" {
			restoreCmd = fmt.Sprintf("bash -c 'source /etc/vps-init/restic.env && restic dump %s %s | docker exec -i %s mongorestore %s --archive'",
				selectedSnap.ID, filename, targetInst.ContainerID, auth)
		} else {
			restoreCmd = fmt.Sprintf("bash -c 'source /etc/vps-init/restic.env && restic dump %s %s | mongorestore %s --archive'",
				selectedSnap.ID, filename, auth)
		}
	}

	if res := conn.RunSudo(restoreCmd, pass); !res.Success {
		return fmt.Errorf("restore failed: %s", res.Stderr)
	}

	fmt.Println("✅ Database restored successfully.")
	return nil
}

func (p *Plugin) snapshotsHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	conn.RunInteractive("sudo bash -c 'source /etc/vps-init/restic.env && restic snapshots'")
	return nil
}

func (p *Plugin) unlockHandler(ctx context.Context, conn *ssh.Connection, args []string, flags map[string]interface{}) error {
	conn.RunInteractive("sudo bash -c 'source /etc/vps-init/restic.env && restic unlock'")
	return nil
}

func getSudoPass(flags map[string]interface{}) string {
	if v, ok := flags["sudo-password"]; ok {
		return v.(string)
	}
	return ""
}
