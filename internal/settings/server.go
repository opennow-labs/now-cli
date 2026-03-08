package settings

import (
	_ "embed"
	"encoding/json"
	"log"
	"net"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
	"time"

	"github.com/nownow-labs/nownow/internal/api"
	"github.com/nownow-labs/nownow/internal/config"
	"github.com/nownow-labs/nownow/internal/open"
	"github.com/nownow-labs/nownow/internal/upgrade"
)

//go:embed ui.html
var uiHTML []byte

const ListenAddr = "127.0.0.1:19191"

var (
	mu      sync.Mutex
	version string

	// Autostart functions injected by caller to avoid import cycle with daemon.
	AutostartIsInstalled func() bool
	AutostartInstall     func() error
	AutostartUninstall   func() error
)

func Start(ver string) error {
	version = ver

	mux := NewMux()

	ln, err := net.Listen("tcp", ListenAddr)
	if err != nil {
		return err
	}

	go func() {
		log.Printf("settings UI listening on http://%s", ListenAddr)
		if err := http.Serve(ln, localOriginOnly(mux)); err != nil {
			log.Printf("settings server: %v", err)
		}
	}()
	return nil
}

// localOriginOnly rejects cross-origin requests to prevent CSRF attacks.
func localOriginOnly(next http.Handler) http.Handler {
	allowed := "http://" + ListenAddr
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		origin := r.Header.Get("Origin")
		if origin != "" && origin != allowed {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

// NewMux creates the HTTP handler for testing or embedding.
func NewMux() *http.ServeMux {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", handleUI)
	mux.HandleFunc("GET /api/config", handleGetConfig)
	mux.HandleFunc("PUT /api/config", handlePutConfig)
	mux.HandleFunc("GET /api/version", handleGetVersion)
	mux.HandleFunc("POST /api/check-update", handleCheckUpdate)
	mux.HandleFunc("POST /api/open-config", handleOpenConfig)
	mux.HandleFunc("GET /api/autostart", handleGetAutostart)
	mux.HandleFunc("POST /api/autostart", handlePostAutostart)
	mux.HandleFunc("POST /api/login", handleLogin)
	mux.HandleFunc("POST /api/login/poll", handleLoginPoll)
	mux.HandleFunc("POST /api/logout", handleLogout)
	mux.HandleFunc("GET /api/permissions", handleGetPermissions)
	mux.HandleFunc("POST /api/open-accessibility", handleOpenAccessibility)
	return mux
}

type configResponse struct {
	Endpoint      string               `json:"endpoint"`
	Token         string               `json:"token"`
	TokenSet      bool                 `json:"token_set"`
	Template      string               `json:"template"`
	Interval      string               `json:"interval"`
	ActivityRules []config.ActivityRule `json:"activity_rules"`
	Ignore        []string             `json:"ignore"`
	Telemetry     bool                 `json:"telemetry"`
	SendApp       bool                 `json:"send_app"`
	SendMusic     bool                 `json:"send_music"`
	SendWatching  bool                 `json:"send_watching"`
	AutoUpdate    bool                 `json:"auto_update"`
}

func maskToken(token string) string {
	if len(token) <= 8 {
		return ""
	}
	return "..." + token[len(token)-4:]
}

func buildConfigResponse(cfg config.Config) configResponse {
	return configResponse{
		Endpoint:      cfg.Endpoint,
		Token:         maskToken(cfg.Token),
		TokenSet:      cfg.Token != "",
		Template:      cfg.Template,
		Interval:      cfg.Interval,
		ActivityRules: cfg.ActivityRules,
		Ignore:        cfg.Ignore,
		Telemetry:     cfg.TelemetryEnabled(),
		SendApp:       cfg.SendAppEnabled(),
		SendMusic:     cfg.SendMusicEnabled(),
		SendWatching:  cfg.SendWatchingEnabled(),
		AutoUpdate:    cfg.AutoUpdateEnabled(),
	}
}

func handleUI(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/" {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write(uiHTML)
}

func handleGetConfig(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildConfigResponse(cfg))
}

type configUpdate struct {
	Template     *string `json:"template"`
	Interval     *string `json:"interval"`
	SendApp      *bool   `json:"send_app"`
	SendMusic    *bool   `json:"send_music"`
	SendWatching *bool   `json:"send_watching"`
	Telemetry    *bool   `json:"telemetry"`
	AutoUpdate   *bool   `json:"auto_update"`
}

func handlePutConfig(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	var update configUpdate
	if err := json.NewDecoder(r.Body).Decode(&update); err != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	if update.Template != nil {
		cfg.Template = *update.Template
	}
	if update.Interval != nil {
		if _, err := time.ParseDuration(*update.Interval); err != nil {
			http.Error(w, "invalid interval", 400)
			return
		}
		cfg.Interval = *update.Interval
	}
	if update.SendApp != nil {
		cfg.SendApp = update.SendApp
	}
	if update.SendMusic != nil {
		cfg.SendMusic = update.SendMusic
	}
	if update.SendWatching != nil {
		cfg.SendWatching = update.SendWatching
	}
	if update.Telemetry != nil {
		cfg.Telemetry = update.Telemetry
	}
	if update.AutoUpdate != nil {
		cfg.AutoUpdate = update.AutoUpdate
	}

	if err := config.Save(cfg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	cfg, _ = config.Load()
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(buildConfigResponse(cfg))
}

func handleGetVersion(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"version": version,
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
	})
}

