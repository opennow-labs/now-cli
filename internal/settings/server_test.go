package settings

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/nownow-labs/nownow/internal/config"
)

func setup(t *testing.T) *httptest.Server {
	t.Helper()
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)
	version = "1.0.0"

	// Reset autostart functions for tests
	AutostartIsInstalled = func() bool { return false }
	AutostartInstall = func() error { return nil }
	AutostartUninstall = func() error { return nil }

	return httptest.NewServer(NewMux())
}

func saveTestConfig(t *testing.T, cfg config.Config) {
	t.Helper()
	if err := config.Save(cfg); err != nil {
		t.Fatalf("Save: %v", err)
	}
}

func TestServeUI(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/html") {
		t.Errorf("Content-Type = %q, want text/html", ct)
	}
}

func TestGetConfig_MasksToken(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_abcdef1234"
	saveTestConfig(t, cfg)

	resp, err := http.Get(srv.URL + "/api/config")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result configResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Token != "...1234" {
		t.Errorf("token = %q, want ...1234", result.Token)
	}
	if !result.TokenSet {
		t.Error("token_set should be true")
	}
}

func TestGetConfig_NoToken(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	saveTestConfig(t, cfg)

	resp, err := http.Get(srv.URL + "/api/config")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result configResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Token != "" {
		t.Errorf("token = %q, want empty", result.Token)
	}
	if result.TokenSet {
		t.Error("token_set should be false")
	}
}

func TestGetConfig_ReturnsAllFields(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test123"
	cfg.Interval = "1m"
	cfg.Template = "{activity}"
	saveTestConfig(t, cfg)

	resp, err := http.Get(srv.URL + "/api/config")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result configResponse
	json.NewDecoder(resp.Body).Decode(&result)

	if result.Interval != "1m" {
		t.Errorf("interval = %q, want 1m", result.Interval)
	}
	if result.Template != "{activity}" {
		t.Errorf("template = %q, want {activity}", result.Template)
	}
	if !result.SendApp {
		t.Error("send_app should default to true")
	}
	if !result.SendMusic {
		t.Error("send_music should default to true")
	}
	if !result.SendWatching {
		t.Error("send_watching should default to true")
	}
	if !result.Telemetry {
		t.Error("telemetry should default to true")
	}
	if !result.AutoUpdate {
		t.Error("auto_update should default to true")
	}
}

func TestPutConfig_UpdatesPrivacy(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test"
	saveTestConfig(t, cfg)

	body := `{"send_app": false}`
	req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Fatalf("status = %d", resp.StatusCode)
	}

	// Verify on disk
	loaded, _ := config.Load()
	if loaded.SendAppEnabled() {
		t.Error("send_app should be false on disk")
	}
}

func TestPutConfig_UpdatesInterval(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test"
	saveTestConfig(t, cfg)

	body := `{"interval": "2m"}`
	req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	loaded, _ := config.Load()
	if loaded.Interval != "2m" {
		t.Errorf("interval = %q, want 2m", loaded.Interval)
	}
}

func TestPutConfig_RejectsInvalidInterval(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test"
	saveTestConfig(t, cfg)

	body := `{"interval": "abc"}`
	req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 400 {
		t.Errorf("status = %d, want 400 for invalid interval", resp.StatusCode)
	}

	loaded, _ := config.Load()
	if loaded.Interval != "30s" {
		t.Errorf("interval changed to %q, should remain 30s", loaded.Interval)
	}
}

func TestPutConfig_UpdatesTemplate(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test"
	saveTestConfig(t, cfg)

	body := `{"template": "{app} · {music}"}`
	req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	loaded, _ := config.Load()
	if loaded.Template != "{app} · {music}" {
		t.Errorf("template = %q, want {app} · {music}", loaded.Template)
	}
}

func TestPutConfig_RejectsToken(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_original"
	saveTestConfig(t, cfg)

	body := `{"token": "now_evil"}`
	req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	loaded, _ := config.Load()
	if loaded.Token != "now_original" {
		t.Errorf("token changed to %q, should remain now_original", loaded.Token)
	}
}

func TestPutConfig_RejectsEndpoint(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test"
	saveTestConfig(t, cfg)

	body := `{"endpoint": "https://evil.com"}`
	req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	loaded, _ := config.Load()
	if loaded.Endpoint != "https://now.ctx.st" {
		t.Errorf("endpoint changed to %q, should remain default", loaded.Endpoint)
	}
}

func TestPutConfig_PartialUpdate(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test"
	cfg.Interval = "1m"
	cfg.Template = "{activity}"
	saveTestConfig(t, cfg)

	body := `{"send_music": false}`
	req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	loaded, _ := config.Load()
	if loaded.Interval != "1m" {
		t.Errorf("interval changed to %q, should remain 1m", loaded.Interval)
	}
	if loaded.Template != "{activity}" {
		t.Errorf("template changed to %q, should remain {activity}", loaded.Template)
	}
	if loaded.SendMusicEnabled() {
		t.Error("send_music should be false")
	}
	if !loaded.SendAppEnabled() {
		t.Error("send_app should still be true")
	}
}

