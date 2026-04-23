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
	checks := []string{startMarker, "* [1. Intro](#intro)", "  * [1.1. API](#api)", `## 1. <a id="intro"></a>Intro`, `### 1.1. <a id="api"></a>API`, "bullets=auto", "state=generated"}
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

// TestGenerateIsIdempotentWithPlusBullets verifies repeated generation for auto-detected plus ToCs.
func TestGenerateIsIdempotentWithPlusBullets(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"+ alpha",
		"+ beta",
		"+ gamma",
		"",
		"## Overview",
		"## Overview",
	}, "\n") + "\n"

	first, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("first Generate error: %v", err)
	}
	if !strings.Contains(first, "+ [1. Overview](#overview)") || !strings.Contains(first, "+ [2. Overview](#overview-1)") {
		t.Fatalf("first generation did not use plus bullets with repeated headings:\n%s", first)
	}

	second, _, err := Generate(first, DefaultOptions())
	if err != nil {
		t.Fatalf("second Generate error: %v", err)
	}
	if first != second {
		t.Fatalf("generate with plus bullets is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	ok, _, err := Check(first)
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !ok {
		t.Fatalf("Check rejected generated plus-bullet output:\n%s", first)
	}
}

// TestGenerateIsIdempotentWithDashBullets verifies repeated generation for auto-detected dash ToCs.
func TestGenerateIsIdempotentWithDashBullets(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"- alpha",
		"- beta",
		"- gamma",
		"",
		"## Overview",
		"## Overview",
	}, "\n") + "\n"

	first, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("first Generate error: %v", err)
	}
	if !strings.Contains(first, "- [1. Overview](#overview)") || !strings.Contains(first, "- [2. Overview](#overview-1)") {
		t.Fatalf("first generation did not use dash bullets with repeated headings:\n%s", first)
	}

	second, _, err := Generate(first, DefaultOptions())
	if err != nil {
		t.Fatalf("second Generate error: %v", err)
	}
	if first != second {
		t.Fatalf("generate with dash bullets is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	ok, _, err := Check(first)
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !ok {
		t.Fatalf("Check rejected generated dash-bullet output:\n%s", first)
	}
}

// TestGenerateIsIdempotentWithStarBullets verifies repeated generation for explicit star ToCs.
func TestGenerateIsIdempotentWithStarBullets(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"- alpha",
		"- beta",
		"- gamma",
		"",
		"## Overview",
		"## Overview",
	}, "\n") + "\n"

	first, _, err := Generate(input, Options{
		Numbering: true,
		MinLevel:  2,
		MaxLevel:  4,
		Anchor:    AnchorGitHub,
		TOC:       true,
		Bullets:   BulletStar,
	})
	if err != nil {
		t.Fatalf("first Generate error: %v", err)
	}
	if !strings.Contains(first, "* [1. Overview](#overview)") || !strings.Contains(first, "* [2. Overview](#overview-1)") {
		t.Fatalf("first generation did not use star bullets with repeated headings:\n%s", first)
	}

	second, _, err := Generate(first, Options{
		Numbering: true,
		MinLevel:  2,
		MaxLevel:  4,
		Anchor:    AnchorGitHub,
		TOC:       true,
		Bullets:   BulletStar,
	})
	if err != nil {
		t.Fatalf("second Generate error: %v", err)
	}
	if first != second {
		t.Fatalf("generate with star bullets is not idempotent\nfirst:\n%s\nsecond:\n%s", first, second)
	}

	ok, _, err := Check(first)
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !ok {
		t.Fatalf("Check rejected generated star-bullet output:\n%s", first)
	}
}

// TestGeneratePreservesForeignTOCContentAsComment verifies preservation of handwritten managed-area content.
func TestGeneratePreservesForeignTOCContentAsComment(t *testing.T) {
	input := strings.Join([]string{startMarker, "Some handwritten note", configStart, "numbering=on", "min-level=2", "max-level=4", "anchor=github", "toc=on", "state=generated", configEnd, endMarker, "", "## Intro"}, "\n") + "\n"
	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, preservedCommentHeader+"\nSome handwritten note\n-->") {
		t.Fatalf("foreign ToC content was not preserved:\n%s", got)
	}
}

