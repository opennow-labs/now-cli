package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type EmojiRule struct {
	Match string `yaml:"match"`
	Emoji string `yaml:"emoji"`
}

type Config struct {
	Endpoint   string      `yaml:"endpoint"`
	Token      string      `yaml:"token"`
	Template   string      `yaml:"template"`
	Interval   string      `yaml:"interval,omitempty"`
	EmojiRules []EmojiRule `yaml:"emoji_rules,omitempty"`
	Ignore     []string    `yaml:"ignore,omitempty"`
	Telemetry  *bool       `yaml:"telemetry,omitempty"`
}

// TelemetryEnabled returns true unless explicitly disabled.
func (c Config) TelemetryEnabled() bool {
	return c.Telemetry == nil || *c.Telemetry
}

func DefaultConfig() Config {
	return Config{
		Endpoint: "https://now.ctx.st",
		Template: "{emoji} {app} · {project} ({branch})",
		Interval: "30s",
		EmojiRules: []EmojiRule{
			{Match: "Code", Emoji: "\U0001F4BB"},
			{Match: "Cursor", Emoji: "\U0001F4BB"},
			{Match: "Terminal", Emoji: "\u26A1"},
			{Match: "iTerm", Emoji: "\u26A1"},
			{Match: "Warp", Emoji: "\u26A1"},
			{Match: "Figma", Emoji: "\U0001F3A8"},
			{Match: "Safari", Emoji: "\U0001F310"},
			{Match: "Chrome", Emoji: "\U0001F310"},
			{Match: "Arc", Emoji: "\U0001F310"},
			{Match: "Slack", Emoji: "\U0001F4AC"},
		},
		Ignore: []string{"1Password", "System Preferences", "System Settings"},
	}
}

// Dir returns the config directory path.
func Dir() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("cannot find home directory: %w", err)
	}

	// Respect XDG_CONFIG_HOME if set
	if xdg := os.Getenv("XDG_CONFIG_HOME"); xdg != "" {
		return filepath.Join(xdg, "nownow"), nil
	}
	return filepath.Join(home, ".config", "nownow"), nil
}

// Path returns the full path to config.yml.
func Path() (string, error) {
	dir, err := Dir()
	if err != nil {
		return "", err
	}
	return filepath.Join(dir, "config.yml"), nil
}

// Load reads config from disk. Returns default config if file doesn't exist.
func Load() (Config, error) {
	cfg := DefaultConfig()

	p, err := Path()
	if err != nil {
		return cfg, err
	}

	data, err := os.ReadFile(p)
	if err != nil {
		if os.IsNotExist(err) {
			return cfg, nil
		}
		return cfg, fmt.Errorf("reading config: %w", err)
	}

	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("parsing config: %w", err)
	}

	// Ensure defaults for empty fields
	if cfg.Endpoint == "" {
		cfg.Endpoint = "https://now.ctx.st"
	}
	if cfg.Template == "" {
		cfg.Template = "{emoji} {app} · {project} ({branch})"
	}
	if cfg.Interval == "" {
		cfg.Interval = "30s"
	}

	return cfg, nil
}

// Save writes config to disk, creating the directory if needed.
func Save(cfg Config) error {
	p, err := Path()
	if err != nil {
		return err
	}

	if err := os.MkdirAll(filepath.Dir(p), 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}

	data, err := yaml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("marshaling config: %w", err)
	}

	if err := os.WriteFile(p, data, 0600); err != nil {
		return fmt.Errorf("writing config: %w", err)
	}

	return nil
}

// HasToken returns true if a token is configured.
func (c Config) HasToken() bool {
	return c.Token != ""
}

// IsIgnored returns true if the app name should be ignored.
func (c Config) IsIgnored(app string) bool {
	for _, name := range c.Ignore {
		if name == app {
			return true
		}
	}
	return false
}

// EmojiFor returns the emoji for a given app name, or fallback if no match.
func (c Config) EmojiFor(app string, fallback string) string {
	for _, rule := range c.EmojiRules {
		// Simple substring match
		if containsInsensitive(app, rule.Match) {
			return rule.Emoji
		}
	}
	return fallback
}

func containsInsensitive(s, substr string) bool {
	// Simple case-insensitive contains using lowercase comparison
	ls := toLower(s)
	lsub := toLower(substr)
	return len(lsub) > 0 && contains(ls, lsub)
}

func toLower(s string) string {
	b := make([]byte, len(s))
	for i := range s {
		c := s[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		b[i] = c
	}
	return string(b)
}

func contains(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
