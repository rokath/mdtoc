package mdtoc

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// fileSystem abstracts CLI file access for deterministic workflow tests.
type fileSystem interface {
	ReadFile(name string) ([]byte, error)
	WriteFile(name string, data []byte, perm os.FileMode) error
}

// osFileSystem implements fileSystem on top of the local OS filesystem.
type osFileSystem struct{}

// ReadFile loads a file from the local filesystem.
func (osFileSystem) ReadFile(name string) ([]byte, error) {
	return os.ReadFile(name)
}

// WriteFile persists a file on the local filesystem.
func (osFileSystem) WriteFile(name string, data []byte, perm os.FileMode) error {
	return os.WriteFile(name, data, perm)
}

// BuildInfo contains release metadata injected by the build system.
type BuildInfo struct {
	Version string
	Commit  string
	Date    string
}

func (b BuildInfo) normalized() BuildInfo {
	if b.Version == "" {
		b.Version = "dev"
	}
	if b.Commit == "" {
		b.Commit = "unknown"
	}
	if b.Date == "" {
		b.Date = "unknown"
	}
	return b
}

// Runner owns the CLI I/O streams. This makes command behavior easy to test.
type Runner struct {
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	buildInfo BuildInfo
	stdinTTY  bool
	fs        fileSystem
}

// NewRunner creates a testable CLI runner.
func NewRunner(stdin io.Reader, stdout, stderr io.Writer) *Runner {
	return NewRunnerWithBuildInfo(stdin, stdout, stderr, BuildInfo{})
}

// NewRunnerWithBuildInfo creates a testable CLI runner with injected build metadata.
func NewRunnerWithBuildInfo(stdin io.Reader, stdout, stderr io.Writer, buildInfo BuildInfo) *Runner {
	return newRunner(stdin, stdout, stderr, buildInfo, isInteractiveInput(stdin))
}

// newRunner builds a runner with an explicitly injected stdin interactivity flag.
func newRunner(stdin io.Reader, stdout, stderr io.Writer, buildInfo BuildInfo, stdinTTY bool) *Runner {
	return newRunnerWithFS(stdin, stdout, stderr, buildInfo, stdinTTY, osFileSystem{})
}

// newRunnerWithFS builds a runner with explicitly injected stdin metadata and filesystem access.
func newRunnerWithFS(stdin io.Reader, stdout, stderr io.Writer, buildInfo BuildInfo, stdinTTY bool, fs fileSystem) *Runner {
	return &Runner{
		stdin:     stdin,
		stdout:    stdout,
		stderr:    stderr,
		buildInfo: buildInfo.normalized(),
		stdinTTY:  stdinTTY,
		fs:        fs,
	}
}

// Run executes the CLI and returns the intended process exit code.
func (r *Runner) Run(args []string) (int, error) {
	if len(args) == 0 {
		fmt.Fprint(r.stdout, shortHelp())
		return 0, nil
	}
	if !isSubcommand(args[0]) {
		return r.runRoot(args)
	}
	switch args[0] {
	case "generate":
		return r.runGenerate(args[1:])
	case "regen":
		return r.runRegen(args[1:])
	case "strip":
		return r.runStrip(args[1:])
	case "check":
		return r.runCheck(args[1:])
	default:
		return 1, fmt.Errorf("unknown command %q", args[0])
	}
}

