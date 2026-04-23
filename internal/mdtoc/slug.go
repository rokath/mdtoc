package mdtoc

import (
	"fmt"
	"strings"
	"unicode"
)

type slugifyFunc func(string) string

// Slugger tracks per-document collisions.
type Slugger struct {
	seen    map[string]int
	slugify slugifyFunc
}

func newSlugger(fn slugifyFunc) *Slugger {
	return &Slugger{seen: map[string]int{}, slugify: fn}
}

// NewSlugger creates a fresh GitHub-compatible collision tracker.
func NewSlugger() *Slugger {
	return newSlugger(slugifyGitHubBase)
}

// NewGitLabSlugger creates a fresh GitLab-compatible collision tracker.
func NewGitLabSlugger() *Slugger {
	return newSlugger(slugifyGitLabBase)
}

// Next returns the deterministic anchor ID for one heading.
func (s *Slugger) Next(title string) string {
	base := s.slugify(title)
	count := s.seen[base]
	s.seen[base] = count + 1
	if count == 0 {
		return base
	}
	return fmt.Sprintf("%s-%d", base, count)
}

// slugifyGitHubBase implements the GitHub-compatible slug/anchor rules.
func slugifyGitHubBase(title string) string {
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

// slugifyGitLabBase implements the GitLab heading-ID rules.
func slugifyGitLabBase(title string) string {
	title = strings.ToLower(title)
	var b strings.Builder
	for _, r := range title {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
		case r == '_' || r == '-':
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteByte('-')
		}
	}
	slug := collapseHyphenRuns(strings.Trim(b.String(), "-"))
	if slug == "" {
		return "section"
	}
	return slug
}

func collapseHyphenRuns(s string) string {
	var b strings.Builder
	lastHyphen := false
	for _, r := range s {
		if r == '-' {
			if lastHyphen {
				continue
			}
			lastHyphen = true
			b.WriteRune(r)
			continue
		}
		lastHyphen = false
		b.WriteRune(r)
	}
	return b.String()
}

// isInWordPunctuation reports whether the rune may stay inside a slugged word.
func isInWordPunctuation(r rune) bool {
	switch r {
	case '\'', '’', '"', '“', '”':
		return true
	default:
		return false
	}
}