func handleCheckUpdate(w http.ResponseWriter, r *http.Request) {
	release, err := upgrade.CheckLatest()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"error": err.Error(),
		})
		return
	}

	latest := upgrade.NormalizeVersion(release.TagName)
	current := upgrade.NormalizeVersion(version)
	newer := upgrade.IsNewer(current, latest)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"current":          current,
		"latest":           latest,
		"update_available": newer,
	})
}

func handleOpenConfig(w http.ResponseWriter, r *http.Request) {
	p, err := config.Path()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	if err := open.File(p); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func handleGetAutostart(w http.ResponseWriter, r *http.Request) {
	installed := false
	if AutostartIsInstalled != nil {
		installed = AutostartIsInstalled()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"installed": installed,
	})
}

func handlePostAutostart(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Enabled bool `json:"enabled"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}

	var err error
	if body.Enabled {
		if AutostartInstall != nil {
			err = AutostartInstall()
		}
	} else {
		if AutostartUninstall != nil {
			err = AutostartUninstall()
		}
	}
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	installed := false
	if AutostartIsInstalled != nil {
		installed = AutostartIsInstalled()
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]bool{
		"installed": installed,
	})
}

func handleLogin(w http.ResponseWriter, r *http.Request) {
	cfg, err := config.Load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	client := api.NewClient(cfg.Endpoint, "")
	deviceResp, err := client.RequestDeviceCode()
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(500)
		json.NewEncoder(w).Encode(map[string]interface{}{"error": err.Error()})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"device_code":      deviceResp.DeviceCode,
		"user_code":        deviceResp.UserCode,
		"verification_url": deviceResp.VerificationURL,
		"expires_in":       deviceResp.ExpiresIn,
		"interval":         deviceResp.Interval,
	})
}

func handleLoginPoll(w http.ResponseWriter, r *http.Request) {
	var body struct {
		DeviceCode string `json:"device_code"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, "invalid JSON", 400)
		return
	}

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	client := api.NewClient(cfg.Endpoint, "")
	tokenResp, err := client.PollDeviceToken(body.DeviceCode)
	if err != nil {
		status := "pending"
		if _, ok := err.(*api.AuthPendingError); !ok {
			status = "error"
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"status": status,
			"error":  err.Error(),
		})
		return
	}

	// Save token — re-load inside lock to avoid overwriting concurrent changes
	mu.Lock()
	cfg, err = config.Load()
	if err == nil {
		cfg.Token = tokenResp.Token
		err = config.Save(cfg)
	}
	mu.Unlock()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"status":    "ok",
		"user_name": tokenResp.User.Name,
	})
}

func handleLogout(w http.ResponseWriter, r *http.Request) {
	mu.Lock()
	defer mu.Unlock()

	cfg, err := config.Load()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	cfg.Token = ""
	if err := config.Save(cfg); err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func handleGetPermissions(w http.ResponseWriter, r *http.Request) {
	resp := map[string]interface{}{
		"os":            runtime.GOOS,
		"accessibility": true,
	}

	if runtime.GOOS == "darwin" {
		// Test accessibility by asking System Events for the frontmost process name.
		err := exec.Command("osascript", "-e",
			`tell application "System Events" to get name of first application process whose frontmost is true`).Run()
		resp["accessibility"] = err == nil
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(resp)
}

func handleOpenAccessibility(w http.ResponseWriter, r *http.Request) {
	if runtime.GOOS != "darwin" {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"ok":false,"error":"not macOS"}`))
		return
	}

	err := exec.Command("open", "x-apple.systempreferences:com.apple.preference.security?Privacy_Accessibility").Run()
	if err != nil {
		http.Error(w, err.Error(), 500)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}
