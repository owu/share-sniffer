# Share Sniffer

English | [中文](./README.md) | [日本語](./README_jp.md)


## 1. Introduction

Share Sniffer is a cross-platform cloud storage sharing link detection tool that supports validity checking for sharing links from various mainstream cloud storage services. The tool provides an intuitive graphical user interface (GUI) and a convenient command-line interface (CLI), allowing users to choose the usage method according to their needs.

### 1.1 Supported Cloud Storage Types

- ✅ Quark Cloud
- ✅ Tianyi Cloud
- ✅ Baidu Cloud
- ✅ Alibaba Cloud
- ✅ 115 Cloud
- ✅ 123 Cloud 
- ✅ UC Cloud
- ✅ Xunlei Cloud
- ✅ 139 Cloud

## 2. Origin

In a movie/TV resource sharing group, there was an online spreadsheet containing thousands of movie/TV resource sharing links. However, these sharing links would sometimes expire, and manual checking was slow, so this tool was developed. Through automated detection, valid sharing links can be quickly filtered out, improving resource management efficiency.

## 3. Technology Stack

- **Development Language**: Go 1.25
- **GUI Framework**: [fyne.io/fyne/v2](https://fyne.io/) - Cross-platform GUI framework
- **CLI Framework**: [github.com/spf13/cobra](https://github.com/spf13/cobra) - Command-line framework


## 4. Development and Running

### 4.1 GUI Mode

#### 4.1.1 Direct Running

```bash
# Initialize dependencies
go mod tidy

# Run GUI application
go run ./launcher/gui/main.go

# Run the GUI application while simultaneously printing all the detailed commands executed during the compilation and linking process.
go clean -cache  && go clean -modcache && go run -x ./launcher/gui/main.go

```

#### 4.1.2 Development Mode

```bash
# Install fyne development tools (optional)
go install fyne.io/tools/cmd/fyne@latest

# Run in fyne development mode (supports hot reload)
fyne serve -src ./launcher/gui
```

### 4.2 CLI Mode

#### 4.2.1 Direct Running

```bash
# Run CLI application
go run ./launcher/cli/main.go [command/URL]
```

#### 4.2.2 CLI Command Examples

```bash
# Show help information
./share-sniffer-cli --help

# Check version
./share-sniffer-cli version

# Check supported link types
./share-sniffer-cli support

# Check project homepage
./share-sniffer-cli home

# Detect a single link
./share-sniffer-cli "https://pan.quark.cn/s/0a6e84c02020"
```

## 5. Packaging and Compilation

The project provides automated packaging scripts located in the `/build` directory, supporting packaging for Windows and Linux platforms.

### 5.1 Packaging Script Description

| Script Name | Platform | Description |
|---------|------|------|
| `build-gui-windows.ps1` | Windows | PowerShell script for building Windows platform GUI executable files |
| `build-gui-linux.sh` | Linux | Bash script for building Linux platform GUI installation packages |
| `build-android.ps1` | Windows | PowerShell script for building Android platform APK |
| `build-android.sh` | Linux | Bash script for building Android platform APK |
| `build-cli-windows.ps1` | Windows | PowerShell script for building Windows platform CLI executable files |
| `build-cli-linux.sh` | Linux | Bash script for building Linux platform CLI executable files |
| `build-all.ps1` | Windows | PowerShell script for batch building all Windows platform executable files |
| `build-all.sh` | Linux | Bash script for batch building all Linux platform executable files |

### 5.2 Using Packaging Scripts

#### 5.2.1 Windows Platform

```powershell
# Build Windows GUI version
cd build/scripts
./build-gui-windows.ps1

# Build Android version
cd build/scripts
./build-android.ps1

# Build CLI tool
cd build/scripts
./build-cli-windows.ps1

# Batch build all Windows packages
cd build/scripts
./build-all.ps1
```

#### 5.2.2 Linux Platform

```bash
# Build Linux GUI version
cd build/scripts
chmod +x *.sh
./build-gui-linux.sh

# Build Android version
./build-android.sh

# Build CLI tool
./build-cli-linux.sh

# Batch build all Linux packages
./build-all.sh
```

### 5.3 Packaging Script Features

- Automatically reads version number from `internal/config/config.go`
- Automatically detects and installs `fyne` tools (if not installed)
- Cleans Go cache to ensure a clean build environment
- Automatically names generated files and outputs them to the `/build/releases/{version}/` directory
- Supports Windows, Linux, and Android platforms
- Provides batch build scripts for one-click compilation

## 6. Installation and Uninstallation

### 6.1 Linux GUI Installation

1. Download the latest installation package `ShareSniffer.v0.2.0.linux-amd64.tar.xz` to any directory

2. Extract the file, enter the directory, and install:

```bash
# Create installation directory
mkdir ./ShareSniffer.linux-amd64 

# Extract installation package
tar -xJf ./ShareSniffer.v0.2.0.linux-amd64.tar.xz -C ./ShareSniffer.linux-amd64 

# Enter installation directory
cd ./ShareSniffer.linux-amd64 

# Execute installation
sudo make install
```

### 6.2 Linux GUI Uninstallation

```bash
# Enter installation directory
cd ./ShareSniffer.linux-amd64 

# Execute uninstallation
sudo make uninstall 

# Return to parent directory
cd ../ 

# Delete installation directory
rm -rf ./ShareSniffer.linux-amd64
```

### 6.3 share-sniffer-cli Installation

#### 6.3.1 Linux Installation

```
Download the latest share-sniffer-cli.v0.2.0.linux-amd64
Rename it to share-sniffer-cli
Move the executable file to the `/usr/local/bin` directory
```

#### 6.3.2 Windows Installation

```
Download the latest share-sniffer-cli.v0.2.0.windows-amd64.exe
Rename it to share-sniffer-cli.exe
Optionally move the executable file to the `C:\Windows\System32` directory
```


## 7. Interface Preview

### 7.1 About Interface

<p align="center">
  <img src="screenshot/about.png" width="48%" />
  <img src="screenshot/update.png" width="48%" />
</p>

### 7.2 Detection Interface

<p align="center">
  <img src="screenshot/check.png" width="48%" />
  <img src="screenshot/open.png" width="48%" />
</p>

<p align="center">
  <img src="screenshot/load.png" width="48%" />
  <img src="screenshot/checking.png" width="48%" />
</p>

### 7.3 Result Interface

<p align="center">
  <img src="screenshot/result.png" width="48%" />
</p>

## 8. CLI Mode Tool Introduction

### 8.1 Command Description

| Command | Description | Example |
|------|------|------|
| `help` | Show help information | `./share-sniffer-cli --help` |
| `version` | Show version information | `./share-sniffer-cli version` |
| `support` | Show supported link types | `./share-sniffer-cli support` |
| `home` | Show project homepage link | `./share-sniffer-cli home` |
| `[URL]` | Detect specified link | `./share-sniffer-cli "https://pan.quark.cn/s/0a6e84c02020"` |

### 8.2 Output Format

The CLI tool returns results in JSON format, which is convenient for calling by other programs:

```json
{
  "error": 0,
  "msg": "valid",
  "data": {
    "url": "https://pan.quark.cn/s/0a6e84c02020",
    "name": "国语动漫",
    "elapsed": 359
  }
}
```

#### 8.2.1 Output Field Description

| Field | Type | Description |
|------|------|------|
| `error` | int | Error code, 0 indicates no errors, meaning the link is valid; 10 indicates an unknown error; 11 indicates the link has expired; 12 indicates parameter errors, etc.; 13 indicates a timeout; 14 indicates an error during the request process. |
| `msg` | string | Status description, "success" indicates success, "failed" indicates failure, "timeout" indicates timeout |
| `data` | object | Detection result details |
| `data.url` | string | Detected URL |
| `data.name` | string | Resource name (if detection is successful) |
| `data.elapsed` | int64 | Detection time (milliseconds) |

### 8.3 Usage Scenarios

- Batch detection of link validity
- Integration into other scripts or programs
- Use in server environments
- Automated detection workflows

### 8.4 Docker Tools and HTTP API

To facilitate containerized deployment and remote calls, the project provides the `docker-tools.sh` script and a matching set of HTTP API interfaces.

#### 8.4.1 Docker Management Script (docker-tools.sh)

`docker-tools.sh` is a convenient Shell script used to manage Docker image building, container starting/stopping, and log viewing.

**Usage:**

```bash
./docker-tools.sh [options]
```

**Options Description:**

| Option | Corresponding Parameter | Description |
|--------|-------------------------|-------------|
| `b` | `build` | Build Docker image (supports interactive input of proxy address) |
| `d` | `down` | Stop and remove container |
| `l` | `logs` | View container real-time logs |
| `m` | `move` | Image migration (export/import image file) |
| `u` | `up` | Start container (prioritizes docker-compose, otherwise uses docker run) |
| `h` | `help` | Show help information |

**Build Example:**

```bash
# Build image (you can input proxy address according to the prompt to accelerate dependency download)
./docker-tools.sh b

# Start container
./docker-tools.sh u
```

#### 8.4.2 HTTP API Interfaces

After the container starts (default port 60204), a set of HTTP interfaces is provided, with functions corresponding one-to-one with CLI commands.

**Base URL:** `http://<IP>:60204`

**Interface List:**

| Interface Path | Method | Corresponding CLI Command | Description |
|----------------|--------|---------------------------|-------------|
| `/api/check` | `POST` | `share-sniffer-cli [URL]` | Detect validity of specified link |
| `/api/version` | `GET` | `share-sniffer-cli version` | Get version information |
| `/api/home` | `GET` | `share-sniffer-cli home` | Get project homepage address |
| `/api/support` | `GET` | `share-sniffer-cli support` | Get list of supported link types |
| `/api/help` | `GET` | `share-sniffer-cli help` | Get help information |

**Call Examples:**

1. **Detect Link (POST /api/check)**

   ```bash
   curl -X POST http://localhost:60204/api/check \
     -H "Content-Type: application/json" \
     -d '{"url": "https://pan.quark.cn/s/0a6e84c02020"}'
   ```

   **Response:**
   ```json
   {
     "error": 0,
     "msg": "valid",
     "data": {
       "url": "https://pan.quark.cn/s/0a6e84c02020",
       "name": "Mandarin Anime",
       "elapsed": 359
     }
   }
   ```

2. **Get Version (GET /api/version)**

   ```bash
   curl http://localhost:60204/api/version
   # Response: 0.2.2
   ```

3. **Get Supported List (GET /api/support)**

   ```bash
   curl http://localhost:60204/api/support
   # Response:
   # https://pan.quark.cn/s/
   # https://pan.baidu.com/s/
   # ...
   ```

## 9. Contribution

Welcome to submit Issues and Pull Requests!

## 10. License

[GNU GPL v3 License](LICENSE)