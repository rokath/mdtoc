package main

import (
	"fmt"
	"io"
	"os"

	"example.com/mdtoc/internal/mdtoc"
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

// main is intentionally tiny. All logic lives in the internal package so the
// command runner can be tested without spawning subprocesses.
func main() {
	exitCode, err := run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr, mdtoc.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	exitFunc(exitCode)
}
