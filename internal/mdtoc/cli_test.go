package mdtoc

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestRunnerRootHelpShortFlagWithoutCommand verifies that root short help is printed for -h.
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

// TestRunnerRootVersionWithoutCommand verifies non-verbose root version output.
func TestRunnerRootVersionWithoutCommand(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := NewRunnerWithBuildInfo(strings.NewReader(""), &stdout, &stderr, BuildInfo{
		Version: "v1.2.3",
		Commit:  "abcdef0",
		Date:    "2026-04-17T12:34:56Z",
	})
	exitCode, err := runner.Run([]string{"--version"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	got := stdout.String()
	for _, want := range []string{"mdtoc v1.2.3", "commit: abcdef0", "date: 2026-04-17T12:34:56Z"} {
		if !strings.Contains(got, want) {
			t.Fatalf("stdout does not contain %q:\n%s", want, got)
		}
	}
}

// TestRunnerRootVersionVerboseWithoutCommand verifies verbose root version output.
func TestRunnerRootVersionVerboseWithoutCommand(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := NewRunnerWithBuildInfo(strings.NewReader(""), &stdout, &stderr, BuildInfo{
		Version: "v1.2.3",
		Commit:  "abcdef0",
		Date:    "2026-04-17T12:34:56Z",
	})
	exitCode, err := runner.Run([]string{"--version", "--verbose"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); !strings.Contains(got, "Go-based Markdown ToC manager") {
		t.Fatalf("stdout does not contain version output:\n%s", got)
	}
}

// TestRunnerSubcommandVerboseHelpIsNotIgnored verifies verbose subcommand help text selection.
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

// TestRunnerSubcommandVerboseWithoutHelpPrintsLongHelp verifies verbose-only rootless subcommand help behavior.
func TestRunnerSubcommandVerboseWithoutHelpPrintsLongHelp(t *testing.T) {
	tests := []struct {
		args []string
		want string
	}{
		{args: []string{"generate", "--verbose"}, want: "Generate or update ToC, heading numbers, and anchors."},
		{args: []string{"regen", "--verbose"}, want: "Regenerate using the persisted config from an existing managed container."},
		{args: []string{"strip", "--verbose"}, want: "Remove managed artifacts and optionally the entire managed container."},
		{args: []string{"check", "--verbose"}, want: "Reconstruct the target document state and compare it byte-for-byte."},
	}

	for _, tc := range tests {
		var stdout, stderr strings.Builder
		runner := NewRunner(strings.NewReader(""), &stdout, &stderr)
		exitCode, err := runner.Run(tc.args)
		if err != nil {
			t.Fatalf("Run(%v) error: %v", tc.args, err)
		}
		if exitCode != 0 {
			t.Fatalf("Run(%v) exit code = %d, want 0", tc.args, exitCode)
		}
		if got := stdout.String(); !strings.Contains(got, tc.want) {
			t.Fatalf("Run(%v) did not print long help:\n%s", tc.args, got)
		}
	}
}

// TestRunnerGenerateFromStdin verifies that generate accepts piped stdin content.
func TestRunnerGenerateFromStdin(t *testing.T) {
	stdin := strings.NewReader("# Title\n\n- item\n- item\n\n## Intro\n")
	var stdout, stderr strings.Builder
	runner := NewRunner(stdin, &stdout, &stderr)
	exitCode, err := runner.Run([]string{"generate"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if !strings.Contains(stdout.String(), "- [1. Intro](#intro)") {
		t.Fatalf("stdout does not contain generated ToC:\n%s", stdout.String())
	}
}

// TestRunnerGenerateAcceptsBulletsOverride verifies the explicit CLI bullet selection.
func TestRunnerGenerateAcceptsBulletsOverride(t *testing.T) {
	stdin := strings.NewReader("# Title\n\n- item\n- item\n\n## Intro\n")
	var stdout, stderr strings.Builder
	runner := NewRunner(stdin, &stdout, &stderr)
	exitCode, err := runner.Run([]string{"generate", "--bullets", "+"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); !strings.Contains(got, "bullets=+") || !strings.Contains(got, "+ [1. Intro](#intro)") {
		t.Fatalf("stdout does not contain forced bullet output:\n%s", got)
	}
}

// TestRunnerRegenFromStdin verifies that regen reuses persisted container config from stdin.
func TestRunnerRegenFromStdin(t *testing.T) {
	input, _, err := Generate("# Title\n\n## Intro\n", Options{
		Numbering: false,
		MinLevel:  2,
		MaxLevel:  4,
		Anchor:    AnchorOff,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	input = strings.Replace(input, "* [Intro](#intro)", "* [BROKEN](#intro)", 1)

	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(input), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"regen"})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}
	if got := stdout.String(); !strings.Contains(got, "* [Intro](#intro)") || strings.Contains(got, "<a id=") {
		t.Fatalf("stdout does not contain regen output honoring stored config:\n%s", got)
	}
}

// TestRunnerGenerateFailsFastOnInteractiveStdinWithoutFile verifies issue #4 behavior for generate.
func TestRunnerGenerateFailsFastOnInteractiveStdinWithoutFile(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := newRunner(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true)
	exitCode, err := runner.Run([]string{"generate"})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if got := err.Error(); !strings.Contains(got, "no input provided") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunnerRegenFailsFastOnInteractiveStdinWithoutFile verifies issue #4 behavior for regen.
func TestRunnerRegenFailsFastOnInteractiveStdinWithoutFile(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := newRunner(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true)
	exitCode, err := runner.Run([]string{"regen"})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
}

// TestRunnerStripFailsFastOnInteractiveStdinWithoutFile verifies issue #4 behavior for strip.
func TestRunnerStripFailsFastOnInteractiveStdinWithoutFile(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := newRunner(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true)
	exitCode, err := runner.Run([]string{"strip"})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
}

// TestRunnerCheckFailsFastOnInteractiveStdinWithoutFile verifies issue #4 behavior for check.
func TestRunnerCheckFailsFastOnInteractiveStdinWithoutFile(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := newRunner(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true)
	exitCode, err := runner.Run([]string{"check"})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
}

// TestRunnerGenerateWithFileDoesNotRequireStdin verifies that -f bypasses interactive stdin checks.
func TestRunnerGenerateWithFileDoesNotRequireStdin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte("# Title\n\n## Intro\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var stdout, stderr strings.Builder
	runner := newRunner(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true)
	exitCode, err := runner.Run([]string{"generate", "-f", path})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if !strings.Contains(string(got), "* [1. Intro](#intro)") {
		t.Fatalf("generated file does not contain ToC:\n%s", string(got))
	}
}

// TestRunnerRejectsMixedFileAndStdinGenerate verifies that generate rejects simultaneous file and stdin input.
func TestRunnerRejectsMixedFileAndStdinGenerate(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	if err := os.WriteFile(path, []byte("# Title\n\n## Intro\n"), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader("# From stdin\n\n## Ignored?\n"), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"generate", "-f", path})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if got := err.Error(); !strings.Contains(got, "cannot use --file together with piped stdin") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunnerRegenWithFileDoesNotRequireStdin verifies that regen honors -f input files.
func TestRunnerRegenWithFileDoesNotRequireStdin(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	input, _, err := Generate("# Title\n\n## Intro\n", Options{
		Numbering: false,
		MinLevel:  2,
		MaxLevel:  4,
		Anchor:    AnchorOff,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	input = strings.Replace(input, "* [Intro](#intro)", "* [BROKEN](#intro)", 1)
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var stdout, stderr strings.Builder
	runner := newRunner(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true)
	exitCode, err := runner.Run([]string{"regen", "-f", path})
	if err != nil {
		t.Fatalf("Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exit code = %d, want 0", exitCode)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	if s := string(got); !strings.Contains(s, "* [Intro](#intro)") || strings.Contains(s, "<a id=") {
		t.Fatalf("regen file output does not honor stored config:\n%s", s)
	}
}

// TestRunnerRejectsMixedFileAndStdinRegen verifies that regen rejects simultaneous file and stdin input.
func TestRunnerRejectsMixedFileAndStdinRegen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	input, _, err := Generate("# Title\n\n## Intro\n", Options{
		Numbering: false,
		MinLevel:  2,
		MaxLevel:  4,
		Anchor:    AnchorOff,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(input), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"regen", "-f", path})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if got := err.Error(); !strings.Contains(got, "cannot use --file together with piped stdin") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunnerRejectsMixedFileAndStdinStrip verifies that strip rejects simultaneous file and stdin input.
func TestRunnerRejectsMixedFileAndStdinStrip(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	input, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(input), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"strip", "-f", path})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if got := err.Error(); !strings.Contains(got, "cannot use --file together with piped stdin") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunnerRejectsMixedFileAndStdinCheck verifies that check rejects simultaneous file and stdin input.
func TestRunnerRejectsMixedFileAndStdinCheck(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "doc.md")
	input, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if err := os.WriteFile(path, []byte(input), 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(input), &stdout, &stderr)
	exitCode, err := runner.Run([]string{"check", "-f", path})
	if err == nil {
		t.Fatalf("Run returned nil error")
	}
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if got := err.Error(); !strings.Contains(got, "cannot use --file together with piped stdin") {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestRunnerGenerateThenCheckWithFixtureFile verifies the real file workflow used by install checks.
func TestRunnerGenerateThenCheckWithFixtureFile(t *testing.T) {
	fixturePath := filepath.Join("..", "..", ".github", "fixtures", "install-smoke.md")
	input, err := os.ReadFile(fixturePath)
	if err != nil {
		t.Fatalf("ReadFile(%q) error: %v", fixturePath, err)
	}

	dir := t.TempDir()
	path := filepath.Join(dir, "install-smoke.md")
	if err := os.WriteFile(path, input, 0o644); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}

	var stdout, stderr strings.Builder
	runner := newRunner(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true)

	exitCode, err := runner.Run([]string{"generate", "-f", path})
	if err != nil {
		t.Fatalf("generate Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("generate exit code = %d, want 0", exitCode)
	}

	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("ReadFile error: %v", err)
	}
	content := string(got)
	for _, want := range []string{
		"bullets=auto",
		"+ [1. 2026 Release Plan](#2026-release-plan)",
		"+ [3. Overview](#overview-1)",
		`## 1. <a id="2026-release-plan"></a>2026 Release Plan`,
		`## 3. <a id="overview-1"></a>Overview`,
		"## Hidden Section",
		"### Hidden Details",
	} {
		if !strings.Contains(content, want) {
			t.Fatalf("generated fixture file missing %q:\n%s", want, content)
		}
	}

	exitCode, err = runner.Run([]string{"check", "-f", path})
	if err != nil {
		t.Fatalf("check Run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("check exit code = %d, want 0", exitCode)
	}
}

// TestRunnerCheckExitCodeOnMismatch verifies the special mismatch exit code from check.
func TestRunnerCheckExitCodeOnMismatch(t *testing.T) {
	stdin := strings.NewReader(strings.Join([]string{startMarker, "* [1. Wrong](#wrong)", configStart, "numbering=on", "min-level=2", "max-level=4", "anchor=github", "toc=on", "state=generated", configEnd, endMarker, "", "## Intro"}, "\n") + "\n")
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
