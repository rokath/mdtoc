package mdtoc

import (
	"strings"
	"testing"
)

func TestRunnerRootHelpShortFlagWithoutCommand(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(""), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"-h"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("stdout does not contain root help:\n%s", stdout.String())
	}
}

func TestRunnerRootVersionWithoutCommand(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(""), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"--version"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); !strings.Contains(got, "mdtoc "+version) {
		t.Fatalf("stdout does not contain version output:\n%s", got)
	}
}

func TestRunnerSubcommandVerboseHelpIsNotIgnored(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(""), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"generate", "--help", "--verbose"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); !strings.Contains(got, "Generate or update ToC, heading numbers, and anchors.") {
		t.Fatalf("stdout does not contain verbose generate help:\n%s", got)
	}
}

func TestRunnerGenerateFromStdin(t *testing.T) {
	stdin := strings.NewReader("# Title\n\n## Intro\n")
	var stdout, stderr strings.Builder
	runner := NewRunner(stdin, &stdout, &stderr)
	exitCode, err := runner.Run([]string{"generate"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), "* [1. Intro](#intro)") {
		t.Fatalf("stdout does not contain generated ToC:\n%s", stdout.String())
	}
}

func TestRunnerCheckExitCodeOnMismatch(t *testing.T) {
	stdin := strings.NewReader(strings.Join([]string{startMarker, "* [1. Wrong](#wrong)", configStart, "numbering=on", "min-level=2", "max-level=4", "anchors=on", "toc=on", "state=generated", configEnd, endMarker, "", "## Intro"}, "\n") + "\n")
	var stdout, stderr strings.Builder
	runner := NewRunner(stdin, &stdout, &stderr)
	exitCode, err := runner.Run([]string{"check"})
	if err == nil {
		t.Fatalf("Run returned nil error on mismatch")
	}
	if exitCode != 2 {
		t.Fatalf("exit code = %d, want 2", exitCode)
	}
}
