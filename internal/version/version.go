package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "source"
)

func HasBuildInfo() bool {
	return Commit != "none"
}
