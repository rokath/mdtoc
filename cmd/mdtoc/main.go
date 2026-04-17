package main

import (
	"fmt"
	"os"

	"example.com/mdtoc/internal/mdtoc"
)

// main is intentionally tiny. All logic lives in the internal package so the
// command runner can be tested without spawning subprocesses.
func main() {
	runner := mdtoc.NewRunner(os.Stdin, os.Stdout, os.Stderr)
	exitCode, err := runner.Run(os.Args[1:])
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	os.Exit(exitCode)
}
