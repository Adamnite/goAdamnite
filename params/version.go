package params

import "fmt"

var (
	VersionMajor = 1
	VersionMinor = 0
	VersionPatch = 1
	VersionMeta  = "stable"
)

// Version holds the textual version string.
func Version() string {
	return fmt.Sprintf("%d.%d.%d", VersionMajor, VersionMinor, VersionPatch)
}

// VersionWithMeta holds the textual version string including the metadata.
func VersionWithMeta() string {
	v := Version()
	if VersionMeta != "" {
		v += "-" + VersionMeta
	}
	return v
}

func VersionWithCommit(gitCommit, gitDate string) string {
	vsn := VersionWithMeta()
	if len(gitCommit) >= 8 {
		vsn += "-" + gitCommit[:8]
	}
	if (VersionMeta != "stable") && (gitDate != "") {
		vsn += "-" + gitDate
	}
	return vsn
}
