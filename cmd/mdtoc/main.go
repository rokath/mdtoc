package main

import (
	"fmt"
	"os"

	"example.com/mdtoc/internal/mdtoc"
)

var (
	version string // do not initialize, goreleaser will handle that
	commit  string // do not initialize, goreleaser will handle that
	date    string // do not initialize, goreleaser will handle that
)

// main is intentionally tiny. All logic lives in the internal package so the
// command runner can be tested without spawning subprocesses.
func main() {
	runner := mdtoc.NewRunnerWithBuildInfo(os.Stdin, os.Stdout, os.Stderr, mdtoc.BuildInfo{
		Version: version,
		Commit:  commit,
		Date:    date,
	})
	exitCode, err := runner.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