func TestPutConfig_ConcurrentWrites(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test"
	saveTestConfig(t, cfg)

	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			var body string
			switch i % 3 {
			case 0:
				body = `{"send_app": false}`
			case 1:
				body = `{"send_music": false}`
			case 2:
				body = `{"interval": "1m"}`
			}
			req, _ := http.NewRequest("PUT", srv.URL+"/api/config", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			resp, err := http.DefaultClient.Do(req)
			if err != nil {
				t.Errorf("concurrent write %d failed: %v", i, err)
				return
			}
			resp.Body.Close()
		}(i)
	}
	wg.Wait()

	// Just verify no corruption - config should load fine
	_, err := config.Load()
	if err != nil {
		t.Fatalf("config corrupted after concurrent writes: %v", err)
	}
}

func TestGetVersion(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/version")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result map[string]string
	json.NewDecoder(resp.Body).Decode(&result)

	if result["version"] != "1.0.0" {
		t.Errorf("version = %q, want 1.0.0", result["version"])
	}
	if result["os"] == "" {
		t.Error("os should not be empty")
	}
	if result["arch"] == "" {
		t.Error("arch should not be empty")
	}
}

func TestCheckUpdate(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	resp, err := http.Post(srv.URL+"/api/check-update", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	// Response should have a known shape (either error or version info)
	var result map[string]interface{}
	json.NewDecoder(resp.Body).Decode(&result)
	if _, ok := result["error"]; !ok {
		if _, ok := result["current"]; !ok {
			t.Error("expected either 'error' or 'current' in response")
		}
	}
}

func TestGetAutostart(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	resp, err := http.Get(srv.URL + "/api/autostart")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()

	var result map[string]bool
	json.NewDecoder(resp.Body).Decode(&result)
	if _, ok := result["installed"]; !ok {
		t.Error("expected 'installed' field in response")
	}
}

func TestPostAutostart(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	installed := false
	AutostartIsInstalled = func() bool { return installed }
	AutostartInstall = func() error { installed = true; return nil }
	AutostartUninstall = func() error { installed = false; return nil }

	// Enable
	body := bytes.NewReader([]byte(`{"enabled": true}`))
	resp, err := http.Post(srv.URL+"/api/autostart", "application/json", body)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	var result map[string]bool
	resp, _ = http.Get(srv.URL + "/api/autostart")
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if !result["installed"] {
		t.Error("expected installed=true after enabling")
	}

	// Disable
	body = bytes.NewReader([]byte(`{"enabled": false}`))
	resp, _ = http.Post(srv.URL+"/api/autostart", "application/json", body)
	resp.Body.Close()

	resp, _ = http.Get(srv.URL + "/api/autostart")
	result = map[string]bool{}
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()
	if result["installed"] {
		t.Error("expected installed=false after disabling")
	}
}

func TestMaskToken(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"now_abcdef1234", "...1234"},
		{"now_ab", ""},
		{"abcd", ""},
		{"abcdefghi", "...fghi"},
		{"12345678", ""},
		{"abc", ""},
		{"", ""},
	}
	for _, tt := range tests {
		got := maskToken(tt.in)
		if got != tt.want {
			t.Errorf("maskToken(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestLogout(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test_token"
	saveTestConfig(t, cfg)

	resp, err := http.Post(srv.URL+"/api/logout", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}

	loaded, _ := config.Load()
	if loaded.Token != "" {
		t.Errorf("token should be empty after logout, got %q", loaded.Token)
	}
}

func TestLogout_AlreadyLoggedOut(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	saveTestConfig(t, cfg)

	resp, err := http.Post(srv.URL+"/api/logout", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("status = %d, want 200", resp.StatusCode)
	}
}

func TestLogout_ConfigReflectsChange(t *testing.T) {
	srv := setup(t)
	defer srv.Close()

	cfg := config.DefaultConfig()
	cfg.Token = "now_test_token"
	saveTestConfig(t, cfg)

	// Logout
	resp, _ := http.Post(srv.URL+"/api/logout", "application/json", nil)
	resp.Body.Close()

	// GET /api/config should show not logged in
	resp, _ = http.Get(srv.URL + "/api/config")
	var result configResponse
	json.NewDecoder(resp.Body).Decode(&result)
	resp.Body.Close()

	if result.TokenSet {
		t.Error("token_set should be false after logout")
	}
	if result.Token != "" {
		t.Errorf("token should be empty after logout, got %q", result.Token)
	}
}

// Verify config file path uses test dir
func TestConfigIsolation(t *testing.T) {
	tmpDir := t.TempDir()
	t.Setenv("XDG_CONFIG_HOME", tmpDir)

	cfg := config.DefaultConfig()
	cfg.Token = "now_isolation_test"
	config.Save(cfg)

	p, _ := config.Path()
	if !strings.HasPrefix(p, tmpDir) {
		t.Errorf("config path %q should be under %q", p, tmpDir)
	}

	_, err := os.Stat(filepath.Join(tmpDir, "nownow", "config.yml"))
	if err != nil {
		t.Errorf("config file should exist: %v", err)
	}
}
