package mdtoc

import "fmt"

const (
	startMarker = "<!-- mdtoc -->"
	endMarker   = "<!-- /mdtoc -->"
	offMarker   = "<!-- mdtoc off -->"
	onMarker    = "<!-- mdtoc on -->"
	configEnd   = "-->"

	preservedCommentHeader = "<!-- preserved by mdtoc"
)

// BulletMode controls how unordered ToC bullets are selected.
type BulletMode string

const (
	BulletAuto BulletMode = "auto"
	BulletStar BulletMode = "*"
	BulletDash BulletMode = "-"
	BulletPlus BulletMode = "+"
)

// SlugMode controls how heading IDs and ToC link targets are derived.
type SlugMode string

const (
	SlugGitHub    SlugMode = "github"
	SlugGitLab    SlugMode = "gitlab"
	SlugCrossnote SlugMode = "crossnote"
)

// Config mirrors the normalized config block managed by the tool.
type Config struct {
	// Numbering controls managed heading numbering inside the active range.
	Numbering bool
	// MinLevel is the smallest heading depth managed by mdtoc.
	MinLevel int
	// MaxLevel is the largest heading depth managed by mdtoc.
	MaxLevel int
	// Slug selects the heading-ID derivation profile.
	Slug SlugMode
	// Anchor controls whether managed inline anchors are rendered.
	Anchor bool
	// TOC controls whether the managed container includes a rendered ToC block.
	TOC bool
	// Link controls whether rendered ToC entries are Markdown links.
	Link bool
	// Bullets selects the unordered-list marker used for rendered ToC entries.
	Bullets BulletMode
}

// DefaultConfig returns the v1 defaults from the specification.
func DefaultConfig() Config {
	return Config{Numbering: true, MinLevel: 2, MaxLevel: 4, Slug: SlugGitHub, Anchor: true, TOC: true, Link: true, Bullets: BulletAuto}
}

// Validate checks the persisted configuration contract.
func (c Config) Validate() error {
	if c.MinLevel < 1 {
		return fmt.Errorf("min-level must be >= 1")
	}
	if c.MaxLevel > 6 {
		return fmt.Errorf("max-level must be <= 6")
	}
	if c.MinLevel > c.MaxLevel {
		return fmt.Errorf("min-level must not be greater than max-level")
	}
	if c.Slug != SlugGitHub && c.Slug != SlugGitLab && c.Slug != SlugCrossnote {
		return fmt.Errorf("slug must be github, gitlab, or crossnote")
	}
	if c.Bullets != BulletAuto && c.Bullets != BulletStar && c.Bullets != BulletDash && c.Bullets != BulletPlus {
		return fmt.Errorf("bullets must be auto, *, -, or +")
	}
	return nil
}

// Options contains only the persisted generate options.
type Options struct {
	// Numbering controls managed heading numbering inside the active range.
	Numbering bool
	// MinLevel is the smallest heading depth managed by mdtoc.
	MinLevel int
	// MaxLevel is the largest heading depth managed by mdtoc.
	MaxLevel int
	// Slug selects the heading-ID derivation profile.
	Slug SlugMode
	// Anchor controls whether managed inline anchors are rendered.
	Anchor bool
	// AnchorSet reports whether Anchor was explicitly provided.
	AnchorSet bool
	// TOC controls whether generation renders a ToC block.
	TOC bool
	// Link controls whether rendered ToC entries are Markdown links.
	Link bool
	// LinkSet reports whether Link was explicitly provided.
	LinkSet bool
	// Bullets selects the unordered-list marker used for rendered ToC entries.
	Bullets BulletMode
}

// DefaultOptions mirrors the generator defaults from the spec.
func DefaultOptions() Options {
	d := DefaultConfig()
	return Options{Numbering: d.Numbering, MinLevel: d.MinLevel, MaxLevel: d.MaxLevel, Slug: d.Slug, Anchor: d.Anchor, AnchorSet: true, TOC: d.TOC, Link: d.Link, LinkSet: true, Bullets: d.Bullets}
}

// ToConfig converts ephemeral generate options to a persisted config.
func (o Options) ToConfig() Config {
	bullets := o.Bullets
	if bullets == "" {
		bullets = BulletAuto
	}
	slug := o.Slug
	if slug == "" {
		slug = SlugGitHub
	}
	anchor := o.Anchor
	if !o.AnchorSet && !anchor {
		anchor = true
	}
	link := o.Link
	if !o.LinkSet && !link {
		link = true
	}
	return Config{Numbering: o.Numbering, MinLevel: o.MinLevel, MaxLevel: o.MaxLevel, Slug: slug, Anchor: anchor, TOC: o.TOC, Link: link, Bullets: bullets}
}

// Heading stores one heading candidate that mdtoc can manage.
type Heading struct {
	// LineIndex points to the heading line in ParsedDocument.Lines.
	LineIndex int
	// Level is the ATX heading depth from 1 to 6.
	Level int
	// Hashes stores the literal leading hash run, such as "##".
	Hashes string
	// TitleMarkup is the heading content as parsed from the Markdown line.
	TitleMarkup string
	// TitleText is the normalized plain-text title used for slugging and ToC text.
	TitleText string
	// ManagedNumber is the computed numbering prefix, or empty when numbering is off/out of range.
	ManagedNumber string
	// ManagedAnchor is the computed inline anchor for this heading under the active slug mode.
	ManagedAnchor string
	// ManagedTOCTarget is the computed ToC link target for this heading.
	ManagedTOCTarget string
}

// InManagedRange reports whether the heading participates in numbering, anchors,
// and ToC rendering under the current config.
func (h Heading) InManagedRange(cfg Config) bool {
	return h.Level >= cfg.MinLevel && h.Level <= cfg.MaxLevel
}

// ParsedDocument stores the structural information gathered during the parse
// pass.
type ParsedDocument struct {
	// Lines stores the original document split into logical lines.
	Lines []string
	// Container points to the managed mdtoc region when present.
	Container *Container
	// Headings holds all parsed heading candidates in document order.
	Headings []Heading
	// Warnings collects non-fatal parse findings for verbose diagnostics.
	Warnings []string
	// TrailingLF reports whether the original document ended with a newline byte.
	TrailingLF bool
}

// Container identifies the managed area and stores the parsed config.
type Container struct {
	// StartLine is the line index of the opening `<!-- mdtoc -->` marker.
	StartLine int
	// ConfigStartLine is the line index of the opening config marker.
	ConfigStartLine int
	// ConfigEndLine is the line index of the closing `-->` for the config block.
	ConfigEndLine int
	// EndLine is the line index of the closing `<!-- /mdtoc -->` marker.
	EndLine int
	// Config is the parsed normalized config stored inside the container.
	Config Config
	// ConfigPresent reports whether the container had an explicit config block.
	ConfigPresent bool
	// ConfigMultiline reports whether the explicit config block used multiline layout.
	ConfigMultiline bool
	// TOCArea stores the managed lines between the config block and closing marker.
	TOCArea []string
}
