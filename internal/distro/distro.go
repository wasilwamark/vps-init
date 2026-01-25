package distro

import (
	"fmt"
	"strings"
)

type PackageManager string

const (
	PackageManagerAPT    PackageManager = "apt"
	PackageManagerDNF    PackageManager = "dnf"
	PackageManagerYUM    PackageManager = "yum"
	PackageManagerPacman PackageManager = "pacman"
	PackageManagerAPK    PackageManager = "apk"
)

type ServiceManager string

const (
	ServiceManagerSystemd ServiceManager = "systemd"
	ServiceManagerInitD   ServiceManager = "init.d"
	ServiceManagerOpenRC  ServiceManager = "openrc"
)

type DistroFamily string

const (
	DistroFamilyDebian DistroFamily = "debian"
	DistroFamilyRedHat DistroFamily = "redhat"
	DistroFamilyArch   DistroFamily = "arch"
	DistroFamilyAlpine DistroFamily = "alpine"
)

type DistroInfo struct {
	ID         string
	IDLike     string
	Name       string
	Version    string
	VersionID  string
	Family     DistroFamily
	PackageMgr PackageManager
	ServiceMgr ServiceManager
	IDLikeList []string
}

type OSRelease struct {
	ID        string
	IDLike    string
	Name      string
	Version   string
	VersionID string
}

func DetectOSRelease(content string) (*OSRelease, error) {
	release := &OSRelease{}
	lines := strings.Split(content, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if !strings.HasPrefix(line, "#") && strings.Contains(line, "=") {
			parts := strings.SplitN(line, "=", 2)
			if len(parts) != 2 {
				continue
			}

			key := strings.TrimSpace(parts[0])
			value := strings.Trim(strings.TrimSpace(parts[1]), "\"")

			switch key {
			case "ID":
				release.ID = value
			case "ID_LIKE":
				release.IDLike = value
			case "NAME":
				release.Name = value
			case "VERSION":
				release.Version = value
			case "VERSION_ID":
				release.VersionID = value
			}
		}
	}

	if release.ID == "" {
		return nil, fmt.Errorf("invalid /etc/os-release: missing ID field")
	}

	return release, nil
}

func GetDistroInfo(osRelease *OSRelease) *DistroInfo {
	info := &DistroInfo{
		ID:         osRelease.ID,
		IDLike:     osRelease.IDLike,
		Name:       osRelease.Name,
		Version:    osRelease.Version,
		VersionID:  osRelease.VersionID,
		IDLikeList: parseIDLike(osRelease.IDLike),
	}

	idLikes := info.IDLikeList
	idLikes = append(idLikes, osRelease.ID)

	for _, id := range idLikes {
		switch {
		case id == "ubuntu" || id == "debian":
			info.Family = DistroFamilyDebian
			info.PackageMgr = PackageManagerAPT
			info.ServiceMgr = ServiceManagerSystemd
			return info
		case id == "rhel" || id == "centos" || id == "fedora":
			info.Family = DistroFamilyRedHat
			if id == "fedora" || (info.VersionID != "" && strings.HasPrefix(info.VersionID, "8") || strings.HasPrefix(info.VersionID, "9")) {
				info.PackageMgr = PackageManagerDNF
			} else {
				info.PackageMgr = PackageManagerYUM
			}
			info.ServiceMgr = ServiceManagerSystemd
			return info
		case id == "arch" || id == "archarm":
			info.Family = DistroFamilyArch
			info.PackageMgr = PackageManagerPacman
			info.ServiceMgr = ServiceManagerSystemd
			return info
		case id == "alpine":
			info.Family = DistroFamilyAlpine
			info.PackageMgr = PackageManagerAPK
			info.ServiceMgr = ServiceManagerOpenRC
			return info
		}
	}

	info.Family = DistroFamilyDebian
	info.PackageMgr = PackageManagerAPT
	info.ServiceMgr = ServiceManagerSystemd
	return info
}

func parseIDLike(idLike string) []string {
	if idLike == "" {
		return []string{}
	}

	var result []string
	for _, item := range strings.Split(idLike, " ") {
		item = strings.TrimSpace(item)
		if item != "" {
			result = append(result, item)
		}
	}
	return result
}

func (d *DistroInfo) IsUbuntu() bool {
	return d.ID == "ubuntu" || contains(d.IDLikeList, "ubuntu")
}

func (d *DistroInfo) IsDebian() bool {
	return d.ID == "debian" || contains(d.IDLikeList, "debian")
}

func (d *DistroInfo) IsCentOS() bool {
	return d.ID == "centos"
}

func (d *DistroInfo) IsRedHat() bool {
	return d.ID == "rhel" || contains(d.IDLikeList, "rhel")
}

func (d *DistroInfo) IsFedora() bool {
	return d.ID == "fedora"
}

func (d *DistroInfo) IsArch() bool {
	return d.ID == "arch" || contains(d.IDLikeList, "arch")
}

func (d *DistroInfo) IsAlpine() bool {
	return d.ID == "alpine"
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
