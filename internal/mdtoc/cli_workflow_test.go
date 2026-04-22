package mdtoc

import (
	"fmt"
	"os"
	"strings"
	"testing"
)

// memoryFileSystem stores test files in memory for end-to-end CLI workflow tests.
type memoryFileSystem struct {
	files map[string][]byte
}

// ReadFile returns a copy of the stored file content.
func (fs *memoryFileSystem) ReadFile(name string) ([]byte, error) {
	data, ok := fs.files[name]
	if !ok {
		return nil, fmt.Errorf("file does not exist: %s", name)
	}
	return append([]byte(nil), data...), nil
}

// WriteFile stores a copy of the written file content.
func (fs *memoryFileSystem) WriteFile(name string, data []byte, _ os.FileMode) error {
	fs.files[name] = append([]byte(nil), data...)
	return nil
}

// newMemoryFileSystem creates a memory-backed file system with initial file content.
func newMemoryFileSystem(files map[string]string) *memoryFileSystem {
	mem := &memoryFileSystem{files: make(map[string][]byte, len(files))}
	for name, content := range files {
		mem.files[name] = []byte(content)
	}
	return mem
}

// fileString returns the stored file as a string for assertions.
func (fs *memoryFileSystem) fileString(name string) string {
	return string(fs.files[name])
}

// runFileCommand executes one file-based CLI command against the provided memory filesystem.
func runFileCommand(t *testing.T, fs *memoryFileSystem, args ...string) string {
	t.Helper()

	var stdout, stderr strings.Builder
	runner := newRunnerWithFS(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true, fs)
	exitCode, err := runner.Run(args)
	if err != nil {
		t.Fatalf("Run(%v) error: %v", args, err)
	}
	if exitCode != 0 {
		t.Fatalf("Run(%v) exit code = %d, want 0", args, exitCode)
	}
	if got := stderr.String(); got != "" {
		t.Fatalf("Run(%v) wrote unexpected stderr:\n%s", args, got)
	}
	return stdout.String()
}

// runFileCommandExpect executes one file-based CLI command and asserts its exact exit code.
func runFileCommandExpect(t *testing.T, fs *memoryFileSystem, wantExit int, args ...string) error {
	t.Helper()

	var stdout, stderr strings.Builder
	runner := newRunnerWithFS(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true, fs)
	exitCode, err := runner.Run(args)
	if exitCode != wantExit {
		t.Fatalf("Run(%v) exit code = %d, want %d, err=%v", args, exitCode, wantExit, err)
	}
	return err
}

// TestRunnerFileWorkflowGenerateStripRegenCheck verifies the full file-based state cycle.
func TestRunnerFileWorkflowGenerateStripRegenCheck(t *testing.T) {
	const path = "README.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	runFileCommand(t, fs, "generate", "-f", path, "--min-level=1")
	generated := fs.fileString(path)
	if !strings.Contains(generated, "state=generated") || !strings.Contains(generated, "<a id=\"title\"></a>") {
		t.Fatalf("generate did not create managed state:\n%s", generated)
	}

	runFileCommand(t, fs, "strip", "-f", path)
	stripped := fs.fileString(path)
	if !strings.Contains(stripped, "state=stripped") || strings.Contains(stripped, "<a id=") {
		t.Fatalf("strip did not produce stripped state:\n%s", stripped)
	}

	runFileCommand(t, fs, "regen", "-f", path)
	regenerated := fs.fileString(path)
	if !strings.Contains(regenerated, "state=generated") || !strings.Contains(regenerated, "<a id=\"title\"></a>") {
		t.Fatalf("regen did not restore generated state:\n%s", regenerated)
	}

	if err := runFileCommandExpect(t, fs, 0, "check", "-f", path); err != nil {
		t.Fatalf("check unexpectedly failed after regen: %v", err)
	}
}

