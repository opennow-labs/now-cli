package upgrade

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

func TestBackgroundCheckerNotifies(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v1.0.0","assets":[]}`))
	}))
	defer server.Close()

	origURL := releasesURL
	setReleasesURL(server.URL)
	defer setReleasesURL(origURL)

	var called atomic.Int32
	checker := NewBackgroundChecker("0.1.0", func(r *Release) {
		called.Add(1)
	})

	// Call check directly instead of Start to avoid timing issues
	checker.check()

	if called.Load() != 1 {
		t.Errorf("expected onUpdate called once, got %d", called.Load())
	}

	if checker.Latest() == nil {
		t.Fatal("Latest() returned nil after finding update")
	}
	if checker.Latest().TagName != "v1.0.0" {
		t.Errorf("Latest().TagName = %q, want %q", checker.Latest().TagName, "v1.0.0")
	}
}

func TestBackgroundCheckerNoUpdateWhenCurrent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v1.0.0","assets":[]}`))
	}))
	defer server.Close()

	origURL := releasesURL
	setReleasesURL(server.URL)
	defer setReleasesURL(origURL)

	var called atomic.Int32
	checker := NewBackgroundChecker("1.0.0", func(r *Release) {
		called.Add(1)
	})

	checker.check()

	if called.Load() != 0 {
		t.Errorf("expected onUpdate not called, got %d", called.Load())
	}
}

func TestBackgroundCheckerDeduplicates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"tag_name":"v2.0.0","assets":[]}`))
	}))
	defer server.Close()

	origURL := releasesURL
	setReleasesURL(server.URL)
	defer setReleasesURL(origURL)

	var called atomic.Int32
	checker := NewBackgroundChecker("1.0.0", func(r *Release) {
		called.Add(1)
	})

	checker.check()
	checker.check()
	checker.check()

	if called.Load() != 1 {
		t.Errorf("expected onUpdate called once (deduplicated), got %d", called.Load())
	}
}

func TestBackgroundCheckerStartCancellation(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	checker := NewBackgroundChecker("1.0.0", nil)

	done := make(chan struct{})
	go func() {
		checker.Start(ctx)
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(2 * time.Second):
		t.Fatal("Start did not return after context cancellation")
	}
}
