package mdtoc

import (
	"strings"
	"testing"
)

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
