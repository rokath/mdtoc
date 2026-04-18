package mdtoc

import (
	"os"
	"strings"
	"testing"
)

// TestRunnerRunAndHelpHelpers covers root CLI helper behavior and simple help builders.
func TestRunnerRunAndHelpHelpers(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(""), &stdout, &stderr)

	exitCode, err := runner.Run(nil)
	if err != nil {
		t.Fatalf("Run(nil) error: %v", err)
	}
	if exitCode != 0 || !strings.Contains(stdout.String(), "Usage:") {
		t.Fatalf("unexpected root help output:\n%s", stdout.String())
	}

	if _, err := runner.Run([]string{"unknown"}); err == nil {
		t.Fatalf("Run should reject unknown command")
	}
	if _, err := runner.Run([]string{"--bogus"}); err == nil {
		t.Fatalf("Run should reject unknown root arg")
	}

	if got := longHelp(); !strings.Contains(got, "Commands:") {
		t.Fatalf("longHelp missing commands section:\n%s", got)
	}
	if got := stripHelp(true); !strings.Contains(got, "Remove managed artifacts") {
		t.Fatalf("stripHelp(verbose) missing verbose text:\n%s", got)
	}
	if got := stripHelp(false); strings.Contains(got, "Remove managed artifacts") {
		t.Fatalf("stripHelp(non-verbose) unexpectedly contains verbose text:\n%s", got)
	}
	if got := checkHelp(true); !strings.Contains(got, "compare it byte-for-byte") {
		t.Fatalf("checkHelp(verbose) missing verbose text:\n%s", got)
	}
	if got := generateHelp(false); strings.Contains(got, "Generate or update") {
		t.Fatalf("generateHelp(non-verbose) unexpectedly contains verbose text:\n%s", got)
	}
	if !hasFlag([]string{"--raw", "--help"}, "--raw") || hasFlag([]string{"--help"}, "--raw") {
		t.Fatalf("hasFlag returned unexpected result")
	}
}

// TestRunnerSubcommandHelpAndVerboseDiagnostics verifies subcommand help rendering and verbose warnings.
func TestRunnerSubcommandHelpAndVerboseDiagnostics(t *testing.T) {
	input := strings.Join([]string{"## <a id=\"foreign\">Intro"}, "\n") + "\n"
	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(input), &stdout, &stderr)

	exitCode, err := runner.Run([]string{"strip", "--help", "--verbose"})
	if err != nil || exitCode != 0 {
		t.Fatalf("strip help failed: exit=%d err=%v", exitCode, err)
	}
	if got := stdout.String(); !strings.Contains(got, "Remove managed artifacts") {
		t.Fatalf("strip help output missing expected text:\n%s", got)
	}

	stdout.Reset()
	stderr.Reset()
	runner = NewRunner(strings.NewReader(input), &stdout, &stderr)
	exitCode, err = runner.Run([]string{"strip", "--raw", "--verbose"})
	if err != nil || exitCode != 0 {
		t.Fatalf("strip --raw failed: exit=%d err=%v", exitCode, err)
	}
	if !strings.Contains(stderr.String(), "non-managed inline anchor") {
		t.Fatalf("expected verbose diagnostics, got:\n%s", stderr.String())
	}
}

