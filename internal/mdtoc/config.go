package mdtoc

import (
	"fmt"
	"strconv"
	"strings"
)

// parseConfig parses the compact key/value config comment.
func parseConfig(lines []string) (Config, error) {
	content, err := configCommentContent(lines)
	if err != nil {
		return Config{}, err
	}
	cfg := DefaultConfig()
	seen := map[string]bool{}
	for _, field := range strings.Fields(content) {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			return Config{}, fmt.Errorf("invalid config token %q", field)
		}
		key, value := parts[0], parts[1]
		if !isKnownConfigKey(key) {
			continue
		}
		if seen[key] {
			return Config{}, fmt.Errorf("duplicate config key %q", key)
		}
		seen[key] = true
		switch key {
		case "numbering":
			v, err := parseBoolValue(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid numbering: %w", err)
			}
			cfg.Numbering = v
		case "min":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid min: %w", err)
			}
			cfg.MinLevel = n
		case "max":
			n, err := strconv.Atoi(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid max: %w", err)
			}
			cfg.MaxLevel = n
		case "slug":
			v, err := parseSlugMode(value)
			if err != nil {
				return Config{}, err
			}
			cfg.Slug = v
		case "anchor":
			v, err := parseBoolValue(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid anchor: %w", err)
			}
			cfg.Anchor = v
		case "link":
			v, err := parseBoolValue(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid link: %w", err)
			}
			cfg.Link = v
		case "toc":
			v, err := parseBoolValue(value)
			if err != nil {
				return Config{}, fmt.Errorf("invalid toc: %w", err)
			}
			cfg.TOC = v
		case "bullets":
			v, err := parseBulletMode(value)
			if err != nil {
				return Config{}, err
			}
			cfg.Bullets = v
		}
	}
	if err := cfg.Validate(); err != nil {
		return Config{}, err
	}
	return cfg, nil
}

func configCommentContent(lines []string) (string, error) {
	if len(lines) == 0 {
		return "", fmt.Errorf("empty config block")
	}
	if len(lines) == 1 {
		trimmed := strings.TrimSpace(lines[0])
		if !strings.HasPrefix(trimmed, "<!--") || !strings.HasSuffix(trimmed, "-->") {
			return "", fmt.Errorf("invalid config block delimiters")
		}
		return strings.TrimSpace(strings.TrimSuffix(strings.TrimPrefix(trimmed, "<!--"), "-->")), nil
	}
	first := strings.TrimSpace(lines[0])
	last := strings.TrimSpace(lines[len(lines)-1])
	if !strings.HasPrefix(first, "<!--") || !strings.HasSuffix(last, "-->") {
		return "", fmt.Errorf("invalid config block delimiters")
	}
	parts := []string{strings.TrimSpace(strings.TrimPrefix(first, "<!--"))}
	for _, line := range lines[1 : len(lines)-1] {
		parts = append(parts, strings.TrimSpace(line))
	}
	parts = append(parts, strings.TrimSpace(strings.TrimSuffix(last, "-->")))
	return strings.TrimSpace(strings.Join(parts, " ")), nil
}

func isKnownConfigKey(key string) bool {
	switch key {
	case "numbering", "min", "max", "slug", "anchor", "link", "toc", "bullets":
		return true
	default:
		return false
	}
}

func configContentLooksLikeConfig(content string) bool {
	fields := strings.Fields(content)
	if len(fields) == 0 {
		return true
	}
	allKeyValue := true
	for _, field := range fields {
		parts := strings.SplitN(field, "=", 2)
		if len(parts) != 2 || parts[0] == "" {
			allKeyValue = false
			continue
		}
		if isKnownConfigKey(parts[0]) {
			return true
		}
	}
	return allKeyValue
}

// parseBoolValue accepts canonical true/false booleans and on/off aliases.
func parseBoolValue(s string) (bool, error) {
	switch s {
	case "on", "true":
		return true, nil
	case "off", "false":
		return false, nil
	default:
		return false, fmt.Errorf("invalid boolean value %q", s)
	}
}

func parseSlugMode(s string) (SlugMode, error) {
	switch SlugMode(s) {
	case SlugGitHub, SlugGitLab, SlugCrossnote:
		return SlugMode(s), nil
	default:
		return "", fmt.Errorf("invalid slug value %q", s)
	}
}

// parseBulletMode validates the configured unordered-list bullet selection mode.
func parseBulletMode(s string) (BulletMode, error) {
	switch BulletMode(s) {
	case BulletAuto, BulletStar, BulletDash, BulletPlus:
		return BulletMode(s), nil
	default:
		return "", fmt.Errorf("invalid bullets value %q", s)
	}
}

// RenderConfig emits the compact normalized config block.
func RenderConfig(cfg Config, multiline bool) []string {
	if multiline {
		return []string{
			"<!--",
			fmt.Sprintf("numbering=%s", boolString(cfg.Numbering)),
			fmt.Sprintf("min=%d", cfg.MinLevel),
			fmt.Sprintf("max=%d", cfg.MaxLevel),
			fmt.Sprintf("slug=%s", cfg.Slug),
			fmt.Sprintf("anchor=%s", boolString(cfg.Anchor)),
			fmt.Sprintf("link=%s", boolString(cfg.Link)),
			fmt.Sprintf("toc=%s", boolString(cfg.TOC)),
			fmt.Sprintf("bullets=%s", cfg.Bullets),
			configEnd,
		}
	}
	return []string{fmt.Sprintf("<!-- numbering=%s min=%d max=%d slug=%s anchor=%s link=%s toc=%s bullets=%s -->",
		boolString(cfg.Numbering),
		cfg.MinLevel,
		cfg.MaxLevel,
		cfg.Slug,
		boolString(cfg.Anchor),
		boolString(cfg.Link),
		boolString(cfg.TOC),
		cfg.Bullets,
	)}
}

func boolString(v bool) string {
	if v {
		return "true"
	}
	return "false"
}

func configEquals(a, b Config) bool {
	return a.Numbering == b.Numbering &&
		a.MinLevel == b.MinLevel &&
		a.MaxLevel == b.MaxLevel &&
		a.Slug == b.Slug &&
		a.Anchor == b.Anchor &&
		a.Link == b.Link &&
		a.TOC == b.TOC &&
		a.Bullets == b.Bullets
}