// TestRegenReusesPersistedContainerConfig verifies that regen reads stored config instead of CLI defaults.
func TestRegenReusesPersistedContainerConfig(t *testing.T) {
	input := "# Title\n\n## Intro\n\n### API\n"
	generated, _, err := Generate(input, Options{
		Numbering: false,
		MinLevel:  2,
		MaxLevel:  3,
		Anchor:    AnchorFalse,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	manual := strings.Replace(generated, "* [Intro](#intro)", "* [WRONG](#intro)", 1)
	manual = strings.Replace(manual, "## Intro", "## 9. Intro", 1)

	got, _, err := Regen(manual)
	if err != nil {
		t.Fatalf("Regen error: %v", err)
	}
	if strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("regen did not honor persisted config:\n%s", got)
	}
	if !strings.Contains(got, "* [Intro](#intro)") || !strings.Contains(got, "## Intro") {
		t.Fatalf("regen did not reconstruct expected output:\n%s", got)
	}
}

// TestRegenRequiresManagedConfig verifies that regen fails without a valid managed container.
func TestRegenRequiresManagedConfig(t *testing.T) {
	if _, _, err := Regen("# Title\n\n## Intro\n"); err == nil {
		t.Fatalf("Regen without config unexpectedly succeeded")
	}
}

// TestRegenRestoresGeneratedStateFromStrippedInput verifies regen after a stripped-state workflow.
func TestRegenRestoresGeneratedStateFromStrippedInput(t *testing.T) {
	generated, _, err := Generate("# Title\n\n## Intro\n", Options{
		Numbering: true,
		MinLevel:  1,
		MaxLevel:  4,
		Anchor:    AnchorGitHub,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	stripped, _, err := Strip(generated)
	if err != nil {
		t.Fatalf("Strip error: %v", err)
	}
	if !strings.Contains(stripped, "state=stripped") {
		t.Fatalf("stripped document missing stripped state:\n%s", stripped)
	}

	manual := strings.Replace(stripped, "## Intro", "## 7. Intro", 1)
	regenerated, _, err := Regen(manual)
	if err != nil {
		t.Fatalf("Regen error: %v", err)
	}
	if strings.Contains(regenerated, "## 7. Intro") {
		t.Fatalf("regen kept manual stripped-state edits:\n%s", regenerated)
	}
	if !strings.Contains(regenerated, "state=generated") {
		t.Fatalf("regen did not restore generated state:\n%s", regenerated)
	}
	if !strings.Contains(regenerated, "<a id=") || !strings.Contains(regenerated, "## 1.1. ") {
		t.Fatalf("regen did not rebuild managed artifacts:\n%s", regenerated)
	}

	ok, _, err := Check(regenerated)
	if err != nil {
		t.Fatalf("Check error: %v", err)
	}
	if !ok {
		t.Fatalf("Check rejected regen output after restoring generated state:\n%s", regenerated)
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

// TestGenerateIgnoresHeadingsInsideMdtocOffRegions verifies explicit mdtoc off/on exclusion blocks.
func TestGenerateIgnoresHeadingsInsideMdtocOffRegions(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"<!-- mdtoc off -->",
		"## Excluded",
		"### Also excluded",
		"<!-- mdtoc on -->",
		"",
		"## Included",
	}, "\n") + "\n"

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if strings.Contains(got, "Excluded](#excluded)") || strings.Contains(got, "Also excluded](#also-excluded)") {
		t.Fatalf("excluded headings leaked into ToC:\n%s", got)
	}
	if strings.Contains(got, "## 1. <a id=\"excluded\"></a>Excluded") || strings.Contains(got, "### 1.1. <a id=\"also-excluded\"></a>Also excluded") {
		t.Fatalf("excluded headings were rewritten:\n%s", got)
	}
	if !strings.Contains(got, "* [1. Included](#included)") || !strings.Contains(got, "## 1. <a id=\"included\"></a>Included") {
		t.Fatalf("included heading was not managed:\n%s", got)
	}
}

// TestGenerateTreatsMdtocOffWithoutOnAsExclusionToEOF verifies issue #6 EOF behavior.
func TestGenerateTreatsMdtocOffWithoutOnAsExclusionToEOF(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"## Managed",
		"",
		"<!-- mdtoc off -->",
		"## Excluded to EOF",
		"### Still excluded",
	}, "\n") + "\n"

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if strings.Contains(got, "Excluded to EOF](#excluded-to-eof)") || strings.Contains(got, "Still excluded](#still-excluded)") {
		t.Fatalf("EOF-excluded headings leaked into ToC:\n%s", got)
	}
	if strings.Contains(got, "## 2. <a id=\"excluded-to-eof\"></a>Excluded to EOF") || strings.Contains(got, "### 2.1. <a id=\"still-excluded\"></a>Still excluded") {
		t.Fatalf("EOF-excluded headings were rewritten:\n%s", got)
	}
	if !strings.Contains(got, "* [1. Managed](#managed)") || !strings.Contains(got, "## 1. <a id=\"managed\"></a>Managed") {
		t.Fatalf("managed heading before exclusion was not preserved:\n%s", got)
	}
}

// TestGenerateAutoDetectsDominantBullet verifies majority-based ToC bullet selection.
func TestGenerateAutoDetectsDominantBullet(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"- alpha",
		"- beta",
		"+ gamma",
		"",
		"## Intro",
	}, "\n") + "\n"

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, "- [1. Intro](#intro)") {
		t.Fatalf("ToC did not use dominant dash bullets:\n%s", got)
	}
	if strings.Contains(got, "* [1. Intro](#intro)") || strings.Contains(got, "+ [1. Intro](#intro)") {
		t.Fatalf("ToC used an unexpected bullet marker:\n%s", got)
	}
}

