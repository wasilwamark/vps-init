package version

import (
	"fmt"
	"runtime"
)

// Version information
var (
	// Version is the current semantic version
	Version = "0.1.0"

	// GitCommit is the git commit hash (set during build)
	GitCommit = ""

	// BuildDate is the build date (set during build)
	BuildDate = ""

	// GoVersion is the Go version used to build
	GoVersion = runtime.Version()
)

// VersionInfo contains detailed version information
type VersionInfo struct {
	Version   string `json:"version"`
	GitCommit string `json:"git_commit,omitempty"`
	BuildDate string `json:"build_date,omitempty"`
	GoVersion string `json:"go_version"`
	Platform  string `json:"platform"`
}

// GetVersion returns detailed version information
func GetVersion() VersionInfo {
	return VersionInfo{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
		GoVersion: GoVersion,
		Platform:  runtime.GOOS + "/" + runtime.GOARCH,
	}
}

// String returns the version string
func (v VersionInfo) String() string {
	version := fmt.Sprintf("vps-init v%s", v.Version)
	if v.GitCommit != "" {
		shortCommit := v.GitCommit
		if len(shortCommit) > 8 {
			shortCommit = shortCommit[:8]
		}
		version += fmt.Sprintf(" (%s)", shortCommit)
	}
	if v.BuildDate != "" {
		version += fmt.Sprintf(" built %s", v.BuildDate)
	}
	return version
}

// ShortVersion returns just the version number
func ShortVersion() string {
	return Version
}