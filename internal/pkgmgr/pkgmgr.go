package pkgmgr

import (
	"fmt"
	"strings"

	"github.com/wasilwamark/vps-init/internal/distro"
)

type PackageManager interface {
	Update(packages ...string) (string, error)
	Install(packages ...string) (string, error)
	Remove(packages ...string) (string, error)
	Upgrade(packages ...string) (string, error)
	DistUpgrade() (string, error)
	Autoremove() (string, error)
	Search(query string) (string, error)
}

type APT struct{}

func NewAPT() *APT {
	return &APT{}
}

func (a *APT) Update(packages ...string) (string, error) {
	cmd := "apt-get update"
	if len(packages) > 0 {
		cmd = fmt.Sprintf("apt-get install -y %s", strings.Join(packages, " "))
	}
	return cmd, nil
}

func (a *APT) Install(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for installation")
	}
	return fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get install -y %s", strings.Join(packages, " ")), nil
}

func (a *APT) Remove(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for removal")
	}
	return fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get remove -y %s", strings.Join(packages, " ")), nil
}

func (a *APT) Upgrade(packages ...string) (string, error) {
	if len(packages) > 0 {
		return fmt.Sprintf("DEBIAN_FRONTEND=noninteractive apt-get upgrade -y %s", strings.Join(packages, " ")), nil
	}
	return "DEBIAN_FRONTEND=noninteractive apt-get upgrade -y", nil
}

func (a *APT) DistUpgrade() (string, error) {
	return "DEBIAN_FRONTEND=noninteractive apt-get dist-upgrade -y", nil
}

func (a *APT) Autoremove() (string, error) {
	return "DEBIAN_FRONTEND=noninteractive apt-get autoremove -y", nil
}

func (a *APT) Search(query string) (string, error) {
	return fmt.Sprintf("apt-cache search %s", query), nil
}

type DNF struct{}

func NewDNF() *DNF {
	return &DNF{}
}

func (d *DNF) Update(packages ...string) (string, error) {
	cmd := "dnf check-update"
	if len(packages) > 0 {
		cmd = fmt.Sprintf("dnf install -y %s", strings.Join(packages, " "))
	}
	return cmd, nil
}

func (d *DNF) Install(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for installation")
	}
	return fmt.Sprintf("dnf install -y %s", strings.Join(packages, " ")), nil
}

func (d *DNF) Remove(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for removal")
	}
	return fmt.Sprintf("dnf remove -y %s", strings.Join(packages, " ")), nil
}

func (d *DNF) Upgrade(packages ...string) (string, error) {
	if len(packages) > 0 {
		return fmt.Sprintf("dnf upgrade -y %s", strings.Join(packages, " ")), nil
	}
	return "dnf upgrade -y", nil
}

func (d *DNF) DistUpgrade() (string, error) {
	return "dnf upgrade -y", nil
}

func (d *DNF) Autoremove() (string, error) {
	return "dnf autoremove -y", nil
}

func (d *DNF) Search(query string) (string, error) {
	return fmt.Sprintf("dnf search %s", query), nil
}

type YUM struct{}

func NewYUM() *YUM {
	return &YUM{}
}

func (y *YUM) Update(packages ...string) (string, error) {
	cmd := "yum check-update"
	if len(packages) > 0 {
		cmd = fmt.Sprintf("yum install -y %s", strings.Join(packages, " "))
	}
	return cmd, nil
}

func (y *YUM) Install(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for installation")
	}
	return fmt.Sprintf("yum install -y %s", strings.Join(packages, " ")), nil
}

func (y *YUM) Remove(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for removal")
	}
	return fmt.Sprintf("yum remove -y %s", strings.Join(packages, " ")), nil
}

func (y *YUM) Upgrade(packages ...string) (string, error) {
	if len(packages) > 0 {
		return fmt.Sprintf("yum update -y %s", strings.Join(packages, " ")), nil
	}
	return "yum update -y", nil
}

func (y *YUM) DistUpgrade() (string, error) {
	return "yum update -y", nil
}

func (y *YUM) Autoremove() (string, error) {
	return "yum autoremove -y", nil
}

func (y *YUM) Search(query string) (string, error) {
	return fmt.Sprintf("yum search %s", query), nil
}

type Pacman struct{}

func NewPacman() *Pacman {
	return &Pacman{}
}

func (p *Pacman) Update(packages ...string) (string, error) {
	cmd := "pacman -Sy"
	if len(packages) > 0 {
		cmd = fmt.Sprintf("pacman -S --noconfirm %s", strings.Join(packages, " "))
	}
	return cmd, nil
}

func (p *Pacman) Install(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for installation")
	}
	return fmt.Sprintf("pacman -S --noconfirm %s", strings.Join(packages, " ")), nil
}

func (p *Pacman) Remove(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for removal")
	}
	return fmt.Sprintf("pacman -R --noconfirm %s", strings.Join(packages, " ")), nil
}

func (p *Pacman) Upgrade(packages ...string) (string, error) {
	if len(packages) > 0 {
		return fmt.Sprintf("pacman -S --noconfirm %s", strings.Join(packages, " ")), nil
	}
	return "pacman -Syu --noconfirm", nil
}

func (p *Pacman) DistUpgrade() (string, error) {
	return "pacman -Syu --noconfirm", nil
}

func (p *Pacman) Autoremove() (string, error) {
	return "pacman -Sc --noconfirm", nil
}

func (p *Pacman) Search(query string) (string, error) {
	return fmt.Sprintf("pacman -Ss %s", query), nil
}

type APK struct{}

func NewAPK() *APK {
	return &APK{}
}

func (a *APK) Update(packages ...string) (string, error) {
	cmd := "apk update"
	if len(packages) > 0 {
		cmd = fmt.Sprintf("apk add %s", strings.Join(packages, " "))
	}
	return cmd, nil
}

func (a *APK) Install(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for installation")
	}
	return fmt.Sprintf("apk add %s", strings.Join(packages, " ")), nil
}

func (a *APK) Remove(packages ...string) (string, error) {
	if len(packages) == 0 {
		return "", fmt.Errorf("no packages specified for removal")
	}
	return fmt.Sprintf("apk del %s", strings.Join(packages, " ")), nil
}

func (a *APK) Upgrade(packages ...string) (string, error) {
	if len(packages) > 0 {
		return fmt.Sprintf("apk add --upgrade %s", strings.Join(packages, " ")), nil
	}
	return "apk upgrade", nil
}

func (a *APK) DistUpgrade() (string, error) {
	return "apk upgrade", nil
}

func (a *APK) Autoremove() (string, error) {
	return "apk cache clean", nil
}

func (a *APK) Search(query string) (string, error) {
	return fmt.Sprintf("apk search %s", query), nil
}

func GetPackageManager(distroInfo *distro.DistroInfo) PackageManager {
	switch distroInfo.PackageMgr {
	case distro.PackageManagerAPT:
		return NewAPT()
	case distro.PackageManagerDNF:
		return NewDNF()
	case distro.PackageManagerYUM:
		return NewYUM()
	case distro.PackageManagerPacman:
		return NewPacman()
	case distro.PackageManagerAPK:
		return NewAPK()
	default:
		return NewAPT()
	}
}
