//go:build !android
// +build !android

package check

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	fyneDialog "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/widget"
	sqweekDialog "github.com/sqweek/dialog"
	"share-sniffer/internal/logger"
)

// 基于 fyneDialog 的桌面端对话框提供者
type DesktopDialogProvider struct {
	window fyne.Window
}

// 为DesktopDialogProvider实现DialogProvider接口
func (d *DesktopDialogProvider) ShowError(message string) {
	// 使用fyneDialog显示错误信息
	fyneDialog.ShowError(fmt.Errorf("%s", message), d.window)
}

func (d *DesktopDialogProvider) ShowInfo(message string, title string) {
	// 使用fyneDialog显示带图标的信息
	fyneDialog.ShowInformation(title, message, d.window)
}

// ShowTxt 显示不带图标的文本对话框
func (d *DesktopDialogProvider) ShowTxt(message string, title string) {
	// 创建不带图标的自定义文本对话框
	content := container.NewVBox(
		widget.NewLabel(message),
	)

	// 创建对话框
	dialog := fyneDialog.NewCustom(title, "确定", content, d.window)
	dialog.Resize(fyne.NewSize(200, 100))
	dialog.Show()
}

// 桌面平台的对话框提供者获取函数
func getDesktopDialogProvider(window fyne.Window) DialogProvider {
	return &DesktopDialogProvider{window: window}
}

// openFileWithSqweekDialog 使用github.com/sqweek/dialog的文件选择对话框（桌面平台）
func (q *CheckUI) openFileWithSqweekDialog() {
	// 使用sqweek/dialog打开文件选择对话框
	filename, err := sqweekDialog.File().Filter("文本文件", "txt").Title("打开分享链接文本文件").Load()
	if err != nil {
		// 检查是否是用户取消操作，不区分大小写
		errMsg := strings.ToLower(err.Error())
		if errMsg != "cancelled" {
			logger.Error("文件选择错误: %v", err)
			q.dialogProvider.ShowError(err.Error())
		} else {
			logger.Debug("用户取消了文件选择")
		}
		return
	}

	if filename == "" {
		logger.Debug("用户取消了文件选择")
		return
	}

	logger.Debug("选择的文件路径: %s", filename)

	q.fileEntry.SetText(filename)

	// 设置FilePath，由于sqweek/dialog直接返回文件路径，不需要URI
	q.state.FilePath = filename
	q.state.FileURI = nil // 桌面平台不需要URI

	q.loadToTable()
}
