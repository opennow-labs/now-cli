package config

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/yaml.v3"
)

type ActivityRule struct {
	Match    []string `yaml:"match"`
	Activity string   `yaml:"activity"`
}

type Config struct {
	Endpoint      string         `yaml:"endpoint"`
	Token         string         `yaml:"token"`
	Template      string         `yaml:"template"`
	Interval      string         `yaml:"interval,omitempty"`
	ActivityRules []ActivityRule `yaml:"activity_rules,omitempty"`
	Ignore        []string       `yaml:"ignore,omitempty"`
	Telemetry     *bool          `yaml:"telemetry,omitempty"`
	SendApp       *bool          `yaml:"send_app,omitempty"`
	SendMusic     *bool          `yaml:"send_music,omitempty"`
	SendWatching  *bool          `yaml:"send_watching,omitempty"`
	AutoUpdate    *bool          `yaml:"auto_update,omitempty"`
}

// TelemetryEnabled returns true unless explicitly disabled.
func (c Config) TelemetryEnabled() bool {
	return c.Telemetry == nil || *c.Telemetry
}

// SendAppEnabled returns true unless explicitly disabled.
func (c Config) SendAppEnabled() bool {
	return c.SendApp == nil || *c.SendApp
}

// SendMusicEnabled returns true unless explicitly disabled.
func (c Config) SendMusicEnabled() bool {
	return c.SendMusic == nil || *c.SendMusic
}

// SendWatchingEnabled returns true unless explicitly disabled.
func (c Config) SendWatchingEnabled() bool {
	return c.SendWatching == nil || *c.SendWatching
}

// AutoUpdateEnabled returns true unless explicitly disabled.
func (c Config) AutoUpdateEnabled() bool {
	return c.AutoUpdate == nil || *c.AutoUpdate
}

