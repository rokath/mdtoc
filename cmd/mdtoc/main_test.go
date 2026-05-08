package main

import (
	"bytes"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/rokath/mdtoc/internal/mdtoc"
)

// TestRunReturnsVersionOutput verifies version output through the testable run helper.
func TestRunReturnsVersionOutput(t *testing.T) {
	var stdout, stderr bytes.Buffer
	exitCode, err := run([]string{"--version"}, strings.NewReader(""), &stdout, &stderr, mdtoc.BuildInfo{
		Version: "v1.2.3",
		Commit:  "abcdef0",
		Date:    "2026-04-18T00:00:00Z",
	})
	if err != nil {
		t.Fatalf("run error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("exitCode = %d, want 0", exitCode)
	}
	if got := stdout.String(); !strings.Contains(got, "mdtoc v1.2.3") {
		t.Fatalf("stdout missing version output:\n%s", got)
	}
}

// TestExecutableBuildDateForPath verifies the non-goreleaser date fallback used by `go install`.
func TestExecutableBuildDateForPath(t *testing.T) {
	path := filepath.Join(t.TempDir(), "mdtoc")
	if err := os.WriteFile(path, []byte("binary"), 0o755); err != nil {
		t.Fatalf("WriteFile error: %v", err)
	}
	modTime := time.Date(2026, 5, 8, 12, 34, 56, 0, time.FixedZone("CEST", 2*60*60))
	if err := os.Chtimes(path, modTime, modTime); err != nil {
		t.Fatalf("Chtimes error: %v", err)
	}

	got := executableBuildDateForPath(path)
	want := "2026-05-08T10:34:56Z"
	if got != want {
		t.Fatalf("executableBuildDateForPath() = %q, want %q", got, want)
	}
	if got := executableBuildDateForPath(filepath.Join(t.TempDir(), "missing")); got != "" {
		t.Fatalf("missing executable build date = %q, want empty", got)
	}
}

// TestMainWritesErrorAndExits verifies that main reports runner errors and exits with the returned code.
func TestMainWritesErrorAndExits(t *testing.T) {
	oldArgs := os.Args
	oldStderr := os.Stderr
	oldExit := exitFunc
	t.Cleanup(func() {
		os.Args = oldArgs
		os.Stderr = oldStderr
		exitFunc = oldExit
	})

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe error: %v", err)
	}
	defer r.Close()

	os.Stderr = w
	os.Args = []string{"mdtoc", "generate"}

	var gotExit int
	exitFunc = func(code int) {
		gotExit = code
		panic(errors.New("exit"))
	}

	func() {
		defer func() {
			if r := recover(); r == nil {
				t.Fatalf("main did not exit")
			}
		}()
		main()
	}()

	if err := w.Close(); err != nil {
		t.Fatalf("stderr pipe close error: %v", err)
	}
	buf := new(bytes.Buffer)
	if _, err := buf.ReadFrom(r); err != nil {
		t.Fatalf("stderr read error: %v", err)
	}
	if gotExit != 1 {
		t.Fatalf("exit code = %d, want 1", gotExit)
	}
	if got := buf.String(); !strings.Contains(got, "no input provided") {
		t.Fatalf("stderr missing expected error:\n%s", got)
	}
}
