// Package version provides version information for bab.
package version

var (
	// Version is the current version of bab.
	Version = "dev"
	// Commit is the git commit hash.
	Commit = "none"
	// Date is the build date.
	Date = "unknown"
	// BuiltBy indicates who/what built the binary.
	BuiltBy = "source"
)

// GetVersion returns the current version.
func GetVersion() string {
	return Version
}

// GetFullVersion returns the version with build metadata.
func GetFullVersion() string {
	return Version + " (" + Commit + ") built on " + Date + " by " + BuiltBy
}
