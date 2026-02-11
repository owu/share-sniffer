# PowerShell script to build Windows version

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
    $gopath = go env GOPATH
    if (-not $gopath) {
        $gopath = Join-Path $env:USERPROFILE "go"
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
        $cmd = Get-Command fyne -ErrorAction SilentlyContinue
        if ($cmd) {
            $fynePath = $cmd.Source
        } else {
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

# Expected output filename
$expectedExecutable = "ShareSniffer.v$version.windows-amd64.exe"

# Run build command
Write-Host "Running build command..."

# Clean Go cache to ensure a clean build environment
Write-Host "Cleaning Go cache..."
go clean -cache

# Store current directory to return to it later
$originalDir = Get-Location

# Change to package directory to ensure generated file is placed in the right location
Set-Location -Path $currentDir

# Execute fyne package command
Write-Host "Command: $fynePath package -os windows -name $name -icon $icon -release -app-version $version -app-id $appId -src $src"
$buildOutput = & $fynePath package -os windows -name $name -icon $icon -release -app-version $version -app-id $appId -src $src 2>&1

Write-Host "Build output:"
Write-Host $buildOutput

if ($LASTEXITCODE -ne 0) {
    Write-Error "Build failed with exit code $LASTEXITCODE"
    # Return to original directory before exiting
    Set-Location -Path $originalDir
    exit 1
}

# Return to original directory
Set-Location -Path $originalDir

# Find the generated executable file
# fyne generates files in different locations depending on platform and command execution directory
# First check src directory (default for Windows)
$generatedExe = Join-Path $src "$name.exe"
if (-not (Test-Path $generatedExe)) {
    # Try to find it in project root directory
    $generatedExe = Join-Path $PSScriptRoot "..\$name.exe"
    if (-not (Test-Path $generatedExe)) {
        # Try to find it in package directory
        $generatedExe = Join-Path $currentDir "$name.exe"
        if (-not (Test-Path $generatedExe)) {
            Write-Error "Failed to find generated executable at $src\$name.exe, $PSScriptRoot\..\$name.exe, or $currentDir\$name.exe"
            exit 1
        }
    }
}

# Create releases directory if it doesn't exist
$releasesDir = Join-Path $currentDir "..\releases\v$version"
if (-not (Test-Path $releasesDir)) {
    New-Item -ItemType Directory -Path $releasesDir -Force | Out-Null
}

# Move and rename the executable to releases directory
$targetExe = Join-Path $releasesDir $expectedExecutable
Move-Item -Path $generatedExe -Destination $targetExe -Force

Write-Host "Build completed successfully. Output file: $targetExe"

