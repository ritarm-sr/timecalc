$repo      = "ritarm-sr/timecalc"
$exeName   = "timecalc.exe"
$batName   = "tc.bat"
$binPath   = Join-Path $HOME ".local\bin"


if (!(Test-Path $binPath)) {
    New-Item -ItemType Directory -Force $binPath | Out-Null
}

$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($userPath -notlike "*$binPath*") {
    Write-Host "Adding $binPath to User PATH..." -ForegroundColor Yellow
    [Environment]::SetEnvironmentVariable("Path", "$userPath;$binPath", "User")
    $env:Path += ";$binPath"
}

try {
    $release = Invoke-RestMethod "https://api.github.com/repos/$repo/releases/latest"
    $exeAsset = $release.assets | Where-Object { $_.name -eq $exeName }
    if ($null -eq $exeAsset) { throw "$exeName がリリースに見つかりません。" }

    Write-Host "Downloading $exeName..." -ForegroundColor Cyan
    Invoke-WebRequest -Uri $exeAsset.browser_download_url -OutFile "$binPath\$exeName"

    Write-Host "Creating $batName..." -ForegroundColor Cyan
    $batContent = @"
@echo off
"%~dp0$exeName" %*
"@
    Set-Content -Path "$binPath\$batName" -Value $batContent -Encoding Ascii

    Write-Host "`nInstallation Complete!" -ForegroundColor Green
    Write-Host "Command: $batName"
} catch {
    Write-Error "Failed to install: $_"
}