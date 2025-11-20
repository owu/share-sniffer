// Copyright 2025 Share Sniffer
//
// main.go 是应用程序的入口文件，负责初始化配置、设置日志和启动主应用程序
package main

import (
	"github.com/owu/share-sniffer/internal/app"
	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/logger"
)

// main 函数是程序的入口点
// 它执行以下操作：
// 1. 获取应用配置单例
// 2. 设置日志级别为Info
// 3. 记录应用启动日志
// 4. 创建并运行ShareSniffer应用实例
func main() {
	// 初始化配置 - 获取全局配置单例
	cfg := config.GetConfig()

	// 设置日志级别为Info，控制日志输出的详细程度
	logger.SetLogLevel(logger.LevelInfo)

	// 记录应用启动信息，包括版本号
	logger.Info("应用启动,名称: %s , 版本: %s", cfg.AppInfo.AppName, cfg.AppInfo.Version)

	// 启动应用 - 创建并运行ShareSniffer应用实例
	app := app.NewShareSnifferApp()

	app.Run()
}