// TestRunnerRootVerboseHelpAndSubcommandErrorPaths exercises root and subcommand CLI error paths.
func TestRunnerRootVerboseHelpAndSubcommandErrorPaths(t *testing.T) {
	var stdout, stderr strings.Builder
	runner := NewRunner(strings.NewReader(""), &stdout, &stderr)

	exitCode, err := runner.Run([]string{"--help", "--verbose"})
	if err != nil || exitCode != 0 {
		t.Fatalf("root verbose help failed: exit=%d err=%v", exitCode, err)
	}
	if got := stdout.String(); !strings.Contains(got, "Commands:") {
		t.Fatalf("verbose root help missing commands section:\n%s", got)
	}

	stdout.Reset()
	exitCode, err = runner.Run([]string{"check", "--help"})
	if err != nil || exitCode != 0 {
		t.Fatalf("check help failed: exit=%d err=%v", exitCode, err)
	}
	if got := stdout.String(); !strings.Contains(got, "mdtoc check") {
		t.Fatalf("check help missing usage:\n%s", got)
	}

	stdout.Reset()
	if exitCode, err = runner.Run([]string{"generate", "--anchors", "maybe"}); err == nil || exitCode != 1 {
		t.Fatalf("generate invalid anchors should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"generate", "--numbering", "maybe"}); err == nil || exitCode != 1 {
		t.Fatalf("generate invalid numbering should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"generate", "--toc", "maybe"}); err == nil || exitCode != 1 {
		t.Fatalf("generate invalid toc should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"strip", "--badflag"}); err == nil || exitCode != 1 {
		t.Fatalf("strip invalid flag should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"check", "--badflag"}); err == nil || exitCode != 1 {
		t.Fatalf("check invalid flag should fail, exit=%d err=%v", exitCode, err)
	}
}

// TestInteractiveInputDetectionBranches verifies interactive-input detection across reader kinds.
func TestInteractiveInputDetectionBranches(t *testing.T) {
	if isInteractiveInput(strings.NewReader("")) {
		t.Fatalf("strings.Reader must not be treated as interactive")
	}

	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("os.Pipe error: %v", err)
	}
	defer r.Close()
	defer w.Close()
	if isInteractiveInput(r) {
		t.Fatalf("pipe must not be treated as interactive")
	}

	badFile := os.NewFile(^uintptr(0), "bad")
	if isInteractiveInput(badFile) {
		t.Fatalf("invalid file descriptor must not be treated as interactive")
	}
}

// TestConfigParsingAndValidationErrors verifies config parsing success and representative error paths.
func TestConfigParsingAndValidationErrors(t *testing.T) {
	valid := []string{
		configStart,
		"numbering=on",
		"min-level=2",
		"max-level=4",
		"anchors=off",
		"toc=on",
		"state=stripped",
		configEnd,
	}
	cfg, err := parseConfig(valid)
	if err != nil {
		t.Fatalf("parseConfig(valid) error: %v", err)
	}
	if cfg.Anchors || cfg.State != StateStripped {
		t.Fatalf("parseConfig(valid) returned unexpected config: %+v", cfg)
	}

	cases := []struct {
		name  string
		lines []string
	}{
		{"badLength", valid[:7]},
		{"badDelimiters", append([]string{"<!-- wrong -->"}, valid[1:]...)},
		{"badKeyOrder", []string{configStart, "anchors=on", "min-level=2", "max-level=4", "numbering=on", "toc=on", "state=generated", configEnd}},
		{"badState", []string{configStart, "numbering=on", "min-level=2", "max-level=4", "anchors=on", "toc=on", "state=weird", configEnd}},
		{"badMin", []string{configStart, "numbering=on", "min-level=x", "max-level=4", "anchors=on", "toc=on", "state=generated", configEnd}},
		{"badMax", []string{configStart, "numbering=on", "min-level=2", "max-level=x", "anchors=on", "toc=on", "state=generated", configEnd}},
	}
	for _, tc := range cases {
		if _, err := parseConfig(tc.lines); err == nil {
			t.Fatalf("parseConfig(%s) unexpectedly succeeded", tc.name)
		}
	}

	if _, err := parseOnOff("maybe"); err == nil {
		t.Fatalf("parseOnOff should reject invalid values")
	}
	if v, err := parseOnOff("off"); err != nil || v {
		t.Fatalf("parseOnOff(off) = %v, %v", v, err)
	}
	if onOff(false) != "off" {
		t.Fatalf("onOff(false) must return off")
	}

	for _, cfg := range []Config{
		{MinLevel: 0, MaxLevel: 4, State: StateGenerated},
		{MinLevel: 2, MaxLevel: 7, State: StateGenerated},
		{MinLevel: 4, MaxLevel: 2, State: StateGenerated},
		{MinLevel: 2, MaxLevel: 4, State: State("broken")},
	} {
		if err := cfg.Validate(); err == nil {
			t.Fatalf("Validate unexpectedly succeeded for %+v", cfg)
		}
	}
}

// TestMarkdownHelpersAndHeadingParserBranches covers helper parsing branches for inline Markdown and headings.
func TestMarkdownHelpersAndHeadingParserBranches(t *testing.T) {
	if got := extractInlineText("![alt *x*](img.png) and [link](x) <b>tag</b> `code`"); got != "alt x and link tag code" {
		t.Fatalf("extractInlineText returned %q", got)
	}
	if label, consumed, ok := parseBracketLinkLike("[link](target)"); !ok || label != "link" || consumed == 0 {
		t.Fatalf("parseBracketLinkLike valid case failed: %q %d %v", label, consumed, ok)
	}
	if _, _, ok := parseBracketLinkLike("nope"); ok {
		t.Fatalf("parseBracketLinkLike should reject non-links")
	}
	if _, _, ok := parseBracketLinkLike("[broken](target"); ok {
		t.Fatalf("parseBracketLinkLike should reject unterminated target")
	}

	h, warning, ok, err := parseHeadingLine("## <a id=\"broken\">Intro", 4)
	if err != nil || !ok || warning == "" {
		t.Fatalf("parseHeadingLine malformed anchor case failed: h=%+v warning=%q ok=%v err=%v", h, warning, ok, err)
	}
	if h.TitleText != "Intro" {
		t.Fatalf("unexpected TitleText: %q", h.TitleText)
	}

	h, warning, ok, err = parseHeadingLine("### 1.2. <a id=\"api\"></a>API", 1)
	if err != nil || !ok || warning != "" {
		t.Fatalf("parseHeadingLine managed heading failed: h=%+v warning=%q ok=%v err=%v", h, warning, ok, err)
	}
	if h.ManagedNumber != "1.2." || h.ManagedAnchor != `<a id="api"></a>` {
		t.Fatalf("managed artifacts not parsed: %+v", h)
	}

	if _, _, ok, err := parseHeadingLine("## ", 1); err != nil || ok {
		t.Fatalf("empty heading should not be accepted")
	}
}

// TestParserAndContainerHelpers covers structural parser helpers and container validation errors.
func TestParserAndContainerHelpers(t *testing.T) {
	if got := splitLines(""); len(got) != 0 {
		t.Fatalf("splitLines(empty) = %v", got)
	}
	if got := splitLines("a\nb\n"); len(got) != 2 || got[1] != "b" {
		t.Fatalf("splitLines returned %v", got)
	}
	if fenceOpen("~~~go") != "~~~" || fenceOpen("plain") != "" {
		t.Fatalf("fenceOpen returned unexpected values")
	}
	if end := findConfigEnd([]string{"x", configEnd}, 0); end != 1 {
		t.Fatalf("findConfigEnd = %d, want 1", end)
	}
	if end := findConfigEnd([]string{"x"}, 0); end != -1 {
		t.Fatalf("findConfigEnd missing terminator = %d, want -1", end)
	}

	lines := []string{startMarker, configStart, "numbering=on", "min-level=2", "max-level=4", "anchors=on", "toc=on", "state=generated", configEnd, endMarker}
	if _, err := buildContainer(lines, -1, 1, -1, -1); err == nil {
		t.Fatalf("buildContainer should reject incomplete container")
	}
	if _, err := buildContainer(lines, 3, 1, 1, 8); err == nil {
		t.Fatalf("buildContainer should reject reversed markers")
	}
	if _, err := buildContainer(lines, 0, 9, -1, -1); err == nil {
		t.Fatalf("buildContainer should reject missing config")
	}
	if _, err := buildContainer(lines, 0, 9, 0, 8); err == nil {
		t.Fatalf("buildContainer should reject config outside managed area")
	}
	if _, err := buildContainer(lines, 0, 9, 1, 7); err == nil {
		t.Fatalf("buildContainer should reject misplaced config end")
	}

	if _, err := ParseDocument(strings.Join([]string{startMarker, startMarker, endMarker}, "\n")); err == nil {
		t.Fatalf("ParseDocument should reject duplicate start marker")
	}
	if _, err := ParseDocument(strings.Join([]string{startMarker, endMarker}, "\n")); err == nil {
		t.Fatalf("ParseDocument should reject missing config block")
	}
}

// TestProcessHelperBranches covers lower-level process helpers and state-specific check behavior.
func TestProcessHelperBranches(t *testing.T) {
	strippedInput := strings.Join([]string{
		startMarker,
		configStart,
		"numbering=on",
		"min-level=2",
		"max-level=4",
		"anchors=on",
		"toc=on",
		"state=stripped",
		configEnd,
		endMarker,
		"",
		"## Intro",
	}, "\n") + "\n"
	ok, _, err := Check(strippedInput)
	if err != nil || !ok {
		t.Fatalf("Check(stripped) = %v, %v", ok, err)
	}
	if ok, _, err = Check("# Title\n"); err == nil || ok {
		t.Fatalf("Check without config should fail, got ok=%v err=%v", ok, err)
	}

	unsupported := strings.Replace(strippedInput, "state=stripped", "state=broken", 1)
	if _, _, err := Check(unsupported); err == nil {
		t.Fatalf("Check should reject unsupported state")
	}

	generatedStateInput := strings.Join([]string{
		startMarker,
		"* [1. Intro](#intro)",
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
		"## 1. <a id=\"intro\"></a>Intro",
	}, "\n") + "\n"
	if ok, _, err = Check(generatedStateInput); err != nil || !ok {
		t.Fatalf("Check(generated) = %v, %v", ok, err)
	}

	if _, _, err := Strip("# Title\n"); err == nil {
		t.Fatalf("Strip without config should fail")
	}

	if got := normalizeInput("a\r\nb"); got != "a\nb\n" {
		t.Fatalf("normalizeInput returned %q", got)
	}
	if got := normalizeInput(""); got != "" {
		t.Fatalf("normalizeInput(empty) = %q", got)
	}

	container := &Container{TOCArea: []string{
		"custom line",
		preservedCommentHeader,
		"kept",
	}}
	preserved := preserveForeignTOC(container)
	if len(preserved) == 0 || !strings.Contains(strings.Join(preserved, "\n"), "custom line") {
		t.Fatalf("preserveForeignTOC did not preserve custom text: %v", preserved)
	}
	container = &Container{TOCArea: []string{
		"* [1. Intro](#intro)",
		"",
		preservedCommentHeader,
		"already wrapped",
		"-->",
		"handwritten",
	}}
	preserved = preserveForeignTOC(container)
	gotPreserved := strings.Join(preserved, "\n")
	if !strings.Contains(gotPreserved, "already wrapped") || !strings.Contains(gotPreserved, "handwritten") {
		t.Fatalf("preserveForeignTOC did not keep expected chunks: %s", gotPreserved)
	}

	if !isGeneratedTOCLine("") || isGeneratedTOCLine("plain text") {
		t.Fatalf("isGeneratedTOCLine returned unexpected result")
	}
	if wrapPreservedComment([]string{"", ""}) != nil {
		t.Fatalf("wrapPreservedComment should drop empty chunks")
	}
	if got := trimBlankEdges([]string{"", "x", ""}); len(got) != 1 || got[0] != "x" {
		t.Fatalf("trimBlankEdges returned %v", got)
	}
	if got := prependContainer([]string{}, []string{"A"}); len(got) != 1 || got[0] != "A" {
		t.Fatalf("prependContainer(empty body) returned %v", got)
	}
	if got := placeContainer([]string{"body"}, []string{"toc"}, &Container{StartLine: 5}); len(got) != 2 || got[1] != "toc" {
		t.Fatalf("placeContainer(out of range) returned %v", got)
	}
	if got := joinLines(nil); got != "" {
		t.Fatalf("joinLines(nil) = %q", got)
	}
}
