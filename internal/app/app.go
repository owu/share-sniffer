// Package app Copyright 2025 Share Sniffer
//
// app.go 实现了Share Sniffer应用程序的核心逻辑和UI初始化
package app

import (
	"fmt"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/logger"
	"github.com/owu/share-sniffer/internal/ui/about"
	"github.com/owu/share-sniffer/internal/ui/check"
	"github.com/owu/share-sniffer/internal/ui/icons"
	"github.com/owu/share-sniffer/internal/ui/state"
	"github.com/owu/share-sniffer/internal/utils"
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

	// 设置窗口图标、大小并使其居中显示
	w.SetIcon(icons.LogoTransparent)
	w.Resize(fyne.NewSize(800, 600))
	w.CenterOnScreen()

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
// 2. 显示窗口并进入主事件循环
// 3. 使用goroutine在窗口显示后立即设置标准时间，避免阻塞UI
func (q *ShareSnifferApp) Run() {
	// 创建窗口内容并设置到窗口中
	q.window.SetContent(q.createContent())

	// 使用goroutine在窗口显示后立即执行StandardTime操作，不阻塞UI
	go func() {
		q.state.StandardTime = utils.StandardTime()
		logger.Info("ShareSnifferApp: 设置标准时间,%d", q.state.StandardTime)
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
		about.NewAboutTab(),
	)
	// 设置标签页位置在窗口左侧
	tabs.SetTabLocation(container.TabLocationLeading)

	// 返回创建的UI内容
	return tabs
}
