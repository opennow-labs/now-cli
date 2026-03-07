# nownow

Live presence for builders and their agents. You're not building alone.

Keep your [now.ctx.st](https://now.ctx.st) status green without thinking about it.

## Install

```bash
# macOS
brew install biao29/tap/nownow

# From source
go install github.com/nownow-labs/nownow@latest
```

## Quick Start

```bash
nownow login    # opens browser for device flow auth
nownow start    # auto-detect context, push every 30s
```

## Commands

| Command | Description |
|---|---|
| `nownow login` | Authenticate via device flow (or `--token` for direct input) |
| `nownow start` | Start daemon — auto-push on interval. `--interval 2m` to customize |
| `nownow stop` | Stop the daemon |
| `nownow status` | Show current status and daemon info |
| `nownow detect` | Print detected context (app, git, music). `--json` for JSON output |
| `nownow push [msg]` | Detect + push status. Pass a message to skip auto-detection |
| `nownow upgrade` | Self-update to the latest release |
| `nownow version` | Print version info |

## Context Detection

| Signal | macOS | Linux | Windows |
|---|---|---|---|
| Active app | lsappinfo | xdotool + xprop | PowerShell |
| Window title | osascript | xdotool | PowerShell |
| Music (Spotify) | osascript | playerctl | — |
| Music (Apple Music) | osascript | — | — |

Missing signals are silently skipped — nownow reports what it can detect.

## Configuration

Config lives at `~/.config/nownow/config.yml`:

```yaml
endpoint: https://now.ctx.st
token: now_xxx

# Status template — available: {app}, {title}, {music}, {music.artist}, {music.track}, {watching}, {emoji}
template: "{emoji} {app}"

# Watch interval
interval: 30s

# Emoji mapping (substring match, case-insensitive)
emoji_rules:
  - match: "Code"
    emoji: "💻"
  - match: "Terminal"
    emoji: "⚡"
  - match: "Figma"
    emoji: "🎨"

# Apps to ignore
ignore:
  - "1Password"
  - "System Settings"
```

## Development

```bash
go build -o nownow .
go test ./...
```

## License

MIT
