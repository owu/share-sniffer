# Share Sniffer（共有リンクスニッファー）

このツールは、複数のクラウドストレージ共有リンクが期限切れかどうかを一括で検出することができます。現在、クアーククラウド、天翼クラウド、百度クラウド、阿里クラウド、115クラウドのリンク検出をサポートしています。

---
[English](./README.md) | [中文](./README_cn.md) | 日本語

## 1、ガイド

[リリースノート](https://github.com/owu/share-sniffer/issues/1) |
[インターフェースプレビュー](https://github.com/owu/share-sniffer/issues/2) |
[最新バージョン](https://github.com/owu/share-sniffer/releases)

## 2、起源

某映画・TVドラマリソース共有グループでは、数千の映画・TVドラマリソース共有リンクが含まれるオンラインスプレッドシートがありました。しかし、共有リンクは時々期限切れになり、手動でチェックするのは遅いため、このツールが開発されました。

## 3、技術スタック

- [golang](https://go.dev/)
- [fyne](https://fyne.io/)

## 4、実行

```bash
go mod vendor  && go run main.go
```

## 5、実行可能ファイルのパッケージ化

### 5.1  ShareSniffer.v0.0.9.windows-amd64.exe
- fyne package -os windows -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.2   ShareSniffer.v0.0.9.linux-amd64.tar.xz
- fyne package -os linux -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.3   ShareSniffer.v0.0.9.darwin-amd64.pkg
- fyne package -os darwin -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io

### 5.4   ShareSniffer.v0.0.9.android-arm64.apk
- fyne package -os android/arm64 -name ShareSniffer -icon logo.png -release -app-version 0.0.9 -app-id owu.github.io
```