// TestGenerateAutoDetectsBulletTiePrecedence verifies tie-breaking order * > - > +.
func TestGenerateAutoDetectsBulletTiePrecedence(t *testing.T) {
	tests := []struct {
		name  string
		input []string
		want  string
	}{
		{name: "allTie", input: []string{"* a", "- b", "+ c"}, want: "* [1. Intro](#intro)"},
		{name: "dashBeatsPlus", input: []string{"- a", "+ b"}, want: "- [1. Intro](#intro)"},
	}

	for _, tc := range tests {
		lines := append([]string{"# Title", ""}, tc.input...)
		lines = append(lines, "", "## Intro")
		got, _, err := Generate(strings.Join(lines, "\n")+"\n", DefaultOptions())
		if err != nil {
			t.Fatalf("%s: Generate error: %v", tc.name, err)
		}
		if !strings.Contains(got, tc.want) {
			t.Fatalf("%s: ToC missing expected bullet choice %q:\n%s", tc.name, tc.want, got)
		}
	}
}

// TestGenerateAutoDetectsBulletsOutsideIgnoredRegions verifies bullet detection ignores fences, comments, and exclusions.
func TestGenerateAutoDetectsBulletsOutsideIgnoredRegions(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"```md",
		"- code bullet",
		"```",
		"",
		"<!--",
		"- comment bullet",
		"-->",
		"",
		"<!-- mdtoc off -->",
		"- excluded bullet",
		"<!-- mdtoc on -->",
		"",
		"+ live bullet",
		"+ another live bullet",
		"",
		"## Intro",
	}, "\n") + "\n"

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, "+ [1. Intro](#intro)") {
		t.Fatalf("ToC did not ignore non-live bullet regions:\n%s", got)
	}
}

// TestGenerateForcedBulletModeOverridesAutoDetection verifies explicit bullet selection.
func TestGenerateForcedBulletModeOverridesAutoDetection(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"- alpha",
		"- beta",
		"",
		"## Intro",
	}, "\n") + "\n"

	got, _, err := Generate(input, Options{
		Numbering: true,
		MinLevel:  2,
		MaxLevel:  4,
		Anchor:    AnchorGitHub,
		TOC:       true,
		Bullets:   BulletPlus,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, "bullets=+") || !strings.Contains(got, "+ [1. Intro](#intro)") {
		t.Fatalf("forced bullet mode was not applied:\n%s", got)
	}
}