// TestRunnerFileWorkflowRegenRestoresStoredFlags verifies regen after stripped input with non-default stored config.
func TestRunnerFileWorkflowRegenRestoresStoredFlags(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n\n### API\n",
	})

	runFileCommand(t, fs, "generate", "-f", path, "--min-level=2", "--max-level=3", "--anchors=off", "--numbering=off")
	runFileCommand(t, fs, "strip", "-f", path)
	runFileCommand(t, fs, "regen", "-f", path)

	got := fs.fileString(path)
	if !strings.Contains(got, "state=generated") {
		t.Fatalf("regen did not restore generated state:\n%s", got)
	}
	if strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") || strings.Contains(got, "### 1.1. ") {
		t.Fatalf("regen ignored stored disabled flags:\n%s", got)
	}
	if !strings.Contains(got, "* [Intro](#intro)") || !strings.Contains(got, "* [API](#api)") {
		t.Fatalf("regen did not rebuild ToC using stored config:\n%s", got)
	}
	if err := runFileCommandExpect(t, fs, 0, "check", "-f", path); err != nil {
		t.Fatalf("check unexpectedly failed after custom-config regen: %v", err)
	}
}

// TestRunnerFileWorkflowAutoBulletsAndForcedOverride verifies both auto detection and explicit bullet mode.
func TestRunnerFileWorkflowAutoBulletsAndForcedOverride(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n- a\n- b\n\n## Intro\n",
	})

	runFileCommand(t, fs, "generate", "-f", path)
	got := fs.fileString(path)
	if !strings.Contains(got, "bullets=auto") || !strings.Contains(got, "- [1. Intro](#intro)") {
		t.Fatalf("auto bullets were not detected from the document:\n%s", got)
	}

	runFileCommand(t, fs, "generate", "-f", path, "--bullets", "+")
	got = fs.fileString(path)
	if !strings.Contains(got, "bullets=+") || !strings.Contains(got, "+ [1. Intro](#intro)") {
		t.Fatalf("forced bullet override was not persisted:\n%s", got)
	}
}

// TestRunnerFileWorkflowAutoBulletsIgnoreManagedTOC verifies that auto detection does not count the existing managed ToC.
func TestRunnerFileWorkflowAutoBulletsIgnoreManagedTOC(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: strings.Join([]string{
			startMarker,
			"* [1. Wrong](#wrong)",
			configStart,
			"numbering=on",
			"min-level=2",
			"max-level=4",
			"anchors=on",
			"toc=on",
			"bullets=auto",
			"state=generated",
			configEnd,
			endMarker,
			"",
			"- local bullet",
			"- local bullet",
			"",
			"## Intro",
		}, "\n") + "\n",
	})

	runFileCommand(t, fs, "generate", "-f", path)
	got := fs.fileString(path)
	if !strings.Contains(got, "bullets=auto") {
		t.Fatalf("generate did not keep auto bullet mode:\n%s", got)
	}
	if !strings.Contains(got, "- [1. Intro](#intro)") {
		t.Fatalf("auto bullet detection did not use surrounding dash lists:\n%s", got)
	}
	if strings.Contains(got, "* [1. Intro](#intro)") {
		t.Fatalf("existing managed ToC was incorrectly counted for auto detection:\n%s", got)
	}
}

