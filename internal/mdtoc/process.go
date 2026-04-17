package mdtoc

import (
	"fmt"
	"strings"
)

// Generate applies the managed transformation from the specification.
func Generate(input string, opts Options) (string, []string, error) {
	cfg := opts.ToConfig()
	if err := cfg.Validate(); err != nil {
		return "", nil, err
	}
	parsed, err := ParseDocument(input)
	if err != nil {
		return "", nil, err
	}
	bodyLines, headings := removeContainerAndNormalizeHeadings(parsed)
	assignDerivedArtifacts(headings, cfg)
	var tocLines []string
	if cfg.TOC {
		tocLines = renderTOC(headings, cfg)
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
		expected, _, err = Generate(input, Options{Numbering: parsed.Container.Config.Numbering, MinLevel: parsed.Container.Config.MinLevel, MaxLevel: parsed.Container.Config.MaxLevel, Anchors: parsed.Container.Config.Anchors, TOC: parsed.Container.Config.TOC})
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

func renderTOC(headings []Heading, cfg Config) []string {
	lines := []string{}
	for _, h := range headings {
		if !h.InManagedRange(cfg) {
			continue
		}
		anchorID := strings.TrimSuffix(strings.TrimPrefix(h.ManagedAnchor, `<a id="`), `"></a>`)
		text := h.TitleText
		if cfg.Numbering && h.ManagedNumber != "" {
			text = h.ManagedNumber + " " + text
		}
		lines = append(lines, fmt.Sprintf("%s* [%s](#%s)", strings.Repeat("  ", h.Level-cfg.MinLevel), text, anchorID))
	}
	return lines
}

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

func isGeneratedTOCLine(line string) bool {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return true
	}
	return strings.HasPrefix(trimmed, "* [") && strings.Contains(trimmed, "](#") && strings.HasSuffix(trimmed, ")")
}

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

func joinLines(lines []string) string {
	if len(lines) == 0 {
		return ""
	}
	return strings.Join(lines, "\n") + "\n"
}
