//go:build android
// +build android

package check

import (
	"fyne.io/fyne/v2"
	"share-sniffer/internal/logger"
)

// Android平台的对话框提供者获取函数
func getDesktopDialogProvider(window fyne.Window) DialogProvider {
	// Android平台不应该调用此函数
	panic("Desktop dialog provider should not be called on Android")
}

// openFileWithSqweekDialog 在Android平台上的安全实现
// 这个方法不应该在Android平台上被调用，因为Android平台会使用openFileWithFyneDialog
func (q *CheckUI) openFileWithSqweekDialog() {
	// Android平台不应该调用此方法，记录错误并使用Fyne原生对话框
	logger.Error("openFileWithSqweekDialog should not be called on Android platform")
	q.openFileWithFyneDialog()
}
