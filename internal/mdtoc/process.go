package mdtoc

import (
	"fmt"
	"strings"
)

// Generate applies the managed transformation from the specification.
func Generate(input string, opts Options) (string, []string, error) {
	return generateWithConfig(input, opts.ToConfig())
}

// Regen rebuilds the generated state from the persisted config of an existing
// managed container.
func Regen(input string) (string, []string, error) {
	parsed, err := ParseDocument(input)
	if err != nil {
		return "", nil, err
	}
	if parsed.Container == nil {
		return "", nil, fmt.Errorf("regen requires a valid mdtoc config block")
	}
	cfg := parsed.Container.Config
	switch cfg.State {
	case StateGenerated, StateStripped:
		cfg.State = StateGenerated
		return generateWithConfig(input, cfg)
	default:
		return "", parsed.Warnings, fmt.Errorf("unsupported state %q", parsed.Container.Config.State)
	}
}

// generateWithConfig renders the managed document state for a fully validated config.
func generateWithConfig(input string, cfg Config) (string, []string, error) {
	parsed, err := ParseDocument(input)
	if err != nil {
		return "", nil, err
	}
	if parsed.Container != nil && !parsed.Container.Config.BulletsExplicit && cfg.Bullets == BulletAuto {
		cfg.Bullets = BulletStar
	}
	if err := cfg.Validate(); err != nil {
		return "", nil, err
	}
	bodyLines, headings := removeContainerAndNormalizeHeadings(parsed)
	assignDerivedArtifacts(headings, cfg)
	var tocLines []string
	if cfg.TOC {
		tocLines = renderTOC(headings, bodyLines, cfg)
	}
	containerLines := renderContainer(cfg, preserveForeignTOC(parsed.Container), tocLines)
	bodyLines = rewriteHeadings(bodyLines, headings, cfg)
	return joinLines(placeContainer(bodyLines, containerLines, parsed.Container)), parsed.Warnings, nil
}

// Strip removes managed artifacts but keeps the outer container and config.
func Strip(input string) (string, []string, error) {
	parsed, err := ParseDocument(input)
	if err != nil {
		return "", nil, err
	}
	if parsed.Container == nil {
		return "", nil, fmt.Errorf("strip requires a valid mdtoc config block")
	}
	cfg := parsed.Container.Config
	cfg.State = StateStripped
	bodyLines, headings := removeContainerAndNormalizeHeadings(parsed)
	bodyLines = rewriteHeadings(bodyLines, headings, Config{MinLevel: 1, MaxLevel: 0})
	containerLines := renderContainer(cfg, nil, nil)
	return joinLines(placeContainer(bodyLines, containerLines, parsed.Container)), parsed.Warnings, nil
}

// StripRaw removes the entire container, managed numbering, and managed anchors.
func StripRaw(input string) (string, []string, error) {
	parsed, err := ParseDocument(input)
	if err != nil {
		return "", nil, err
	}
	bodyLines, headings := removeContainerAndNormalizeHeadings(parsed)
	bodyLines = rewriteHeadings(bodyLines, headings, Config{MinLevel: 1, MaxLevel: 0})
	return joinLines(bodyLines), parsed.Warnings, nil
}

// Check reconstructs the target state indicated by the stored config and
// compares it byte-for-byte against the current document.
func Check(input string) (bool, []string, error) {
	parsed, err := ParseDocument(input)
	if err != nil {
		return false, nil, err
	}
	if parsed.Container == nil {
		return false, nil, fmt.Errorf("check requires a valid mdtoc config block")
	}
	var expected string
	switch parsed.Container.Config.State {
	case StateGenerated:
		expected, _, err = Generate(input, Options{Numbering: parsed.Container.Config.Numbering, MinLevel: parsed.Container.Config.MinLevel, MaxLevel: parsed.Container.Config.MaxLevel, Anchors: parsed.Container.Config.Anchors, TOC: parsed.Container.Config.TOC, Bullets: parsed.Container.Config.Bullets})
	case StateStripped:
		expected, _, err = Strip(input)
	default:
		err = fmt.Errorf("unsupported state %q", parsed.Container.Config.State)
	}
	if err != nil {
		return false, parsed.Warnings, err
	}
	return expected == normalizeInput(input), parsed.Warnings, nil
}

// normalizeInput canonicalizes line endings and ensures a trailing newline.
func normalizeInput(input string) string {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	if input == "" {
		return ""
	}
	if !strings.HasSuffix(input, "\n") {
		input += "\n"
	}
	return input
}

