package tray

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"time"

	"fyne.io/systray"
	"github.com/nownow-labs/nownow/internal/api"
	"github.com/nownow-labs/nownow/internal/config"
	"github.com/nownow-labs/nownow/internal/detect"
	"github.com/nownow-labs/nownow/internal/open"
	"github.com/nownow-labs/nownow/internal/template"
	"github.com/nownow-labs/nownow/internal/upgrade"
)

// Version is set by the caller before Run.
var Version = "dev"

// RestartFunc is called to restart the daemon after an upgrade.
// Set by the caller before Run to avoid import cycle with daemon package.
var RestartFunc func() error

// SettingsAvailable indicates whether the settings HTTP server started successfully.
// When false, "Settings..." falls back to opening config.yml in an editor.
var SettingsAvailable bool

// Run starts the systray menubar and push loop.
// This function blocks until the user quits.
func Run(interval time.Duration) {
	systray.Run(func() { onReady(interval) }, onExit)
}

var (
	mu         sync.Mutex
	paused     bool
	lastStatus string
	mStatus    *systray.MenuItem
	mMusic     *systray.MenuItem
	mPause     *systray.MenuItem
	mUpdate    *systray.MenuItem

	updateCancel context.CancelFunc
)

func onReady(interval time.Duration) {
	// SetTemplateIcon: macOS treats the image as a template (auto light/dark)
	// First arg = icon, second arg = selected icon
	// For template to work, the PNG must be black + alpha only
	systray.SetTemplateIcon(IconDark, IconDark)
	systray.SetTooltip("nownow")

	mStatus = systray.AddMenuItem("starting...", "Current status")
	mStatus.Disable()

	mMusic = systray.AddMenuItem("", "Now playing")
	mMusic.Disable()
	mMusic.Hide()

	systray.AddSeparator()

	mPause = systray.AddMenuItem("Pause", "Pause auto-detection")
	mSettings := systray.AddMenuItem("Settings...", "Open settings")
	mBoard := systray.AddMenuItem("Open Board", "Open now.ctx.st in browser")

	systray.AddSeparator()

	mUpdate = systray.AddMenuItem("", "Update available")
	mUpdate.Hide()

	systray.AddSeparator()
	mQuit := systray.AddMenuItem("Quit", "Stop nownow")

	// Initial push
	pushAndUpdate()

	// Start background update checker (only if auto-update enabled)
	cfg2, _ := config.Load()
	var ctx context.Context
	ctx, updateCancel = context.WithCancel(context.Background())
	checker := upgrade.NewBackgroundChecker(Version, func(release *upgrade.Release) {
		v := upgrade.NormalizeVersion(release.TagName)
		mUpdate.SetTitle(fmt.Sprintf("\u2193 Update available (v%s)", v))
		mUpdate.Show()
	})
	if cfg2.AutoUpdateEnabled() {
		go checker.Start(ctx)
	}

	// Push loop
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				mu.Lock()
				p := paused
				mu.Unlock()
				if !p {
					pushAndUpdate()
				}
			case <-mPause.ClickedCh:
				mu.Lock()
				paused = !paused
				if paused {
					mPause.SetTitle("Resume")
					mStatus.SetTitle("paused")
					systray.SetTitle("⏸")
				} else {
					mPause.SetTitle("Pause")
					systray.SetTitle("")
					pushAndUpdate()
				}
				mu.Unlock()
			case <-mSettings.ClickedCh:
				if SettingsAvailable {
					open.URL("http://127.0.0.1:19191")
				} else {
					if p, err := config.Path(); err == nil {
						open.File(p)
					}
				}
			case <-mBoard.ClickedCh:
				open.URL("https://now.ctx.st")
			case <-mUpdate.ClickedCh:
				go performUpgrade(checker)
			case <-mQuit.ClickedCh:
				systray.Quit()
				return
			}
		}
	}()
}

func onExit() {
	if updateCancel != nil {
		updateCancel()
	}
}

func performUpgrade(checker *upgrade.BackgroundChecker) {
	// Disable immediately to prevent concurrent clicks
	mUpdate.Disable()

	release := checker.Latest()
	if release == nil {
		mUpdate.Enable()
		return
	}

	mUpdate.SetTitle("Downloading update...")

	asset, err := upgrade.FindAsset(release)
	if err != nil {
		mUpdate.SetTitle("Update failed: no asset")
		mUpdate.Enable()
		return
	}

	execPath, err := os.Executable()
	if err != nil {
		mUpdate.SetTitle("Update failed")
		mUpdate.Enable()
		return
	}

	if err := upgrade.Download(asset, execPath); err != nil {
		mUpdate.SetTitle("Update failed: download error")
		mUpdate.Enable()
		return
	}

	mUpdate.SetTitle("Restarting...")
	if RestartFunc != nil {
		if err := RestartFunc(); err != nil {
			mUpdate.SetTitle("Restart failed")
			mUpdate.Enable()
		}
	}
}

func pushAndUpdate() {
	cfg, err := config.Load()
	if err != nil {
		updateStatus("config error", "")
		return
	}
	if !cfg.HasToken() {
		updateStatus("not logged in", "")
		return
	}

	ctx := detect.Detect()

	if cfg.IsIgnored(ctx.App) {
		return
	}

	// Sanitize context before rendering so privacy-disabled fields
	// never leak into activity/content strings.
	if !cfg.SendMusicEnabled() {
		ctx.MusicArtist = ""
		ctx.MusicTrack = ""
	}
	if !cfg.SendWatchingEnabled() {
		ctx.Watching = ""
	}
	if !cfg.SendAppEnabled() {
		ctx.App = ""
	}

	activity := cfg.ResolveActivity(ctx.App, ctx.Watching, ctx.Music())

	content := template.Render(cfg.Template, ctx, activity)
	if content == "" {
		return
	}

	client := api.NewClient(cfg.Endpoint, cfg.Token)
	client.Version = Version
	client.Telemetry = cfg.TelemetryEnabled()
	client.SendApp = cfg.SendAppEnabled()
	client.SendMusic = cfg.SendMusicEnabled()
	client.SendWatching = cfg.SendWatchingEnabled()
	err = client.PushStatus(api.StatusRequest{
		Content:     content,
		App:         ctx.App,
		Activity:    activity,
		MusicArtist: ctx.MusicArtist,
		MusicTrack:  ctx.MusicTrack,
		Watching:    ctx.Watching,
	})
	if err != nil {
		var rle *api.RateLimitError
		if errors.As(err, &rle) {
			updateStatus("rate limited", "")
			return
		}
		updateStatus("push error", "")
		return
	}

	music := ""
	if ctx.HasMusic() {
		music = fmt.Sprintf("\U0001F3B5 %s", ctx.Music())
	}
	updateStatus(content, music)
}

func updateStatus(status, music string) {
	mu.Lock()
	defer mu.Unlock()

	lastStatus = status
	mStatus.SetTitle(status)

	if music != "" {
		mMusic.SetTitle(music)
		mMusic.Show()
	} else {
		mMusic.Hide()
	}
}

