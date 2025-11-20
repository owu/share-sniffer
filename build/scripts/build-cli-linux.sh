#!/bin/bash

# 构建CLI工具的打包脚本（Linux Bash版）

# 设置工作目录为脚本所在目录
cd "$(dirname "$0")"

# 定义常量
CONFIG_FILE="../../internal/config/config.go"
CLI_MAIN="../../launcher/cli/main.go"

# 读取版本号
echo "正在读取版本号..."
version=$(grep -E 'q\.AppInfo\.Version\s*=\s*"[^"]+"' "$CONFIG_FILE" | sed -E 's/.*"([^"]+)".*/\1/' | tr -d '[:space:]')

if [ -z "$version" ]; then
    echo "错误：无法从 $CONFIG_FILE 中读取版本号"
    exit 1
fi
echo "版本号: $version"

# 创建输出目录
OUTPUT_DIR="../releases/v$version"
mkdir -p "$OUTPUT_DIR"

# 编译Linux版本
echo "正在编译Linux版本..."
linux_output="share-sniffer-cli.v$version.linux-amd64"
linux_output_path="$OUTPUT_DIR/$linux_output"
go build -o "$linux_output_path" -ldflags "-s -w" "$CLI_MAIN"

if [ $? -ne 0 ]; then
    echo "错误：Linux版本编译失败"
    exit 1
fi
echo "Linux版本编译成功: $linux_output"

echo ""
echo "=== 编译完成 ==="
echo "Linux可执行文件: $linux_output_path"
echo ""
