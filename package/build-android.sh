#!/bin/bash

# Bash script to build Android version

# Get the current directory
current_dir=$(dirname "$0")

# Path to config.go
config_path="$current_dir/../internal/config/config.go"

# Read version from config.go
version_line=$(grep -E 'q\.AppInfo\.Version\s*=\s*"[^"]+"' "$config_path")
if [ -z "$version_line" ]; then
    echo "Failed to find version in $config_path"
    exit 1
fi

version=$(echo "$version_line" | sed -E 's/^[[:space:]]*q\.AppInfo\.Version[[:space:]]*=[[:space:]]*"([^"]+)"[[:space:]]*/\1/' | tr -d '[:space:]')
echo "Found version: $version"

# Check and install fyne if needed
echo "Checking for fyne command..."

# Try to find fyne in PATH
if command -v fyne &> /dev/null; then
    fyne_path="fyne"
    echo "Found fyne in PATH"
else
    # Try to find fyne in GOPATH/bin
    gopath=${GOPATH:-$HOME/go}
    fyne_in_gopath="$gopath/bin/fyne"
    if [ -f "$fyne_in_gopath" ]; then
        fyne_path="$fyne_in_gopath"
        echo "Found fyne in GOPATH: $fyne_path"
    else
        # Install fyne
        echo "Fyne not found, installing..."
        go install fyne.io/tools/cmd/fyne@latest
        
        if [ $? -ne 0 ]; then
            echo "Failed to install fyne"
            exit 1
        fi
        
        # Check again after installation
        if command -v fyne &> /dev/null; then
            fyne_path="fyne"
        else
            fyne_path="$fyne_in_gopath"
        fi
        
        if [ ! -f "$fyne_path" ]; then
            echo "Failed to find fyne after installation"
            exit 1
        fi
        
        echo "Fyne installed successfully: $fyne_path"
    fi
fi

# Set parameters
name="ShareSniffer"
icon=$(readlink -f "$current_dir/../logo.png")
src=$(readlink -f "$current_dir/../launcher/gui")
app_id="owu.github.io"

# Expected output filename
expected_executable="ShareSniffer.v$version.android-arm64.apk"

# Run build command
echo "Running build command..."

# Clean Go cache to ensure a clean build environment
echo "Cleaning Go cache..."
go clean -cache

# Run build command from the original directory (project root)
echo "Command: $fyne_path package -os android/arm64 -name $name -icon $icon -release -app-version $version -app-id $app_id -src $src"
"$fyne_path" package -os android/arm64 -name "$name" -icon "$icon" -release -app-version "$version" -app-id "$app_id" -src "$src"

if [ $? -ne 0 ]; then
    echo "Build failed with exit code $?"
    exit 1
fi

# Find the generated APK file (fyne creates it in different locations depending on platform)
# First check current directory (project root)
generated_apk="$name.apk"
if [ ! -f "$generated_apk" ]; then
    # Try to find it in the package directory
    generated_apk="$current_dir/$name.apk"
    if [ ! -f "$generated_apk" ]; then
        # Try to find it in the src directory as last resort
        generated_apk="$src/$name.apk"
        if [ ! -f "$generated_apk" ]; then
            echo "Failed to find generated APK at $name.apk, $current_dir/$name.apk, or $src/$name.apk"
            exit 1
        fi
    fi
fi

# Rename the APK to the expected filename (it's already in the package directory)
target_apk="$current_dir/$expected_executable"
mv -f "$generated_apk" "$target_apk"

echo "Build completed successfully. Output file: $target_apk"

