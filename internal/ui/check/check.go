package check

import (
	"image/color"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/theme"
	"fyne.io/fyne/v2/widget"
	"github.com/owu/share-sniffer/internal/ui/state"
)

// CheckUI 负责检测功能的用户界面和逻辑
type CheckUI struct {
	window      fyne.Window
	state       *state.AppState
	resultTable *fyne.Container
	isChecking  bool
	stopChan    chan struct{}
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
		window:     window,
		state:      state,
		isChecking: false,
		stopChan:   make(chan struct{}),
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
	table := widget.NewTable(
		func() (int, int) {
			return 0, 0 // 0行0列，完全空表格
		},
		func() fyne.CanvasObject {
			return widget.NewLabel("")
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			// 空实现，不会被调用因为表格大小为0
		},
	)

	return table
}
