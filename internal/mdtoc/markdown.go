package mdtoc

import "strings"

// ExtractPlainText derives the visible heading text from inline title markup.
// The spec allows an internal parser choice as long as the external behavior is
// deterministic. This implementation stays self-contained and supports the
// subset used by the tests and the examples in the specification.
func ExtractPlainText(titleMarkup string) (string, error) {
	return strings.TrimSpace(extractInlineText(titleMarkup)), nil
}

// extractInlineText removes supported inline Markdown markup while keeping text.
func extractInlineText(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		switch {
		case strings.HasPrefix(s[i:], "`"):
			end := strings.Index(s[i+1:], "`")
			if end >= 0 {
				b.WriteString(s[i+1 : i+1+end])
				i += end + 2
				continue
			}
		case strings.HasPrefix(s[i:], "!["):
			if alt, consumed, ok := parseBracketLinkLike(s[i+1:]); ok {
				b.WriteString(alt)
				i += consumed + 1
				continue
			}
		case strings.HasPrefix(s[i:], "["):
			if label, consumed, ok := parseBracketLinkLike(s[i:]); ok {
				b.WriteString(label)
				i += consumed
				continue
			}
		case strings.HasPrefix(s[i:], "<"):
			if end := strings.IndexByte(s[i:], '>'); end >= 0 {
				i += end + 1
				continue
			}
		}
		if isFormattingMarker(rune(s[i])) {
			i++
			continue
		}
		b.WriteByte(s[i])
		i++
	}
	return collapseWhitespace(b.String())
}

// parseBracketLinkLike parses a Markdown link or image-like construct at s[0].
func parseBracketLinkLike(s string) (label string, consumed int, ok bool) {
	if len(s) == 0 || s[0] != '[' {
		return "", 0, false
	}
	closeLabel := strings.IndexByte(s, ']')
	if closeLabel < 0 || closeLabel+1 >= len(s) || s[closeLabel+1] != '(' {
		return "", 0, false
	}
	closeTarget := strings.IndexByte(s[closeLabel+2:], ')')
	if closeTarget < 0 {
		return "", 0, false
	}
	label = extractInlineText(s[1:closeLabel])
	consumed = closeLabel + 2 + closeTarget + 1
	return label, consumed, true
}

// isFormattingMarker reports whether the rune is stripped as inline formatting.
func isFormattingMarker(r rune) bool {
	switch r {
	case '*', '_', '~':
		return true
	default:
		return false
	}
}

// collapseWhitespace folds runs of whitespace into single spaces.
func collapseWhitespace(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
