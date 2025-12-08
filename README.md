# Share Sniffer

This tool can batch check whether various cloud storage sharing links have expired. Currently supported cloud storage services include Quark Cloud, Tianyi Cloud, Baidu Cloud, Ali Cloud, and 115 Cloud.

---
English | [中文](./README_cn.md) | [日本語](./README_jp.md)

## 1. Guide

[Release Notes](https://github.com/owu/share-sniffer/issues/1) |
[Interface Preview](https://github.com/owu/share-sniffer/issues/2) |
[Latest Version](https://github.com/owu/share-sniffer/releases)

## 2. Origin

In a movie/TV resource sharing group, there was an online spreadsheet containing thousands of movie/TV resource sharing links. However, these sharing links would sometimes expire, and manual checking was slow, so this tool was developed.

## 3. Technology Stack

- [golang](https://go.dev/)
- [fyne](https://fyne.io/)

## 4. Running

```bash
go mod vendor && go run main.go
```

## 5. Packaging Executable Files

### 5.1 ShareSniffer.v0.0.9.windows-amd64.exe
- fyne package -os windows -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.2 ShareSniffer.v0.0.9.linux-amd64.tar.xz
- fyne package -os linux -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.3 ShareSniffer.v0.0.9.darwin-amd64.pkg
- fyne package -os darwin -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.4 ShareSniffer.v0.0.9.android-arm64.apk
- fyne package -os android/arm64 -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io
```