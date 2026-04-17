package mdtoc

import (
	"fmt"
	"strings"
	"unicode"
)

// Slugger tracks per-document collisions.
type Slugger struct {
	seen map[string]int
}

// NewSlugger creates a fresh collision tracker.
func NewSlugger() *Slugger {
	return &Slugger{seen: map[string]int{}}
}

// Next returns the deterministic anchor ID for one heading.
func (s *Slugger) Next(title string) string {
	base := slugifyBase(title)
	count := s.seen[base]
	s.seen[base] = count + 1
	if count == 0 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, count)
}

// slugifyBase implements the shared slug/anchor rules from the specification.
func slugifyBase(title string) string {
	title = strings.ToLower(title)
	var b strings.Builder
	hasContent := false
	pendingDash := false
	for _, r := range title {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			if pendingDash && hasContent {
				b.WriteByte('-')
			}
			b.WriteRune(r)
			hasContent = true
			pendingDash = false
		case unicode.IsSpace(r):
			if hasContent {
				pendingDash = true
			}
		case unicode.IsPunct(r):
			if isInWordPunctuation(r) {
				continue
			}
			if hasContent {
				pendingDash = true
			}
		default:
			if hasContent {
				pendingDash = true
			}
		}
	}
	if b.Len() == 0 {
		return "section"
	}
	return b.String()
}

func isInWordPunctuation(r rune) bool {
	switch r {
	case '\'', '’', '"', '“', '”':
		return true
	default:
		return false
	}
}