func DefaultConfig() Config {
	return Config{
		Endpoint: "https://now.ctx.st",
		Template: "{activity}",
		Interval: "30s",
		ActivityRules: []ActivityRule{
			// Dev tools
			{Match: []string{"Visual Studio Code", "Code", "Cursor", "Windsurf", "Zed", "Sublime Text", "Nova"}, Activity: "Coding"},
			{Match: []string{"Xcode", "Android Studio"}, Activity: "Building an app"},
			{Match: []string{"IntelliJ IDEA", "GoLand", "PyCharm", "WebStorm", "RustRover", "CLion", "PhpStorm", "Rider"}, Activity: "Coding"},
			{Match: []string{"Terminal", "iTerm2", "Warp", "Alacritty", "kitty", "Hyper", "WezTerm", "Rio"}, Activity: "In the terminal"},
			{Match: []string{"Docker Desktop", "Podman Desktop"}, Activity: "Managing containers"},
			{Match: []string{"TablePlus", "Postico", "DataGrip", "DBeaver", "Sequel Pro", "pgAdmin 4"}, Activity: "Querying database"},
			{Match: []string{"Postman", "Insomnia", "HTTPie", "RapidAPI"}, Activity: "Testing APIs"},
			// Browsers
			{Match: []string{"Google Chrome", "Safari", "Arc", "Firefox", "Brave Browser", "Microsoft Edge", "Opera", "Vivaldi", "Orion", "Zen Browser"}, Activity: "Browsing"},
			// Design & creative
			{Match: []string{"Figma", "Sketch", "Framer"}, Activity: "Designing"},
			{Match: []string{"Adobe Photoshop", "Pixelmator Pro", "Affinity Photo 2", "GIMP"}, Activity: "Editing photos"},
			{Match: []string{"Adobe Illustrator", "Affinity Designer 2", "Vectornator", "Linearity Curve"}, Activity: "Drawing vectors"},
			{Match: []string{"Final Cut Pro", "Adobe Premiere Pro", "DaVinci Resolve", "CapCut", "iMovie"}, Activity: "Editing video"},
			{Match: []string{"Logic Pro", "Ableton Live", "GarageBand", "FL Studio"}, Activity: "Making music"},
			{Match: []string{"Blender", "Cinema 4D", "Maya"}, Activity: "3D modeling"},
			// Communication
			{Match: []string{"Slack", "Discord", "Telegram", "WeChat", "Messages", "WhatsApp", "Signal"}, Activity: "Chatting"},
			{Match: []string{"Zoom", "Google Meet", "Microsoft Teams", "Lark", "Feishu", "腾讯会议", "钉钉"}, Activity: "In a meeting"},
			{Match: []string{"Mail", "Outlook", "Spark", "Airmail", "Mimestream"}, Activity: "Reading email"},
			// Writing & knowledge
			{Match: []string{"Notion", "Obsidian", "Logseq", "Craft", "Bear", "Notes", "Apple Notes"}, Activity: "Taking notes"},
			{Match: []string{"iA Writer", "Ulysses", "Typora", "marktext"}, Activity: "Writing"},
			{Match: []string{"Microsoft Word", "Pages", "Google Docs"}, Activity: "Writing a doc"},
			// Productivity
			{Match: []string{"Microsoft Excel", "Numbers", "Google Sheets"}, Activity: "Working on a spreadsheet"},
			{Match: []string{"Keynote", "Microsoft PowerPoint", "Google Slides"}, Activity: "Making slides"},
			{Match: []string{"Linear", "Jira", "Asana", "Trello", "Todoist", "Things"}, Activity: "Managing tasks"},
			{Match: []string{"Calendar", "Fantastical", "Cron"}, Activity: "Checking calendar"},
			// Reading & learning
			{Match: []string{"Kindle", "Books", "Apple Books"}, Activity: "Reading"},
			{Match: []string{"Reeder", "NetNewsWire", "Readwise Reader", "Feedly"}, Activity: "Reading feeds"},
			{Match: []string{"Preview", "PDF Expert", "Skim"}, Activity: "Reading a PDF"},
			// Media
			{Match: []string{"Spotify", "Apple Music", "NetEase Music", "QQ Music", "网易云音乐"}, Activity: "Listening to music"},
			{Match: []string{"IINA", "VLC", "Infuse", "mpv"}, Activity: "Watching something"},
			// Gaming
			{Match: []string{"Steam", "Epic Games Launcher"}, Activity: "Gaming"},
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
		cfg.Template = "{activity}"
	}
	if cfg.Interval == "" {
		cfg.Interval = "30s"
	}

	// Migrate: if telemetry was explicitly false and send_app not set, inherit
	if cfg.Telemetry != nil && !*cfg.Telemetry && cfg.SendApp == nil {
		f := false
		cfg.SendApp = &f
	}

	// Migrate legacy templates: strip removed {project}/{branch} placeholders
	cfg.Template = migrateTemplate(cfg.Template)

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

// ActivityFor returns the activity label for a given app name via exact case-insensitive match.
func (c Config) ActivityFor(app string) string {
	for _, rule := range c.ActivityRules {
		for _, m := range rule.Match {
			if strings.EqualFold(app, m) {
				return rule.Activity
			}
		}
	}
	return ""
}

// ResolveActivity returns the core activity description.
// Priority: watching > matched activity > "Using {app}".
// Does not append music/watching details — the frontend displays those separately.
func (c Config) ResolveActivity(app, watching string) string {
	if watching != "" {
		return "Watching"
	}
	activity := c.ActivityFor(app)
	if activity == "" && app != "" {
		return "Using " + app
	}
	return activity
}

// migrateTemplate strips removed {project}/{branch} placeholders and legacy emoji references from templates.
func migrateTemplate(tmpl string) string {
	// Migrate legacy emoji placeholders to activity
	tmpl = strings.ReplaceAll(tmpl, "{emoji} {app}", "{activity}")
	tmpl = strings.ReplaceAll(tmpl, "{emoji}", "{activity}")

	tmpl = strings.ReplaceAll(tmpl, "{project}", "")
	tmpl = strings.ReplaceAll(tmpl, "{branch}", "")

	// Clean up artifacts: empty parens/brackets, collapse spaces first, then separators
	tmpl = strings.ReplaceAll(tmpl, "()", "")
	tmpl = strings.ReplaceAll(tmpl, "[]", "")
	for strings.Contains(tmpl, "  ") {
		tmpl = strings.ReplaceAll(tmpl, "  ", " ")
	}
	for strings.Contains(tmpl, "· ·") {
		tmpl = strings.ReplaceAll(tmpl, "· ·", "·")
	}
	// Re-collapse spaces after middot cleanup
	for strings.Contains(tmpl, "  ") {
		tmpl = strings.ReplaceAll(tmpl, "  ", " ")
	}
	tmpl = strings.TrimRight(tmpl, " ·")
	tmpl = strings.TrimLeft(tmpl, " ·")
	return strings.TrimSpace(tmpl)
}
