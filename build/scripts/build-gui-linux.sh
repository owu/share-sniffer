#!/bin/bash

# Bash script to build Linux version

# Get the current directory
current_dir=$(dirname "$0")

# Path to config.go
config_path="$current_dir/../../internal/config/config.go"

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
icon=$(readlink -f "$current_dir/../../static/linux.png")
src=$(readlink -f "$current_dir/../../launcher/gui")
app_id="owu.github.io"

# Expected output filename
expected_executable="ShareSniffer.v$version.linux-amd64.tar.xz"

# Run build command
echo "Running build command..."

# Clean Go cache to ensure a clean build environment
echo "Cleaning Go cache..."
go clean -cache

# Execute fyne package command from the project root directory to ensure consistent behavior
echo "Command: $fyne_path package -os linux -name $name -icon $icon -release -app-version $version -app-id $app_id -src $src"
"$fyne_path" package -os linux -name "$name" -icon "$icon" -release -app-version "$version" -app-id "$app_id" -src "$src"

if [ $? -ne 0 ]; then
    echo "Build failed with exit code $?"
    exit 1
fi

# Find the generated tar.xz file (fyne creates it in the project root directory on Linux)
# First check current directory (project root where the script is executed from)
generated_tar="$name.tar.xz"
if [ ! -f "$generated_tar" ]; then
    # Try to find it in the package directory
    generated_tar="$current_dir/$name.tar.xz"
    if [ ! -f "$generated_tar" ]; then
        # Try to find it in the src directory as last resort
        generated_tar="$src/$name.tar.xz"
        if [ ! -f "$generated_tar" ]; then
            echo "Failed to find generated tar.xz at $name.tar.xz, $current_dir/$name.tar.xz, or $src/$name.tar.xz"
            exit 1
        fi
    fi
fi

# Create releases directory if it doesn't exist
releases_dir="$current_dir/../releases/v$version"
mkdir -p "$releases_dir"

# Move and rename the tar.xz to releases directory
target_tar="$releases_dir/$expected_executable"
mv -f "$generated_tar" "$target_tar"

echo "Build completed successfully. Output file: $target_tar"

