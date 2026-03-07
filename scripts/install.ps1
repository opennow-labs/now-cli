$ErrorActionPreference = "Stop"

$Repo = "nownow-labs/nownow"
$Binary = "nownow"

# Detect architecture
$Arch = switch ($env:PROCESSOR_ARCHITECTURE) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { Write-Error "Unsupported architecture: $env:PROCESSOR_ARCHITECTURE"; exit 1 }
}

# Get latest release
$Release = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
$Version = $Release.tag_name
Write-Host "Installing $Binary $Version (windows/$Arch)..."

# Download
$AssetName = "${Binary}_windows_${Arch}.zip"
$Asset = $Release.assets | Where-Object { $_.name -eq $AssetName }
if (-not $Asset) {
    Write-Error "No release asset found: $AssetName"
    exit 1
}

$InstallDir = Join-Path $env:LOCALAPPDATA $Binary
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$TempZip = Join-Path $env:TEMP "$Binary.zip"
Invoke-WebRequest -Uri $Asset.browser_download_url -OutFile $TempZip

# Extract
Expand-Archive -Path $TempZip -DestinationPath $InstallDir -Force
Remove-Item $TempZip -Force

# Add to PATH if not already there
$UserPath = [Environment]::GetEnvironmentVariable("PATH", "User")
if ($UserPath -notlike "*$InstallDir*") {
    [Environment]::SetEnvironmentVariable("PATH", "$UserPath;$InstallDir", "User")
    $env:PATH = "$env:PATH;$InstallDir"
    Write-Host "Added $InstallDir to PATH"
}

# Verify
$Installed = & (Join-Path $InstallDir "$Binary.exe") version 2>&1
Write-Host "Installed $Installed"
Write-Host ""
Write-Host "Next steps:"
Write-Host "  nownow login"
