package check

import (
	"bufio"
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/storage"
	"fyne.io/fyne/v2/widget"
	"github.com/samber/lo"
	"share-sniffer/internal/config"
	"share-sniffer/internal/core"
	"share-sniffer/internal/logger"
	"share-sniffer/internal/utils"
	"share-sniffer/internal/workerpool"
)

// taskResult 表示检测任务的结果
type taskResult struct {
	index  int
	result utils.Result
}

// headerLayout 实现固定列宽的表头布局
type headerLayout struct {
	widths []float32
}

func (l *headerLayout) Layout(objects []fyne.CanvasObject, size fyne.Size) {
	x := float32(0)
	for i, obj := range objects {
		if i < len(l.widths) {
			w := l.widths[i]
			obj.Move(fyne.NewPos(x, 0))
			obj.Resize(fyne.NewSize(w, size.Height))
			x += w
		}
	}
}

func (l *headerLayout) MinSize(objects []fyne.CanvasObject) fyne.Size {
	w := float32(0)
	for _, width := range l.widths {
		w += width
	}
	return fyne.NewSize(w, 30) // 默认高度30
}

// 创建表头容器
func (q *CheckUI) createHeaderContainer() *fyne.Container {
	// 定义列宽，需与dataTable.SetColumnWidth保持一致
	colWidths := []float32{50, 400, 70, 70, 100}

	// 创建布局
	layout := &headerLayout{widths: colWidths}

	headerContainer := container.New(layout)

	// 创建并添加表头标签，设置样式
	// 使用NewPadded与表格内容对齐（表格内容使用了NewPadded）
	indexHeaderLabel := widget.NewLabel("序号")
	indexHeaderLabel.Importance = widget.HighImportance
	indexHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	indexHeaderLabel.Alignment = fyne.TextAlignLeading
	headerContainer.Add(container.NewPadded(indexHeaderLabel))

	urlHeaderLabel := widget.NewLabel("网址")
	urlHeaderLabel.Importance = widget.HighImportance
	urlHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	urlHeaderLabel.Alignment = fyne.TextAlignCenter
	headerContainer.Add(container.NewPadded(urlHeaderLabel))

	statusHeaderLabel := widget.NewLabel("状态")
	statusHeaderLabel.Importance = widget.HighImportance
	statusHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	statusHeaderLabel.Alignment = fyne.TextAlignCenter
	headerContainer.Add(container.NewPadded(statusHeaderLabel))

	timeHeaderLabel := widget.NewLabel("耗时ms")
	timeHeaderLabel.Importance = widget.HighImportance
	timeHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	timeHeaderLabel.Alignment = fyne.TextAlignCenter
	headerContainer.Add(container.NewPadded(timeHeaderLabel))

	noteHeaderLabel := widget.NewLabel("信息")
	noteHeaderLabel.Importance = widget.HighImportance
	noteHeaderLabel.TextStyle = fyne.TextStyle{Bold: true}
	noteHeaderLabel.Alignment = fyne.TextAlignCenter
	headerContainer.Add(container.NewPadded(noteHeaderLabel))

	return headerContainer
}

