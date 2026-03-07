package open

import (
	"os/exec"
	"runtime"
)

// URL opens a URL in the default browser.
func URL(url string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", url).Start()
	case "linux":
		return exec.Command("xdg-open", url).Start()
	case "windows":
		return exec.Command("cmd", "/c", "start", "", url).Start()
	}
	return nil
}

// File opens a file in the default application/editor.
func File(path string) error {
	switch runtime.GOOS {
	case "darwin":
		return exec.Command("open", path).Start()
	case "linux":
		return exec.Command("xdg-open", path).Start()
	case "windows":
		return exec.Command("notepad", path).Start()
	}
	return nil
}
