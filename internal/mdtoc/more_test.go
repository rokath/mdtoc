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

	stdout.Reset()
	exitCode, err = runner.Run([]string{"--verbose"})
	if err != nil || exitCode != 0 {
		t.Fatalf("Run(--verbose) failed: exit=%d err=%v", exitCode, err)
	}
	if got := stdout.String(); !strings.Contains(got, "Generate options:") {
		t.Fatalf("Run(--verbose) did not print long help:\n%s", got)
	}

	if got := longHelp(); !strings.Contains(got, "Usage:") {
		t.Fatalf("longHelp missing usage section:\n%s", got)
	}
	if got := longHelp(); !strings.Contains(got, "Help:") {
		t.Fatalf("longHelp missing help section:\n%s", got)
	}
	if got := longHelp(); !strings.Contains(got, "Generate options:") {
		t.Fatalf("longHelp missing generate options section:\n%s", got)
	}
	if got := longHelp(); !strings.Contains(got, "--bullets=auto") {
		t.Fatalf("longHelp missing bullets option:\n%s", got)
	}
	if got := longHelp(); !strings.Contains(got, "Info: https://github.com/rokath/mdtoc/") {
		t.Fatalf("longHelp missing project info link:\n%s", got)
	}
	if got := longHelp(); !strings.Contains(got, "mdtoc <command> --help [--verbose]") {
		t.Fatalf("longHelp missing subcommand help usage:\n%s", got)
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
	if got := regenHelp(true); !strings.Contains(got, "persisted config") {
		t.Fatalf("regenHelp(verbose) missing verbose text:\n%s", got)
	}
	if got := generateHelp(false); strings.Contains(got, "Generate or update") {
		t.Fatalf("generateHelp(non-verbose) unexpectedly contains verbose text:\n%s", got)
	}
	if got := generateHelp(true); !strings.Contains(got, "--bullets, -b <auto|*|-|+>") {
		t.Fatalf("generateHelp(verbose) missing bullets option:\n%s", got)
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
	if got := stdout.String(); !strings.Contains(got, "Help:") {
		t.Fatalf("verbose root help missing help section:\n%s", got)
	}
	if got := stdout.String(); !strings.Contains(got, "Generate options:") {
		t.Fatalf("verbose root help missing generate options section:\n%s", got)
	}
	if got := stdout.String(); !strings.Contains(got, "mdtoc <command> --help [--verbose]") {
		t.Fatalf("verbose root help missing subcommand help hint:\n%s", got)
	}
	if got := stdout.String(); !strings.Contains(got, "check    [--file <name>] [--verbose]") {
		t.Fatalf("verbose root help missing reordered check usage:\n%s", got)
	}

	stdout.Reset()
	exitCode, err = runner.Run([]string{"regen", "--help"})
	if err != nil || exitCode != 0 {
		t.Fatalf("regen help failed: exit=%d err=%v", exitCode, err)
	}
	if got := stdout.String(); !strings.Contains(got, "mdtoc regen") {
		t.Fatalf("regen help missing usage:\n%s", got)
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
	if exitCode, err = runner.Run([]string{"generate", "--anchor", "maybe"}); err == nil || exitCode != 1 {
		t.Fatalf("generate invalid anchor should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"generate", "--numbering", "maybe"}); err == nil || exitCode != 1 {
		t.Fatalf("generate invalid numbering should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"generate", "--toc", "maybe"}); err == nil || exitCode != 1 {
		t.Fatalf("generate invalid toc should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"generate", "--bullets", "x"}); err == nil || exitCode != 1 {
		t.Fatalf("generate invalid bullets should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"strip", "--badflag"}); err == nil || exitCode != 1 {
		t.Fatalf("strip invalid flag should fail, exit=%d err=%v", exitCode, err)
	}
	if exitCode, err = runner.Run([]string{"regen", "--badflag"}); err == nil || exitCode != 1 {
		t.Fatalf("regen invalid flag should fail, exit=%d err=%v", exitCode, err)
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
		"container-version=v2",
		"numbering=true",
		"min-level=2",
		"max-level=4",
		"anchor=off",
		"toc=true",
		"bullets=auto",
		"state=stripped",
		configEnd,
	}
	cfg, err := parseConfig(valid)
	if err != nil {
		t.Fatalf("parseConfig(valid) error: %v", err)
	}
	if cfg.Anchor != AnchorOff || cfg.State != StateStripped || cfg.Bullets != BulletAuto {
		t.Fatalf("parseConfig(valid) returned unexpected config: %+v", cfg)
	}
	if cfg.ContainerVersion != ContainerVersionV2 {
		t.Fatalf("parseConfig(valid) did not record v2 container version: %+v", cfg)
	}

	legacy := []string{
		configStart,
		"numbering=on",
		"min-level=2",
		"max-level=4",
		"anchors=off",
		"toc=on",
		"state=stripped",
		configEnd,
	}
	cfg, err = parseConfig(legacy)
	if err != nil {
		t.Fatalf("parseConfig(legacy) error: %v", err)
	}
	if cfg.Bullets != BulletStar || cfg.BulletsExplicit {
		t.Fatalf("legacy config should default bullets=* : %+v", cfg)
	}
	if cfg.ContainerVersion != ContainerVersionV1 {
		t.Fatalf("legacy config should default to implicit v1: %+v", cfg)
	}

	cases := []struct {
		name  string
		lines []string
	}{
		{"badLength", valid[:8]},
		{"badDelimiters", append([]string{"<!-- wrong -->"}, valid[1:]...)},
		{"badKeyOrder", []string{configStart, "anchors=on", "min-level=2", "max-level=4", "numbering=on", "toc=on", "bullets=auto", "state=generated", configEnd}},
		{"badVersion", []string{configStart, "container-version=v3", "numbering=true", "min-level=2", "max-level=4", "anchor=github", "toc=true", "bullets=auto", "state=generated", configEnd}},
		{"badState", []string{configStart, "container-version=v2", "numbering=true", "min-level=2", "max-level=4", "anchor=github", "toc=true", "bullets=auto", "state=weird", configEnd}},
		{"badMin", []string{configStart, "container-version=v2", "numbering=true", "min-level=x", "max-level=4", "anchor=github", "toc=true", "bullets=auto", "state=generated", configEnd}},
		{"badMax", []string{configStart, "container-version=v2", "numbering=true", "min-level=2", "max-level=x", "anchor=github", "toc=true", "bullets=auto", "state=generated", configEnd}},
		{"badBullets", []string{configStart, "container-version=v2", "numbering=true", "min-level=2", "max-level=4", "anchor=github", "toc=true", "bullets=x", "state=generated", configEnd}},
		{"missingV2Bullets", []string{configStart, "container-version=v2", "numbering=true", "min-level=2", "max-level=4", "anchor=github", "toc=true", "state=generated", configEnd}},
	}
	for _, tc := range cases {
		if _, err := parseConfig(tc.lines); err == nil {
			t.Fatalf("parseConfig(%s) unexpectedly succeeded", tc.name)
		}
	}
	if _, err := parseConfig([]string{configStart, "container-version=v2", "numbering=true", "min-level=2", "max-level=4", "anchor=github", "toc=true", "bullets=auto", "state=generated", "extra=line", configEnd}); err == nil || !strings.Contains(err.Error(), "please update mdtoc") {
		t.Fatalf("parseConfig should hint about newer mdtoc versions for versioned length mismatches: %v", err)
	}

	if _, err := parseBoolValue("maybe"); err == nil {
		t.Fatalf("parseBoolValue should reject invalid values")
	}
	if v, err := parseBoolValue("off"); err != nil || v {
		t.Fatalf("parseBoolValue(off) = %v, %v", v, err)
	}
	if v, err := parseBoolValue("true"); err != nil || !v {
		t.Fatalf("parseBoolValue(true) = %v, %v", v, err)
	}
	if mode, err := parseAnchorMode("gitlab"); err != nil || mode != AnchorGitLab {
		t.Fatalf("parseAnchorMode(gitlab) = %q, %v", mode, err)
	}
	if mode, err := parseAnchorMode("off"); err != nil || mode != AnchorOff {
		t.Fatalf("parseAnchorMode(off) = %q, %v", mode, err)
	}
	if mode, err := parseAnchorMode("false"); err != nil || mode != AnchorOff {
		t.Fatalf("parseAnchorMode(false) = %q, %v", mode, err)
	}
	if version, err := parseContainerVersion("v2"); err != nil || version != ContainerVersionV2 {
		t.Fatalf("parseContainerVersion(v2) = %q, %v", version, err)
	}
	if _, err := parseContainerVersion("v3"); err == nil {
		t.Fatalf("parseContainerVersion should reject invalid values")
	}
	if _, err := parseAnchorMode("maybe"); err == nil {
		t.Fatalf("parseAnchorMode should reject invalid values")
	}
	if boolString(false) != "false" {
		t.Fatalf("boolString(false) must return false")
	}
	if mode, err := parseBulletMode("-"); err != nil || mode != BulletDash {
		t.Fatalf("parseBulletMode(-) = %q, %v", mode, err)
	}
	if _, err := parseBulletMode("x"); err == nil {
		t.Fatalf("parseBulletMode should reject invalid values")
	}

	for _, cfg := range []Config{
		{ContainerVersion: "", MinLevel: 0, MaxLevel: 4, Bullets: BulletAuto, State: StateGenerated},
		{ContainerVersion: ContainerVersionV2, MinLevel: 2, MaxLevel: 7, Bullets: BulletAuto, State: StateGenerated},
		{ContainerVersion: ContainerVersionV2, MinLevel: 4, MaxLevel: 2, Bullets: BulletAuto, State: StateGenerated},
		{ContainerVersion: ContainerVersionV2, MinLevel: 2, MaxLevel: 4, Bullets: BulletMode("x"), State: StateGenerated},
		{ContainerVersion: ContainerVersionV2, MinLevel: 2, MaxLevel: 4, Bullets: BulletAuto, State: State("broken")},
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
	if bullet, ok := detectListBullet("  - item"); !ok || bullet != BulletDash {
		t.Fatalf("detectListBullet dash case failed: %q %v", bullet, ok)
	}
	if bullet, ok := detectListBullet("\t+ item"); !ok || bullet != BulletPlus {
		t.Fatalf("detectListBullet plus case failed: %q %v", bullet, ok)
	}
	if _, ok := detectListBullet("***"); ok {
		t.Fatalf("detectListBullet should reject thematic breaks")
	}
	if got := detectDominantBullet([]string{"* a", "- b", "+ c"}); got != BulletStar {
		t.Fatalf("detectDominantBullet tie = %q, want *", got)
	}
	if got := detectDominantBullet([]string{"- a", "+ b"}); got != BulletDash {
		t.Fatalf("detectDominantBullet dash-plus tie = %q, want -", got)
	}
}

// TestParseDocumentIgnoresMdtocOffMarkersInsideFences verifies that off/on markers only act outside ignored regions.
func TestParseDocumentIgnoresMdtocOffMarkersInsideFences(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"```md",
		"<!-- mdtoc off -->",
		"## Still code",
		"```",
		"",
		"## Real heading",
	}, "\n") + "\n"

	parsed, err := ParseDocument(input)
	if err != nil {
		t.Fatalf("ParseDocument error: %v", err)
	}
	if len(parsed.Headings) != 2 {
		t.Fatalf("unexpected headings parsed: %+v", parsed.Headings)
	}
	if parsed.Headings[1].TitleText != "Real heading" {
		t.Fatalf("second heading = %q, want Real heading", parsed.Headings[1].TitleText)
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

	lines := []string{startMarker, configStart, "container-version=v2", "numbering=true", "min-level=2", "max-level=4", "anchor=github", "toc=true", "bullets=auto", "state=generated", configEnd, endMarker}
	if _, err := buildContainer(lines, -1, 1, -1, -1); err == nil {
		t.Fatalf("buildContainer should reject incomplete container")
	}
	if _, err := buildContainer(lines, 3, 1, 1, 9); err == nil {
		t.Fatalf("buildContainer should reject reversed markers")
	}
	if _, err := buildContainer(lines, 0, 10, -1, -1); err == nil {
		t.Fatalf("buildContainer should reject missing config")
	}
	if _, err := buildContainer(lines, 0, 10, 0, 9); err == nil {
		t.Fatalf("buildContainer should reject config outside managed area")
	}
	if _, err := buildContainer(lines, 0, 10, 1, 8); err == nil {
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
		"container-version=v2",
		"numbering=true",
		"min-level=2",
		"max-level=4",
		"anchor=github",
		"toc=true",
		"bullets=auto",
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
		"container-version=v2",
		"numbering=true",
		"min-level=2",
		"max-level=4",
		"anchor=github",
		"toc=true",
		"bullets=auto",
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
