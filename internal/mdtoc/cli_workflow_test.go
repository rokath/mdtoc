package mdtoc

import (
	"fmt"
	"io"
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

// runCommandWithFS executes one CLI invocation with explicit stdin metadata so
// tests can cover root convenience mode, positional files, and mixed-input
// conflicts against the virtual filesystem.
func runCommandWithFS(t *testing.T, fs *memoryFileSystem, stdin io.Reader, stdinTTY bool, args ...string) (string, string, int, error) {
	t.Helper()

	var stdout, stderr strings.Builder
	runner := newRunnerWithFS(stdin, &stdout, &stderr, BuildInfo{}, stdinTTY, fs)
	exitCode, err := runner.Run(args)
	return stdout.String(), stderr.String(), exitCode, err
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

// TestRunnerFileWorkflowGenerateWithEmptyRedirectedStdin verifies that file
// workflows still run when CI provides non-interactive stdin that is already at EOF.
func TestRunnerFileWorkflowGenerateWithEmptyRedirectedStdin(t *testing.T) {
	const path = "README.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	stdout, stderr, exitCode, err := runCommandWithFS(t, fs, strings.NewReader(""), false, "generate", "-f", path)
	if err != nil {
		t.Fatalf("Run returned error: %v", err)
	}
	if exitCode != 0 {
		t.Fatalf("Run exit code = %d, want 0", exitCode)
	}
	if stdout != "" {
		t.Fatalf("Run wrote unexpected stdout:\n%s", stdout)
	}
	if stderr != "" {
		t.Fatalf("Run wrote unexpected stderr:\n%s", stderr)
	}

	got := fs.fileString(path)
	if !strings.Contains(got, startMarker) || !strings.Contains(got, "state=generated") {
		t.Fatalf("generate did not create managed state:\n%s", got)
	}
}

// TestRunnerRootFileWorkflowGenerateFromPositional verifies root-mode generate
// dispatch for unmanaged files passed positionally.
func TestRunnerRootFileWorkflowGenerateFromPositional(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	runFileCommand(t, fs, path)
	got := fs.fileString(path)
	if !strings.Contains(got, startMarker) || !strings.Contains(got, "state=generated") {
		t.Fatalf("root generate did not create managed state:\n%s", got)
	}
}

// TestRunnerRootFileWorkflowRegenFromPositional verifies root-mode regen
// dispatch for already managed files passed positionally.
func TestRunnerRootFileWorkflowRegenFromPositional(t *testing.T) {
	const path = "doc.md"
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	broken := strings.Replace(generated, "* [1. Intro](#intro)", "* [BROKEN](#intro)", 1)
	fs := newMemoryFileSystem(map[string]string{
		path: broken,
	})

	runFileCommand(t, fs, path)
	got := fs.fileString(path)
	if !strings.Contains(got, "* [1. Intro](#intro)") || strings.Contains(got, "* [BROKEN](#intro)") {
		t.Fatalf("root regen did not rebuild the managed state:\n%s", got)
	}
}

// TestRunnerRootFileWorkflowGenerateOverridesForceGenerate verifies that root
// mode switches to generate when explicit generate-control flags are present.
func TestRunnerRootFileWorkflowGenerateOverridesForceGenerate(t *testing.T) {
	const path = "doc.md"
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	fs := newMemoryFileSystem(map[string]string{
		path: generated,
	})

	runFileCommand(t, fs, path, "-a", "off", "-n", "off")
	got := fs.fileString(path)
	if !strings.Contains(got, "anchor=off") || !strings.Contains(got, "numbering=false") {
		t.Fatalf("root generate did not honor explicit override flags:\n%s", got)
	}
	if strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("root generate did not disable managed artifacts as requested:\n%s", got)
	}
}

// TestRunnerRootFileWorkflowSingleDashGenerateOverrides verifies that the
// tolerated one-dash long generate-control flags all force root mode to run
// generate instead of regen.
func TestRunnerRootFileWorkflowSingleDashGenerateOverrides(t *testing.T) {
	const path = "doc.md"
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	tests := []struct {
		name  string
		args  []string
		check func(t *testing.T, got string)
	}{
		{
			name: "numbering",
			args: []string{path, "-numbering", "off"},
			check: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "numbering=false") {
					t.Fatalf("root generate did not persist numbering=false:\n%s", got)
				}
				if strings.Contains(got, "## 1. ") {
					t.Fatalf("root generate left numbering enabled:\n%s", got)
				}
			},
		},
		{
			name: "anchor",
			args: []string{path, "-anchor", "off"},
			check: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "anchor=off") {
					t.Fatalf("root generate did not persist anchor=off:\n%s", got)
				}
				if strings.Contains(got, "<a id=") {
					t.Fatalf("root generate left anchors enabled:\n%s", got)
				}
			},
		},
		{
			name: "min-level",
			args: []string{path, "-min-level", "3"},
			check: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "min-level=3") {
					t.Fatalf("root generate did not persist min-level=3:\n%s", got)
				}
				if strings.Contains(got, "* [1. Intro](#intro)") {
					t.Fatalf("root generate still included the level-2 heading in the ToC:\n%s", got)
				}
			},
		},
		{
			name: "max-level",
			args: []string{path, "-max-level", "2"},
			check: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "max-level=2") {
					t.Fatalf("root generate did not persist max-level=2:\n%s", got)
				}
				if strings.Contains(got, "### 1.1. ") {
					t.Fatalf("root generate did not stop numbering at level 2:\n%s", got)
				}
			},
		},
		{
			name: "bullets",
			args: []string{path, "-bullets", "+"},
			check: func(t *testing.T, got string) {
				t.Helper()
				if !strings.Contains(got, "bullets=+") {
					t.Fatalf("root generate did not persist bullets=+:\n%s", got)
				}
				if !strings.Contains(got, "+ [1. Intro](#intro)") {
					t.Fatalf("root generate did not render plus bullets:\n%s", got)
				}
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := newMemoryFileSystem(map[string]string{
				path: generated,
			})
			runFileCommand(t, fs, tc.args...)
			tc.check(t, fs.fileString(path))
		})
	}
}

