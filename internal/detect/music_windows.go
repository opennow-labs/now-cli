//go:build windows

package detect

import (
	"context"
	"encoding/json"
	"os/exec"
	"strings"
	"time"
)

type gsmtcResult struct {
	Artist string `json:"artist"`
	Title  string `json:"title"`
	Status string `json:"status"`
}

func detectMusic() (artist, track string) {
	script := `
Add-Type -AssemblyName System.Runtime.WindowsRuntime
$null = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager, Windows.Media.Control, ContentType=WindowsRuntime]
$async = [Windows.Media.Control.GlobalSystemMediaTransportControlsSessionManager]::RequestAsync()
$op = [System.WindowsRuntimeSystemExtensions]::AsTask($async)
$op.Wait()
$mgr = $op.Result
$session = $mgr.GetCurrentSession()
if ($null -eq $session) { exit 0 }
$infoAsync = $session.TryGetMediaPropertiesAsync()
$infoOp = [System.WindowsRuntimeSystemExtensions]::AsTask($infoAsync)
$infoOp.Wait()
$props = $infoOp.Result
$info = $session.GetPlaybackInfo()
$status = $info.PlaybackStatus.ToString()
@{ artist = $props.Artist; title = $props.Title; status = $status } | ConvertTo-Json -Compress
`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
	if err != nil {
		return "", ""
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return "", ""
	}
	var result gsmtcResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return "", ""
	}
	if result.Status != "Playing" {
		return "", ""
	}
	return result.Artist, result.Title
}
