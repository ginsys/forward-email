package version

import (
	"fmt"
	"runtime"
)

var (
	// Version is the semantic version of the application
	Version = "dev"
	// Commit is the git commit hash
	Commit = "none"
	// Date is the build date
	Date = "unknown"
	// GoVersion is the Go version used to build the binary
	GoVersion = runtime.Version()
)

// Info represents version information
type Info struct {
	Version   string `json:"version"`
	Commit    string `json:"commit"`
	Date      string `json:"date"`
	GoVersion string `json:"go_version"`
	OS        string `json:"os"`
	Arch      string `json:"arch"`
}

// Get returns version information
func Get() Info {
	return Info{
		Version:   Version,
		Commit:    Commit,
		Date:      Date,
		GoVersion: GoVersion,
		OS:        runtime.GOOS,
		Arch:      runtime.GOARCH,
	}
}

// String returns a formatted version string
func (i Info) String() string {
	return fmt.Sprintf("forward-email version %s\ncommit: %s\nbuilt: %s\ngo: %s\nos/arch: %s/%s",
		i.Version, i.Commit, i.Date, i.GoVersion, i.OS, i.Arch)
}
