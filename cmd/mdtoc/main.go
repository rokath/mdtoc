package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/rokath/mdtoc/internal/mdtoc"
)

var (
	version string // do not initialize, goreleaser will handle that
	commit  string // do not initialize, goreleaser will handle that
	date    string // do not initialize, goreleaser will handle that
)

var exitFunc = os.Exit

// run wires the CLI runner to the provided streams and build metadata.
func run(args []string, stdin io.Reader, stdout, stderr io.Writer, buildInfo mdtoc.BuildInfo) (int, error) {
	runner := mdtoc.NewRunnerWithBuildInfo(stdin, stdout, stderr, buildInfo)
	return runner.Run(args)
}

// buildInfoFromLinkerVars combines release metadata injected through ldflags
// with local-build fallbacks. GoReleaser provides all three package variables;
// plain `go install` leaves them empty, so only the date needs an executable
// timestamp fallback before the internal CLI applies its normal defaults.
func buildInfoFromLinkerVars() mdtoc.BuildInfo {
	info := mdtoc.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	}
	if info.Date == "" {
		info.Date = executableBuildDate()
	}
	return info
}

// executableBuildDate returns the modification time of the running executable
// in the same RFC3339 UTC format used by release builds. An empty result lets
// the internal BuildInfo normalization keep its existing `unknown` fallback for
// unusual platforms where the executable cannot be resolved or statted.
func executableBuildDate() string {
	path, err := os.Executable()
	if err != nil {
		return ""
	}
	return executableBuildDateForPath(path)
}

// executableBuildDateForPath is split out so tests can verify the timestamp
// conversion deterministically without spawning a real installed binary.
func executableBuildDateForPath(path string) string {
	stat, err := os.Stat(path)
	if err != nil {
		return ""
	}
	return stat.ModTime().UTC().Format(time.RFC3339)
}

// main is intentionally tiny. All logic lives in the internal package so the
// command runner can be tested without spawning subprocesses.
func main() {
	exitCode, err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, buildInfoFromLinkerVars())
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	exitFunc(exitCode)
}
