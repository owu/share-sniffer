package workerpool

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/owu/share-sniffer/internal/config"
	"github.com/owu/share-sniffer/internal/logger"
)

// Task 表示一个工作任务
type Task struct {
	URL  string
	Func func(ctx context.Context) interface{}
}

// Result 表示任务执行结果
type Result struct {
	Value interface{}
	Err   error
}

// WorkerPool 工作池结构
type WorkerPool struct {
	workers           int
	queueSize         int
	taskQueue         chan Task
	resultChan        chan Result
	wg                sync.WaitGroup
	ctx               context.Context
	cancel            context.CancelFunc
	// 长耗时任务并发控制
	longTaskSemaphore   chan struct{}
	maxLongConcurrent int
}

// NewWorkerPool 创建新的工作池
func NewWorkerPool() *WorkerPool {
	cfg := config.GetConfig()
	workers := cfg.CheckConfig.MaxConcurrentTasks
	// 大幅增加队列容量，使其能够处理最多10000个任务
	// 保持结果通道容量适中，避免内存占用过高
	queueSize := 10000     // 足够处理最多9999个链接
	resultQueueSize := 100 // 结果通道保持合理容量
	
	// 长耗时任务并发控制
	maxLongConcurrent := config.GetLongMaxConcurrent()

	ctx, cancel := context.WithCancel(context.Background())
	pool := &WorkerPool{
		workers:             workers,
		queueSize:           queueSize,
		taskQueue:           make(chan Task, queueSize),
		resultChan:          make(chan Result, resultQueueSize),
		ctx:                 ctx,
		cancel:              cancel,
		longTaskSemaphore:     make(chan struct{}, maxLongConcurrent),
		maxLongConcurrent: maxLongConcurrent,
	}

	return pool
}

// Start 启动工作池
func (q *WorkerPool) Start() {
	startTime := time.Now()
	logger.Debug("启动工作池: workers=%d, queue_size=%d", q.workers, q.queueSize)

	// 启动工作协程
	for i := 0; i < q.workers; i++ {
		q.wg.Add(1)
		go q.worker(i)
	}
	logger.Debug("工作池启动完成，耗时: %v", time.Since(startTime))
}

// worker 工作协程
func (q *WorkerPool) worker(id int) {
	defer q.wg.Done()
	logger.Debug("工作协程 %d 启动", id)

	for {
		select {
		case <-q.ctx.Done():
			logger.Debug("工作协程 %d 收到停止信号", id)
			return
		case task, ok := <-q.taskQueue:
			if !ok {
				logger.Debug("工作协程 %d 任务队列关闭", id)
				return
			}

			logger.Debug("工作协程 %d 收到任务: %s", id, task.URL)
			
			// 判断是否为长耗时任务（迅雷和139云盘链接）
	isLongTask := false
	for _, prefix := range config.GetSupportedXunlei() {
		if strings.HasPrefix(task.URL, prefix) {
			isLongTask = true
			break
		}
	}
	// 检查是否为139云盘链接
	if !isLongTask {
		for _, prefix := range config.GetSupportedYd() {
			if strings.HasPrefix(task.URL, prefix) {
				isLongTask = true
				break
			}
		}
	}
		
		// 为每个任务创建带超时的context，直接使用工作池上下文，不创建新的超时
		taskCtx := q.ctx
		if isLongTask {
			// 使用信号量控制长耗时任务并发数
			logger.Debug("尝试获取长耗时任务并发信号量，当前信号量队列长度: %d/%d", len(q.longTaskSemaphore), cap(q.longTaskSemaphore))
			
			// 等待获取信号量，这里会阻塞直到有可用资源
			q.longTaskSemaphore <- struct{}{}
			logger.Debug("成功获取长耗时任务并发信号量，当前信号量队列长度: %d/%d", len(q.longTaskSemaphore), cap(q.longTaskSemaphore))
			
			// 执行任务
			result := q.executeTask(taskCtx, task)
			
			// 释放信号量
			<-q.longTaskSemaphore
			logger.Debug("释放长耗时任务并发信号量，当前信号量队列长度: %d/%d", len(q.longTaskSemaphore), cap(q.longTaskSemaphore))
				
				// 发送结果，即使上下文已取消也要尝试发送
				select {
				case q.resultChan <- result:
					logger.Debug("工作协程 %d 任务结果已发送", id)
				case <-q.ctx.Done():
					logger.Debug("工作协程 %d 任务结果丢弃（工作池已关闭）", id)
				}
			} else {
				// 执行任务
				result := q.executeTask(taskCtx, task)

				// 发送结果，即使上下文已取消也要尝试发送
				select {
				case q.resultChan <- result:
					logger.Debug("工作协程 %d 任务结果已发送", id)
				case <-q.ctx.Done():
					logger.Debug("工作协程 %d 任务结果丢弃（工作池已关闭）", id)
				}
			}
		}
	}
}

// executeTask 执行任务并处理错误
func (q *WorkerPool) executeTask(ctx context.Context, task Task) Result {
	startTime := time.Now()
	logger.Debug("开始执行任务: %s", task.URL)
	defer func() {
		if r := recover(); r != nil {
			logger.Error("任务执行panic: %v, 耗时: %v", r, time.Since(startTime))
		}
		logger.Debug("任务执行结束，耗时: %v", time.Since(startTime))
	}()

	value := task.Func(ctx)
	return Result{Value: value, Err: nil}
}

// Submit 提交任务到工作池
func (q *WorkerPool) Submit(task Task) bool {
	submitStart := time.Now()
	// 使用非阻塞方式提交任务，如果队列已满，会返回false
	select {
	case q.taskQueue <- task:
		logger.Debug("任务提交成功，队列当前长度: %d/%d, 耗时: %v", len(q.taskQueue), cap(q.taskQueue), time.Since(submitStart))
		return true
	default:
		// 队列已满，记录警告
		logger.Warn("任务队列已满，无法提交新任务，队列长度: %d/%d", len(q.taskQueue), cap(q.taskQueue))
		return false
	}
}

// Results 获取结果通道
func (q *WorkerPool) Results() <-chan Result {
	return q.resultChan
}

// Stop 停止工作池
func (q *WorkerPool) Stop() {
	startTime := time.Now()
	logger.Debug("开始停止工作池")
	q.cancel()
	close(q.taskQueue)
	logger.Debug("等待所有工作协程结束...")
	q.wg.Wait()
	close(q.resultChan)
	logger.Debug("工作池已完全停止，耗时: %v", time.Since(startTime))
}

// Wait 等待所有任务完成
func (q *WorkerPool) Wait() {
	startTime := time.Now()
	logger.Debug("开始等待所有任务完成")
	close(q.taskQueue)
	q.wg.Wait()
	close(q.resultChan)
	logger.Debug("所有任务已完成，等待结束，耗时: %v", time.Since(startTime))
}
