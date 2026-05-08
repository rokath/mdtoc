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
	checks := []string{
		startMarker + "\n\n* [1. Intro](#intro)",
		"  * [1.1. API](#api)\n\n<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->",
		`## 1. <a id="intro"></a>Intro`,
		`### 1.1. <a id="api"></a>API`,
		"bullets=auto",
		"anchor=true",
		"slug=github",
	}
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
		Slug:      SlugGitHub,
		Anchor:    true,
		AnchorSet: true,
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
		Slug:      SlugGitHub,
		Anchor:    true,
		AnchorSet: true,
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
	input := strings.Join([]string{startMarker, "Some handwritten note", "<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->", endMarker, "", "## Intro"}, "\n") + "\n"
	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, preservedCommentHeader+"\nSome handwritten note\n-->") {
		t.Fatalf("foreign ToC content was not preserved:\n%s", got)
	}
	if !strings.Contains(got, startMarker+"\n"+preservedCommentHeader+"\nSome handwritten note\n-->\n\n* [1. Intro](#intro)") {
		t.Fatalf("generated ToC was not surrounded by a leading blank line:\n%s", got)
	}
	if !strings.Contains(got, "* [1. Intro](#intro)\n\n<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->") {
		t.Fatalf("generated ToC was not surrounded by a trailing blank line:\n%s", got)
	}
}

