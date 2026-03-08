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
| `nownow start` | Start daemon — auto-push on interval. `--interval 2m` to customize. `--no-autostart` to skip autostart installation |
| `nownow stop` | Stop the daemon |
| `nownow status` | Show current status on the board |
| `nownow detect` | Print detected context (app, music, video). `--json` for JSON output |
| `nownow push [msg]` | Detect + push status. Pass a message to skip auto-detection |
| `nownow hook` | Manage git hooks for automatic status updates |
| `nownow wrap` | Run a command and push its result as status |
| `nownow config` | Open config file in your editor |
| `nownow upgrade` | Self-update to the latest release. `--restart` to restart daemon after upgrade |
| `nownow version` | Print version info |

## Context Detection

| Signal | macOS | Linux | Windows |
|---|---|---|---|
| Active app | lsappinfo | xdotool + xprop | PowerShell |
| Window title | osascript | xdotool | PowerShell |
| Music | nowplaying-helper / osascript | playerctl | GlobalSystemMediaTransportControls |
| Video | nowplaying-helper / window title | window title | window title |

Music sources: Spotify, Apple Music, Tidal, Amazon Music, Deezer, QQ Music, NetEase, and more.
Video detection: YouTube, Netflix, Twitch, Disney+, Prime Video, Bilibili, VLC, IINA, mpv, etc.

Missing signals are silently skipped — nownow reports what it can detect.

## System Tray

When running as a daemon, nownow shows a system tray icon with:

- Current status display
- Now playing music info
- Pause / Resume auto-detection
- Settings UI (opens in browser at `127.0.0.1:19191`)
- Open Board (opens [now.ctx.st](https://now.ctx.st))
- Update notifications

## Configuration

Config lives at `~/.config/nownow/config.yml` (or `$XDG_CONFIG_HOME/nownow/config.yml`):

```yaml
endpoint: https://now.ctx.st
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

The default config includes 40+ activity rules covering dev tools, browsers, design apps, communication, writing, media, and more. Run `nownow config` to customize.

## Git Hooks

Automatically push status on git events:

```bash
nownow hook install                          # install post-commit hook
nownow hook install --hooks post-commit,pre-push  # install multiple hooks
nownow hook install --template "Shipped: {commit_msg}"  # custom message
nownow hook list                             # list installed hooks
nownow hook remove                           # remove all nownow hooks
```

Hooks are appended to existing hook files (never overwritten) and managed via `# nownow:start` / `# nownow:end` markers. Works with worktrees and submodules.

**Default messages:**
- `post-commit`: "Just committed: {commit_msg}"
- `pre-push`: "Pushing to {branch}"

## Command Wrapper

Run any command and push its outcome as status:

```bash
nownow wrap -- make build                   # "make completed" or "make failed (exit 2)"
nownow wrap --name "Deploy" -- ./deploy.sh  # "Deploy completed"
nownow wrap --on-success "Ship it!" --on-failure "Broke it ({exit_code})" -- make test
nownow wrap --quiet -- backup.sh            # push without printing nownow output
```

**Template variables:** `{cmd}`, `{name}`, `{exit_code}`, `{duration}`

The wrapped command's stdin/stdout/stderr are fully transparent, and its exit code is preserved.

## Development

```bash
go build -o nownow .
go test ./...
```

## License

MIT
