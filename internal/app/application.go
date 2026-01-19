// Package app Copyright 2025 Share Sniffer
//
// app.go 实现了Share Sniffer应用程序的核心逻辑和UI初始化
package app

import (
	"fmt"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"share-sniffer/internal/config"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/ui/about"
	"share-sniffer/internal/ui/check"
	"share-sniffer/internal/ui/state"
	"share-sniffer/internal/utils"
)

// ShareSnifferApp 是应用程序的主结构
// 它封装了Fyne应用实例、主窗口和应用状态
//
// 字段:
// - app: Fyne框架的应用实例
// - window: 应用的主窗口
// - state: 应用的全局状态

type ShareSnifferApp struct {
	app    fyne.App
	window fyne.Window
	state  *state.AppState
}

// NewShareSnifferApp 创建并初始化ShareSnifferApp的新实例
// 功能:
// 1. 获取应用配置
// 2. 创建Fyne应用实例
// 3. 创建并配置主窗口
// 4. 返回初始化后的应用实例
//
// 返回值:
// - *ShareSnifferApp: 初始化完成的应用实例
func NewShareSnifferApp() *ShareSnifferApp {
	// 获取配置 - 从全局配置单例获取应用配置信息
	cfg := config.GetConfig()

	// 创建Fyne应用实例，并设置应用ID
	a := app.NewWithID(cfg.AppInfo.AppName)

	// 创建应用主窗口
	w := a.NewWindow(fmt.Sprintf("%s v%s", cfg.AppInfo.AppName, cfg.AppInfo.Version))

	// 先设置窗口大小和居中，暂时不设置图标以提高启动速度
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()
	// 注释掉图标设置，避免可能的卡顿
	// w.SetIcon(icons.LogoTransparent)

	// 创建并返回应用实例
	return &ShareSnifferApp{
		app:    a,
		window: w,
		state:  state.NewAppState(),
	}
}

// Run 启动应用程序
// 功能:
// 1. 创建并设置窗口内容
// 2. 在窗口显示前启动goroutine执行时间同步
// 3. 启动版本检查
// 4. 显示窗口并进入主事件循环
func (q *ShareSnifferApp) Run() {
	// 创建窗口内容并设置到窗口中
	q.window.SetContent(q.createContent())

	// 启动goroutine执行时间同步，在窗口显示前就开始执行
	// 但使用延迟确保UI有足够时间加载
	go func() {
		// 延迟1秒执行，确保UI已经完全初始化和显示
		time.Sleep(1 * time.Second)
		q.state.StandardTime = utils.StandardTime()
		logger.Info("ShareSnifferApp: 设置标准时间,%d", q.state.StandardTime)
	}()

	// 启动版本检查，在协程中执行，避免阻塞UI
	go func() {
		// 延迟2秒执行，确保主窗口已经完全初始化
		time.Sleep(2 * time.Second)
		logger.Info("ShareSnifferApp: 开始检查版本更新")
		about.CheckUpdate(q.window, false)
	}()

	// 显示窗口并启动应用的主事件循环（阻塞操作）
	q.window.ShowAndRun()
}

// createContent 创建应用的主界面内容
// 功能:
// 1. 创建带有多个标签页的容器
// 2. 添加检查标签页和关于标签页
// 3. 设置标签页位置为左侧
//
// 返回值:
// - fyne.CanvasObject: 可添加到窗口的UI对象
func (q *ShareSnifferApp) createContent() fyne.CanvasObject {
	// 使用默认的Tabs布局 - 创建标签页容器
	tabs := container.NewAppTabs(
		// 添加检查标签页，用于检查分享链接
		check.NewCheckTab(q.window, q.state),
		// 添加关于标签页，显示应用信息
		about.NewAboutTab(q.window),
	)
	// 设置标签页位置在窗口左侧
	tabs.SetTabLocation(container.TabLocationLeading)

	// 返回创建的UI内容
	return tabs
}