// TestRunnerFileWorkflowLegacyContainerKeepsStarBullets verifies backward-compatible handling of pre-bullets containers.
func TestRunnerFileWorkflowLegacyContainerKeepsStarBullets(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: strings.Join([]string{
			startMarker,
			"* [1. Wrong](#wrong)",
			configStart,
			"numbering=on",
			"min-level=2",
			"max-level=4",
			"anchors=on",
			"toc=on",
			"state=generated",
			configEnd,
			endMarker,
			"",
			"- local bullet",
			"- local bullet",
			"",
			"## Intro",
		}, "\n") + "\n",
	})

	runFileCommand(t, fs, "generate", "-f", path)
	got := fs.fileString(path)
	if !strings.Contains(got, "bullets=*") {
		t.Fatalf("legacy container was not normalized to star bullets:\n%s", got)
	}
	if !strings.Contains(got, "* [1. Intro](#intro)") {
		t.Fatalf("legacy star bullets were not preserved:\n%s", got)
	}
	if strings.Contains(got, "- [1. Intro](#intro)") {
		t.Fatalf("legacy container unexpectedly switched to auto-detected dash bullets:\n%s", got)
	}

	runFileCommand(t, fs, "strip", "-f", path)
	runFileCommand(t, fs, "regen", "-f", path)
	got = fs.fileString(path)
	if !strings.Contains(got, "bullets=*") || !strings.Contains(got, "* [1. Intro](#intro)") {
		t.Fatalf("regen did not preserve normalized legacy star bullets:\n%s", got)
	}
	if err := runFileCommandExpect(t, fs, 0, "check", "-f", path); err != nil {
		t.Fatalf("check unexpectedly failed after legacy star regen: %v", err)
	}
}

// TestRunnerFileWorkflowStripCheckThenRegenCheck verifies both persisted target states on the same file.
func TestRunnerFileWorkflowStripCheckThenRegenCheck(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	runFileCommand(t, fs, "generate", "-f", path)
	runFileCommand(t, fs, "strip", "-f", path)
	if err := runFileCommandExpect(t, fs, 0, "check", "-f", path); err != nil {
		t.Fatalf("check unexpectedly failed for stripped state: %v", err)
	}

	runFileCommand(t, fs, "regen", "-f", path)
	if err := runFileCommandExpect(t, fs, 0, "check", "-f", path); err != nil {
		t.Fatalf("check unexpectedly failed after regen from stripped state: %v", err)
	}
}

// TestRunnerFileWorkflowRawStripRejectsRegen verifies that raw stripping removes the data regen requires.
func TestRunnerFileWorkflowRawStripRejectsRegen(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	runFileCommand(t, fs, "generate", "-f", path)
	runFileCommand(t, fs, "strip", "--raw", "-f", path)

	err := runFileCommandExpect(t, fs, 1, "regen", "-f", path)
	if err == nil || !strings.Contains(err.Error(), "valid mdtoc config block") {
		t.Fatalf("regen after raw strip returned unexpected error: %v", err)
	}
}

// TestRunnerFileWorkflowRawStripRecoversMalformedContainer verifies fallback raw stripping through the file-based CLI path.
func TestRunnerFileWorkflowRawStripRecoversMalformedContainer(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: strings.Join([]string{
			startMarker,
			"* [1. Intro](#intro)",
			configStart,
			"container-version=v2",
			"numbering=on",
			"min-level=2",
			"max-level=4",
			"anchor=github",
			"toc=on",
			"bullets=auto",
			"state=generated",
			endMarker,
			"",
			"## 1. <a id=\"intro\"></a>Intro",
		}, "\n") + "\n",
	})

	var stdout, stderr strings.Builder
	runner := newRunnerWithFS(strings.NewReader(""), &stdout, &stderr, BuildInfo{}, true, fs)
	exitCode, err := runner.Run([]string{"strip", "--raw", "--verbose", "-f", path})
	if err != nil {
		t.Fatalf("Run(strip --raw --verbose -f) error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("Run(strip --raw --verbose -f) exit code = %d, want 0", exitCode)
	}
	got := fs.fileString(path)
	if strings.Contains(got, startMarker) || strings.Contains(got, endMarker) || strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("raw strip left malformed-container artifacts behind:\n%s", got)
	}
	if !strings.Contains(got, "## Intro") {
		t.Fatalf("raw strip removed heading text:\n%s", got)
	}
	if !strings.Contains(stderr.String(), "fallback parsing") {
		t.Fatalf("strip --raw did not report fallback parsing:\n%s", stderr.String())
	}
}
