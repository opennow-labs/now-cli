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
	Artist   string `json:"artist"`
	Title    string `json:"title"`
	Album    string `json:"album"`
	SourceID string `json:"source_id"`
	Status   string `json:"status"`
}

func detectMedia() MediaResult {
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
@{ artist = $props.Artist; title = $props.Title; album = $props.AlbumTitle; source_id = $session.SourceAppUserModelId; status = $status } | ConvertTo-Json -Compress
`
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	out, err := exec.CommandContext(ctx, "powershell", "-NoProfile", "-ExecutionPolicy", "Bypass", "-Command", script).Output()
	if err != nil {
		return MediaResult{}
	}
	raw := strings.TrimSpace(string(out))
	if raw == "" {
		return MediaResult{}
	}
	var result gsmtcResult
	if err := json.Unmarshal([]byte(raw), &result); err != nil {
		return MediaResult{}
	}

	return ClassifyMedia(&MediaInfo{
		Title:     result.Title,
		Artist:    result.Artist,
		Album:     result.Album,
		SourceID:  result.SourceID,
		IsPlaying: result.Status == "Playing",
	})
}
