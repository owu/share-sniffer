#!/bin/bash

# Bash script to build Android version

# Set working directory to the script's directory
cd "$(dirname "$0")"

# Path to config.go (relative to script directory)
config_path="../../internal/config/config.go"

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
    gopath=$(go env GOPATH)
    if [ -z "$gopath" ]; then
        gopath="$HOME/go"
    fi
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
            fyne_path=$(command -v fyne)
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
# Resolve absolute paths for fyne command to avoid issues
current_dir=$(pwd)
icon=$(readlink -f "../../static/logo.png")
src=$(readlink -f "../../launcher/gui")
app_id="owu.github.io"

# Architectures to build for
architectures=("arm64" "amd64")

# Clean Go cache to ensure a clean build environment
echo "Cleaning Go cache..."
go clean -cache

# Create output directory
OUTPUT_DIR="../releases/v$version"
mkdir -p "$OUTPUT_DIR"

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
    # Since we are in build/scripts, and src is absolute path
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
    
    # Move and rename the APK to releases directory with architecture suffix
    target_apk="$OUTPUT_DIR/$expected_executable"
    mv -f "$generated_apk" "$target_apk"
    
    echo "Build for $arch completed successfully. Output file: $target_apk"
done

echo -e "\n=== All builds completed ==="

