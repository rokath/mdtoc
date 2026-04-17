package mdtoc

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

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
}

// NewRunner creates a testable CLI runner.
func NewRunner(stdin io.Reader, stdout, stderr io.Writer) *Runner {
	return NewRunnerWithBuildInfo(stdin, stdout, stderr, BuildInfo{})
}

// NewRunnerWithBuildInfo creates a testable CLI runner with injected build metadata.
func NewRunnerWithBuildInfo(stdin io.Reader, stdout, stderr io.Writer, buildInfo BuildInfo) *Runner {
	return &Runner{stdin: stdin, stdout: stdout, stderr: stderr, buildInfo: buildInfo.normalized()}
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
	case "strip":
		return r.runStrip(args[1:])
	case "check":
		return r.runCheck(args[1:])
	default:
		return 1, fmt.Errorf("unknown command %q", args[0])
	}
}

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
	fmt.Fprint(r.stdout, shortHelp())
	return 0, nil
}

func (r *Runner) runGenerate(args []string) (int, error) {
	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	numbering := fs.String("numbering", "on", "")
	numberingS := fs.String("n", "", "")
	minLevel := fs.Int("min-level", 2, "")
	maxLevel := fs.Int("max-level", 4, "")
	anchors := fs.String("anchors", "on", "")
	anchorsS := fs.String("a", "", "")
	toc := fs.String("toc", "on", "")
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
	if *numberingS != "" {
		*numbering = *numberingS
	}
	if *anchorsS != "" {
		*anchors = *anchorsS
	}
	if *fileS != "" {
		*file = *fileS
	}
	numberingB, err := parseOnOff(*numbering)
	if err != nil {
		return 1, err
	}
	anchorsB, err := parseOnOff(*anchors)
	if err != nil {
		return 1, err
	}
	tocB, err := parseOnOff(*toc)
	if err != nil {
		return 1, err
	}
	input, err := r.readInput(*file)
	if err != nil {
		return 1, err
	}
	result, warnings, err := Generate(input, Options{Numbering: numberingB, MinLevel: *minLevel, MaxLevel: *maxLevel, Anchors: anchorsB, TOC: tocB})
	if err != nil {
		return 1, err
	}
	if err := r.writeOutput(*file, result); err != nil {
		return 1, err
	}
	r.writeDiagnostics(*verbose, warnings)
	return 0, nil
}

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
	if *fileS != "" {
		*file = *fileS
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
	if *fileS != "" {
		*file = *fileS
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

func (r *Runner) readInput(file string) (string, error) {
	if file != "" {
		b, err := os.ReadFile(file)
		return string(b), err
	}
	b, err := io.ReadAll(r.stdin)
	return string(b), err
}

func (r *Runner) writeOutput(file, content string) error {
	if file != "" {
		return os.WriteFile(file, []byte(content), 0o644)
	}
	_, err := io.WriteString(r.stdout, content)
	return err
}

func (r *Runner) writeDiagnostics(verbose bool, warnings []string) {
	if !verbose {
		return
	}
	for _, w := range warnings {
		fmt.Fprintln(r.stderr, w)
	}
}

func hasFlag(args []string, flag string) bool {
	for _, arg := range args {
		if arg == flag {
			return true
		}
	}
	return false
}

func isSubcommand(arg string) bool {
	switch arg {
	case "generate", "strip", "check":
		return true
	default:
		return false
	}
}

func shortHelp() string {
	return strings.TrimSpace(`mdtoc - deterministic Markdown ToC manager

Usage:
  mdtoc --help [--verbose]
  mdtoc --version [--verbose]
  mdtoc generate [OPTIONS]
  mdtoc strip [--raw] [--file <name>] [--verbose]
  mdtoc check [--file <name>] [--verbose]
`) + "\n"
}

func longHelp() string {
	return shortHelp() + "\nCommands:\n  generate   generate or update ToC, numbers, and anchors\n  strip      remove managed artifacts and keep the container\n  check      validate that the document matches its persisted state\n"
}

func generateHelp(verbose bool) string {
	if verbose {
		return strings.TrimSpace(`mdtoc generate

Generate or update ToC, heading numbers, and anchors.

Options:
  --numbering, -n <on|off>
  --min-level <N>
  --max-level <N>
  --anchors, -a <on|off>
  --toc <on|off>
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
  --anchors, -a <on|off>
  --toc <on|off>
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
}

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
