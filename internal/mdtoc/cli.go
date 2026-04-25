package mdtoc

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
)

// fileSystem abstracts CLI file access so workflow tests can validate the
// command behavior without touching the host filesystem.
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

// normalized returns a fully populated build-info struct so the CLI never has
// to special-case unset goreleaser metadata.
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

// generateInvocation captures one normalized `generate` command request after
// CLI parsing but before document execution.
type generateInvocation struct {
	input       inputSource
	options     Options
	verbose     bool
	help        bool
	verboseOnly bool
}

// simpleInvocation captures one normalized `regen` or `check` request.
type simpleInvocation struct {
	input       inputSource
	verbose     bool
	help        bool
	verboseOnly bool
}

// stripInvocation captures one normalized `strip` request, including the
// command-specific `--raw` switch.
type stripInvocation struct {
	input       inputSource
	raw         bool
	verbose     bool
	help        bool
	verboseOnly bool
}

// rootInvocation captures the richer root-command mode, where help/version and
// smart `generate`/`regen` dispatch coexist.
type rootInvocation struct {
	input             inputSource
	options           Options
	verbose           bool
	help              bool
	showVersion       bool
	verboseOnly       bool
	generateOverrides bool
}

// Runner owns the CLI streams and the input metadata needed to distinguish
// interactive terminal use from piped stdin use.
type Runner struct {
	stdin     io.Reader
	stdout    io.Writer
	stderr    io.Writer
	buildInfo BuildInfo
	stdinTTY  bool
	fs        fileSystem
}

// NewRunner creates a testable CLI runner with default build metadata.
func NewRunner(stdin io.Reader, stdout, stderr io.Writer) *Runner {
	return NewRunnerWithBuildInfo(stdin, stdout, stderr, BuildInfo{})
}

// NewRunnerWithBuildInfo creates a testable CLI runner with injected build metadata.
func NewRunnerWithBuildInfo(stdin io.Reader, stdout, stderr io.Writer, buildInfo BuildInfo) *Runner {
	return newRunner(stdin, stdout, stderr, buildInfo, isInteractiveInput(stdin))
}

// newRunner builds a runner with explicitly injected stdin interactivity. Tests
// use this to model terminal invocations without touching real file
// descriptors.
func newRunner(stdin io.Reader, stdout, stderr io.Writer, buildInfo BuildInfo, stdinTTY bool) *Runner {
	return newRunnerWithFS(stdin, stdout, stderr, buildInfo, stdinTTY, osFileSystem{})
}

// newRunnerWithFS injects both stdin metadata and a mockable filesystem.
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