// TestGenerateTreatsLegacyConfigWithoutBulletsAsStar verifies backward-compatible normalization of old containers.
func TestGenerateTreatsLegacyConfigWithoutBulletsAsStar(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"- [1. Wrong](#wrong)",
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
	}, "\n") + "\n"

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, "bullets=*") {
		t.Fatalf("legacy config was not normalized to bullets=*:\n%s", got)
	}
	if !strings.Contains(got, "* [1. Intro](#intro)") {
		t.Fatalf("legacy config did not preserve star bullets:\n%s", got)
	}
	if strings.Contains(got, "- [1. Intro](#intro)") || strings.Contains(got, "+ [1. Intro](#intro)") {
		t.Fatalf("legacy config unexpectedly switched to non-star bullets:\n%s", got)
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

// TestStripRawRecoversFromFutureContainerVersion verifies raw stripping for an unsupported future container format.
func TestStripRawRecoversFromFutureContainerVersion(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"+ [1. Intro](#intro)",
		configStart,
		"container-version=v3",
		"numbering=on",
		"min-level=2",
		"max-level=4",
		"anchor-profile=github",
		"toc=on",
		"bullets=auto",
		"state=generated",
		configEnd,
		endMarker,
		"",
		"## 1. <a id=\"intro\"></a>Intro",
	}, "\n") + "\n"

	got, warnings, err := StripRaw(input)
	if err != nil {
		t.Fatalf("StripRaw error: %v", err)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "fallback parsing") {
		t.Fatalf("StripRaw did not report fallback parsing: %v", warnings)
	}
	if strings.Contains(got, startMarker) || strings.Contains(got, endMarker) || strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("raw strip left future-format managed artifacts behind:\n%s", got)
	}
	if !strings.Contains(got, "## Intro") {
		t.Fatalf("raw strip removed heading text for future container:\n%s", got)
	}
}

// TestGenerateNormalizesLegacyContainerToExplicitV2 verifies that rewrites upgrade legacy containers.
func TestGenerateNormalizesLegacyContainerToExplicitV2(t *testing.T) {
	input := strings.Join([]string{
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

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, "container-version=v2") {
		t.Fatalf("legacy container was not upgraded to explicit v2:\n%s", got)
	}
	if !strings.Contains(got, "anchor=github") {
		t.Fatalf("legacy anchor config was not normalized:\n%s", got)
	}
	if !strings.Contains(got, "bullets=*") {
		t.Fatalf("legacy bullets were not normalized into explicit v2 output:\n%s", got)
	}
}

// TestStripRawRecoversFromUnknownConfigKey verifies raw stripping with a not-yet-known config key.
func TestStripRawRecoversFromUnknownConfigKey(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"* [1. Intro](#intro)",
		configStart,
		"numbering=on",
		"min-level=2",
		"max-level=4",
		"anchor-profile=github",
		"toc=on",
		"bullets=auto",
		"state=generated",
		configEnd,
		endMarker,
		"",
		"## 1. <a id=\"intro\"></a>Intro",
	}, "\n") + "\n"

	got, warnings, err := StripRaw(input)
	if err != nil {
		t.Fatalf("StripRaw error: %v", err)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "fallback parsing") {
		t.Fatalf("StripRaw did not report fallback parsing: %v", warnings)
	}
	if strings.Contains(got, startMarker) || strings.Contains(got, endMarker) || strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("raw strip left unknown-key managed artifacts behind:\n%s", got)
	}
	if !strings.Contains(got, "## Intro") {
		t.Fatalf("raw strip removed heading text for unknown-key container:\n%s", got)
	}
}

// TestStripRawRecoversFromUnterminatedConfigBlock verifies raw stripping when the config block is malformed.
func TestStripRawRecoversFromUnterminatedConfigBlock(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"* [1. Intro](#intro)",
		configStart,
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
	}, "\n") + "\n"

	got, warnings, err := StripRaw(input)
	if err != nil {
		t.Fatalf("StripRaw error: %v", err)
	}
	if !strings.Contains(strings.Join(warnings, "\n"), "fallback parsing") {
		t.Fatalf("StripRaw did not report fallback parsing: %v", warnings)
	}
	if strings.Contains(got, startMarker) || strings.Contains(got, endMarker) || strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("raw strip left malformed-container artifacts behind:\n%s", got)
	}
	if !strings.Contains(got, "## Intro") {
		t.Fatalf("raw strip removed heading text for malformed container:\n%s", got)
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
