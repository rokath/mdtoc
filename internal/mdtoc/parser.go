package mdtoc

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
)

var (
	headingStartRE  = regexp.MustCompile(`^(#{1,6}) `)
	managedNumberRE = regexp.MustCompile(`^(\d+(?:\.\d+)*)\. `)
	managedAnchorRE = regexp.MustCompile(`^<a id="([^"]+)"></a>`)
	closingATXRE    = regexp.MustCompile(`[ \t]+#+[ \t]*$`)
)

// ParseDocument performs the line-oriented structural parse required by the
// specification.
func ParseDocument(input string) (*ParsedDocument, error) {
	input = strings.ReplaceAll(input, "\r\n", "\n")
	parsed := &ParsedDocument{TrailingLF: strings.HasSuffix(input, "\n")}
	parsed.Lines = splitLines(input)

	startLine, endLine := -1, -1

	inFence := false
	fenceMarker := ""
	inGenericComment := false

	for i := 0; i < len(parsed.Lines); i++ {
		line := parsed.Lines[i]
		trimmed := strings.TrimSpace(line)
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
		switch trimmed {
		case startMarker:
			if startLine != -1 {
				return nil, fmt.Errorf("duplicate %s marker", startMarker)
			}
			startLine = i
			continue
		case endMarker:
			if endLine != -1 {
				return nil, fmt.Errorf("duplicate %s marker", endMarker)
			}
			endLine = i
			continue
		}
		if startsGenericHTMLComment(trimmed) {
			if !strings.Contains(trimmed, "-->") {
				inGenericComment = true
			}
		}
	}

	container, err := buildContainer(parsed.Lines, startLine, endLine)
	if err != nil {
		return nil, err
	}
	parsed.Container = container
	parsed.Headings, parsed.Warnings, err = parseHeadings(parsed.Lines)
	if err != nil {
		return nil, err
	}
	return parsed, nil
}

// splitLines normalizes an input string into logical lines without the final LF.
func splitLines(input string) []string {
	if input == "" {
		return []string{}
	}
	input = strings.TrimSuffix(input, "\n")
	if input == "" {
		return []string{}
	}
	return strings.Split(input, "\n")
}

// startsGenericHTMLComment reports whether a line opens a non-mdtoc HTML comment.
func startsGenericHTMLComment(trimmed string) bool {
	if !strings.HasPrefix(trimmed, "<!--") {
		return false
	}
	return trimmed != startMarker && trimmed != endMarker && trimmed != offMarker && trimmed != onMarker
}

// fenceOpen reports the full supported fence run that starts on the given line.
func fenceOpen(trimmed string) string {
	return leadingFenceRun(trimmed)
}

// isFenceClose reports whether the line closes the active fence type.
func isFenceClose(trimmed, marker string) bool {
	if marker == "" {
		return false
	}
	run := leadingFenceRun(trimmed)
	if len(run) < len(marker) || run[0] != marker[0] {
		return false
	}
	return strings.TrimSpace(trimmed[len(run):]) == ""
}

func leadingFenceRun(trimmed string) string {
	if trimmed == "" {
		return ""
	}
	switch trimmed[0] {
	case '`', '~':
	default:
		return ""
	}
	i := 0
	for i < len(trimmed) && trimmed[i] == trimmed[0] {
		i++
	}
	if i < 3 {
		return ""
	}
	return trimmed[:i]
}

// buildContainer validates and materializes the parsed managed container metadata.
func buildContainer(lines []string, startLine, endLine int) (*Container, error) {
	if startLine == -1 && endLine == -1 {
		return nil, nil
	}
	if startLine == -1 || endLine == -1 {
		return nil, errors.New("incomplete mdtoc container")
	}
	if startLine > endLine {
		return nil, errors.New("mdtoc start marker appears after end marker")
	}
	cfg := DefaultConfig()
	configStartLine, configEndLine := -1, -1
	configPresent, configMultiline := false, false
	tocEnd := endLine
	if startLine+1 < endLine {
		start, end, multiline, ok, err := findTrailingConfigBlock(lines, startLine+1, endLine-1)
		if err != nil {
			return nil, err
		}
		if ok {
			cfg, err = parseConfig(lines[start : end+1])
			if err != nil {
				return nil, err
			}
			configStartLine, configEndLine = start, end
			configPresent, configMultiline = true, multiline
			tocEnd = start
		}
	}
	return &Container{
		StartLine:       startLine,
		ConfigStartLine: configStartLine,
		ConfigEndLine:   configEndLine,
		EndLine:         endLine,
		Config:          cfg,
		ConfigPresent:   configPresent,
		ConfigMultiline: configMultiline,
		TOCArea:         append([]string(nil), lines[startLine+1:tocEnd]...),
	}, nil
}

