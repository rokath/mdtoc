package mdtoc

import "fmt"

const (
	startMarker = "<!-- mdtoc -->"
	endMarker   = "<!-- /mdtoc -->"
	offMarker   = "<!-- mdtoc off -->"
	onMarker    = "<!-- mdtoc on -->"
	configStart = "<!-- mdtoc-config"
	configEnd   = "-->"

	preservedCommentHeader = "<!-- preserved by mdtoc"
)

// State is the persisted document state in the config block.
type State string

const (
	StateGenerated State = "generated"
	StateStripped  State = "stripped"
)

// ContainerVersion identifies the managed config-block format.
type ContainerVersion string

const (
	ContainerVersionV1 ContainerVersion = "v1"
	ContainerVersionV2 ContainerVersion = "v2"
)

// BulletMode controls how unordered ToC bullets are selected.
type BulletMode string

const (
	BulletAuto BulletMode = "auto"
	BulletStar BulletMode = "*"
	BulletDash BulletMode = "-"
	BulletPlus BulletMode = "+"
)

// AnchorMode controls how anchor IDs are interpreted and whether inline anchors are rendered.
type AnchorMode string

const (
	AnchorGitHub AnchorMode = "github"
	AnchorGitLab AnchorMode = "gitlab"
	AnchorFalse  AnchorMode = "false"
)

// RendersInline reports whether the mode should emit inline anchor HTML.
func (m AnchorMode) RendersInline() bool {
	return m != AnchorFalse
}

// Config mirrors the normalized config block managed by the tool.
type Config struct {
	ContainerVersion ContainerVersion
	Numbering        bool
	MinLevel         int
	MaxLevel         int
	Anchor           AnchorMode
	TOC              bool
	Bullets          BulletMode
	State            State

	// BulletsExplicit reports whether the parsed config block already contained
	// an explicit bullets line. Legacy configs omit it.
	BulletsExplicit bool
}

// DefaultConfig returns the v1 defaults from the specification.
func DefaultConfig() Config {
	return Config{ContainerVersion: ContainerVersionV2, Numbering: true, MinLevel: 2, MaxLevel: 4, Anchor: AnchorGitHub, TOC: true, Bullets: BulletAuto, State: StateGenerated}
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
	if c.ContainerVersion != ContainerVersionV1 && c.ContainerVersion != ContainerVersionV2 {
		return fmt.Errorf("container-version must be v1 or v2")
	}
	if c.Anchor != AnchorGitHub && c.Anchor != AnchorGitLab && c.Anchor != AnchorFalse {
		return fmt.Errorf("anchor must be github, gitlab, or false")
	}
	if c.Bullets != BulletAuto && c.Bullets != BulletStar && c.Bullets != BulletDash && c.Bullets != BulletPlus {
		return fmt.Errorf("bullets must be auto, *, -, or +")
	}
	if c.State != StateGenerated && c.State != StateStripped {
		return fmt.Errorf("state must be generated or stripped")
	}
	return nil
}

// Options contains only the persisted generate options.
type Options struct {
	Numbering bool
	MinLevel  int
	MaxLevel  int
	Anchor    AnchorMode
	TOC       bool
	Bullets   BulletMode
}

// DefaultOptions mirrors the generator defaults from the spec.
func DefaultOptions() Options {
	d := DefaultConfig()
	return Options{Numbering: d.Numbering, MinLevel: d.MinLevel, MaxLevel: d.MaxLevel, Anchor: d.Anchor, TOC: d.TOC, Bullets: d.Bullets}
}

// ToConfig converts ephemeral generate options to a persisted config.
func (o Options) ToConfig() Config {
	bullets := o.Bullets
	if bullets == "" {
		bullets = BulletAuto
	}
	anchor := o.Anchor
	if anchor == "" {
		anchor = AnchorGitHub
	}
	return Config{ContainerVersion: ContainerVersionV2, Numbering: o.Numbering, MinLevel: o.MinLevel, MaxLevel: o.MaxLevel, Anchor: anchor, TOC: o.TOC, Bullets: bullets, State: StateGenerated}
}

// Heading stores one heading candidate that mdtoc can manage.
type Heading struct {
	LineIndex     int
	Level         int
	Hashes        string
	TitleMarkup   string
	TitleText     string
	ManagedNumber string
	ManagedAnchor string
}

// InManagedRange reports whether the heading participates in numbering, anchors,
// and ToC rendering under the current config.
func (h Heading) InManagedRange(cfg Config) bool {
	return h.Level >= cfg.MinLevel && h.Level <= cfg.MaxLevel
}

// ParsedDocument stores the structural information gathered during the parse
// pass.
type ParsedDocument struct {
	Lines      []string
	Container  *Container
	Headings   []Heading
	Warnings   []string
	TrailingLF bool
}

// Container identifies the managed area and stores the parsed config.
type Container struct {
	StartLine       int
	ConfigStartLine int
	ConfigEndLine   int
	EndLine         int
	Config          Config
	TOCArea         []string
}
