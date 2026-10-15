package mdtoc

import (
	"html"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ExtractPlainText derives the visible heading text from inline title markup.
// The spec allows an internal parser choice as long as the external behavior is
// deterministic. This implementation stays self-contained and supports the
// subset used by the tests and the examples in the specification.
func ExtractPlainText(titleMarkup string) (string, error) {
	return strings.TrimSpace(extractInlineText(titleMarkup, true)), nil
}

// ExtractPlainTextPreservingWhitespace derives heading text from inline title
// markup while keeping literal whitespace runs intact for slug generation.
func ExtractPlainTextPreservingWhitespace(titleMarkup string) (string, error) {
	return strings.TrimSpace(extractInlineText(titleMarkup, false)), nil
}

// extractInlineText removes supported inline Markdown markup while keeping text.
// When collapseWhitespace is true, the result is normalized for human-readable
// heading text. When false, literal whitespace runs are preserved so slug
// generation can distinguish repeated spaces from single spaces.
func extractInlineText(s string, collapseWhitespace bool) string {
	text := transformInlineMarkdown(s, false)
	if collapseWhitespace {
		return collapseWhitespaceRuns(text)
	}
	return text
}

// renderTOCLabel keeps authored inline formatting instead of reusing slug text.
// Links and images contribute their labels to avoid nested links, HTML tags are
// omitted to avoid copying anchors, and literal brackets are escaped outside code.
func renderTOCLabel(titleMarkup string) string {
	return transformInlineMarkdown(strings.TrimSpace(stripClosingATX(titleMarkup)), true)
}

// transformInlineMarkdown shares structural parsing between semantic slug text
// and Markdown labels. Escapes and entities are decoded only for semantic text;
// code content is always literal and is never parsed again as Markdown.
func transformInlineMarkdown(s string, preserveMarkup bool) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		switch {
		case s[i] == '\\' && i+1 < len(s) && isMarkdownEscapable(s[i+1]):
			if preserveMarkup {
				b.WriteByte('\\')
			}
			b.WriteByte(s[i+1])
			i += 2
			continue
		case strings.HasPrefix(s[i:], "`"):
			if content, consumed, ok := parseCodeSpan(s[i:]); ok {
				if preserveMarkup {
					b.WriteString(s[i : i+consumed])
				} else {
					b.WriteString(content)
				}
				i += consumed
				continue
			}
			// An unmatched backtick run is literal; do not reinterpret its tail
			// as a shorter opening delimiter on the next iteration.
			run := delimiterRun(s[i:], '`')
			b.WriteString(s[i : i+run])
			i += run
			continue
		case strings.HasPrefix(s[i:], "!["):
			if alt, consumed, ok := parseInlineLink(s[i+1:]); ok {
				b.WriteString(transformInlineMarkdown(alt, preserveMarkup))
				i += consumed + 1
				continue
			}
		case strings.HasPrefix(s[i:], "["):
			if label, consumed, ok := parseInlineLink(s[i:]); ok {
				b.WriteString(transformInlineMarkdown(label, preserveMarkup))
				i += consumed
				continue
			}
		case strings.HasPrefix(s[i:], "<"):
			if end := strings.IndexByte(s[i:], '>'); end >= 0 {
				i += end + 1
				continue
			}
		case !preserveMarkup && s[i] == '&':
			if decoded, consumed, ok := parseInlineEntity(s[i:]); ok {
				b.WriteString(decoded)
				i += consumed
				continue
			}
		}
		if !preserveMarkup && isFormattingMarker(rune(s[i])) {
			if content, consumed, ok := parseFormattingSpan(s, i); ok {
				b.WriteString(transformInlineMarkdown(content, false))
				i += consumed
				continue
			}
			// Unpaired delimiters and intraword underscores are visible text.
			run := delimiterRun(s[i:], s[i])
			b.WriteString(s[i : i+run])
			i += run
			continue
		}
		if preserveMarkup && (s[i] == '[' || s[i] == ']') {
			b.WriteByte('\\')
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// decodeHeadingEntities preserves Crossnote's markup-based slug input while
// decoding entity spellings outside code. Unlike decoding the whole title, this
// keeps entity-like code and escaped ampersands literal, and never decodes twice.
func decodeHeadingEntities(s string) string {
	var b strings.Builder
	for i := 0; i < len(s); {
		if s[i] == '\\' && i+1 < len(s) && isMarkdownEscapable(s[i+1]) {
			b.WriteString(s[i : i+2])
			i += 2
			continue
		}
		if s[i] == '`' {
			if _, consumed, ok := parseCodeSpan(s[i:]); ok {
				b.WriteString(s[i : i+consumed])
				i += consumed
				continue
			}
			run := delimiterRun(s[i:], '`')
			b.WriteString(s[i : i+run])
			i += run
			continue
		}
		if s[i] == '&' {
			if decoded, consumed, ok := parseInlineEntity(s[i:]); ok {
				b.WriteString(decoded)
				i += consumed
				continue
			}
		}
		b.WriteByte(s[i])
		i++
	}
	return b.String()
}

// parseInlineEntity decodes one complete reference without swallowing preceding
// literal text or a later escape. The semicolon is required by Markdown.
func parseInlineEntity(s string) (decoded string, consumed int, ok bool) {
	end := strings.IndexByte(s, ';')
	if end < 0 {
		return "", 0, false
	}
	for _, c := range s[1:end] {
		if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' || c >= '0' && c <= '9' || c == '#') {
			return "", 0, false
		}
	}
	entity := s[:end+1]
	decoded = html.UnescapeString(entity)
	return decoded, len(entity), decoded != entity
}

// parseBracketLinkLike exposes the semantic label used by existing text callers.
func parseBracketLinkLike(s string, collapseWhitespace bool) (label string, consumed int, ok bool) {
	label, consumed, ok = parseInlineLink(s)
	if ok {
		label = extractInlineText(label, collapseWhitespace)
	}
	return label, consumed, ok
}

// parseInlineLink finds an inline link while retaining the label's own markup.
// Balanced brackets/parentheses, escapes, and code spans must not truncate the
// label or leave part of a destination in the generated ToC.
func parseInlineLink(s string) (label string, consumed int, ok bool) {
	if len(s) == 0 || s[0] != '[' {
		return "", 0, false
	}
	closeLabel := findClosingDelimiter(s, '[', ']', true)
	if closeLabel < 0 || closeLabel+1 >= len(s) || s[closeLabel+1] != '(' {
		return "", 0, false
	}
	closeTarget := findClosingDelimiter(s[closeLabel+1:], '(', ')', false)
	if closeTarget < 0 {
		return "", 0, false
	}
	return s[1:closeLabel], closeLabel + closeTarget + 2, true
}

// findClosingDelimiter matches a bracketed label or parenthesized destination.
// Code spans are opaque inside labels, but backticks in destinations are literal.
func findClosingDelimiter(s string, open, close byte, skipCode bool) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		if s[i] == '\\' && i+1 < len(s) && isMarkdownEscapable(s[i+1]) {
			i++
			continue
		}
		if skipCode && s[i] == '`' {
			if _, consumed, ok := parseCodeSpan(s[i:]); ok {
				i += consumed - 1
				continue
			}
		}
		switch s[i] {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

// parseCodeSpan requires equal-length backtick runs, allowing code that contains
// single backticks, brackets, underscores, escapes, and entity spellings.
func parseCodeSpan(s string) (content string, consumed int, ok bool) {
	run := delimiterRun(s, '`')
	for i := run; i < len(s); {
		if s[i] != '`' {
			i++
			continue
		}
		closingRun := delimiterRun(s[i:], '`')
		if closingRun == run {
			content = s[run:i]
			// Markdown removes one padding space at each end of nonblank code.
			if strings.HasPrefix(content, " ") && strings.HasSuffix(content, " ") && strings.Trim(content, " ") != "" {
				content = content[1 : len(content)-1]
			}
			return content, i + run, true
		}
		i += closingRun
	}
	return "", 0, false
}

// delimiterRun counts a contiguous ASCII delimiter without splitting a run into
// shorter, independently interpreted markers.
func delimiterRun(s string, marker byte) int {
	i := 0
	for i < len(s) && s[i] == marker {
		i++
	}
	return i
}

// parseFormattingSpan removes paired emphasis/strikethrough markers for slugs.
// Flanking checks keep identifier underscores and unmatched delimiters literal;
// code spans and escaped punctuation cannot accidentally close formatting.
func parseFormattingSpan(s string, start int) (content string, consumed int, ok bool) {
	marker := s[start]
	run := delimiterRun(s[start:], marker)
	canOpen, _ := delimiterFlanking(s, start, start+run)
	if !canOpen {
		return "", 0, false
	}
	for i := start + run; i < len(s); {
		if s[i] == '\\' && i+1 < len(s) && isMarkdownEscapable(s[i+1]) {
			i += 2
			continue
		}
		if s[i] == '`' {
			if _, consumed, ok := parseCodeSpan(s[i:]); ok {
				i += consumed
				continue
			}
		}
		if s[i] == marker {
			closingRun := delimiterRun(s[i:], marker)
			_, canClose := delimiterFlanking(s, i, i+closingRun)
			if closingRun >= run && canClose {
				end := i + closingRun - run
				return s[start+run : end], i + closingRun - start, true
			}
			i += closingRun
			continue
		}
		i++
	}
	return "", 0, false
}

// delimiterFlanking applies Markdown's word-boundary distinction between '*'
// and '_', including Unicode neighbors. Underscores inside identifiers cannot
// open or close emphasis even when another underscore appears later in a word.
func delimiterFlanking(s string, start, end int) (canOpen, canClose bool) {
	before, _ := utf8.DecodeLastRuneInString(s[:start])
	after, _ := utf8.DecodeRuneInString(s[end:])
	beforeSpace := start == 0 || unicode.IsSpace(before)
	afterSpace := end == len(s) || unicode.IsSpace(after)
	beforePunct := unicode.IsPunct(before) || unicode.IsSymbol(before)
	afterPunct := unicode.IsPunct(after) || unicode.IsSymbol(after)
	left := !afterSpace && (!afterPunct || beforeSpace || beforePunct)
	right := !beforeSpace && (!beforePunct || afterSpace || afterPunct)
	if s[start] == '_' {
		return left && (!right || beforePunct), right && (!left || afterPunct)
	}
	return left, right
}

// isMarkdownEscapable limits backslash decoding to Markdown's ASCII punctuation
// so ordinary backslashes (including those in paths) remain visible.
func isMarkdownEscapable(c byte) bool {
	return c >= '!' && c <= '/' || c >= ':' && c <= '@' || c >= '[' && c <= '`' || c >= '{' && c <= '~'
}

// isFormattingMarker identifies supported emphasis and strikethrough delimiters;
// callers still have to check pairing before discarding visible punctuation.
func isFormattingMarker(r rune) bool {
	switch r {
	case '*', '_', '~':
		return true
	default:
		return false
	}
}

// collapseWhitespaceRuns folds runs of whitespace into single spaces.
func collapseWhitespaceRuns(s string) string {
	return strings.Join(strings.Fields(s), " ")
}