// runRoot handles root-level flags such as --help and --version.
func (r *Runner) runRoot(args []string) (int, error) {
	fs := flag.NewFlagSet("mdtoc", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	help := fs.Bool("help", false, "")
	helpShort := fs.Bool("h", false, "")
	showVersion := fs.Bool("version", false, "")
	verbose := fs.Bool("verbose", false, "")
	verboseShort := fs.Bool("v", false, "")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if *verboseShort {
		*verbose = true
	}
	if fs.NArg() > 0 {
		return 1, fmt.Errorf("unknown command %q", fs.Arg(0))
	}
	if *help || *helpShort {
		if *verbose {
			fmt.Fprint(r.stdout, longHelp())
		} else {
			fmt.Fprint(r.stdout, shortHelp())
		}
		return 0, nil
	}
	if *showVersion {
		if *verbose {
			fmt.Fprintf(r.stdout, "mdtoc %s\ncommit: %s\ndate: %s\nGo-based Markdown ToC manager\n", r.buildInfo.Version, r.buildInfo.Commit, r.buildInfo.Date)
		} else {
			fmt.Fprintf(r.stdout, "mdtoc %s\ncommit: %s\ndate: %s\n", r.buildInfo.Version, r.buildInfo.Commit, r.buildInfo.Date)
		}
		return 0, nil
	}
	if *verbose {
		fmt.Fprint(r.stdout, longHelp())
		return 0, nil
	}
	fmt.Fprint(r.stdout, shortHelp())
	return 0, nil
}

// runGenerate parses and executes the generate subcommand.
func (r *Runner) runGenerate(args []string) (int, error) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	numbering := fs.String("numbering", "on", "")
	numberingS := fs.String("n", "", "")
	minLevel := fs.Int("min-level", 2, "")
	maxLevel := fs.Int("max-level", 4, "")
	anchor := fs.String("anchor", string(AnchorGitHub), "")
	anchorS := fs.String("a", "", "")
	toc := fs.String("toc", "on", "")
	bullets := fs.String("bullets", "auto", "")
	bulletsS := fs.String("b", "", "")
	file := fs.String("file", "", "")
	fileS := fs.String("f", "", "")
	verbose := fs.Bool("verbose", false, "")
	verboseS := fs.Bool("v", false, "")
	help := fs.Bool("help", false, "")
	helpS := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if *verboseS {
		*verbose = true
	}
	if *help || *helpS {
		fmt.Fprint(r.stdout, generateHelp(*verbose))
		return 0, nil
	}
	if *verbose && fs.NFlag() == 1 {
		fmt.Fprint(r.stdout, generateHelp(true))
		return 0, nil
	}
	if *numberingS != "" {
		*numbering = *numberingS
	}
	if *anchorS != "" {
		*anchor = *anchorS
	}
	if *bulletsS != "" {
		*bullets = *bulletsS
	}
	if *fileS != "" {
		*file = *fileS
	}
	if err := r.requireInputSource(*file); err != nil {
		return 1, err
	}
	numberingB, err := parseBoolValue(*numbering)
	if err != nil {
		return 1, err
	}
	anchorMode, err := parseAnchorMode(*anchor)
	if err != nil {
		return 1, err
	}
	tocB, err := parseBoolValue(*toc)
	if err != nil {
		return 1, err
	}
	bulletMode, err := parseBulletMode(*bullets)
	if err != nil {
		return 1, err
	}
	input, err := r.readInput(*file)
	if err != nil {
		return 1, err
	}
	result, warnings, err := Generate(input, Options{Numbering: numberingB, MinLevel: *minLevel, MaxLevel: *maxLevel, Anchor: anchorMode, TOC: tocB, Bullets: bulletMode})
	if err != nil {
		return 1, err
	}
	if err := r.writeOutput(*file, result); err != nil {
		return 1, err
	}
	r.writeDiagnostics(*verbose, warnings)
	return 0, nil
}