// TestRunnerRootFileWorkflowTOCOverrideOrders verifies that root mode honors
// both canonical and one-dash `toc` overrides regardless of whether the
// positional file appears before or after the flag pair.
func TestRunnerRootFileWorkflowTOCOverrideOrders(t *testing.T) {
	const path = "doc.md"
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	tests := []struct {
		name       string
		args       []string
		enableArgs []string
	}{
		{name: "double-dash-file-first", args: []string{path, "--toc", "off"}, enableArgs: []string{path, "--toc", "on"}},
		{name: "double-dash-flag-first", args: []string{"--toc", "off", path}, enableArgs: []string{"--toc", "on", path}},
		{name: "single-dash-file-first", args: []string{path, "-toc", "off"}, enableArgs: []string{path, "-toc", "on"}},
		{name: "single-dash-flag-first", args: []string{"-toc", "off", path}, enableArgs: []string{"-toc", "on", path}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := newMemoryFileSystem(map[string]string{
				path: generated,
			})

			runFileCommand(t, fs, tc.args...)
			got := fs.fileString(path)
			if !strings.Contains(got, "toc=false") {
				t.Fatalf("root generate did not persist toc=false for %s:\n%s", tc.name, got)
			}
			if strings.Contains(got, "* [1. Intro](#intro)") {
				t.Fatalf("root generate left ToC content behind for %s:\n%s", tc.name, got)
			}

			runFileCommand(t, fs, tc.enableArgs...)
			regenerated := fs.fileString(path)
			if !strings.Contains(regenerated, "toc=true") || !strings.Contains(regenerated, "* [1. Intro](#intro)") {
				t.Fatalf("root generate did not restore toc=true after override for %s:\n%s", tc.name, regenerated)
			}
		})
	}
}

// TestRunnerGenerateSubcommandAcceptsPositionalFile verifies that the explicit
// generate command accepts the file path without --file.
func TestRunnerGenerateSubcommandAcceptsPositionalFile(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	runFileCommand(t, fs, "generate", path, "--min-level=1")
	got := fs.fileString(path)
	if !strings.Contains(got, "<a id=\"title\"></a>") || !strings.Contains(got, "state=generated") {
		t.Fatalf("generate positional file did not rewrite the document:\n%s", got)
	}
}

// TestRunnerRegenSubcommandAcceptsPositionalFile verifies that the explicit
// regen command accepts the file path without --file.
func TestRunnerRegenSubcommandAcceptsPositionalFile(t *testing.T) {
	const path = "doc.md"
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	broken := strings.Replace(generated, "* [1. Intro](#intro)", "* [BROKEN](#intro)", 1)
	fs := newMemoryFileSystem(map[string]string{
		path: broken,
	})

	runFileCommand(t, fs, "regen", path)
	got := fs.fileString(path)
	if strings.Contains(got, "* [BROKEN](#intro)") || !strings.Contains(got, "* [1. Intro](#intro)") {
		t.Fatalf("regen positional file did not rebuild the managed state:\n%s", got)
	}
}

