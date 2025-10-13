param(
    [string]$Version,
    [string]$InstallPath
)

$Owner = "bab-sh"
$Repo = "bab"
$BinaryName = "bab"

[Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12

function Get-LatestRelease {
    $apiUrl = "https://api.github.com/repos/$Owner/$Repo/releases/latest"
    try {
        $response = Invoke-RestMethod -Uri $apiUrl -Method Get -ErrorAction Stop
        return $response.tag_name
    }
    catch {
        Write-Error "Failed to fetch latest release: $_"
        Write-Error "Please check your internet connection or visit https://github.com/$Owner/$Repo/releases"
        exit 1
    }
}

function Get-Architecture {
    $arch = $env:PROCESSOR_ARCHITECTURE
    switch ($arch) {
        "AMD64" { return "x86_64" }
        "ARM64" { return "arm64" }
        "x86" { return "i386" }
        default {
            Write-Error "Unsupported architecture: $arch"
            Write-Error "Supported architectures: AMD64, ARM64, x86"
            exit 1
        }
    }
}

function Test-CommandExists {
    param($command)
    $null = Get-Command $command -ErrorAction SilentlyContinue
    return $?
}

function Verify-Checksum {
    param(
        [string]$FilePath,
        [string]$ChecksumFile
    )

    $fileName = Split-Path $FilePath -Leaf
    $content = Get-Content $ChecksumFile
    $checksumLine = $content | Where-Object { $_ -like "*$fileName*" }

    if (-not $checksumLine) {
        Write-Warning "Checksum not found for $fileName, skipping verification"
        return $true
    }

    $expectedChecksum = ($checksumLine -split '\s+')[0]
    $fileHash = (Get-FileHash -Path $FilePath -Algorithm SHA256).Hash.ToLower()

    if ($fileHash -eq $expectedChecksum.ToLower()) {
        Write-Host "Checksum verified" -ForegroundColor Green
        return $true
    }
    else {
        Write-Error "Checksum verification failed"
        Write-Error "Expected: $expectedChecksum"
        Write-Error "Got: $fileHash"
        return $false
    }
}

if (-not $Version) {
    Write-Host "Fetching latest release..." -ForegroundColor Cyan
    $Version = Get-LatestRelease
}

Write-Host "`nInstalling $BinaryName $Version" -ForegroundColor Green

$arch = Get-Architecture
$os = "Windows"

$versionNumber = $Version -replace '^v', ''
$fileName = "${BinaryName}_${versionNumber}_${os}_${arch}.zip"
$downloadUrl = "https://github.com/$Owner/$Repo/releases/download/$Version/$fileName"
$checksumUrl = "https://github.com/$Owner/$Repo/releases/download/$Version/checksums.txt"

if (-not $InstallPath) {
    $InstallPath = "$env:LOCALAPPDATA\$BinaryName\bin"
}

$tempDir = Join-Path $env:TEMP ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Path $tempDir -Force | Out-Null

try {
    $zipPath = Join-Path $tempDir $fileName
    Write-Host "Downloading from GitHub releases..." -ForegroundColor Cyan
    try {
        Invoke-WebRequest -Uri $downloadUrl -OutFile $zipPath -UseBasicParsing -ErrorAction Stop
    }
    catch {
        Write-Error "Failed to download $fileName"
        Write-Error "URL: $downloadUrl"
        Write-Error "Error: $_"
        exit 1
    }

    $checksumPath = Join-Path $tempDir "checksums.txt"
    Write-Host "Verifying checksum..." -ForegroundColor Cyan
    try {
        Invoke-WebRequest -Uri $checksumUrl -OutFile $checksumPath -UseBasicParsing -ErrorAction Stop
        if (-not (Verify-Checksum -FilePath $zipPath -ChecksumFile $checksumPath)) {
            exit 1
        }
    }
    catch {
        Write-Warning "Could not verify checksum: $_"
    }

    Write-Host "Extracting..." -ForegroundColor Cyan
    Expand-Archive -Path $zipPath -DestinationPath $tempDir -Force

    if (-not (Test-Path $InstallPath)) {
        New-Item -ItemType Directory -Path $InstallPath -Force | Out-Null
    }

    $binaryPath = Join-Path $tempDir "$BinaryName.exe"
    $destPath = Join-Path $InstallPath "$BinaryName.exe"

    if (-not (Test-Path $binaryPath)) {
        Write-Error "Binary not found in archive: $binaryPath"
        Write-Error "Archive contents:"
        Get-ChildItem $tempDir | ForEach-Object { Write-Error "  $_" }
        exit 1
    }

    Copy-Item -Path $binaryPath -Destination $destPath -Force
    Write-Host "Installed to: $destPath" -ForegroundColor Green

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    if ($userPath -notlike "*$InstallPath*") {
        Write-Host "Adding to PATH..." -ForegroundColor Cyan
        [Environment]::SetEnvironmentVariable(
            "Path",
            "$userPath;$InstallPath",
            "User"
        )
        $env:Path = "$env:Path;$InstallPath"
        Write-Host "Added $InstallPath to PATH" -ForegroundColor Green
        Write-Host "`nRestart your terminal for PATH changes to take effect" -ForegroundColor Yellow
    }

    Write-Host "`n$BinaryName installed successfully!" -ForegroundColor Green

    $env:Path = "$env:Path;$InstallPath"
    if (Test-CommandExists $BinaryName) {
        Write-Host "`nVersion:" -ForegroundColor Cyan
        try {
            & $BinaryName --version
        }
        catch {
            Write-Host "Run '$BinaryName --version' to verify" -ForegroundColor Yellow
        }
    }
    else {
        Write-Host "`nRun 'refreshenv' or restart your terminal" -ForegroundColor Yellow
    }

    Write-Host "`nQuick start:" -ForegroundColor Cyan
    Write-Host "  1. Create a Babfile in your project"
    Write-Host "  2. Run 'bab' to list tasks"
    Write-Host "  3. Run 'bab <task>' to execute"
}
catch {
    Write-Error "Installation failed: $_"
    exit 1
}
finally {
    if (Test-Path $tempDir) {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
}
