package tray

import _ "embed"

// icon_dark.png: black icon for light mode
// icon_light.png: white icon for dark mode
// SetTemplateIcon(iconBytes, selectedIconBytes) — macOS uses template rendering

//go:embed icon_light.png
var IconLight []byte

//go:embed icon_dark.png
var IconDark []byte