// runRegen parses and executes the regen subcommand.
func (r *Runner) runRegen(args []string) (int, error) {
	fs := flag.NewFlagSet("regen", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	file := fs.String("file", "", "")
	fileS := fs.String("f", "", "")
	verbose := fs.Bool("verbose", false, "")
	verboseS := fs.Bool("v", false, "")
	help := fs.Bool("help", false, "")
	helpS := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if *verboseS {
		*verbose = true
	}
	if *help || *helpS {
		fmt.Fprint(r.stdout, regenHelp(*verbose))
		return 0, nil
	}
	if *verbose && fs.NFlag() == 1 {
		fmt.Fprint(r.stdout, regenHelp(true))
		return 0, nil
	}
	if *fileS != "" {
		*file = *fileS
	}
	if err := r.requireInputSource(*file); err != nil {
		return 1, err
	}
	input, err := r.readInput(*file)
	if err != nil {
		return 1, err
	}
	result, warnings, err := Regen(input)
	if err != nil {
		return 1, err
	}
	if err := r.writeOutput(*file, result); err != nil {
		return 1, err
	}
	r.writeDiagnostics(*verbose, warnings)
	return 0, nil
}

// runStrip parses and executes the strip subcommand.
func (r *Runner) runStrip(args []string) (int, error) {
	fs := flag.NewFlagSet("strip", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	raw := fs.Bool("raw", false, "")
	file := fs.String("file", "", "")
	fileS := fs.String("f", "", "")
	verbose := fs.Bool("verbose", false, "")
	verboseS := fs.Bool("v", false, "")
	help := fs.Bool("help", false, "")
	helpS := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if *verboseS {
		*verbose = true
	}
	if *help || *helpS {
		fmt.Fprint(r.stdout, stripHelp(*verbose))
		return 0, nil
	}
	if *verbose && fs.NFlag() == 1 {
		fmt.Fprint(r.stdout, stripHelp(true))
		return 0, nil
	}
	if *fileS != "" {
		*file = *fileS
	}
	if err := r.requireInputSource(*file); err != nil {
		return 1, err
	}
	input, err := r.readInput(*file)
	if err != nil {
		return 1, err
	}
	var result string
	var warnings []string
	if *raw {
		result, warnings, err = StripRaw(input)
	} else {
		result, warnings, err = Strip(input)
	}
	if err != nil {
		return 1, err
	}
	if err := r.writeOutput(*file, result); err != nil {
		return 1, err
	}
	r.writeDiagnostics(*verbose, warnings)
	return 0, nil
}

// runCheck parses and executes the check subcommand.
func (r *Runner) runCheck(args []string) (int, error) {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	file := fs.String("file", "", "")
	fileS := fs.String("f", "", "")
	verbose := fs.Bool("verbose", false, "")
	verboseS := fs.Bool("v", false, "")
	help := fs.Bool("help", false, "")
	helpS := fs.Bool("h", false, "")
	if err := fs.Parse(args); err != nil {
		return 1, err
	}
	if *verboseS {
		*verbose = true
	}
	if *help || *helpS {
		fmt.Fprint(r.stdout, checkHelp(*verbose))
		return 0, nil
	}
	if *verbose && fs.NFlag() == 1 {
		fmt.Fprint(r.stdout, checkHelp(true))
		return 0, nil
	}
	if *fileS != "" {
		*file = *fileS
	}
	if err := r.requireInputSource(*file); err != nil {
		return 1, err
	}
	input, err := r.readInput(*file)
	if err != nil {
		return 1, err
	}
	ok, warnings, err := Check(input)
	if err != nil {
		return 1, err
	}
	r.writeDiagnostics(*verbose, warnings)
	if ok {
		return 0, nil
	}
	return 2, errors.New("document does not match the reconstructed target state")
}

// readInput loads document content from a file or from stdin.
func (r *Runner) readInput(file string) (string, error) {
	if file != "" {
		b, err := r.fs.ReadFile(file)
		return string(b), err
	}
	b, err := io.ReadAll(r.stdin)
	return string(b), err
}

// requireInputSource rejects interactive invocations that forgot both -f and piped stdin.
func (r *Runner) requireInputSource(file string) error {
	if file != "" || !r.stdinTTY {
		return nil
	}
	return errors.New("no input provided; use --file <name> or pipe Markdown via stdin")
}

// writeOutput writes transformed content either back to a file or to stdout.
func (r *Runner) writeOutput(file, content string) error {
	if file != "" {
		return r.fs.WriteFile(file, []byte(content), 0o644)
	}
	_, err := io.WriteString(r.stdout, content)
	return err
}

// writeDiagnostics emits collected warnings only in verbose mode.
func (r *Runner) writeDiagnostics(verbose bool, warnings []string) {
	if !verbose {
		return
	}
	for _, w := range warnings {
		fmt.Fprintln(r.stderr, w)
	}
}

// hasFlag reports whether args contains the exact flag token.
func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

// isSubcommand reports whether the first CLI token names a supported subcommand.
func isSubcommand(arg string) bool {
	switch arg {
	case "generate", "regen", "strip", "check":
		return true
	default:
		return false
	}
}

// isInteractiveInput reports whether the reader is backed by a terminal device.
func isInteractiveInput(r io.Reader) bool {
	f, ok := r.(*os.File)
	if !ok {
		return false
	}
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

// shortHelp returns the compact root help text.
func shortHelp() string {
	return strings.TrimSpace(`Usage: mdtoc <command>

Commands:
  generate [--file <name>] [--verbose] [OPTIONS]  generate or update ToC, numbers, and anchors
  check    [--file <name>] [--verbose]            validate that the document matches its persisted state
  regen    [--file <name>] [--verbose]            regenerate using persisted container config
  strip    [--file <name>] [--verbose] [--raw]    remove managed artifacts and keep the container

Details: mdtoc -v   short for mdtoc --help --verbose
`) + "\n"
}

// longHelp returns the root help text with the command summary section.
func longHelp() string {
	return strings.TrimSpace(`mdtoc - deterministic Markdown ToC manager

Usage: mdtoc <command>

Commands:
  generate [--file <name>] [--verbose] [OPTIONS]  generate or update ToC, numbers, and anchors
  check    [--file <name>] [--verbose]            validate that the document matches its persisted state
  regen    [--file <name>] [--verbose]            regenerate using persisted container config
  strip    [--file <name>] [--verbose] [--raw]    remove managed artifacts and keep the container

Generate options:
  --numbering=on  heading numbers on or off
  --min-level=2   minimum heading level (valid 1 to --max-level)
  --max-level=4   maximum heading level (valid --min-level to 6)
  --anchor=github anchor profile: github, gitlab, or off
  --toc=on        table of contents on or off
  --bullets=auto  ToC bullets auto, *, -, or +

Help:
  mdtoc [--help] [--verbose]         general help
  mdtoc <command> --help [--verbose] specific help
  mdtoc --version [--verbose]        version info

Info: https://github.com/rokath/mdtoc/
`) + "\n"
}

// generateHelp returns the generate subcommand help text.
func generateHelp(verbose bool) string {
	if verbose {
		return strings.TrimSpace(`mdtoc generate

Generate or update ToC, heading numbers, and anchors.

Options:
  --numbering, -n <on|off>
  --min-level <N>
  --max-level <N>
  --anchor, -a <github|gitlab|off>
  --toc <on|off>
  --bullets, -b <auto|*|-|+>
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
	}
	return strings.TrimSpace(`mdtoc generate

Options:
  --numbering, -n <on|off>
  --min-level <N>
  --max-level <N>
  --anchor, -a <github|gitlab|off>
  --toc <on|off>
  --bullets, -b <auto|*|-|+>
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
}

// regenHelp returns the regen subcommand help text.
func regenHelp(verbose bool) string {
	if verbose {
		return strings.TrimSpace(`mdtoc regen

Regenerate using the persisted config from an existing managed container.

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
	}
	return strings.TrimSpace(`mdtoc regen

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
}

// stripHelp returns the strip subcommand help text.
func stripHelp(verbose bool) string {
	if verbose {
		return strings.TrimSpace(`mdtoc strip

Remove managed artifacts and optionally the entire managed container.

Options:
  --raw
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
	}
	return strings.TrimSpace(`mdtoc strip

Options:
  --raw
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
}

// checkHelp returns the check subcommand help text.
func checkHelp(verbose bool) string {
	if verbose {
		return strings.TrimSpace(`mdtoc check

Reconstruct the target document state and compare it byte-for-byte.

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
	}
	return strings.TrimSpace(`mdtoc check

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
}