// TestRunnerStripSubcommandAcceptsPositionalFile verifies that the explicit
// strip command accepts the file path without --file.
func TestRunnerStripSubcommandAcceptsPositionalFile(t *testing.T) {
	const path = "doc.md"
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	fs := newMemoryFileSystem(map[string]string{
		path: generated,
	})

	runFileCommand(t, fs, "strip", path)
	got := fs.fileString(path)
	if !strings.Contains(got, "state=stripped") || strings.Contains(got, "<a id=") {
		t.Fatalf("strip positional file did not produce stripped state:\n%s", got)
	}
}

// TestRunnerCheckSubcommandAcceptsPositionalFile verifies that the explicit
// check command accepts the file path without --file.
func TestRunnerCheckSubcommandAcceptsPositionalFile(t *testing.T) {
	const path = "doc.md"
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	fs := newMemoryFileSystem(map[string]string{
		path: generated,
	})

	if err := runFileCommandExpect(t, fs, 0, "check", path); err != nil {
		t.Fatalf("check positional file unexpectedly failed: %v", err)
	}
}

// TestRunnerRootRejectsConflictingFileSources verifies that root mode rejects
// simultaneous positional and --file input sources.
func TestRunnerRootRejectsConflictingFileSources(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	_, _, exitCode, err := runCommandWithFS(t, fs, strings.NewReader(""), true, path, "--file", path)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if err == nil || !strings.Contains(err.Error(), "provide exactly one input source") {
		t.Fatalf("root conflict returned unexpected error: %v", err)
	}
}

// TestRunnerGenerateRejectsConflictingFileSources verifies that explicit
// subcommands reject simultaneous positional and --file input sources.
func TestRunnerGenerateRejectsConflictingFileSources(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	_, _, exitCode, err := runCommandWithFS(t, fs, strings.NewReader(""), true, "generate", path, "--file", path)
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if err == nil || !strings.Contains(err.Error(), "provide exactly one input source") {
		t.Fatalf("generate conflict returned unexpected error: %v", err)
	}
}

// TestRunnerRootRejectsMultiplePositionalFiles verifies that root mode keeps
// the single-file contract even after positional files are accepted.
func TestRunnerRootRejectsMultiplePositionalFiles(t *testing.T) {
	fs := newMemoryFileSystem(map[string]string{
		"a.md": "# A\n",
		"b.md": "# B\n",
	})

	_, _, exitCode, err := runCommandWithFS(t, fs, strings.NewReader(""), true, "a.md", "b.md")
	if exitCode != 1 {
		t.Fatalf("exit code = %d, want 1", exitCode)
	}
	if err == nil || !strings.Contains(err.Error(), "provide exactly one input source") {
		t.Fatalf("multiple positional files returned unexpected error: %v", err)
	}
}

// TestRunnerFileWorkflowRegenRestoresStoredFlags verifies regen after stripped input with non-default stored config.
func TestRunnerFileWorkflowRegenRestoresStoredFlags(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n\n### API\n",
	})

	runFileCommand(t, fs, "generate", "-f", path, "--min-level=2", "--max-level=3", "--anchor=false", "--numbering=off")
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

// TestRunnerFileWorkflowPersistsExplicitAnchorProfile verifies the profile-based anchor setting.
func TestRunnerFileWorkflowPersistsExplicitAnchorProfile(t *testing.T) {
	const path = "doc.md"
	fs := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Intro\n",
	})

	runFileCommand(t, fs, "generate", "-f", path, "--anchor", "gitlab")
	got := fs.fileString(path)
	if !strings.Contains(got, "anchor=gitlab") {
		t.Fatalf("generate did not persist the explicit anchor profile:\n%s", got)
	}
	if !strings.Contains(got, `<a id="intro"></a>`) {
		t.Fatalf("generate did not render inline anchors for the gitlab profile:\n%s", got)
	}
}