// removeContainerAndNormalizeHeadings returns body lines without the managed container.
func removeContainerAndNormalizeHeadings(parsed *ParsedDocument) ([]string, []Heading) {
	lines := append([]string(nil), parsed.Lines...)
	if parsed.Container != nil {
		lines = append(append([]string{}, lines[:parsed.Container.StartLine]...), lines[parsed.Container.EndLine+1:]...)
	}
	headings := make([]Heading, len(parsed.Headings))
	for i, h := range parsed.Headings {
		headings[i] = h
	}
	if parsed.Container != nil {
		removed := parsed.Container.EndLine - parsed.Container.StartLine + 1
		for i := range headings {
			if headings[i].LineIndex > parsed.Container.EndLine {
				headings[i].LineIndex -= removed
			}
		}
	}
	return lines, headings
}

// assignDerivedArtifacts computes managed numbers and anchors for eligible headings.
func assignDerivedArtifacts(headings []Heading, cfg Config) {
	slugger := NewSlugger()
	counters := make([]int, 7)
	for i := range headings {
		h := &headings[i]
		if !h.InManagedRange(cfg) {
			h.ManagedNumber, h.ManagedAnchor = "", ""
			continue
		}
		anchorID := slugger.Next(h.TitleText)
		h.ManagedAnchor = fmt.Sprintf(`<a id="%s"></a>`, anchorID)
		if !cfg.Numbering {
			h.ManagedNumber = ""
			continue
		}
		counters[h.Level]++
		for j := h.Level + 1; j <= 6; j++ {
			counters[j] = 0
		}
		parts := []string{}
		for j := cfg.MinLevel; j <= h.Level; j++ {
			if counters[j] == 0 {
				continue
			}
			parts = append(parts, fmt.Sprintf("%d", counters[j]))
		}
		h.ManagedNumber = strings.Join(parts, ".") + "."
	}
}

// rewriteHeadings renders the managed heading state back into the document lines.
func rewriteHeadings(lines []string, headings []Heading, cfg Config) []string {
	out := append([]string(nil), lines...)
	for _, h := range headings {
		if h.LineIndex < 0 || h.LineIndex >= len(out) {
			continue
		}
		line := h.Hashes + " "
		if h.InManagedRange(cfg) && h.ManagedNumber != "" {
			line += h.ManagedNumber + " "
		}
		if h.InManagedRange(cfg) && cfg.Anchors && h.ManagedAnchor != "" {
			line += h.ManagedAnchor
		}
		line += h.TitleMarkup
		out[h.LineIndex] = line
	}
	return out
}

// renderTOC converts managed headings into Markdown list entries.
func renderTOC(headings []Heading, bodyLines []string, cfg Config) []string {
	lines := []string{}
	bullet := resolveTOCBullet(bodyLines, cfg)
	for _, h := range headings {
		if !h.InManagedRange(cfg) {
			continue
		}
		anchorID := strings.TrimSuffix(strings.TrimPrefix(h.ManagedAnchor, `<a id="`), `"></a>`)
		text := h.TitleText
		if cfg.Numbering && h.ManagedNumber != "" {
			text = h.ManagedNumber + " " + text
		}
		lines = append(lines, fmt.Sprintf("%s%s [%s](#%s)", strings.Repeat("  ", h.Level-cfg.MinLevel), bullet, text, anchorID))
	}
	return lines
}

// resolveTOCBullet returns the configured or auto-detected unordered-list marker.
func resolveTOCBullet(bodyLines []string, cfg Config) BulletMode {
	if cfg.Bullets != BulletAuto {
		return cfg.Bullets
	}
	return detectDominantBullet(bodyLines)
}

// detectDominantBullet counts list markers outside ignored regions and applies the tie-break order * > - > +.
func detectDominantBullet(lines []string) BulletMode {
	counts := map[BulletMode]int{
		BulletStar: 0,
		BulletDash: 0,
		BulletPlus: 0,
	}
	inFence := false
	fenceMarker := ""
	inGenericComment := false
	inExcludedRegion := false
	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if inExcludedRegion {
			if trimmed == onMarker {
				inExcludedRegion = false
			}
			continue
		}
		if inFence {
			if isFenceClose(trimmed, fenceMarker) {
				inFence, fenceMarker = false, ""
			}
			continue
		}
		if inGenericComment {
			if strings.Contains(line, "-->") {
				inGenericComment = false
			}
			continue
		}
		if marker := fenceOpen(trimmed); marker != "" {
			inFence, fenceMarker = true, marker
			continue
		}
		if trimmed == offMarker {
			inExcludedRegion = true
			continue
		}
		if startsGenericHTMLComment(trimmed) {
			if !strings.Contains(trimmed, "-->") {
				inGenericComment = true
			}
			continue
		}
		if bullet, ok := detectListBullet(line); ok {
			counts[bullet]++
		}
	}

	best := BulletStar
	for _, candidate := range []BulletMode{BulletDash, BulletPlus} {
		if counts[candidate] > counts[best] {
			best = candidate
		}
	}
	return best
}

