// Command codeforge is the command line tool of codeforge.
package main

import (
	"os"
	"runtime/debug"
	"strings"

	"github.com/pointjungle/codeforge-cli/internal/cli"
)

// Set by GoReleaser (-ldflags "-X main.version=…").
var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	app := cli.NewApp(cli.Build{Version: buildVersion(), Commit: commit, Date: date})
	os.Exit(app.Run(os.Args[1:]))
}

// buildVersion is GoReleaser's version, or the module version recorded by
// `go install …@vX.Y.Z` when the binary was not built by GoReleaser.
func buildVersion() string {
	if version != "dev" {
		return version
	}
	if info, ok := debug.ReadBuildInfo(); ok && strings.HasPrefix(info.Main.Version, "v") {
		return strings.TrimPrefix(info.Main.Version, "v")
	}
	return version
}