// 创建数据表格
func (q *CheckUI) createDataTable(tableData [][]string, mutex *sync.RWMutex) *container.Scroll {
	logger.Debug("开始创建数据表格，行数: %d", len(tableData))

	startTime := time.Now()
	defer logger.Debug("数据表格创建完成，耗时: %v", time.Since(startTime))

	// 自定义表头
	headers := []string{"序号", "网址", "状态", "耗时ms", "信息"}

	// 手动创建表格
	logger.Debug("创建Table组件")
	dataTable := widget.NewTable(
		func() (int, int) {
			logger.Debug("Table尺寸回调: 行=%d, 列=%d", len(tableData), len(headers))
			return len(tableData), len(headers)
		},
		func() fyne.CanvasObject {
			logger.Debug("创建表格单元格组件")
			return container.NewPadded(widget.NewLabel(""))
		},
		func(id widget.TableCellID, obj fyne.CanvasObject) {
			logger.Debug("更新表格单元格: 行=%d, 列=%d", id.Row, id.Col)
			containerObj := obj.(*fyne.Container)

			// 使用锁保护tableData读取
			if mutex != nil {
				mutex.RLock()
				defer mutex.RUnlock()
			}

			// 数据行
			if id.Row < len(tableData) && id.Col < len(tableData[id.Row]) {
				// 序号列（第0列）
				if id.Col == 0 {
					// 如果序号列已经有值，使用它；否则根据行号生成
					value := tableData[id.Row][id.Col]
					if value == "" {
						value = fmt.Sprintf("%d", id.Row+1)
					}
					containerObj.Objects = []fyne.CanvasObject{
						widget.NewLabel(value),
					}
				} else {
					value := tableData[id.Row][id.Col]

					// 特殊处理网址列（第1列）
					if id.Col == 1 {
						// 创建可点击的超链接
						hyperlink := widget.NewHyperlink(utils.Substr(value, 50, ""), nil)
						// 设置左对齐
						hyperlink.Alignment = fyne.TextAlignLeading
						// 使用标准库解析URL
						parsedURL, err := url.Parse(value)
						if err == nil {
							hyperlink.URL = parsedURL
						}
						hyperlink.OnTapped = func() {
							if hyperlink.URL != nil {
								// 尝试打开链接，失败时不处理错误
								fyne.CurrentApp().OpenURL(hyperlink.URL)
							}
						}
						containerObj.Objects = []fyne.CanvasObject{hyperlink}
					} else {
						// 其他列使用普通标签
						containerObj.Objects = []fyne.CanvasObject{
							widget.NewLabel(value),
						}
						label := containerObj.Objects[0].(*widget.Label)

						// 根据状态设置不同颜色
						if id.Col == 2 { // 状态列
							if value == utils.ValidTxt {
								label.Importance = widget.SuccessImportance // 绿色
							} else if value == utils.TimeoutTxt || value == utils.MalformedTxt || value == utils.FatalTxt {
								label.Importance = widget.HighImportance // 红色高亮
							} else if value == utils.InvalidTxt {
								label.Importance = widget.DangerImportance // 红色高亮
							} else if value == utils.UnknownTxt {
								label.Importance = widget.LowImportance // 红色高亮
							} else if value == utils.DoingTxt {
								label.Importance = widget.MediumImportance // 黄色
							} else if value == utils.StopTxt {
								label.Importance = widget.WarningImportance // 橙色
							}
						}
					}
				}
				containerObj.Refresh()
			}
		})

	// 调整数据表格列宽
	dataTable.SetColumnWidth(0, 50)  // 序号列
	dataTable.SetColumnWidth(1, 400) // 网址列
	dataTable.SetColumnWidth(2, 70)  // 状态列
	dataTable.SetColumnWidth(3, 70)  // 耗时列
	dataTable.SetColumnWidth(4, 100) // 备注列

	// 创建可滚动的数据表格容器
	dataTableContainer := container.NewScroll(dataTable)
	dataTableContainer.SetMinSize(fyne.NewSize(720, 500))

	return dataTableContainer
}

// 更新表格显示
func (q *CheckUI) updateTableDisplay(headerContainer *fyne.Container, dataTableContainer *container.Scroll) {
	logger.Debug("开始更新表格显示")

	startTime := time.Now()
	defer logger.Debug("表格显示更新完成，耗时: %v", time.Since(startTime))

	// 创建包含表头和可滚动表格内容的垂直容器
	// 使用Border布局，让dataTableContainer填充剩余垂直空间
	newTableContainer := container.NewBorder(
		headerContainer,    // top
		nil,                // bottom
		nil,                // left
		nil,                // right
		dataTableContainer, // center (expands)
	)

	if q.resultTable != nil {
		logger.Debug("更新现有表格容器内容")
		fyne.Do(func() {
			logger.Debug("在GUI线程中更新表格内容")

			// 检查当前容器是否是Stack布局（有水印）
			_, isStack := q.resultTable.Objects[0].(*fyne.Container)
			if isStack && len(q.resultTable.Objects[0].(*fyne.Container).Objects) > 1 {
				// 如果是Stack布局且有水印，只更新表格部分
				stackContainer := q.resultTable.Objects[0].(*fyne.Container)
				// 替换Stack中的第二个元素（表格容器）
				if len(stackContainer.Objects) >= 2 {
					stackContainer.Objects[1] = newTableContainer
					stackContainer.Refresh()
				}
			} else {
				// 否则直接替换整个容器
				q.resultTable.Objects = []fyne.CanvasObject{newTableContainer}
			}
			q.resultTable.Refresh()
		})
	} else {
		// 首次创建表格容器
		logger.Debug("首次创建表格容器")
		fyne.Do(func() {
			logger.Debug("在GUI线程中创建新表格容器")
			q.resultTable = newTableContainer
		})
	}
}

