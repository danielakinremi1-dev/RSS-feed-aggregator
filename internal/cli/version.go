package cli

import "runtime/debug"

// Version is set at link time by release builds:
//
//	-ldflags "-X github.com/danielakinremi1-dev/RSS-feed-aggregator/internal/cli.Version=v1.2.3"
var Version = "dev"

// version reports Version, falling back to the module version Go embeds for
// `go install ...@vX.Y.Z` builds, which don't set ldflags.
func version() string {
	if Version != "dev" {
		return Version
	}
	if info, ok := debug.ReadBuildInfo(); ok && info.Main.Version != "" && info.Main.Version != "(devel)" {
		return info.Main.Version
	}
	return Version
}