func findTrailingConfigBlock(lines []string, first, last int) (int, int, bool, bool, error) {
	if last < first {
		return 0, 0, false, false, nil
	}
	trimmedLast := strings.TrimSpace(lines[last])
	if strings.HasPrefix(trimmedLast, "<!--") && strings.HasSuffix(trimmedLast, "-->") {
		content, err := configCommentContent([]string{lines[last]})
		if err != nil {
			return 0, 0, false, false, err
		}
		if configContentLooksLikeConfig(content) {
			return last, last, false, true, nil
		}
		return 0, 0, false, false, nil
	}
	if !strings.HasSuffix(trimmedLast, "-->") {
		return 0, 0, false, false, nil
	}
	for i := last - 1; i >= first; i-- {
		if strings.HasPrefix(strings.TrimSpace(lines[i]), "<!--") {
			content, err := configCommentContent(lines[i : last+1])
			if err != nil {
				return 0, 0, false, false, err
			}
			if configContentLooksLikeConfig(content) {
				return i, last, true, true, nil
			}
			return 0, 0, false, false, nil
		}
	}
	return 0, 0, false, false, fmt.Errorf("unterminated trailing config block")
}

// parseHeadings scans the document for managed heading candidates and warnings.
func parseHeadings(lines []string) ([]Heading, []string, error) {
	headings := []Heading{}
	warnings := []string{}
	inFence := false
	fenceMarker := ""
	inGenericComment := false
	inExcludedRegion := false
	for i, line := range lines {
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
		h, warning, ok, err := parseHeadingLine(line, i)
		if err != nil {
			return nil, nil, err
		}
		if warning != "" {
			warnings = append(warnings, warning)
		}
		if ok {
			headings = append(headings, h)
		}
	}
	return headings, warnings, nil
}

// parseHeadingLine parses one heading line into its managed and semantic parts.
func parseHeadingLine(line string, lineIndex int) (Heading, string, bool, error) {
	m := headingStartRE.FindStringSubmatch(line)
	if m == nil {
		return Heading{}, "", false, nil
	}
	hashes := m[1]
	rest := line[len(hashes)+1:]
	h := Heading{LineIndex: lineIndex, Level: len(hashes), Hashes: hashes}
	if nm := managedNumberRE.FindStringSubmatch(rest); nm != nil {
		h.ManagedNumber = nm[1] + "."
		rest = rest[len(nm[0]):]
	}
	if am := managedAnchorRE.FindStringSubmatch(rest); am != nil {
		h.ManagedAnchor = am[0]
		rest = rest[len(am[0]):]
	} else if strings.HasPrefix(rest, "<a id=") {
		warning := fmt.Sprintf("warning: heading line %d contains a non-managed inline anchor; raw stripping will leave it unchanged", lineIndex+1)
		h.TitleMarkup = rest
		text, err := ExtractPlainText(stripClosingATX(rest))
		if err != nil {
			return Heading{}, "", false, err
		}
		h.TitleText = text
		return h, warning, true, nil
	}
	if rest == "" {
		return Heading{}, "", false, nil
	}
	h.TitleMarkup = rest
	text, err := ExtractPlainText(stripClosingATX(rest))
	if err != nil {
		return Heading{}, "", false, err
	}
	h.TitleText = text
	return h, "", true, nil
}

func stripClosingATX(title string) string {
	return closingATXRE.ReplaceAllString(title, "")
}
