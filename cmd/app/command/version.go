package command

import (
	"fmt"
	"runtime"
)

// version is stamped at build time from `git describe`; see bin/build.sh. The
// fallback is what a plain `go build` produces, and says so rather than naming
// a revision the binary may not have been built from.
var version = "dev"

// printVersion answers --version on stdout: it is the output that was asked
// for, not a usage hint, so it pipes like any other payload. The Go version and
// platform come free and are the first thing worth knowing when one machine
// behaves differently from another.
func printVersion() {
	fmt.Printf("%s %s\n", defaultProgramName, version)
	fmt.Printf("%s %s/%s\n", runtime.Version(), runtime.GOOS, runtime.GOARCH)
}
