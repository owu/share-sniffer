package app

import (
	"share-sniffer/internal/config"
	"share-sniffer/internal/logger"
)

// 它执行以下操作：
// 1. 获取应用配置单例
// 2. 设置日志级别为Info
// 3. 记录应用启动日志
// 4. 创建并运行ShareSniffer应用实例

func Launcher() {
	// 初始化配置 - 获取全局配置单例
	cfg := config.GetConfig()

	// 设置日志级别为Info，控制日志输出的详细程度
	logger.SetLogLevel(logger.LevelInfo)

	// 记录应用启动信息，包括版本号
	logger.Info("应用启动,名称: %s , 版本: %s", cfg.AppInfo.AppName, cfg.AppInfo.Version)

	// 启动应用 - 创建并运行ShareSniffer应用实例
	app := NewShareSnifferApp()

	app.Run()
}
