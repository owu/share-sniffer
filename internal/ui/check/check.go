package check

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/owu/share-sniffer/internal/logger"
	"github.com/owu/share-sniffer/internal/ui/state"
)

// DialogProvider 是对话框功能的抽象接口
type DialogProvider interface {
	ShowError(message string)
	ShowInfo(message string)
}

// CheckUI 负责检测功能的用户界面和逻辑
type CheckUI struct {
	window         fyne.Window
	state          *state.AppState
	resultTable    *fyne.Container
	isChecking     bool
	stopChan       chan struct{}
	dialogProvider DialogProvider
	// UI组件
	fileEntry       *EntryWithEnterKeyEvent
	fileOpenButton  *widget.Button
	fileCheckButton *widget.Button
}

// EntryWithEnterKeyEvent 是一个自定义的输入框组件，支持回车键事件
type EntryWithEnterKeyEvent struct {
	widget.Entry
	OnEnterKey func() // 回车键回调函数
}

// KeyDown 处理按键事件，当按下回车键时触发OnEnterKey回调
func (q *EntryWithEnterKeyEvent) KeyDown(key *fyne.KeyEvent) {
	q.Entry.KeyDown(key)
	if key.Name == fyne.KeyReturn || key.Name == fyne.KeyEnter {
		if q.OnEnterKey != nil {
			q.OnEnterKey()
		}
	}
}

func NewCheckTab(window fyne.Window, state *state.AppState) *container.TabItem {
	ui := &CheckUI{
		window:         window,
		state:          state,
		isChecking:     false,
		stopChan:       make(chan struct{}),
		dialogProvider: getDialogProvider(window),
	}
	return ui.createTab()
}

func (q *CheckUI) createTab() *container.TabItem {

	spacer := canvas.NewRectangle(color.Transparent)
	spacer.SetMinSize(fyne.NewSize(1, 1))

	// 初始化UI组件作为结构体字段
	q.fileEntry = &EntryWithEnterKeyEvent{}
	q.fileEntry.ExtendBaseWidget(q.fileEntry)
	q.fileEntry.SetPlaceHolder("打开分享链接文本文件(.txt),每行一条分享链接（单次上限9999条）")
	q.fileOpenButton = &widget.Button{Text: "打开", OnTapped: q.OpenFile,
		Icon: theme.FileIcon()}
	q.fileCheckButton = &widget.Button{Text: "检测", OnTapped: q.CheckFile,
		Icon: theme.SearchIcon()}
	fileHbox := container.NewBorder(
		nil, nil,
		container.NewHBox(spacer, q.fileOpenButton, spacer),
		container.NewHBox(spacer, q.fileCheckButton, spacer),
		q.fileEntry,
	)

	// 创建表格容器并保存引用
	// 设置最小高度确保表格有足够的显示空间
	tableContainer := container.NewScroll(createEmptyTable())
	tableContainer.SetMinSize(fyne.NewSize(0, 400)) // 设置最小高度
	q.resultTable = container.NewPadded(tableContainer)

	// 使用BorderLayout让表格占满剩余空间
	content := container.NewBorder(
		fileHbox,      // 顶部
		nil,           // 底部
		nil,           // 左侧
		nil,           // 右侧
		q.resultTable, // 中间（会占满剩余空间）
	)

	return container.NewTabItemWithIcon(
		"检测",
		theme.SearchIcon(),
		container.NewPadded(content))
}

// 创建空表格（不渲染表头）
func createEmptyTable() *widget.Table {
	// 返回一个空表格，不显示任何内容
	// 使用更简单的实现方式创建空表格，减少初始化开销
	table := widget.NewTable(
		func() (int, int) { return 0, 0 }, // 0行0列，完全空表格
		func() fyne.CanvasObject { return widget.NewLabel("") },
		func(id widget.TableCellID, obj fyne.CanvasObject) {},
	)
	return table
}

// 默认的对话框提供者实现（作为后备方案）
type DefaultDialogProvider struct {
	window fyne.Window
}

// 默认的getDialogProvider函数实现（作为后备方案）
func getDialogProvider(window fyne.Window) DialogProvider {
	return &DefaultDialogProvider{window: window}
}

// 为DefaultDialogProvider实现DialogProvider接口
func (d *DefaultDialogProvider) ShowError(message string) {
	// 使用Fyne原生的错误对话框
	dialog.ShowError(fmt.Errorf(message), d.window)
}

func (d *DefaultDialogProvider) ShowInfo(message string) {
	// 使用Fyne原生的信息对话框
	dialog.ShowInformation("信息", message, d.window)
}

// OpenFile 打开文件选择对话框（使用Fyne原生的dialog库）
func (q *CheckUI) OpenFile() {
	startTime := time.Now()
	defer logger.Debug("OpenFile方法执行完毕，耗时: %v", time.Since(startTime).Milliseconds())

	// 使用Fyne原生的dialog库创建文件选择对话框
	fileDialog := dialog.NewFileOpen(
		func(uri fyne.URIReadCloser, err error) {
			if err != nil {
				logger.Error("文件对话框错误: %v", err)
				q.dialogProvider.ShowError(err.Error())
				return
			}
			if uri == nil {
				logger.Debug("用户取消了文件选择")
				return
			}
			defer uri.Close()

			// 获取文件URI
			fileURI := uri.URI()
			uriStr := fileURI.String()

			logger.Debug("选择的文件URI: %s", uriStr)

			// 处理Windows路径格式
			filename := uriStr

			// 移除file://前缀
			if strings.HasPrefix(filename, "file://") {
				filename = filename[7:]
			}

			// 处理Windows路径分隔符
			filename = strings.ReplaceAll(filename, "/", "\\")

			// 处理UNC路径（Windows网络路径）
			if strings.HasPrefix(filename, "\\\\") {
				// 保留UNC路径格式
			} else if !strings.HasPrefix(filename, "\\") && !strings.Contains(filename, ":") {
				// 处理相对路径，添加当前工作目录
				cwd, _ := os.Getwd()
				filename = cwd + "\\" + filename
			}

			logger.Debug("解析后的文件路径: %s", filename)

			q.fileEntry.SetText(filename)

			// 同时设置FilePath和FileURI
			q.state.FilePath = filename
			q.state.FileURI = fileURI

			q.loadToTable()
		},
		q.window,
	)

	// 设置文件过滤器
	fileDialog.SetFilter(storage.NewExtensionFileFilter([]string{".txt"}))

	// 显示文件选择对话框
	fileDialog.Show()
}
