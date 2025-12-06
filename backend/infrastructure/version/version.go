package version

// Version information - set via ldflags during build
// go build -ldflags "-X shotgun_code/infrastructure/version.Version=v1.0.0 -X shotgun_code/infrastructure/version.GitCommit=abc123"
var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildDate = "unknown"
)

// Info returns version information
type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildDate string `json:"buildDate"`
}

// GetInfo returns current version info
func GetInfo() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildDate: BuildDate,
	}
}
