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

// Config mirrors the normalized config block managed by the tool.
type Config struct {
	Numbering bool
	MinLevel  int
	MaxLevel  int
	Anchors   bool
	TOC       bool
	State     State
}

// DefaultConfig returns the v1 defaults from the specification.
func DefaultConfig() Config {
	return Config{Numbering: true, MinLevel: 2, MaxLevel: 4, Anchors: true, TOC: true, State: StateGenerated}
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
	Anchors   bool
	TOC       bool
}

// DefaultOptions mirrors the generator defaults from the spec.
func DefaultOptions() Options {
	d := DefaultConfig()
	return Options{Numbering: d.Numbering, MinLevel: d.MinLevel, MaxLevel: d.MaxLevel, Anchors: d.Anchors, TOC: d.TOC}
}

// ToConfig converts ephemeral generate options to a persisted config.
func (o Options) ToConfig() Config {
	return Config{Numbering: o.Numbering, MinLevel: o.MinLevel, MaxLevel: o.MaxLevel, Anchors: o.Anchors, TOC: o.TOC, State: StateGenerated}
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