// TestRegenReusesPersistedContainerConfig verifies that regen reads stored config instead of CLI defaults.
func TestRegenReusesPersistedContainerConfig(t *testing.T) {
	input := "# Title\n\n## Intro\n\n### API\n"
	generated, _, err := Generate(input, Options{
		Numbering: false,
		MinLevel:  2,
		MaxLevel:  3,
		Anchor:    false,
		AnchorSet: true,
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

// TestRegenUsesDefaultsWhenContainerHasNoConfig verifies the optional config block contract.
func TestRegenUsesDefaultsWhenContainerHasNoConfig(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"* [Wrong](#wrong)",
		endMarker,
		"",
		"## Intro",
	}, "\n") + "\n"

	got, _, err := Regen(input)
	if err != nil {
		t.Fatalf("Regen error: %v", err)
	}
	if !strings.Contains(got, "* [1. Intro](#intro)") || !strings.Contains(got, `## 1. <a id="intro"></a>Intro`) {
		t.Fatalf("regen did not apply default config to config-less container:\n%s", got)
	}
	if strings.Contains(got, "numbering=") || strings.Contains(got, "anchor=") {
		t.Fatalf("regen should preserve absent default config block:\n%s", got)
	}
	ok, _, err := Check(got)
	if err != nil || !ok {
		t.Fatalf("Check(config-less generated) = %v, %v", ok, err)
	}
}

// TestGenerateLinkFalseRendersPlainTOC verifies link=false keeps ToC text but omits link targets.
func TestGenerateLinkFalseRendersPlainTOC(t *testing.T) {
	got, _, err := Generate("# Title\n\n## Intro\n\n### API\n", Options{
		Numbering: true,
		MinLevel:  2,
		MaxLevel:  3,
		Slug:      SlugGitHub,
		Anchor:    true,
		AnchorSet: true,
		Link:      false,
		LinkSet:   true,
		TOC:       true,
		Bullets:   BulletAuto,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, "* 1. Intro") || !strings.Contains(got, "  * 1.1. API") {
		t.Fatalf("link=false did not render plain ToC entries:\n%s", got)
	}
	if strings.Contains(got, "](#intro)") || strings.Contains(got, "](#api)") {
		t.Fatalf("link=false rendered linked ToC entries:\n%s", got)
	}
	if !strings.Contains(got, "link=false") {
		t.Fatalf("link=false was not persisted:\n%s", got)
	}
	ok, _, err := Check(got)
	if err != nil || !ok {
		t.Fatalf("Check(link=false output) = %v, %v", ok, err)
	}
	second, _, err := Generate(got, Options{
		Numbering: true,
		MinLevel:  2,
		MaxLevel:  3,
		Slug:      SlugGitHub,
		Anchor:    true,
		AnchorSet: true,
		Link:      false,
		LinkSet:   true,
		TOC:       true,
		Bullets:   BulletAuto,
	})
	if err != nil {
		t.Fatalf("second Generate error: %v", err)
	}
	if got != second {
		t.Fatalf("link=false generate is not idempotent\nfirst:\n%s\nsecond:\n%s", got, second)
	}
}

// TestGenerateAnchorOffNumberingUsesRenderedHeadingSlugForTOC verifies issue #75.
func TestGenerateAnchorOffNumberingUsesRenderedHeadingSlugForTOC(t *testing.T) {
	input := "# Title\n\n## Intro\n\n### API\n"

	got, _, err := Generate(input, Options{
		Numbering: true,
		MinLevel:  2,
		MaxLevel:  3,
		Anchor:    false,
		AnchorSet: true,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	checks := []string{
		"* [1. Intro](#1-intro)",
		"  * [1.1. API](#1-1-api)",
		"## 1. Intro",
		"### 1.1. API",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("generated output missing %q:\n%s", check, got)
		}
	}
	if strings.Contains(got, "<a id=") {
		t.Fatalf("anchor=off unexpectedly rendered inline anchors:\n%s", got)
	}
}

// TestGenerateAnchorOffCrossnoteClosingATXSpaces verifies the MPE/Crossnote
// slug behavior for ATX closing markers preceded by multiple spaces.
func TestGenerateAnchorOffCrossnoteClosingATXSpaces(t *testing.T) {
	got, _, err := Generate("# Title\n\n## An ATX title with closing hash markers  ####\n", Options{
		Numbering: false,
		MinLevel:  2,
		MaxLevel:  2,
		Slug:      SlugCrossnote,
		Anchor:    false,
		AnchorSet: true,
		Link:      true,
		LinkSet:   true,
		TOC:       true,
		Bullets:   BulletAuto,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	want := "* [An ATX title with closing hash markers](#an-atx-title-with-closing-hash-markers--)"
	if !strings.Contains(got, want) {
		t.Fatalf("generated ToC missing %q:\n%s", want, got)
	}
	if strings.Contains(got, "<a id=") {
		t.Fatalf("anchor=false rendered inline anchors:\n%s", got)
	}
}

// TestGenerateAnchorOffNumberingPreservesTitleNumberBoundary verifies that title
// numbers do not get merged into the numbering prefix for renderer-derived IDs.
func TestGenerateAnchorOffNumberingPreservesTitleNumberBoundary(t *testing.T) {
	input := "# Title\n\n## Intro\n\n### 2025 Roadmap\n"

	got, _, err := Generate(input, Options{
		Numbering: true,
		MinLevel:  2,
		MaxLevel:  3,
		Anchor:    false,
		AnchorSet: true,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	checks := []string{
		"* [1. Intro](#1-intro)",
		"  * [1.1. 2025 Roadmap](#1-1-2025-roadmap)",
		"### 1.1. 2025 Roadmap",
	}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("generated output missing %q:\n%s", check, got)
		}
	}
	if strings.Contains(got, "  * [1.1. 2025 Roadmap](#112025-roadmap)") {
		t.Fatalf("renderer-derived ToC target merged numbering with title digits:\n%s", got)
	}
}

// TestRegenRequiresManagedConfig verifies that regen fails without a valid managed container.
func TestRegenRequiresManagedConfig(t *testing.T) {
	if _, _, err := Regen("# Title\n\n## Intro\n"); err == nil {
		t.Fatalf("Regen without config unexpectedly succeeded")
	}
}

// TestRegenRestoresGeneratedStateFromStrippedInput verifies regen after a stripped-state workflow.
func TestRegenRestoresGeneratedOutputFromStrippedInput(t *testing.T) {
	generated, _, err := Generate("# Title\n\n## Intro\n", Options{
		Numbering: true,
		MinLevel:  1,
		MaxLevel:  4,
		Slug:      SlugGitHub,
		Anchor:    true,
		AnchorSet: true,
		TOC:       true,
	})
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}

	stripped, _, err := Strip(generated)
	if err != nil {
		t.Fatalf("Strip error: %v", err)
	}
	if strings.Contains(stripped, "<a id=") || strings.Contains(stripped, "## 1.1. ") {
		t.Fatalf("strip kept managed artifacts:\n%s", stripped)
	}

	manual := strings.Replace(stripped, "## Intro", "## 7. Intro", 1)
	regenerated, _, err := Regen(manual)
	if err != nil {
		t.Fatalf("Regen error: %v", err)
	}
	if strings.Contains(regenerated, "## 7. Intro") {
		t.Fatalf("regen kept manual stripped-state edits:\n%s", regenerated)
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

// TestGenerateIgnoresHeadingsInsideNestedFenceContent verifies issue #77:
// a shorter inner fence marker inside a longer fenced code block must not end
// the outer ignored region early.
func TestGenerateIgnoresHeadingsInsideNestedFenceContent(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"````md",
		"```",
		"## Still code",
		"```",
		"````",
		"",
		"## Real heading",
	}, "\n") + "\n"

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if strings.Contains(got, "Still code](#still-code)") {
		t.Fatalf("nested fence content leaked into ToC:\n%s", got)
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

// TestGenerateAutoDetectsBulletsOutsideNestedFenceContent verifies that a
// shorter inner fence inside a longer outer fence does not re-enable bullet
// detection too early.
func TestGenerateAutoDetectsBulletsOutsideNestedFenceContent(t *testing.T) {
	input := strings.Join([]string{
		"# Title",
		"",
		"````md",
		"```",
		"- not live",
		"```",
		"````",
		"",
		"+ live bullet",
		"",
		"## Real heading",
	}, "\n") + "\n"

	got, _, err := Generate(input, DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	if !strings.Contains(got, "+ [1. Real heading](#real-heading)") {
		t.Fatalf("ToC bullet auto-detection counted nested-fence bullets:\n%s", got)
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
		Slug:      SlugGitHub,
		Anchor:    true,
		AnchorSet: true,
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

// TestRegenUsesStoredBulletMode verifies stored compact config controls regen.
func TestRegenUsesStoredBulletMode(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"- [1. Wrong](#wrong)",
		"<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=* -->",
		endMarker,
		"",
		"- local bullet",
		"- local bullet",
		"",
		"## Intro",
	}, "\n") + "\n"

	got, _, err := Regen(input)
	if err != nil {
		t.Fatalf("Regen error: %v", err)
	}
	if !strings.Contains(got, "bullets=*") {
		t.Fatalf("stored bullet mode was not preserved:\n%s", got)
	}
	if !strings.Contains(got, "* [1. Intro](#intro)") {
		t.Fatalf("stored bullet mode did not render star bullets:\n%s", got)
	}
	if strings.Contains(got, "- [1. Intro](#intro)") || strings.Contains(got, "+ [1. Intro](#intro)") {
		t.Fatalf("stored bullet mode unexpectedly switched to non-star bullets:\n%s", got)
	}
}

// TestStripKeepsContainerAndConfig verifies strip removes artifacts while retaining the container.
func TestStripKeepsContainerAndConfig(t *testing.T) {
	generated, _, err := Generate("# Title\n\n## Intro\n", DefaultOptions())
	if err != nil {
		t.Fatalf("Generate error: %v", err)
	}
	got, _, err := Strip(generated)
	if err != nil {
		t.Fatalf("Strip error: %v", err)
	}
	checks := []string{startMarker, endMarker, "anchor=true", "## Intro"}
	for _, check := range checks {
		if !strings.Contains(got, check) {
			t.Fatalf("stripped output missing %q:\n%s", check, got)
		}
	}
	if strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("managed artifacts were not removed:\n%s", got)
	}
}

// TestStripPreservesForeignTOCContent verifies that strip keeps authored
// content in the managed ToC area instead of dropping it together with
// generated ToC lines.
func TestStripPreservesForeignTOCContent(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"* [1. Intro](#intro)",
		"Keep this manual note",
		"<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto -->",
		endMarker,
		"",
		"# Title",
		"",
		"## 1. <a id=\"intro\"></a>Intro",
	}, "\n") + "\n"

	got, _, err := Strip(input)
	if err != nil {
		t.Fatalf("Strip error: %v", err)
	}
	if !strings.Contains(got, preservedCommentHeader+"\nKeep this manual note\n-->") {
		t.Fatalf("strip did not preserve foreign ToC content:\n%s", got)
	}
	if strings.Contains(got, "* [1. Intro](#intro)") {
		t.Fatalf("strip kept generated ToC content:\n%s", got)
	}
	if strings.Contains(got, "<a id=") || strings.Contains(got, "## 1. ") {
		t.Fatalf("strip kept managed heading artifacts:\n%s", got)
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
		"<!-- mdtoc-config",
		"container-version=v3",
		"numbering=true",
		"min-level=2",
		"max-level=4",
		"anchor-profile=github",
		"toc=true",
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

// TestGenerateRejectsLegacyConfig verifies old mdtoc-config blocks are no longer supported.
func TestGenerateRejectsLegacyConfig(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"* [1. Intro](#intro)",
		"<!-- mdtoc-config",
		"numbering=true",
		"min-level=2",
		"max-level=4",
		"anchors=on",
		"toc=true",
		"state=generated",
		configEnd,
		endMarker,
		"",
		"## 1. <a id=\"intro\"></a>Intro",
	}, "\n") + "\n"

	if _, _, err := Generate(input, DefaultOptions()); err == nil {
		t.Fatalf("Generate unexpectedly accepted legacy mdtoc-config")
	}
}

// TestStripRawRecoversFromUnknownConfigKey verifies raw stripping with a not-yet-known config key.
func TestStripRawRecoversFromUnknownConfigKey(t *testing.T) {
	input := strings.Join([]string{
		startMarker,
		"* [1. Intro](#intro)",
		"<!-- mdtoc-config",
		"numbering=true",
		"min-level=2",
		"max-level=4",
		"anchor-profile=github",
		"toc=true",
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
		"<!-- numbering=true min=2 max=4 slug=github anchor=true link=true toc=true bullets=auto",
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
