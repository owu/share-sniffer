# Share Sniffer (分享嗅探器)

该工具可以批量检测多种网盘分享链接是否过期，已支持 夸克网盘、天翼云盘、百度网盘、阿里云盘、115网盘 的检测。

## 1、指南

[发布日志](https://github.com/owu/share-sniffer/issues/1) |
[界面预览](https://github.com/owu/share-sniffer/issues/2) |
[最新版本](https://github.com/owu/share-sniffer/releases)

## 2、起源

某影视资源分享群，有一份在线表格，包含数几千个影视资源分享链接，但分享链接时有过期，人工检查慢，因此有了这个工具。

## 3、技术栈

- [golang](https://go.dev/)
- [fyne](https://fyne.io/)

## 4、运行

```bash
go mod vendor  && go run main.go
```

## 5、打包可执行文件

### 5.1  ShareSniffer.v0.0.9.windows-amd64.exe
- fyne package -os windows -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.2   ShareSniffer.v0.0.9.linux-amd64.tar.xz
- fyne package -os linux -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.3   ShareSniffer.v0.0.9.darwin-amd64.pkg
- fyne package -os darwin -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.4   ShareSniffer.v0.0.9.android-arm64.apk
- fyne package -os android/arm64 -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

