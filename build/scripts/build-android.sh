#!/bin/bash

# Bash script to build Android version

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
icon=$(readlink -f "$current_dir/../../static/logo.png")
src=$(readlink -f "$current_dir/../../launcher/gui")
app_id="owu.github.io"

# Architectures to build for
architectures=("arm64" "amd64")

# Clean Go cache to ensure a clean build environment
echo "Cleaning Go cache..."
go clean -cache

# Store current directory to return to it later
original_dir=$(pwd)

# Change to package directory to ensure generated file is placed in the right location
cd "$current_dir"

# Loop through each architecture
for arch in "${architectures[@]}"; do
    echo -e "\n=== Building for $arch architecture ==="
    
    # Run build command for current architecture
    echo "Running build command for $arch..."
    
    # Expected output filename with architecture suffix
    expected_executable="ShareSniffer.v$version.android-$arch.apk"
    
    # Execute fyne package command for the current architecture
    os_arg="android/$arch"
    echo "Command: $fyne_path package -os $os_arg -name $name -icon $icon -release -app-version $version -app-id $app_id -src $src"
    "$fyne_path" package -os "$os_arg" -name "$name" -icon "$icon" -release -app-version "$version" -app-id "$app_id" -src "$src"
    
    if [ $? -ne 0 ]; then
        echo "Build failed for $arch with exit code $?"
        continue
    fi
    
    # Find the generated APK file
    # fyne generates files in different locations depending on platform and command execution directory
    generated_apk=""
    
    # Try multiple possible locations for the generated APK
    possible_locations=("$name.apk" "../$name.apk" "$src/$name.apk")
    
    for location in "${possible_locations[@]}"; do
        if [ -f "$location" ]; then
            generated_apk="$location"
            break
        fi
    done
    
    if [ -z "$generated_apk" ]; then
        echo "Failed to find generated APK for $arch at any of the possible locations"
        continue
    fi
    
    # Create releases directory if it doesn't exist
    releases_dir="$current_dir/../releases/v$version"
    mkdir -p "$releases_dir"
    
    # Move and rename the APK to releases directory with architecture suffix
    target_apk="$releases_dir/$expected_executable"
    mv -f "$generated_apk" "$target_apk"
    
    echo "Build for $arch completed successfully. Output file: $target_apk"
done

# Return to original directory
cd "$original_dir"

echo -e "\n=== All builds completed ==="

