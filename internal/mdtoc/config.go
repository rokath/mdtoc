package mdtoc

import (
	"fmt"
	"strconv"
	"strings"
)

// parseConfig parses the exact fixed-order config block described by the spec.
func parseConfig(lines []string) (Config, error) {
	expectedKeys := []string{"numbering", "min-level", "max-level", "anchors", "toc", "state"}
	if len(lines) != 8 {
		return Config{}, fmt.Errorf("invalid mdtoc config block length")
	}
	if strings.TrimSpace(lines[0]) != configStart || strings.TrimSpace(lines[7]) != configEnd {
		return Config{}, fmt.Errorf("invalid mdtoc config block delimiters")
	}
	cfg := DefaultConfig()
	for i, key := range expectedKeys {
		line := strings.TrimSpace(lines[i+1])
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 || parts[0] != key {
			return Config{}, fmt.Errorf("invalid config line %q: expected key %q", line, key)
		}
		value := parts[1]
		switch key {
		case "numbering":
			v, err := parseOnOff(value)
			if err != nil {
				return Config{}, err
			}
			cfg.Numbering = v
		case "min-level":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid min-level: %w", err)
			}
			cfg.MinLevel = n
		case "max-level":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid max-level: %w", err)
			}
			cfg.MaxLevel = n
		case "anchors":
			v, err := parseOnOff(value)
			if err != nil {
				return Config{}, err
			}
			cfg.Anchors = v
		case "toc":
			v, err := parseOnOff(value)
			if err != nil {
				return Config{}, err
			}
			cfg.TOC = v
		case "state":
			cfg.State = State(value)
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

// parseOnOff converts the persisted on/off syntax into a boolean.
func parseOnOff(s string) (bool, error) {
	switch s {
	case "on":
		return true, nil
	case "off":
		return false, nil
	default:
		return false, fmt.Errorf("invalid on/off value %q", s)
	}
}

// RenderConfig emits the exact normalized config block.
func RenderConfig(cfg Config) []string {
	return []string{
		configStart,
		fmt.Sprintf("numbering=%s", onOff(cfg.Numbering)),
		fmt.Sprintf("min-level=%d", cfg.MinLevel),
		fmt.Sprintf("max-level=%d", cfg.MaxLevel),
		fmt.Sprintf("anchors=%s", onOff(cfg.Anchors)),
		fmt.Sprintf("toc=%s", onOff(cfg.TOC)),
		fmt.Sprintf("state=%s", cfg.State),
		configEnd,
	}
}

// onOff converts a boolean back to the persisted on/off syntax.
func onOff(v bool) string {
	if v {
		return "on"
	}
	return "off"
}
