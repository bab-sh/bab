package version

var (
	Version = "dev"
	Commit  = "none"
	Date    = "unknown"
	BuiltBy = "source"
)

func GetVersion() string {
	return Version
}

func GetFullVersion() string {
	return Version + " (" + Commit + ") built on " + Date + " by " + BuiltBy
}
