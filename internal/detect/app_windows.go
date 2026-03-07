//go:build windows

package detect

import (
	"os/exec"
	"strings"
)

// detectApp returns the focused application and window title on Windows via PowerShell.
func detectApp() (app, title string) {
	// Get foreground window process name
	out, err := exec.Command("powershell", "-NoProfile", "-Command",
		`(Get-Process | Where-Object {$_.MainWindowHandle -eq (Add-Type -MemberDefinition '[DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();' -Name Win32 -Namespace Temp -PassThru)::GetForegroundWindow()}).ProcessName`).Output()
	if err == nil {
		app = strings.TrimSpace(string(out))
	}

	// Get window title
	out, err = exec.Command("powershell", "-NoProfile", "-Command",
		`(Get-Process | Where-Object {$_.MainWindowHandle -eq (Add-Type -MemberDefinition '[DllImport("user32.dll")] public static extern IntPtr GetForegroundWindow();' -Name Win32 -Namespace Temp -PassThru)::GetForegroundWindow()}).MainWindowTitle`).Output()
	if err == nil {
		title = strings.TrimSpace(string(out))
	}

	return app, title
}
