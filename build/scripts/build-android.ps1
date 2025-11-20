# PowerShell script to build Android version

# Get the current directory
$currentDir = Split-Path -Parent $MyInvocation.MyCommand.Path

# Path to config.go
$configPath = Join-Path $currentDir "..\..\internal\config\config.go"

# Read version from config.go
$versionLine = Get-Content $configPath | Where-Object { $_ -match 'q\.AppInfo\.Version\s*=\s*"(.*)"' }
if (-not $versionLine) {
    Write-Error "Failed to find version in $configPath"
    exit 1
}

$version = $matches[1].Trim()
Write-Host "Found version: $version"

# Check and install fyne if needed
Write-Host "Checking for fyne command..."

# Try to find fyne in PATH
$fynePath = Get-Command fyne -ErrorAction SilentlyContinue
if ($fynePath) {
    $fynePath = $fynePath.Source
    Write-Host "Found fyne in PATH: $fynePath"
} else {
    # Try to find fyne in GOPATH/bin
    $gopath = if ($env:GOPATH) {
        $env:GOPATH
    } else {
        Join-Path $env:USERPROFILE "go"
    }
    
    $fyneInGopath = Join-Path $gopath "bin\fyne.exe"
    if (Test-Path $fyneInGopath) {
        $fynePath = $fyneInGopath
        Write-Host "Found fyne in GOPATH: $fynePath"
    } else {
        # Install fyne
        Write-Host "Fyne not found, installing..."
        go install fyne.io/tools/cmd/fyne@latest
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Failed to install fyne"
            exit 1
        }
        
        # Check again after installation
        $fynePath = Get-Command fyne -ErrorAction SilentlyContinue
        if (-not $fynePath) {
            $fynePath = $fyneInGopath
        }
        
        if (-not $fynePath -or (-not (Test-Path $fynePath))) {
            Write-Error "Failed to find fyne after installation"
            exit 1
        }
        
        Write-Host "Fyne installed successfully: $fynePath"
    }
}

# Set parameters
$name = "ShareSniffer"
$icon = Resolve-Path (Join-Path $currentDir "..\..\static\logo.png") -ErrorAction Stop
$src = Resolve-Path (Join-Path $currentDir "..\..\launcher\gui") -ErrorAction Stop
$appId = "owu.github.io"

# Architectures to build for
$architectures = @("arm64", "amd64")

# Clean Go cache to ensure a clean build environment
Write-Host "Cleaning Go cache..."
go clean -cache

# Store current directory to return to it later
$originalDir = Get-Location

# Change to package directory to ensure generated file is placed in the right location
Set-Location -Path $currentDir

# Loop through each architecture
try {
    foreach ($arch in $architectures) {
        Write-Host "\n=== Building for $arch architecture ==="
        
        # Run build command
        Write-Host "Running build command for $arch..."
        
        # Expected output filename with architecture suffix
        $expectedExecutable = "ShareSniffer.v$version.android-$arch.apk"
        
        # Execute fyne package command for the current architecture
        $osArg = "android/$arch"
        Write-Host "Command: $fynePath package -os $osArg -name $name -icon $icon -release -app-version $version -app-id $appId -src $src"
        $buildOutput = & $fynePath package -os $osArg -name $name -icon $icon -release -app-version $version -app-id $appId -src $src 2>&1
        
        Write-Host "Build output for ${arch}:"
        Write-Host $buildOutput
        
        if ($LASTEXITCODE -ne 0) {
            Write-Error "Build failed for $arch with exit code $LASTEXITCODE"
            continue
        }
        
        # Find the generated APK file
        # fyne generates files in different locations depending on platform and command execution directory
        $generatedApk = $null
        
        # Try multiple possible locations for the generated APK
        $possibleLocations = @(
            Join-Path $src "$name.apk"
            Join-Path $PSScriptRoot "..\$name.apk"
            Join-Path $currentDir "$name.apk"
        )
        
        foreach ($location in $possibleLocations) {
            if (Test-Path $location) {
                $generatedApk = $location
                break
            }
        }
        
        if (-not $generatedApk) {
            Write-Error "Failed to find generated APK for $arch at any of the possible locations"
            continue
        }
        
        # Create releases directory if it doesn't exist
        $releasesDir = Join-Path $currentDir "..\releases\v$version"
        if (-not (Test-Path $releasesDir)) {
            New-Item -ItemType Directory -Path $releasesDir -Force | Out-Null
        }
        
        # Move and rename the APK to releases directory with architecture suffix
        $targetApk = Join-Path $releasesDir $expectedExecutable
        Move-Item -Path $generatedApk -Destination $targetApk -Force
        
        Write-Host "Build for $arch completed successfully. Output file: $targetApk"
    }
} finally {
    # Return to original directory
    Set-Location -Path $originalDir
}

Write-Host "\n=== All builds completed ==="

