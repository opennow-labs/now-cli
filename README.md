# now

[![Go](https://img.shields.io/badge/Go-1.25-00ADD8?logo=go&logoColor=white)](https://go.dev)
[![Release](https://img.shields.io/github/v/release/opennow-labs/now-cli?color=blue)](https://github.com/opennow-labs/now-cli/releases)
[![License](https://img.shields.io/badge/license-O--Saasy-green)](LICENSE.md)
[![macOS](https://img.shields.io/badge/macOS-supported-black?logo=apple)](https://github.com/opennow-labs/now-cli)
[![Linux](https://img.shields.io/badge/Linux-supported-FCC624?logo=linux&logoColor=black)](https://github.com/opennow-labs/now-cli)
[![Windows](https://img.shields.io/badge/Windows-supported-0078D4?logo=windows&logoColor=white)](https://github.com/opennow-labs/now-cli)

Live presence for builders and their agents. You're not building alone.

Keep your [opennow.dev](https://opennow.dev) status green without thinking about it.

## Install

```bash
# macOS
brew install opennow-labs/tap/now-cli

# Linux / macOS (script)
curl -fsSL https://opennow.dev/install.sh | sh

# Windows
irm https://opennow.dev/install.ps1 | iex

# From source
go install github.com/opennow-labs/now-cli@latest
```

## Quick Start

```bash
now login    # opens browser for device flow auth
now start    # auto-detect context, push every 30s
```

## Commands

| Command | Description |
|---|---|
| `now login` | Authenticate via device flow (or `--token` for direct input) |
| `now start` | Start daemon — auto-push on interval. `--interval 2m` to customize. `--no-autostart` to skip autostart installation |
| `now stop` | Stop the daemon |
| `now status` | Show current status on the board |
| `now detect` | Print detected context (app, music, video). `--json` for JSON output |
| `now push [msg]` | Detect + push status. Pass a message to skip auto-detection |
| `now hook` | Manage git hooks for automatic status updates |
| `now wrap` | Run a command and push its result as status |
| `now config` | Open config file in your editor |
| `now upgrade` | Self-update to the latest release. `--restart` to restart daemon after upgrade |
| `now version` | Print version info |

## Context Detection

| Signal | macOS | Linux | Windows |
|---|---|---|---|
| Active app | lsappinfo | xdotool + xprop | PowerShell |
| Window title | osascript | xdotool | PowerShell |
| Music | nowplaying-helper / osascript | playerctl | GlobalSystemMediaTransportControls |
| Video | nowplaying-helper / window title | window title | window title |

Music sources: Spotify, Apple Music, Tidal, Amazon Music, Deezer, QQ Music, NetEase, and more.
Video detection: YouTube, Netflix, Twitch, Disney+, Prime Video, Bilibili, VLC, IINA, mpv, etc.

Missing signals are silently skipped — now reports what it can detect.

## System Tray

When running as a daemon, now shows a system tray icon with:

- Current status display
- Now playing music info
- Pause / Resume auto-detection
- Settings UI (opens in browser at `127.0.0.1:19191`)
- Open Board (opens [opennow.dev](https://opennow.dev))
- Update notifications

## Configuration

Config lives at `~/.config/now/config.yml` (or `$XDG_CONFIG_HOME/now/config.yml`):

```yaml
endpoint: https://opennow.dev
token: now_xxx

# Status template — available: {app}, {title}, {music}, {music.artist}, {music.track}, {watching}, {activity}
template: "{activity}"

# Watch interval
interval: 30s

# Activity rules (exact match, case-insensitive)
activity_rules:
  - match: ["Visual Studio Code", "Code", "Cursor", "Windsurf", "Zed"]
    activity: "Vibe coding"
  - match: ["Xcode", "Android Studio"]
    activity: "Building an app"
  - match: ["Terminal", "iTerm2", "Warp", "Alacritty", "kitty"]
    activity: "Hacking away"
  - match: ["Google Chrome", "Safari", "Arc", "Firefox", "Brave Browser"]
    activity: "Down the rabbit hole"
  - match: ["Figma", "Sketch", "Framer"]
    activity: "Pushing pixels"
  - match: ["Slack", "Discord", "Telegram", "WeChat"]
    activity: "In conversation"
  - match: ["Notion", "Obsidian", "Bear", "Notes"]
    activity: "Capturing thoughts"

# Privacy controls (all enabled by default)
telemetry: true       # overall telemetry
send_app: true        # send app name
send_music: true      # send music info
send_watching: true   # send video content

# Automatic update checks
auto_update: true

# Apps to ignore (case-insensitive)
ignore:
  - "1Password"
  - "System Preferences"
  - "System Settings"
```

The default config includes 40+ activity rules covering dev tools, browsers, design apps, communication, writing, media, and more. Run `now config` to customize.

## Git Hooks

Automatically push status on git events:

```bash
now hook install                          # install post-commit hook
now hook install --hooks post-commit,pre-push  # install multiple hooks
now hook install --template "Shipped: {commit_msg}"  # custom message
now hook list                             # list installed hooks
now hook remove                           # remove all now hooks
```

Hooks are appended to existing hook files (never overwritten) and managed via `# now:start` / `# now:end` markers. Works with worktrees and submodules.

**Default messages:**
- `post-commit`: "Just committed: {commit_msg}"
- `pre-push`: "Pushing to {branch}"

## Command Wrapper

Run any command and push its outcome as status:

```bash
now wrap -- make build                   # "make completed" or "make failed (exit 2)"
now wrap --name "Deploy" -- ./deploy.sh  # "Deploy completed"
now wrap --on-success "Ship it!" --on-failure "Broke it ({exit_code})" -- make test
now wrap --quiet -- backup.sh            # push without printing CLI output
```

**Template variables:** `{cmd}`, `{name}`, `{exit_code}`, `{duration}`

The wrapped command's stdin/stdout/stderr are fully transparent, and its exit code is preserved.

## Privacy

now is designed with privacy as a first-class concern. You stay in full control of what leaves your machine.

### What is collected vs. what is sent

| Data | Detected locally | Sent to server | Toggle |
|---|---|---|---|
| Active app name | Yes | Only if `send_app: true` | `send_app` |
| Activity label | Yes | Only if `send_app: true` | `send_app` |
| Window title | Yes | **Never** | — |
| Music artist & track | Yes | Only if `send_music: true` | `send_music` |
| Video content | Yes | Only if `send_watching: true` | `send_watching` |
| OS & architecture | Yes | Only if `telemetry: true` | `telemetry` |

**Window titles are never transmitted** — they are only used locally for video detection and template rendering.

### Granular opt-out

Each data type can be independently disabled in `~/.config/now/config.yml`:

```yaml
send_app: false       # stop sending app name and activity
send_music: false     # stop sending music info
send_watching: false  # stop sending video content
telemetry: false      # stop sending OS/arch in User-Agent
```

When a toggle is off, the corresponding fields are cleared **before** any network request is made. The data never leaves your machine.

### Ignore list

Block specific apps from being reported entirely (case-insensitive, supports prefix matching):

```yaml
ignore:
  - "1Password"
  - "System Preferences"
```

When an ignored app is in the foreground, no status update is sent — your previous status is preserved.

### Where data goes

All status updates are sent to a single endpoint (default `https://opennow.dev`). You can point it to a self-hosted instance by changing the `endpoint` field. The settings UI is local-only at `127.0.0.1:19191` and is not exposed to the network.

## Development

```bash
go build -o now .
go test ./...
```

## License

[O-Saasy](LICENSE.md)
