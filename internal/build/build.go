package build

import "runtime/debug"

// These values are set at release build time via -ldflags; see the Makefile.
// For example:
//
//	-X github.com/goodblaster/vertigo/internal/build.Version=v1.2.3
var (
	Version string
	Commit  string
	Date    string
)

// BuildInfo is a snapshot of the build-time variables.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

// Info returns the build-time variables. Fields not set via -ldflags fall
// back to the Go toolchain's embedded build info, so a plain `go install`
// build still reports a version and commit.
func Info() BuildInfo {
	info := BuildInfo{
		Version: Version,
		Commit:  Commit,
		Date:    Date,
	}

	if bi, ok := debug.ReadBuildInfo(); ok {
		if info.Version == "" && bi.Main.Version != "" {
			info.Version = bi.Main.Version
		}
		for _, s := range bi.Settings {
			switch s.Key {
			case "vcs.revision":
				if info.Commit == "" && len(s.Value) >= 7 {
					info.Commit = s.Value[:7]
				}
			case "vcs.time":
				if info.Date == "" {
					info.Date = s.Value
				}
			}
		}
	}

	if info.Version == "" {
		info.Version = "dev"
	}
	return info
}
