# nownow

Keep your [now.ctx.st](https://now.ctx.st) status green without thinking about it.

## Install

```bash
# macOS
brew install biao29/tap/nownow

# Linux
curl -fsSL https://now.ctx.st/install.sh | sh

# Windows
scoop bucket add biao29 https://github.com/biao29/scoop-bucket
scoop install nownow

# From source
go install github.com/ctx-st/nownow@latest
```

## Quick Start

```bash
nownow login    # paste your API token from now.ctx.st/admin.html
nownow start    # auto-detect context, push every 5 minutes
```

## Commands

| Command | Description |
|---|---|
| `nownow detect` | Print detected context (app, git, music). `--json` for JSON output |
| `nownow push [msg]` | Detect + push status. Pass a message to skip auto-detection |
| `nownow login` | Store and verify your API token |
| `nownow start` | Watch mode — auto-push on interval. `--interval 2m` to customize |
| `nownow status` | Show your current status on the board |
| `nownow version` | Print version info |

## Context Detection

| Signal | macOS | Linux | Windows |
|---|---|---|---|
| Active app | osascript | xdotool + xprop | PowerShell |
| Window title | osascript | xdotool | PowerShell |
| Git repo/branch | git | git | git |
| Music (Spotify) | osascript | playerctl | — |
| Music (Apple Music) | osascript | — | — |

Missing signals are silently skipped — nownow reports what it can detect.

## Configuration

Config lives at `~/.config/nownow/config.yml`:

```yaml
endpoint: https://now.ctx.st
token: now_xxx

# Status template — available: {app}, {title}, {project}, {branch}, {music}, {emoji}
template: "{emoji} {app} · {project} ({branch})"

# Watch interval
interval: 5m

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
