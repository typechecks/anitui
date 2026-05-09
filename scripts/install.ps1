param(
	[string]$InstallDir = "$HOME\.anitui"
)

$ErrorActionPreference = "Stop"
$repo = "typechecks/anitui"

if (-not (Test-Path $InstallDir)) {
	New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

Write-Host "Detecting latest version..." -ForegroundColor Cyan
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$tag = $release.tag_name

if (-not $tag) {
	Write-Error "Could not find latest release tag"
	exit 1
}

$arch = "amd64"
$binaryName = "anitui_windows_$arch.zip"
$asset = $release.assets | Where-Object { $_.name -eq $binaryName }

if (-not $asset) {
	Write-Error "Could not find asset $binaryName for version $tag"
	exit 1
}

$url = $asset.browser_download_url
$zipPath = "$env:TEMP\anitui.zip"

Write-Host "Downloading anitui $tag..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $url -OutFile $zipPath

Write-Host "Extracting to $InstallDir..." -ForegroundColor Cyan
Expand-Archive -Path $zipPath -DestinationPath $InstallDir -Force

$exePath = Join-Path $InstallDir "anitui.exe"
Remove-Item "$exePath.old" -ErrorAction SilentlyContinue

if ($InstallDir -eq "$HOME\.anitui") {
	$path = [Environment]::GetEnvironmentVariable("Path", "User")
	if ($path -notlike "*$InstallDir*") {
		Write-Host "Adding $InstallDir to PATH..." -ForegroundColor Cyan
		[Environment]::SetEnvironmentVariable("Path", $path + ";$InstallDir", "User")
		$env:Path += ";$InstallDir"
	}
}

Remove-Item $zipPath
Write-Host "Done! Restart your terminal and run 'anitui' to start." -ForegroundColor Green
