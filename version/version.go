package version

var (
	Version   = "dev"
	GitCommit = "unknown"
	BuildTime = "unknown"
)

type Info struct {
	Version   string `json:"version"`
	GitCommit string `json:"gitCommit"`
	BuildTime string `json:"buildTime"`
}

func Get() Info {
	return Info{
		Version:   Version,
		GitCommit: GitCommit,
		BuildTime: BuildTime,
	}
}