func supportedLinks(url string) bool {
	return lo.ContainsBy(config.GetSupportedLinks(), func(prefix string) bool {
		return strings.HasPrefix(url, prefix)
	})
}

// loadToTable 加载文件并渲染表格
func (q *CheckUI) loadToTable() {
	logger.Debug("开始执行LoadToTable方法，文件路径: %s, 文件URI: %v", q.state.FilePath, q.state.FileURI)

	startTime := time.Now()
	defer logger.Debug("LoadToTable方法执行完毕，耗时: %v", time.Since(startTime))
	// 读取文件内容
	var links []string
	var scanner *bufio.Scanner

	// 根据平台选择不同的文件读取方式
	if q.state.FileURI != nil {
		// Android平台或支持URI的平台，使用storage包读取
		reader, readErr := storage.Reader(q.state.FileURI)
		if readErr != nil {
			logger.Warn("打开文件失败: %v", readErr)
			fyne.Do(func() {
				q.dialogProvider.ShowError(fmt.Sprintf("打开文件失败:%v", readErr))
			})
			return
		}
		defer reader.Close()
		scanner = bufio.NewScanner(reader)
	} else {
		// 非Android平台，使用os.Open读取
		file, openErr := os.Open(q.state.FilePath)
		if openErr != nil {
			logger.Warn("打开文件失败: %v", openErr)
			fyne.Do(func() {
				q.dialogProvider.ShowError(fmt.Sprintf("打开文件失败:%v", openErr))
			})
			return
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	scanStart := time.Now()
	logger.Debug("开始扫描文件内容")
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && supportedLinks(line) {
			links = append(links, line)
		}
	}

	if err := scanner.Err(); err != nil {
		logger.Warn("读取文件错误: %v", err)
		fyne.Do(func() {
			q.dialogProvider.ShowError(fmt.Sprintf("读取文件错误:%v", err))
		})
		return
	}
	logger.Debug("文件扫描完成，耗时: %v", time.Since(scanStart))

	if len(links) == 0 {
		logger.Warn("未找到有效链接，文件内容格式错误")
		fyne.Do(func() {
			q.dialogProvider.ShowError("打开的分享链接文件错误")
		})
		return
	}
	logger.Debug("找到 %d 个有效链接", len(links))

	// 准备表格数据 - 初始状态为空
	dataPrepareStart := time.Now()
	logger.Debug("开始准备表格数据")
	tableData := make([][]string, len(links))
	for i, link := range links {
		tableData[i] = []string{"", link, "", "", ""} // 初始状态字段为空，序号会在渲染时自动生成
	}
	logger.Debug("表格数据准备完成，耗时: %v", time.Since(dataPrepareStart))

	// 保存表格数据到CheckUI实例的tableDataWrapper中
	logger.Debug("保存表格数据到实例的tableDataWrapper中")
	q.tableDataWrapper.Mutex.Lock()
	q.tableDataWrapper.Data = tableData
	q.tableDataWrapper.Mutex.Unlock()

	// 创建表头和数据表格
	tableCreateStart := time.Now()
	logger.Debug("开始创建表格组件")
	headerContainer := q.createHeaderContainer()
	logger.Debug("表头容器创建完成")
	dataTableContainer := q.createDataTable(q.tableDataWrapper.Data, &q.tableDataWrapper.Mutex)
	logger.Debug("数据表格容器创建完成")

	// 更新表格显示
	logger.Debug("开始更新表格显示")
	q.updateTableDisplay(headerContainer, dataTableContainer)
	logger.Debug("表格创建和显示完成，耗时: %v", time.Since(tableCreateStart))
}

func (q *CheckUI) CheckFile() {
	logger.Debug("开始执行CheckFile方法")
	startTime := time.Now()
	defer logger.Debug("CheckFile方法执行完毕，总耗时: %v", time.Since(startTime))

	// 定义共享的完成计数变量
	var completedCount int32 = 0

	if q.state.StandardTime > config.ExpirationDate() {
		logger.Warn("该版本已过期，请升级后再试")
		q.dialogProvider.ShowInfo(fmt.Sprintf("该版本已过期，请升级后再试"), "提示")
		return
	}

	// 如果正在检测，则停止检测
	if q.isChecking {
		logger.Debug("正在检测中，用户点击停止")
		// 发送停止信号
		select {
		case <-q.stopChan:
			// 通道已经关闭，避免重复关闭
		default:
			close(q.stopChan)
			logger.Debug("停止通道已关闭")
		}

		// 更新按钮状态和文本
		fyne.Do(func() {
			logger.Debug("更新UI：恢复按钮状态")
			q.fileCheckButton.SetText("检测")
			q.fileEntry.Enable()
			q.fileOpenButton.Enable()
		})

		q.isChecking = false
		logger.Debug("检测已停止")
		return
	}

	// 初始化停止通道
	q.stopChan = make(chan struct{})
	q.isChecking = true
	logger.Debug("初始化检测环境完成")

	// 确保在GUI线程中禁用控件并更改按钮文本
	fyne.Do(func() {
		logger.Debug("更新UI：禁用控件并更改按钮文本为停止")
		q.fileEntry.Disable()
		q.fileOpenButton.Disable()
		q.fileCheckButton.SetText("停止")
	})

	// 从文件中加载链接
	var links []string
	fileLoadStart := time.Now()
	logger.Debug("开始从文件加载链接: %s, URI: %v", q.state.FilePath, q.state.FileURI)

	// 检查是否已打开文件
	if q.state.FilePath == "" && q.state.FileURI == nil {
		logger.Warn("未选择任何文件")
		fyne.Do(func() {
			q.dialogProvider.ShowError("请先打开包含分享链接的文件")
			q.fileCheckButton.SetText("检测")
			q.fileEntry.Enable()
			q.fileOpenButton.Enable()
		})
		q.isChecking = false
		return
	}

	// 根据平台选择不同的文件读取方式
	var scanner *bufio.Scanner

	if q.state.FileURI != nil {
		// Android平台或支持URI的平台，使用storage包读取
		reader, readErr := storage.Reader(q.state.FileURI)
		if readErr != nil {
			logger.Error("打开文件失败: %v", readErr)
			fyne.Do(func() {
				q.dialogProvider.ShowError("打开分享链接文件失败")
				q.fileCheckButton.SetText("检测")
				q.fileEntry.Enable()
				q.fileOpenButton.Enable()
			})
			q.isChecking = false
			return
		}
		defer reader.Close()
		scanner = bufio.NewScanner(reader)
	} else {
		// 非Android平台，使用os.Open读取
		file, openErr := os.Open(q.state.FilePath)
		if openErr != nil {
			logger.Error("打开文件失败: %v", openErr)
			fyne.Do(func() {
				q.dialogProvider.ShowError("打开分享链接文件失败")
				q.fileCheckButton.SetText("检测")
				q.fileEntry.Enable()
				q.fileOpenButton.Enable()
			})
			q.isChecking = false
			return
		}
		defer file.Close()
		scanner = bufio.NewScanner(file)
	}

	// 优化大文件读取，支持最多9999个链接
	linkCount := 0
	maxLinks := 9999 // 限制最大处理链接数

	// 增加scanner的缓冲区大小，优化大文件读取
	scannerBuf := make([]byte, 64*1024)   // 64KB缓冲区
	scanner.Buffer(scannerBuf, 1024*1024) // 最大行长度1MB

	for scanner.Scan() && linkCount < maxLinks {
		line := strings.TrimSpace(scanner.Text())
		if line != "" && supportedLinks(line) {
			links = append(links, line)
			linkCount++
		}
	}

	// 检查是否有错误或是否达到最大链接数
	if err := scanner.Err(); err != nil {
		logger.Error("读取文件错误: %v", err)
		fyne.Do(func() {
			q.dialogProvider.ShowError(fmt.Sprintf("读取文件错误: %v", err))
			q.fileCheckButton.SetText("检测")
			q.fileEntry.Enable()
			q.fileOpenButton.Enable()
		})
		q.isChecking = false
		return
	}

	// 如果文件中的链接超过最大限制，给用户提示
	if linkCount >= maxLinks {
		logger.Warn("文件中链接数量超过最大限制 %d，仅处理前 %d 个链接", maxLinks, maxLinks)
		fyne.Do(func() {
			q.dialogProvider.ShowInfo(fmt.Sprintf("文件中链接数量超过最大限制 %d，仅处理前 %d 个链接", maxLinks, maxLinks), "提示")
		})
	}

	logger.Debug("文件加载完成，共读取 %d 个链接，耗时: %v", len(links), time.Since(fileLoadStart))

	if len(links) == 0 {
		logger.Warn("未找到有效链接")
		fyne.Do(func() {
			q.dialogProvider.ShowError("请打开包含分享链接的文件")
			q.fileCheckButton.SetText("检测")
			q.fileEntry.Enable()
			q.fileOpenButton.Enable()
		})
		q.isChecking = false
		return
	}

	// 初始化表格数据到实例的tableDataWrapper中
	logger.Debug("开始初始化表格数据")
	q.tableDataWrapper.Mutex.Lock()
	q.tableDataWrapper.Data = make([][]string, len(links))
	for i := 0; i < len(links); i++ {
		q.tableDataWrapper.Data[i] = []string{fmt.Sprintf("%d", i+1), links[i], utils.DoingTxt, "", ""}
	}
	q.tableDataWrapper.Mutex.Unlock()
	logger.Debug("表格数据初始化完成，共 %d 行数据", len(q.tableDataWrapper.Data))

	// 更新表格显示 - 使用抽象的方法创建表头和数据表格
	tableUpdateStart := time.Now()
	fyne.Do(func() {
		logger.Debug("开始更新表格显示")
		// 创建表头和数据表格
		headerContainer := q.createHeaderContainer()
		dataTableContainer := q.createDataTable(q.tableDataWrapper.Data, &q.tableDataWrapper.Mutex)

		// 更新表格显示
		q.updateTableDisplay(headerContainer, dataTableContainer)
		logger.Debug("表格显示更新完成，耗时: %v", time.Since(tableUpdateStart))
	})

	// 创建工作池并启动
	logger.Debug("开始创建并启动工作池")
	pool := workerpool.NewWorkerPool()
	pool.Start()
	logger.Debug("工作池启动成功")

	// 统计变量
	var (
		n_total   int32
		n_valid   int32
		n_invalid int32
		n_error   int32
	)

	// 提交所有任务到工作池，分批处理以优化性能和内存使用
	taskSubmitStart := time.Now()
	totalLinks := len(links)
	logger.Debug("开始提交 %d 个任务到工作池，分批处理优化性能", totalLinks)

	// 分批提交任务，每批处理一定数量，避免一次性提交所有任务导致内存压力
	batchSize := 500
	for batchStart := 0; batchStart < totalLinks; batchStart += batchSize {
		batchEnd := batchStart + batchSize
		if batchEnd > totalLinks {
			batchEnd = totalLinks
		}

		// 提交当前批次的任务
		for i := batchStart; i < batchEnd; i++ {
			// 检查是否已停止
			select {
			case <-q.stopChan:
				logger.Info("检测到停止信号，停止提交更多任务")
				goto stopSubmission
			default:
				// 继续提交
			}

			index := i
			url := links[i]

			// 创建任务
			task := workerpool.Task{
				URL: url,
				Func: func(ctx context.Context) interface{} {
					taskStartTime := time.Now()
					// 对于大量任务，降低日志级别以减少日志开销
					if totalLinks < 1000 {
						logger.Debug("开始执行任务 #%d: %s", index+1, url)
					} else {
						logger.Debug("开始执行任务 #%d: %s", index+1, url)
					}

					// 首先更新状态为检测中（确保UI显示正确）
					q.tableDataWrapper.Mutex.Lock()
					if q.tableDataWrapper.Data[index][2] != utils.StopTxt {
						q.tableDataWrapper.Data[index][2] = utils.DoingTxt
					}
					q.tableDataWrapper.Mutex.Unlock()

					// 检查是否收到停止信号
					select {
					case <-q.stopChan:
						logger.Debug("任务 #%d 收到停止信号", index+1)
						return taskResult{index: index, result: utils.Result{Error: utils.Stop}} // 表示已停止
					case <-ctx.Done():
						logger.Debug("任务 #%d 上下文已取消", index+1)
						return taskResult{index: index, result: utils.Result{Error: utils.Done}} // 上下文取消也视为停止
					default:
						// 继续检测
					}

					// 调用core包中的Check方法检测网址
					result := core.Adapter(ctx, url)

					// 根据任务数量调整日志级别
					if totalLinks < 1000 {
						logger.Debug("任务 #%d 检测完成，状态: %d, 耗时: %v", index+1, result.Error, time.Since(taskStartTime))
					} else {
						logger.Debug("任务 #%d 检测完成，状态: %d, 耗时: %v", index+1, result.Error, time.Since(taskStartTime))
					}

					// 再次检查停止信号
					select {
					case <-q.stopChan:
						logger.Debug("任务 #%d 结果处理前收到停止信号", index+1)
						return taskResult{index: index, result: utils.Result{Error: utils.Stop}}
					case <-ctx.Done():
						logger.Debug("任务 #%d 结果处理前上下文已取消", index+1)
						return taskResult{index: index, result: utils.Result{Error: utils.Done}}
					default:
						// 继续处理，返回实际结果
						return taskResult{index: index, result: result}
					}
				},
			}

			// 提交任务到工作池，带重试逻辑
			submitSuccess := false
			maxRetries := 5
			retryDelay := 300 * time.Millisecond

			for attempt := 0; attempt < maxRetries; attempt++ {
				success := pool.Submit(task)
				if success {
					submitSuccess = true
					break
				}

				// 如果提交失败且不是最后一次尝试，等待后重试
				if attempt < maxRetries-1 {
					logger.Warn("任务 #%d 提交失败，正在重试 (尝试 %d/%d)", index+1, attempt+1, maxRetries)
					time.Sleep(retryDelay)
					// 指数退避策略
					retryDelay *= 2
				}
			}

			if !submitSuccess {
				logger.Error("任务 #%d 提交失败，已达最大重试次数", index+1)
				// 更新任务状态为失败
				go func(idx int) {
					q.tableDataWrapper.Mutex.Lock()
					defer q.tableDataWrapper.Mutex.Unlock()
					if q.tableDataWrapper.Data[idx][2] == utils.DoingTxt {
						q.tableDataWrapper.Data[idx][2] = utils.MalformedTxt
						q.tableDataWrapper.Data[idx][4] = "任务提交失败"
					}

					// 在GUI线程中刷新表格
					fyne.Do(func() {
						if q.resultTable != nil {
							q.resultTable.Refresh()
						}
					})

					// 增加完成计数
					atomic.AddInt32(&completedCount, 1)
				}(index)
			}
		}

		// 每批次提交后短暂暂停，避免系统资源占用过高
		if batchEnd < totalLinks {
			logger.Debug("批次提交完成 (批次 %d-%d/%d)，短暂暂停以优化资源使用", batchStart+1, batchEnd, totalLinks)
			time.Sleep(500 * time.Millisecond)
		}
	}

stopSubmission:
	logger.Debug("所有任务提交完成，耗时: %v", time.Since(taskSubmitStart))

	totalTasks := len(links)

	// 处理任务结果
	go func() {
		logger.Debug("开始处理任务结果，预计处理 %d 个任务", len(links))
		resultProcessStart := time.Now()
		resultsChan := pool.Results()

		// 监听停止信号的goroutine
		go func() {
			logger.Debug("启动停止信号监听协程")
			select {
			case <-q.stopChan:
				logger.Debug("收到停止信号，开始更新所有剩余任务状态")
				// 更新所有剩余的检测状态为"已停止"
				q.tableDataWrapper.Mutex.Lock()
				pendingTasks := 0
				for i := range q.tableDataWrapper.Data {
					if q.tableDataWrapper.Data[i][2] == utils.DoingTxt {
						q.tableDataWrapper.Data[i][2] = utils.StopTxt
						pendingTasks++
					}
				}
				q.tableDataWrapper.Mutex.Unlock()
				logger.Debug("已将 %d 个待处理任务标记为已停止", pendingTasks)

				// 刷新表格
				fyne.Do(func() {
					logger.Debug("更新UI：刷新表格以显示已停止状态")
					if q.resultTable != nil {
						q.resultTable.Refresh()
					}
				})

				// 停止工作池，关闭结果通道
				logger.Debug("停止工作池")
				pool.Stop()
			}
		}()

		// 等待工作池完成所有任务
		go func() {
			logger.Debug("启动任务完成监控协程，总任务数: %d", totalTasks)
			// 等待所有任务完成
			for atomic.LoadInt32(&completedCount) < int32(totalTasks) {
				// 检查是否已停止
				select {
				case <-q.stopChan:
					logger.Debug("任务监控协程收到停止信号，退出")
					return
				default:
					time.Sleep(300 * time.Millisecond)
					logger.Debug("任务进度: %d/%d", atomic.LoadInt32(&completedCount), totalTasks)
				}
			}

			// 确保工作池完成清理
			logger.Debug("所有任务已完成，开始清理工作池")
			pool.Wait()

			// 计算总链接数
			n_total = int32(len(links))

			// 所有任务完成后，如果仍在检测中，恢复按钮状态
			if q.isChecking {
				logger.Debug("所有任务完成，恢复UI状态")
				fyne.Do(func() {
					logger.Debug("更新UI：恢复按钮和输入框状态")
					q.fileCheckButton.SetText("检测")
					q.fileEntry.Enable()
					q.fileOpenButton.Enable()

					// 显示统计数据
					q.dialogProvider.ShowInfo(fmt.Sprintf("总数:%d, 有效:%d, 失效:%d, 其他:%d", n_total, n_valid, n_invalid, n_error), "提示")

					q.isChecking = false
				})
			}
		}()

		// 处理结果通道中的所有结果
		resultProcessStart = time.Now()
		processedCount := 0
		// 当处理大量任务时，降低日志频率
		logInterval := 1
		if totalTasks > 1000 {
			logInterval = 100 // 每100个任务记录一次日志
		}

		// 确保处理所有结果，包括可能在通道中的所有任务
		timeoutCtx, timeoutCancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer timeoutCancel()

		// 使用带有超时的循环来处理结果，避免无限阻塞
		for processedCount < totalTasks {
			select {
			case <-timeoutCtx.Done():
				logger.Warn("结果处理超时，已处理 %d/%d 个任务", processedCount, totalTasks)
				goto resultProcessDone
			case <-q.stopChan:
				logger.Info("结果处理时检测到停止信号，退出结果处理循环")
				goto resultProcessDone
			case result, ok := <-resultsChan:
				if !ok {
					logger.Debug("结果通道已关闭，退出结果处理循环")
					break
				}
				processedCount++

				// 根据任务数量调整日志频率
				if processedCount%logInterval == 0 {
					logger.Debug("已处理 %d/%d 个任务，进度: %.1f%%", processedCount, totalTasks, float64(processedCount)/float64(totalTasks)*100)
				}

				// 检查是否已停止
				select {
				case <-q.stopChan:
					// 已经停止，跳过处理
					if processedCount%logInterval == 0 {
						logger.Debug("结果处理时检测到已停止状态，跳过")
					}
					continue
				default:
					// 继续处理
				}

				if result.Err != nil {
					// 记录错误但继续处理其他结果
					logger.Error("任务执行出错: %v", result.Err)
					continue
				}

				// 类型断言获取任务结果
				taskRes, ok := result.Value.(taskResult)
				if !ok {
					logger.Warn("无法将结果转换为taskResult类型")
					continue
				}

				index := taskRes.index
				checkResult := taskRes.result

				// 检查索引是否有效
				q.tableDataWrapper.Mutex.RLock()
				indexValid := index >= 0 && index < len(q.tableDataWrapper.Data)
				q.tableDataWrapper.Mutex.RUnlock()
				if !indexValid {
					logger.Warn("无效的任务索引: %d", index)
					continue
				}

				// 根据结果状态更新表格
				q.tableDataWrapper.Mutex.Lock()
				// 只有当前状态不是已停止时才更新
				if q.tableDataWrapper.Data[index][2] != "已停止" {
					if checkResult.Error == utils.Stop || checkResult.Error == utils.Done {
						q.tableDataWrapper.Data[index][2] = "已停止"
						logger.Debug("任务 #%d 已停止", index+1)
					} else {
						statusText := utils.UnknownTxt
						if checkResult.Error == utils.Valid {
							statusText = utils.ValidTxt
							logger.Debug("任务 #%d 检测正常: %s", index+1, checkResult.Data.Name)
							atomic.AddInt32(&n_valid, 1)
						} else if checkResult.Error == utils.Invalid {
							statusText = utils.InvalidTxt
							logger.Debug("任务 #%d 检测失败", index+1)
							atomic.AddInt32(&n_invalid, 1)
						} else if checkResult.Error == utils.Malformed || checkResult.Error == utils.Timeout || checkResult.Error == utils.Fatal {
							if checkResult.Error == utils.Malformed {
								statusText = utils.MalformedTxt
								logger.Debug("任务 #%d 检测参数错误", index+1)
							} else if checkResult.Error == utils.Timeout {
								statusText = utils.TimeoutTxt
								logger.Debug("任务 #%d 检测超时", index+1)
							} else if checkResult.Error == utils.Fatal {
								statusText = utils.FatalTxt
								logger.Debug("任务 #%d 检测异常", index+1)
							}
							atomic.AddInt32(&n_error, 1)
						}
						q.tableDataWrapper.Data[index][2] = statusText
						q.tableDataWrapper.Data[index][3] = fmt.Sprintf("%d", checkResult.Data.Elapsed)
						if checkResult.Error == utils.Valid {
							q.tableDataWrapper.Data[index][4] = checkResult.Data.Name
						} else {
							q.tableDataWrapper.Data[index][4] = checkResult.Msg
						}

					}
				}
				q.tableDataWrapper.Mutex.Unlock()

				// 立即刷新UI，确保状态及时更新
				fyne.Do(func() {
					// 对于大量任务，降低UI刷新日志的详细程度
					if totalTasks < 1000 {
						logger.Debug("更新UI：刷新表格显示任务 #%d 结果", index+1)
					}
					if q.resultTable != nil {
						q.resultTable.Refresh()
						// 额外触发子组件刷新，确保所有元素都正确更新
						for _, obj := range q.resultTable.Objects {
							if scrollObj, ok := obj.(*container.Scroll); ok {
								scrollObj.Refresh()
							}
						}
					}
				})

				// 增加完成计数（线程安全）
				completed := atomic.AddInt32(&completedCount, 1)
				logger.Debug("任务 #%d 处理完成，进度: %d/%d", index+1, completed, totalTasks)

				// 轻量级限速，避免请求过快
				time.Sleep(100 * time.Millisecond)
			}
		}

	resultProcessDone:
		logger.Debug("所有结果处理完成，已处理 %d/%d 个任务，耗时: %v", processedCount, totalTasks, time.Since(resultProcessStart))
		logger.Debug("所有结果处理完成，总耗时: %v", time.Since(resultProcessStart))

		// 检查是否已被外部停止
		if q.isChecking {
			// 所有检测完成后的处理
			logger.Debug("所有检测任务处理完成，准备恢复UI状态")
			// 恢复按钮状态
			fyne.Do(func() {
				logger.Debug("更新UI：所有任务完成，恢复按钮和输入框状态")
				q.fileCheckButton.SetText("检测")
				q.fileEntry.Enable()
				q.fileOpenButton.Enable()
			})
			q.isChecking = false
			logger.Debug("UI状态恢复完成")
		}
		logger.Debug("结果处理协程退出，总耗时: %v", time.Since(resultProcessStart))
	}()

	// 确保无论如何都会清理资源
	defer func() {
		// 确保工作池被停止，但避免重复停止
		// pool.Stop() 已经在停止信号处理和任务完成时被调用
		logger.Debug("CheckFile方法defer执行")
	}()
}
