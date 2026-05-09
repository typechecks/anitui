# anitui install script for Windows
# downloads the latest binary and adds to path

$repo = "typechecks/anitui"
$installDir = "$HOME\.anitui"

if (!(Test-Path $installDir)) {
    New-Item -ItemType Directory -Force -Path $installDir
}

Write-Host "Detecting latest version..." -ForegroundColor Cyan
$release = Invoke-RestMethod -Uri "https://api.github.com/repos/$repo/releases/latest"
$tag = $release.tag_name

if (!$tag) {
    Write-Error "Could not find latest release tag"
    exit
}

$arch = "amd64" # Default to amd64 for Windows
$binaryName = "anitui_windows_$arch.zip"
$asset = $release.assets | Where-Object { $_.name -eq $binaryName }

if (!$asset) {
    Write-Error "Could not find asset $binaryName for version $tag"
    exit
}

$url = $asset.browser_download_url
$zipPath = "$env:TEMP\anitui.zip"

Write-Host "Downloading anitui $tag..." -ForegroundColor Cyan
Invoke-WebRequest -Uri $url -OutFile $zipPath

Write-Host "Extracting to $installDir..." -ForegroundColor Cyan
Expand-Archive -Path $zipPath -DestinationPath $installDir -Force

$exePath = Join-Path $installDir "anitui.exe"

# Add to User PATH if not already there
$path = [Environment]::GetEnvironmentVariable("Path", "User")
if ($path -notlike "*$installDir*") {
    Write-Host "Adding $installDir to PATH..." -ForegroundColor Cyan
    [Environment]::SetEnvironmentVariable("Path", $path + ";$installDir", "User")
    $env:Path += ";$installDir"
}

Remove-Item $zipPath
Write-Host "Done! Restart your terminal and run 'anitui' to start." -ForegroundColor Green
