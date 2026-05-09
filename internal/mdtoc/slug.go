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

// NewCrossnoteSlugger creates a fresh collision tracker for Crossnote /
// Markdown Preview Enhanced heading IDs.
func NewCrossnoteSlugger() *Slugger {
	return newSlugger(slugifyCrossnoteBase)
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
	title = strings.TrimSpace(title)
	title = strings.ToLower(title)
	var b strings.Builder
	hasContent := false
	for _, r := range title {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			b.WriteRune(r)
			hasContent = true
		case r == '-' || r == '_':
			b.WriteRune(r)
			hasContent = true
		case r == ' ':
			if hasContent {
				b.WriteByte('-')
			}
		case unicode.IsSpace(r):
			continue
		default:
			continue
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

// slugifyCrossnoteBase models the Crossnote / Markdown Preview Enhanced
// heading ID pipeline.
func slugifyCrossnoteBase(title string) string {
	title = strings.TrimSpace(title)
	title = strings.ReplaceAll(title, "~", "")
	title = strings.ReplaceAll(title, "。", "")
	var withTildes strings.Builder
	for _, r := range title {
		if unicode.IsSpace(r) {
			withTildes.WriteByte('~')
			continue
		}
		withTildes.WriteRune(r)
	}
	slug := slugifyUSlugLike(withTildes.String())
	slug = strings.ReplaceAll(slug, "~", "-")
	if slug == "" {
		return "section"
	}
	return slug
}

func slugifyUSlugLike(title string) string {
	title = strings.ToLower(title)
	var b strings.Builder
	for _, r := range title {
		switch {
		case r == '-' || r == '_' || r == '~':
			b.WriteRune(r)
		case unicode.IsLetter(r) || unicode.IsDigit(r) || unicode.IsMark(r):
			b.WriteRune(r)
		case unicode.IsSpace(r):
			b.WriteByte(' ')
		}
	}
	return collapseHyphenRuns(strings.TrimSpace(b.String()))
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