// Run executes the CLI and returns the process exit code that main should use.
func (r *Runner) Run(args []string) (int, error) {
	if len(args) == 0 {
		if r.stdinTTY {
			fmt.Fprint(r.stdout, shortHelp())
			return 0, nil
		}
		return r.runRoot(args)
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

// runRoot handles help/version requests and the root convenience mode that
// auto-dispatches to `generate` or `regen`.
func (r *Runner) runRoot(args []string) (int, error) {
	inv, err := parseRootInvocation(args)
	if err != nil {
		return 1, err
	}
	if inv.help {
		if inv.verbose {
			fmt.Fprint(r.stdout, longHelp())
		} else {
			fmt.Fprint(r.stdout, shortHelp())
		}
		return 0, nil
	}
	if inv.showVersion {
		r.writeVersion(inv.verbose)
		return 0, nil
	}
	source, err := r.resolveRequestedInput(inv.input)
	if err != nil {
		return 1, err
	}
	if source.kind == inputSourceNone {
		if inv.verboseOnly {
			fmt.Fprint(r.stdout, longHelp())
			return 0, nil
		}
		fmt.Fprint(r.stdout, shortHelp())
		return 0, nil
	}

	input, err := r.readInput(source)
	if err != nil {
		return 1, err
	}
	if shouldUseGenerateInRootMode(input, inv.generateOverrides) {
		return r.executeGenerate(source, input, inv.options, inv.verbose)
	}
	return r.executeRegen(source, input, inv.verbose)
}

// runGenerate parses and executes the explicit `generate` subcommand.
func (r *Runner) runGenerate(args []string) (int, error) {
	inv, err := parseGenerateInvocation(args)
	if err != nil {
		return 1, err
	}
	if inv.help {
		fmt.Fprint(r.stdout, generateHelp(inv.verbose))
		return 0, nil
	}
	source, err := r.resolveRequestedInput(inv.input)
	if err != nil {
		return 1, err
	}
	if source.kind == inputSourceNone {
		if inv.verboseOnly {
			fmt.Fprint(r.stdout, generateHelp(true))
			return 0, nil
		}
		return 1, r.missingInputError()
	}

	input, err := r.readInput(source)
	if err != nil {
		return 1, err
	}
	return r.executeGenerate(source, input, inv.options, inv.verbose)
}

// runRegen parses and executes the explicit `regen` subcommand.
func (r *Runner) runRegen(args []string) (int, error) {
	inv, err := parseRegenInvocation(args)
	if err != nil {
		return 1, err
	}
	if inv.help {
		fmt.Fprint(r.stdout, regenHelp(inv.verbose))
		return 0, nil
	}
	source, err := r.resolveRequestedInput(inv.input)
	if err != nil {
		return 1, err
	}
	if source.kind == inputSourceNone {
		if inv.verboseOnly {
			fmt.Fprint(r.stdout, regenHelp(true))
			return 0, nil
		}
		return 1, r.missingInputError()
	}

	input, err := r.readInput(source)
	if err != nil {
		return 1, err
	}
	return r.executeRegen(source, input, inv.verbose)
}

// runStrip parses and executes the explicit `strip` subcommand.
func (r *Runner) runStrip(args []string) (int, error) {
	inv, err := parseStripInvocation(args)
	if err != nil {
		return 1, err
	}
	if inv.help {
		fmt.Fprint(r.stdout, stripHelp(inv.verbose))
		return 0, nil
	}
	source, err := r.resolveRequestedInput(inv.input)
	if err != nil {
		return 1, err
	}
	if source.kind == inputSourceNone {
		if inv.verboseOnly {
			fmt.Fprint(r.stdout, stripHelp(true))
			return 0, nil
		}
		return 1, r.missingInputError()
	}

	input, err := r.readInput(source)
	if err != nil {
		return 1, err
	}
	return r.executeStrip(source, input, inv.raw, inv.verbose)
}

// runCheck parses and executes the explicit `check` subcommand.
func (r *Runner) runCheck(args []string) (int, error) {
	inv, err := parseCheckInvocation(args)
	if err != nil {
		return 1, err
	}
	if inv.help {
		fmt.Fprint(r.stdout, checkHelp(inv.verbose))
		return 0, nil
	}
	source, err := r.resolveRequestedInput(inv.input)
	if err != nil {
		return 1, err
	}
	if source.kind == inputSourceNone {
		if inv.verboseOnly {
			fmt.Fprint(r.stdout, checkHelp(true))
			return 0, nil
		}
		return 1, r.missingInputError()
	}

	input, err := r.readInput(source)
	if err != nil {
		return 1, err
	}
	return r.executeCheck(input, inv.verbose)
}

// parseRootInvocation parses the top-level invocation. It intentionally accepts
// both file selection and generate control flags because root mode may need to
// dispatch to either `generate` or `regen`.
func parseRootInvocation(args []string) (rootInvocation, error) {
	normalized, err := normalizeCLIArgs(args, rootCommandArgSpec())
	if err != nil {
		return rootInvocation{}, err
	}

	fs := flag.NewFlagSet("mdtoc", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	numbering := fs.String("numbering", "on", "")
	numberingShort := fs.String("n", "", "")
	minLevel := fs.Int("min-level", 2, "")
	maxLevel := fs.Int("max-level", 4, "")
	anchor := fs.String("anchor", string(AnchorGitHub), "")
	anchorShort := fs.String("a", "", "")
	toc := fs.String("toc", "on", "")
	bullets := fs.String("bullets", "auto", "")
	bulletsShort := fs.String("b", "", "")
	help := fs.Bool("help", false, "")
	helpShort := fs.Bool("h", false, "")
	showVersion := fs.Bool("version", false, "")
	verbose := fs.Bool("verbose", false, "")
	verboseShort := fs.Bool("v", false, "")
	if err := fs.Parse(normalized.parseArgs); err != nil {
		return rootInvocation{}, err
	}

	if *verboseShort {
		*verbose = true
	}
	if *helpShort {
		*help = true
	}
	if *numberingShort != "" {
		*numbering = *numberingShort
	}
	if *anchorShort != "" {
		*anchor = *anchorShort
	}
	if *bulletsShort != "" {
		*bullets = *bulletsShort
	}

	options, err := buildGenerateOptions(*numbering, *minLevel, *maxLevel, *anchor, *toc, *bullets)
	if err != nil {
		return rootInvocation{}, err
	}
	return rootInvocation{
		input:             normalized.input,
		options:           options,
		verbose:           *verbose,
		help:              *help,
		showVersion:       *showVersion,
		verboseOnly:       isVerboseOnlyInvocation(normalized.parseArgs),
		generateOverrides: hasExplicitGenerateControlFlag(normalized.parseArgs),
	}, nil
}

// parseGenerateInvocation parses the explicit `generate` subcommand after
// positional files and `--file` were normalized into a single input source.
func parseGenerateInvocation(args []string) (generateInvocation, error) {
	normalized, err := normalizeCLIArgs(args, generateCommandArgSpec())
	if err != nil {
		return generateInvocation{}, err
	}

	fs := flag.NewFlagSet("generate", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	numbering := fs.String("numbering", "on", "")
	numberingShort := fs.String("n", "", "")
	minLevel := fs.Int("min-level", 2, "")
	maxLevel := fs.Int("max-level", 4, "")
	anchor := fs.String("anchor", string(AnchorGitHub), "")
	anchorShort := fs.String("a", "", "")
	toc := fs.String("toc", "on", "")
	bullets := fs.String("bullets", "auto", "")
	bulletsShort := fs.String("b", "", "")
	verbose := fs.Bool("verbose", false, "")
	verboseShort := fs.Bool("v", false, "")
	help := fs.Bool("help", false, "")
	helpShort := fs.Bool("h", false, "")
	if err := fs.Parse(normalized.parseArgs); err != nil {
		return generateInvocation{}, err
	}

	if *verboseShort {
		*verbose = true
	}
	if *helpShort {
		*help = true
	}
	if *numberingShort != "" {
		*numbering = *numberingShort
	}
	if *anchorShort != "" {
		*anchor = *anchorShort
	}
	if *bulletsShort != "" {
		*bullets = *bulletsShort
	}

	options, err := buildGenerateOptions(*numbering, *minLevel, *maxLevel, *anchor, *toc, *bullets)
	if err != nil {
		return generateInvocation{}, err
	}
	return generateInvocation{
		input:       normalized.input,
		options:     options,
		verbose:     *verbose,
		help:        *help,
		verboseOnly: isVerboseOnlyInvocation(normalized.parseArgs),
	}, nil
}

// parseRegenInvocation parses the explicit `regen` subcommand.
func parseRegenInvocation(args []string) (simpleInvocation, error) {
	return parseSimpleInvocation("regen", args, regenCommandArgSpec())
}

// parseCheckInvocation parses the explicit `check` subcommand.
func parseCheckInvocation(args []string) (simpleInvocation, error) {
	return parseSimpleInvocation("check", args, checkCommandArgSpec())
}

// parseSimpleInvocation parses a subcommand that only accepts help, verbose,
// and an input source.
func parseSimpleInvocation(name string, args []string, spec argumentSpec) (simpleInvocation, error) {
	normalized, err := normalizeCLIArgs(args, spec)
	if err != nil {
		return simpleInvocation{}, err
	}

	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	verbose := fs.Bool("verbose", false, "")
	verboseShort := fs.Bool("v", false, "")
	help := fs.Bool("help", false, "")
	helpShort := fs.Bool("h", false, "")
	if err := fs.Parse(normalized.parseArgs); err != nil {
		return simpleInvocation{}, err
	}

	if *verboseShort {
		*verbose = true
	}
	if *helpShort {
		*help = true
	}
	return simpleInvocation{
		input:       normalized.input,
		verbose:     *verbose,
		help:        *help,
		verboseOnly: isVerboseOnlyInvocation(normalized.parseArgs),
	}, nil
}

// parseStripInvocation parses the explicit `strip` subcommand.
func parseStripInvocation(args []string) (stripInvocation, error) {
	normalized, err := normalizeCLIArgs(args, stripCommandArgSpec())
	if err != nil {
		return stripInvocation{}, err
	}

	fs := flag.NewFlagSet("strip", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	raw := fs.Bool("raw", false, "")
	verbose := fs.Bool("verbose", false, "")
	verboseShort := fs.Bool("v", false, "")
	help := fs.Bool("help", false, "")
	helpShort := fs.Bool("h", false, "")
	if err := fs.Parse(normalized.parseArgs); err != nil {
		return stripInvocation{}, err
	}

	if *verboseShort {
		*verbose = true
	}
	if *helpShort {
		*help = true
	}
	return stripInvocation{
		input:       normalized.input,
		raw:         *raw,
		verbose:     *verbose,
		help:        *help,
		verboseOnly: isVerboseOnlyInvocation(normalized.parseArgs),
	}, nil
}

// buildGenerateOptions converts the CLI flag strings into the validated
// execution options consumed by Generate.
func buildGenerateOptions(numbering string, minLevel, maxLevel int, anchor string, toc string, bullets string) (Options, error) {
	numberingValue, err := parseBoolValue(numbering)
	if err != nil {
		return Options{}, err
	}
	anchorValue, err := parseAnchorMode(anchor)
	if err != nil {
		return Options{}, err
	}
	tocValue, err := parseBoolValue(toc)
	if err != nil {
		return Options{}, err
	}
	bulletValue, err := parseBulletMode(bullets)
	if err != nil {
		return Options{}, err
	}
	return Options{
		Numbering: numberingValue,
		MinLevel:  minLevel,
		MaxLevel:  maxLevel,
		Anchor:    anchorValue,
		TOC:       tocValue,
		Bullets:   bulletValue,
	}, nil
}

// resolveRequestedInput finalizes the input source after argv parsing by
// combining the requested file/none state with knowledge about whether stdin is
// interactive.
func (r *Runner) resolveRequestedInput(requested inputSource) (inputSource, error) {
	if requested.kind == inputSourceFile {
		if !r.stdinTTY {
			if requested.viaFlag {
				return inputSource{}, errors.New("cannot use --file together with piped stdin")
			}
			return inputSource{}, errors.New("provide exactly one input source: positional file, --file, or stdin")
		}
		return requested, nil
	}
	if !r.stdinTTY {
		return inputSource{kind: inputSourceStdin}, nil
	}
	return inputSource{kind: inputSourceNone}, nil
}

// missingInputError reports the user-facing guidance when a command that needs
// document input was invoked interactively without a file and without piped stdin.
func (r *Runner) missingInputError() error {
	return errors.New("no input provided; use --file <name>, a positional file, or pipe Markdown via stdin")
}

// shouldUseGenerateInRootMode applies the issue #53 dispatch rule. Explicit
// generate-shaping flags always force `generate`; otherwise only a valid managed
// container selects `regen`.
func shouldUseGenerateInRootMode(input string, generateOverrides bool) bool {
	if generateOverrides {
		return true
	}
	parsed, err := ParseDocument(input)
	return err != nil || parsed.Container == nil
}

// executeGenerate runs the already-decided generate workflow against document
// text that was read by the caller from the resolved input source.
func (r *Runner) executeGenerate(source inputSource, input string, opts Options, verbose bool) (int, error) {
	result, warnings, err := Generate(input, opts)
	if err != nil {
		return 1, err
	}
	if err := r.writeOutput(source, result); err != nil {
		return 1, err
	}
	r.writeDiagnostics(verbose, warnings)
	return 0, nil
}

// executeRegen runs the already-decided regen workflow against document text
// that was read by the caller from the resolved input source.
func (r *Runner) executeRegen(source inputSource, input string, verbose bool) (int, error) {
	result, warnings, err := Regen(input)
	if err != nil {
		return 1, err
	}
	if err := r.writeOutput(source, result); err != nil {
		return 1, err
	}
	r.writeDiagnostics(verbose, warnings)
	return 0, nil
}

// executeStrip runs the already-decided strip workflow against document text
// that was read by the caller from the resolved input source.
func (r *Runner) executeStrip(source inputSource, input string, raw bool, verbose bool) (int, error) {
	var (
		result   string
		warnings []string
		err      error
	)
	if raw {
		result, warnings, err = StripRaw(input)
	} else {
		result, warnings, err = Strip(input)
	}
	if err != nil {
		return 1, err
	}
	if err := r.writeOutput(source, result); err != nil {
		return 1, err
	}
	r.writeDiagnostics(verbose, warnings)
	return 0, nil
}

// executeCheck runs the check workflow. Unlike the mutating commands it never
// writes document content back, but it still emits warnings in verbose mode.
func (r *Runner) executeCheck(input string, verbose bool) (int, error) {
	ok, warnings, err := Check(input)
	if err != nil {
		return 1, err
	}
	r.writeDiagnostics(verbose, warnings)
	if ok {
		return 0, nil
	}
	return 2, errors.New("document does not match the reconstructed target state")
}

// readInput loads the document from the already-resolved input source.
func (r *Runner) readInput(source inputSource) (string, error) {
	if source.kind == inputSourceFile {
		b, err := r.fs.ReadFile(source.file)
		return string(b), err
	}
	b, err := io.ReadAll(r.stdin)
	return string(b), err
}

// writeOutput routes the transformed content back to the matching destination
// for the resolved input source.
func (r *Runner) writeOutput(source inputSource, content string) error {
	if source.kind == inputSourceFile {
		return r.fs.WriteFile(source.file, []byte(content), 0o644)
	}
	_, err := io.WriteString(r.stdout, content)
	return err
}

// writeDiagnostics emits collected warnings only in verbose mode.
func (r *Runner) writeDiagnostics(verbose bool, warnings []string) {
	if !verbose {
		return
	}
	for _, warning := range warnings {
		fmt.Fprintln(r.stderr, warning)
	}
}

// writeVersion renders the short or verbose version output.
func (r *Runner) writeVersion(verbose bool) {
	if verbose {
		fmt.Fprintf(r.stdout, "mdtoc %s\ncommit: %s\ndate: %s\nGo-based Markdown ToC manager\n", r.buildInfo.Version, r.buildInfo.Commit, r.buildInfo.Date)
		return
	}
	fmt.Fprintf(r.stdout, "mdtoc %s\ncommit: %s\ndate: %s\n", r.buildInfo.Version, r.buildInfo.Commit, r.buildInfo.Date)
}

// isSubcommand reports whether the first CLI token names a supported explicit
// subcommand. Everything else is now handled by root convenience mode.
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
       mdtoc [--file <name> | <name>] [GENERATE OPTIONS]
       mdtoc [GENERATE OPTIONS] < INPUT.md

Commands:
  generate [--file <name> | <name>] [--verbose] [OPTIONS]  generate or update ToC, numbers, and anchors
  check    [--file <name> | <name>] [--verbose]            validate that the document matches its persisted state
  regen    [--file <name> | <name>] [--verbose]            regenerate using persisted container config
  strip    [--file <name> | <name>] [--verbose] [--raw]    remove managed artifacts and keep the container

Details: mdtoc -v   short for mdtoc --help --verbose
`) + "\n"
}

// longHelp returns the root help text with the command summary section.
func longHelp() string {
	return strings.TrimSpace(`mdtoc - deterministic Markdown ToC manager

Usage: mdtoc <command>
       mdtoc [--file <name> | <name>] [GENERATE OPTIONS]
       mdtoc [GENERATE OPTIONS] < INPUT.md

Without a subcommand, mdtoc chooses between regen and generate.
If the input already contains a valid managed container and no generate options
are provided, it behaves like regen. Otherwise it behaves like generate.

Commands:
  generate [--file <name> | <name>] [--verbose] [OPTIONS]  generate or update ToC, numbers, and anchors
  check    [--file <name> | <name>] [--verbose]            validate that the document matches its persisted state
  regen    [--file <name> | <name>] [--verbose]            regenerate using persisted container config
  strip    [--file <name> | <name>] [--verbose] [--raw]    remove managed artifacts and keep the container

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
		return strings.TrimSpace(`mdtoc generate [--file <name> | <name>] [--verbose] [OPTIONS]

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
	return strings.TrimSpace(`mdtoc generate [--file <name> | <name>] [--verbose] [OPTIONS]

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
		return strings.TrimSpace(`mdtoc regen [--file <name> | <name>] [--verbose]

Regenerate using the persisted config from an existing managed container.

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
	}
	return strings.TrimSpace(`mdtoc regen [--file <name> | <name>] [--verbose]

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
}

// stripHelp returns the strip subcommand help text.
func stripHelp(verbose bool) string {
	if verbose {
		return strings.TrimSpace(`mdtoc strip [--file <name> | <name>] [--verbose] [--raw]

Remove managed artifacts and optionally the entire managed container.

Options:
  --raw
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
	}
	return strings.TrimSpace(`mdtoc strip [--file <name> | <name>] [--verbose] [--raw]

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
		return strings.TrimSpace(`mdtoc check [--file <name> | <name>] [--verbose]

Reconstruct the target document state and compare it byte-for-byte.

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
	}
	return strings.TrimSpace(`mdtoc check [--file <name> | <name>] [--verbose]

Options:
  --file, -f <name>
  --verbose, -v
  --help, -h
`) + "\n"
}
