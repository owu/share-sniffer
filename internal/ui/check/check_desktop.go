//go:build !android
// +build !android

package check

import (
	"strings"

	sqweekDialog "github.com/sqweek/dialog"
	"github.com/owu/share-sniffer/internal/logger"
)

// 基于 github.com/sqweek/dialog 的桌面端对话框提供者
type DesktopDialogProvider struct{}

// 为DesktopDialogProvider实现DialogProvider接口
func (d *DesktopDialogProvider) ShowError(message string) {
	// 使用 github.com/sqweek/dialog 显示错误信息
	sqweekDialog.Message("错误", message).Error()
}

func (d *DesktopDialogProvider) ShowInfo(message string) {
	// 使用 github.com/sqweek/dialog 显示信息
	sqweekDialog.Message("信息", message).Info()
}

// 桌面平台的对话框提供者获取函数
func getDesktopDialogProvider() DialogProvider {
	return &DesktopDialogProvider{}
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
