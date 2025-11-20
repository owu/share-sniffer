# 构建CLI工具的打包脚本（Windows PowerShell版）

# 设置工作目录为脚本所在目录
Set-Location -Path (Split-Path -Parent $MyInvocation.MyCommand.Path)

# 定义常量
$CONFIG_FILE = "../../internal/config/config.go"
$CLI_MAIN = "../../launcher/cli/main.go"

# 读取版本号
Write-Output "正在读取版本号..."
$versionLine = Get-Content $CONFIG_FILE | Where-Object { $_ -match 'q\.AppInfo\.Version\s*=\s*"([^"]+)"' }
if ($versionLine -eq $null) {
    Write-Error "无法从 $CONFIG_FILE 中读取版本号"
    exit 1
}
$version = $matches[1].Trim()
Write-Output "版本号: $version"

# 创建输出目录
$OUTPUT_DIR = "..\releases\v$version"
if (-not (Test-Path $OUTPUT_DIR)) {
    New-Item -ItemType Directory -Path $OUTPUT_DIR -Force | Out-Null
}

# 编译Windows版本
Write-Output "正在编译Windows版本..."
$windowsOutput = "share-sniffer-cli.v$version.windows-amd64.exe"
$windowsOutputPath = Join-Path $OUTPUT_DIR $windowsOutput
go build -o $windowsOutputPath -ldflags "-s -w" $CLI_MAIN

if ($LASTEXITCODE -ne 0) {
    Write-Error "Windows版本编译失败"
    exit 1
}
Write-Output "Windows版本编译成功: $windowsOutput"

Write-Output ""
Write-Output "=== 编译完成 ==="
Write-Output "Windows可执行文件: $windowsOutputPath"
Write-Output ""