// TestRunnerFileWorkflowGitLabAnchorProfileDiffersFromGitHub verifies a heading where GitLab and GitHub anchor IDs differ.
func TestRunnerFileWorkflowGitLabAnchorProfileDiffersFromGitHub(t *testing.T) {
	const path = "doc.md"
	gitlabFS := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Version 3.5\n",
	})
	githubFS := newMemoryFileSystem(map[string]string{
		path: "# Title\n\n## Version 3.5\n",
	})

	runFileCommand(t, gitlabFS, "generate", "-f", path, "--anchor", "gitlab")
	gotGitLab := gitlabFS.fileString(path)
	if !strings.Contains(gotGitLab, "anchor=gitlab") || !strings.Contains(gotGitLab, `<a id="version-35"></a>`) {
		t.Fatalf("generate did not render the GitLab-specific anchor ID:\n%s", gotGitLab)
	}
	if !strings.Contains(gotGitLab, "* [1. Version 3.5](#version-35)") {
		t.Fatalf("generate did not render the GitLab-specific ToC target:\n%s", gotGitLab)
	}

	runFileCommand(t, githubFS, "generate", "-f", path, "--anchor", "github")
	gotGitHub := githubFS.fileString(path)
	if !strings.Contains(gotGitHub, `<a id="version-3-5"></a>`) {
		t.Fatalf("generate did not render the GitHub-specific anchor ID:\n%s", gotGitHub)
	}
	if strings.Contains(gotGitHub, `<a id="version-35"></a>`) {
		t.Fatalf("generate unexpectedly rendered the GitLab anchor ID under the GitHub profile:\n%s", gotGitHub)
	}
}

// TestRunnerFileWorkflowNormalizesAnchorFalseVariants verifies canonical off persistence for accepted inputs.
func TestRunnerFileWorkflowNormalizesAnchorFalseVariants(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "anchor-false", args: []string{"generate", "-f", "doc.md", "--anchor", "false"}},
		{name: "anchor-off", args: []string{"generate", "-f", "doc.md", "--anchor", "off"}},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fs := newMemoryFileSystem(map[string]string{
				"doc.md": "# Title\n\n## Intro\n",
			})

			runFileCommand(t, fs, tc.args...)
			got := fs.fileString("doc.md")
			if !strings.Contains(got, "anchor=off") {
				t.Fatalf("generate did not normalize %s to anchor=off:\n%s", tc.name, got)
			}
			if strings.Contains(got, "anchor=false") || strings.Contains(got, "<a id=") {
				t.Fatalf("generate left a non-canonical false anchor state for %s:\n%s", tc.name, got)
			}
			if !strings.Contains(got, "* [1. Intro](#intro)") {
				t.Fatalf("generate did not preserve ToC targets for %s:\n%s", tc.name, got)
			}
		})
	}
}

// TestRunnerFileWorkflowNormalizesBooleanContainerValues verifies true/false persistence for bool settings.
func TestRunnerFileWorkflowNormalizesBooleanContainerValues(t *testing.T) {
	fs := newMemoryFileSystem(map[string]string{
		"doc.md": "# Title\n\n## Intro\n",
	})

	runFileCommand(t, fs, "generate", "-f", "doc.md", "--numbering", "on", "--toc", "off")
	got := fs.fileString("doc.md")
	if !strings.Contains(got, "numbering=true") {
		t.Fatalf("generate did not persist numbering=true:\n%s", got)
	}
	if !strings.Contains(got, "toc=false") {
		t.Fatalf("generate did not persist toc=false:\n%s", got)
	}
	if strings.Contains(got, "numbering=on") || strings.Contains(got, "toc=off") {
		t.Fatalf("generate left non-canonical boolean values in the container:\n%s", got)
	}
}

// TestRunnerFileWorkflowRejectsRemovedLegacyAnchorsFlag verifies that --anchors is no longer accepted.
func TestRunnerFileWorkflowRejectsRemovedLegacyAnchorsFlag(t *testing.T) {
	fs := newMemoryFileSystem(map[string]string{
		"doc.md": "# Title\n\n## Intro\n",
	})

	err := runFileCommandExpect(t, fs, 1, "generate", "-f", "doc.md", "--anchors", "off")
	if err == nil {
		t.Fatalf("generate unexpectedly accepted removed --anchors flag")
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
			"numbering=true",
			"min-level=2",
			"max-level=4",
			"anchors=on",
			"toc=true",
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
			"numbering=true",
			"min-level=2",
			"max-level=4",
			"anchors=on",
			"toc=true",
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
			"numbering=true",
			"min-level=2",
			"max-level=4",
			"anchor=github",
			"toc=true",
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
