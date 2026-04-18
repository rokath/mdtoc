package mdtoc

import (
	"fmt"
	"strings"
	"testing"
)

// TestGenerateCreatesContainerAndDerivedArtifacts verifies the default generated container output.
func TestGenerateCreatesContainerAndDerivedArtifacts(t *testing.T) {
	input := "# Title\n\n## Intro\n\n### API\n"
	got, warnings, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if len(warnings) != 0 {
		t.Fatalf("unexpected warnings: %v", warnings)
	}
	checks := []string{startMarker, "* [1. Intro](#intro)", "  * [1.1. API](#api)", `## 1. <a id="intro"></a>Intro`, `### 1.1. <a id="api"></a>API`, "state=generated"}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("generated output missing %q:\n%s", check, got)
		}
	}
}

// TestGenerateIsIdempotent verifies that repeated generation does not change the document again.
func TestGenerateIsIdempotent(t *testing.T) {
	input := "# Title\n\n## Intro\n"
	first, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("first Generate error: %v", err)
	}
	second, _, err := Generate(first, DefaultOptions())
	if err != nil {
		t.Fatalf("second Generate error: %v", err)
	}
	if first != second {
		t.Fatalf("generate is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}
}

// TestGeneratePreservesForeignTOCContentAsComment verifies preservation of handwritten managed-area content.
func TestGeneratePreservesForeignTOCContentAsComment(t *testing.T) {
	input := strings.Join([]string{startMarker, "Some handwritten note", configStart, "numbering=on", "min-level=2", "max-level=4", "anchors=on", "toc=on", "state=generated", configEnd, endMarker, "", "## Intro"}, "\n") + "\n"
	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, preservedCommentHeader+"\nSome handwritten note\n-->") {
		t.Fatalf("foreign ToC content was not preserved:\n%s", got)
	}
}

// TestGenerateIgnoresHeadingsInsideFencesAndComments verifies ignored-region parsing behavior.
func TestGenerateIgnoresHeadingsInsideFencesAndComments(t *testing.T) {
	input := strings.Join([]string{"# Title", "", "```md", "## Code heading", "```", "", "<!--", "## Comment heading", "-->", "", "## Real heading"}, "\n") + "\n"
	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if strings.Contains(got, "Code heading](#code-heading)") || strings.Contains(got, "Comment heading](#comment-heading)") {
		t.Fatalf("ignored regions leaked into ToC:\n%s", got)
	}
	if !strings.Contains(got, "* [1. Real heading](#real-heading)") {
		t.Fatalf("real heading missing from ToC:\n%s", got)
	}
}

// TestStripKeepsContainerAndMarksStateStripped verifies stripped-state rendering with the container retained.
func TestStripKeepsContainerAndMarksStateStripped(t *testing.T) {
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	got, _, err := Strip(generated)
	if err != nil {
		t.Fatalf("Strip error: %v", err)
	}
	checks := []string{startMarker, endMarker, "state=stripped", "## Intro"}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("stripped output missing %q:\n%s", check, got)
		}
	}
	if strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("managed artifacts were not removed:\n%s", got)
	}
}

// TestStripRawRemovesContainerAndManagedArtifacts verifies full raw stripping behavior.
func TestStripRawRemovesContainerAndManagedArtifacts(t *testing.T) {
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	got, _, err := StripRaw(generated)
	if err != nil {
		t.Fatalf("StripRaw error: %v", err)
	}
	if strings.Contains(got, startMarker) || strings.Contains(got, endMarker) || strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("raw strip left managed artifacts behind:\n%s", got)
	}
	if !strings.Contains(got, "## Intro") {
		t.Fatalf("raw strip removed heading text:\n%s", got)
	}
}

// TestCheckMatchesAndDetectsMismatch verifies both matching and mismatching check outcomes.
func TestCheckMatchesAndDetectsMismatch(t *testing.T) {
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	ok, _, err := Check(generated)
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !ok {
		t.Fatalf("Check reported mismatch for generated document")
	}
	broken := strings.Replace(generated, "Intro", "Changed", 1)
	ok, _, err = Check(broken)
	if err != nil {
		t.Fatalf("Check on broken document error: %v", err)
	}
	if ok {
		t.Fatalf("Check did not detect mismatch")
	}
}

// TestGenerateAndCheckPreserveRelocatedContainerPosition verifies stable behavior for moved containers.
func TestGenerateAndCheckPreserveRelocatedContainerPosition(t *testing.T) {
	source := "# Title\n\nIntro paragraph.\n\n## Intro\n"
	generated, _, err := Generate(source, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	relocated, err := relocateContainerAfterParagraph(generated, "Intro paragraph.")
	if err != nil {
		t.Fatalf("relocateContainerAfterParagraph error: %v", err)
	}

	regenerated, _, err := Generate(relocated, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate on relocated document error: %v", err)
	}
	if strings.Index(regenerated, startMarker) < strings.Index(regenerated, "Intro paragraph.") {
		t.Fatalf("container moved before relocated position:\n%s", regenerated)
	}
	if strings.Index(regenerated, startMarker) > strings.Index(regenerated, "## 1. <a id=\"intro\"></a>Intro") {
		t.Fatalf("container moved after the managed heading:\n%s", regenerated)
	}

	ok, _, err := Check(relocated)
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !ok {
		t.Fatalf("Check rejected a document with a valid relocated container:\n%s", relocated)
	}
}

func relocateContainerAfterParagraph(input, marker string) (string, error) {
	parsed, err := ParseDocument(input)
	if err != nil {
		return "", err
	}
	if parsed.Container == nil {
		return "", fmt.Errorf("document does not contain an mdtoc container")
	}

	containerLines := append([]string(nil), parsed.Lines[parsed.Container.StartLine:parsed.Container.EndLine+1]...)
	bodyLines := append(append([]string{}, parsed.Lines[:parsed.Container.StartLine]...), parsed.Lines[parsed.Container.EndLine+1:]...)

	insertAt := -1
	for i, line := range bodyLines {
		if line == marker {
			insertAt = i + 1
			break
		}
	}
	if insertAt == -1 {
		return "", fmt.Errorf("marker %q not found", marker)
	}

	relocated := append([]string{}, bodyLines[:insertAt]...)
	relocated = append(relocated, "")
	relocated = append(relocated, containerLines...)
	relocated = append(relocated, "")
	relocated = append(relocated, bodyLines[insertAt:]...)
	return joinLines(relocated), nil
}