// detectListBullet reports a supported unordered-list bullet at the logical start of a line.
func detectListBullet(line string) (BulletMode, bool) {
	trimmedLeft := strings.TrimLeft(line, " \t")
	if len(trimmedLeft) < 2 || trimmedLeft[1] != ' ' {
		return "", false
	}
	switch BulletMode(trimmedLeft[:1]) {
	case BulletStar, BulletDash, BulletPlus:
		return BulletMode(trimmedLeft[:1]), true
	default:
		return "", false
	}
}

// renderContainer renders the normalized managed container for the current config.
func renderContainer(cfg Config, preserved, toc []string) []string {
	lines := []string{startMarker}
	if len(preserved) > 0 {
		lines = append(lines, preserved...)
	}
	if len(toc) > 0 {
		if len(preserved) > 0 {
			lines = append(lines, "")
		}
		lines = append(lines, toc...)
	}
	lines = append(lines, RenderConfig(cfg)...)
	lines = append(lines, endMarker)
	return lines
}

// preserveForeignTOC keeps handwritten content from the managed ToC area as comments.
func preserveForeignTOC(container *Container) []string {
	if container == nil || len(container.TOCArea) == 0 {
		return nil
	}
	var preserved []string
	for i := 0; i < len(container.TOCArea); {
		line := container.TOCArea[i]
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
			i++
		case trimmed == preservedCommentHeader:
			j := i + 1
			for ; j < len(container.TOCArea); j++ {
				if strings.TrimSpace(container.TOCArea[j]) == "-->" {
					break
				}
			}
			if j < len(container.TOCArea) {
				preserved = append(preserved, container.TOCArea[i:j+1]...)
				i = j + 1
			} else {
				preserved = append(preserved, wrapPreservedComment(container.TOCArea[i:])...)
				return preserved
			}
		case isGeneratedTOCLine(line):
			i++
		default:
			j := i
			var chunk []string
			for ; j < len(container.TOCArea); j++ {
				if isGeneratedTOCLine(container.TOCArea[j]) {
					break
				}
				chunk = append(chunk, container.TOCArea[j])
			}
			preserved = append(preserved, wrapPreservedComment(chunk)...)
			i = j
		}
	}
	return trimBlankEdges(preserved)
}

// isGeneratedTOCLine reports whether the line matches the generated ToC shape.
func isGeneratedTOCLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	return strings.HasPrefix(trimmed, "* [") && strings.Contains(trimmed, "](#") && strings.HasSuffix(trimmed, ")")
}

// wrapPreservedComment wraps preserved foreign lines in a generated HTML comment.
func wrapPreservedComment(lines []string) []string {
	chunk := trimBlankEdges(lines)
	if len(chunk) == 0 {
		return nil
	}
	out := []string{preservedCommentHeader}
	out = append(out, chunk...)
	out = append(out, "-->")
	return out
}

// trimBlankEdges removes leading and trailing blank lines from a slice.
func trimBlankEdges(lines []string) []string {
	start := 0
	for start < len(lines) && strings.TrimSpace(lines[start]) == "" {
		start++
	}
	end := len(lines)
	for end > start && strings.TrimSpace(lines[end-1]) == "" {
		end--
	}
	return append([]string(nil), lines[start:end]...)
}

// prependContainer places a new container at the top of the document body.
func prependContainer(body, container []string) []string {
	if len(body) == 0 {
		return container
	}
	out := append([]string{}, container...)
	if strings.TrimSpace(body[0]) != "" {
		out = append(out, "")
	}
	out = append(out, body...)
	return out
}

// placeContainer reinserts the managed container at its existing document position.
func placeContainer(body, container []string, existing *Container) []string {
	if existing == nil {
		return prependContainer(body, container)
	}
	insertAt := existing.StartLine
	if insertAt < 0 {
		insertAt = 0
	}
	if insertAt > len(body) {
		insertAt = len(body)
	}
	out := append([]string{}, body[:insertAt]...)
	out = append(out, container...)
	out = append(out, body[insertAt:]...)
	return out
}

// joinLines joins logical lines into normalized document text.
func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
