package upgrade

import (
	"context"
	"sync"
	"time"
)

// BackgroundChecker periodically checks for new releases.
type BackgroundChecker struct {
	interval time.Duration
	current  string
	onUpdate func(*Release)

	mu     sync.Mutex
	latest *Release
}

// NewBackgroundChecker creates a checker that calls onUpdate when a newer version is found.
func NewBackgroundChecker(current string, onUpdate func(*Release)) *BackgroundChecker {
	return &BackgroundChecker{
		interval: 1 * time.Hour,
		current:  current,
		onUpdate: onUpdate,
	}
}

// Start begins periodic checking. It delays 10 seconds before the first check.
// Blocks until ctx is cancelled.
func (c *BackgroundChecker) Start(ctx context.Context) {
	// Initial delay to avoid startup storm
	select {
	case <-time.After(10 * time.Second):
	case <-ctx.Done():
		return
	}

	c.check()

	ticker := time.NewTicker(c.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.check()
		case <-ctx.Done():
			return
		}
	}
}

func (c *BackgroundChecker) check() {
	release, err := CheckLatest()
	if err != nil {
		return
	}

	latest := NormalizeVersion(release.TagName)
	current := NormalizeVersion(c.current)

	if !IsNewer(current, latest) {
		return
	}

	c.mu.Lock()
	alreadyNotified := c.latest != nil && c.latest.TagName == release.TagName
	c.latest = release
	c.mu.Unlock()

	if !alreadyNotified && c.onUpdate != nil {
		c.onUpdate(release)
	}
}

// Latest returns the most recently discovered release, if any.
func (c *BackgroundChecker) Latest() *Release {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.latest
}
