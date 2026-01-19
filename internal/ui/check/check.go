// Package check 提供了检测功能的用户界面和逻辑

package check

import (
	"fmt"
	"image/color"
	"os"
	"strings"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	fyneDialog "fyne.io/fyne/v2/dialog"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/ui/icons"
	"share-sniffer/internal/ui/state"
	"share-sniffer/internal/utils"
)

// DialogProvider 是对话框功能的抽象接口
type DialogProvider interface {
	ShowError(message string)
	ShowInfo(message string, title string)
}

// CheckUI 负责检测功能的用户界面和逻辑
type CheckUI struct {
	window         fyne.Window
	state          *state.AppState
	resultTable    *fyne.Container
	isChecking     bool
	stopChan       chan struct{}
	dialogProvider DialogProvider
	// 表格数据，用于UI和检测结果共享
	tableDataWrapper struct {
		Data  [][]string
		Mutex sync.RWMutex
	}
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

	// 加载水印图片
	bg := loadBg()

	// 如果水印图片加载成功，使用Stack布局将水印放在表格下方
	if bg != nil {
		// 创建Stack布局，水印放在底层，表格放在上层
		stackContainer := container.NewStack(
			bg,
			tableContainer,
		)
		q.resultTable = container.NewPadded(stackContainer)
	} else {
		// 如果水印图片加载失败，使用默认布局
		q.resultTable = container.NewPadded(tableContainer)
	}

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

// 加载水印图片
func loadBg() *canvas.Image {
	// 使用内嵌资源加载水印图片
	bg := canvas.NewImageFromResource(icons.LogoBg)
	if bg == nil {
		logger.Error("加载水印图片失败")
		return nil
	}

	// 设置图片属性
	bg.FillMode = canvas.ImageFillContain // 保持比例填充

	return bg
}

// 基于 Fyne 原生对话框的提供者（用于Android平台）
type FyneDialogProvider struct {
	window fyne.Window
}

// 为FyneDialogProvider实现DialogProvider接口
func (d *FyneDialogProvider) ShowError(message string) {
	// 使用Fyne原生的错误对话框
	fyneDialog.ShowError(fmt.Errorf("%s", message), d.window)
}

func (d *FyneDialogProvider) ShowInfo(message string, title string) {
	// 使用Fyne原生的信息对话框（带图标）
	fyneDialog.ShowInformation(title, message, d.window)
}

// ShowTxt 显示不带图标的文本对话框
func (d *FyneDialogProvider) ShowTxt(message string, title string) {
	// 创建不带图标的自定义文本对话框
	content := container.NewVBox(
		widget.NewLabel(message),
	)

	// 创建对话框
	dialog := fyneDialog.NewCustom(title, "确定", content, d.window)
	dialog.Resize(fyne.NewSize(200, 100))
	dialog.Show()
}

// 根据平台获取合适的对话框提供者
func getDialogProvider(window fyne.Window) DialogProvider {
	// 如果是Android平台，使用Fyne原生对话框
	if !utils.IsDesktop() {
		return &FyneDialogProvider{window: window}
	}
	// 其他平台（Windows、Linux等）使用DesktopDialogProvider
	return getDesktopDialogProvider(window)
}

// OpenFile 打开文件选择对话框，根据平台选择不同的实现
func (q *CheckUI) OpenFile() {
	startTime := time.Now()
	defer logger.Debug("OpenFile方法执行完毕，耗时: %v", time.Since(startTime).Milliseconds())

	// 根据平台选择不同的文件选择对话框实现
	if !utils.IsDesktop() {
		// Android平台使用Fyne原生的文件选择对话框
		q.openFileWithFyneDialog()
	} else {
		// 桌面平台使用github.com/sqweek/dialog
		q.openFileWithSqweekDialog()
	}
}

// openFileWithFyneDialog 使用Fyne原生的文件选择对话框（Android平台）
func (q *CheckUI) openFileWithFyneDialog() {
	fileDialog := fyneDialog.NewFileOpen(
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
