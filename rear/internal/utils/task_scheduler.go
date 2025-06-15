package utils

import (
	"context"
	"fmt"
	"log"
	"os/exec"
	"sync"
	"sync/atomic"
	"time"
)

// Task 任务接口
type Task interface {
	// Execute 执行任务
	Execute(ctx context.Context) error
	// ID 获取任务ID
	ID() string
	// Type 获取任务类型
	Type() string

	/*
		紧急任务：优先级 8-10
		正常任务：优先级 4-7
		后台任务：优先级 1-3
	*/
	// Priority 获取任务优先级 (数字越大优先级越高)
	Priority() int
}

// TaskResult 任务执行结果
type TaskResult struct {
	TaskID   string
	TaskType string
	Success  bool
	Error    error
	Duration time.Duration
}

// TaskScheduler 任务调度器
type TaskScheduler struct {
	workers     int
	taskQueue   chan Task
	resultQueue chan TaskResult
	workerPool  chan struct{}
	wg          sync.WaitGroup
	ctx         context.Context
	cancel      context.CancelFunc

	// 统计信息
	totalTasks     int64
	completedTasks int64
	failedTasks    int64

	// 任务类型统计
	taskStats  map[string]*TaskTypeStats
	statsMutex sync.RWMutex
}

// TaskTypeStats 任务类型统计
type TaskTypeStats struct {
	Total     int64
	Completed int64
	Failed    int64
	AvgTime   time.Duration
}

// NewTaskScheduler 创建任务调度器
func NewTaskScheduler(workers int, queueSize int) *TaskScheduler {
	ctx, cancel := context.WithCancel(context.Background())

	ts := &TaskScheduler{
		workers:     workers,
		taskQueue:   make(chan Task, queueSize),
		resultQueue: make(chan TaskResult, workers),
		workerPool:  make(chan struct{}, workers),
		ctx:         ctx,
		cancel:      cancel,
		taskStats:   make(map[string]*TaskTypeStats),
	}

	// 初始化worker池
	for i := 0; i < workers; i++ {
		ts.workerPool <- struct{}{}
	}

	// 启动worker
	ts.startWorkers()

	// 启动结果处理器
	go ts.resultProcessor()

	return ts
}

// startWorkers 启动worker
func (ts *TaskScheduler) startWorkers() {
	for i := 0; i < ts.workers; i++ {
		ts.wg.Add(1)
		go ts.worker(i)
	}
}

// worker 工作协程
func (ts *TaskScheduler) worker(id int) {
	defer ts.wg.Done()

	for {
		select {
		case <-ts.ctx.Done():
			return
		case task, ok := <-ts.taskQueue:
			if !ok {
				return
			}

			// 获取worker令牌
			<-ts.workerPool

			// 执行任务
			start := time.Now()
			err := task.Execute(ts.ctx)
			duration := time.Since(start)

			// 发送结果
			ts.resultQueue <- TaskResult{
				TaskID:   task.ID(),
				TaskType: task.Type(),
				Success:  err == nil,
				Error:    err,
				Duration: duration,
			}

			// 归还worker令牌
			ts.workerPool <- struct{}{}
		}
	}
}

// resultProcessor 结果处理器
func (ts *TaskScheduler) resultProcessor() {
	for result := range ts.resultQueue {
		atomic.AddInt64(&ts.completedTasks, 1)

		if !result.Success {
			atomic.AddInt64(&ts.failedTasks, 1)
			log.Printf("Task %s failed: %v", result.TaskID, result.Error)
		}

		// 更新统计信息
		ts.updateStats(result)
	}
}

// updateStats 更新统计信息
func (ts *TaskScheduler) updateStats(result TaskResult) {
	ts.statsMutex.Lock()
	defer ts.statsMutex.Unlock()

	stats, exists := ts.taskStats[result.TaskType]
	if !exists {
		stats = &TaskTypeStats{}
		ts.taskStats[result.TaskType] = stats
	}

	stats.Total++
	if result.Success {
		stats.Completed++
	} else {
		stats.Failed++
	}

	// 更新平均执行时间
	if stats.Completed > 0 {
		stats.AvgTime = time.Duration(
			(int64(stats.AvgTime)*(stats.Completed-1) + int64(result.Duration)) / stats.Completed,
		)
	}
}

// Submit 提交任务
func (ts *TaskScheduler) Submit(task Task) error {
	select {
	case <-ts.ctx.Done():
		return fmt.Errorf("scheduler is shutting down")
	case ts.taskQueue <- task:
		atomic.AddInt64(&ts.totalTasks, 1)
		return nil
	default:
		return fmt.Errorf("task queue is full")
	}
}

// SubmitBatch 批量提交任务
func (ts *TaskScheduler) SubmitBatch(tasks []Task) error {
	for _, task := range tasks {
		if err := ts.Submit(task); err != nil {
			return fmt.Errorf("failed to submit task %s: %w", task.ID(), err)
		}
	}
	return nil
}

// GetStats 获取统计信息
func (ts *TaskScheduler) GetStats() map[string]interface{} {
	ts.statsMutex.RLock()
	defer ts.statsMutex.RUnlock()

	stats := make(map[string]interface{})
	stats["total_tasks"] = atomic.LoadInt64(&ts.totalTasks)
	stats["completed_tasks"] = atomic.LoadInt64(&ts.completedTasks)
	stats["failed_tasks"] = atomic.LoadInt64(&ts.failedTasks)
	stats["pending_tasks"] = len(ts.taskQueue)
	stats["active_workers"] = ts.workers - len(ts.workerPool)

	// 复制任务类型统计
	typeStats := make(map[string]*TaskTypeStats)
	for k, v := range ts.taskStats {
		typeStats[k] = &TaskTypeStats{
			Total:     v.Total,
			Completed: v.Completed,
			Failed:    v.Failed,
			AvgTime:   v.AvgTime,
		}
	}
	stats["type_stats"] = typeStats

	return stats
}

// Shutdown 关闭调度器
func (ts *TaskScheduler) Shutdown() {
	ts.cancel()
	close(ts.taskQueue)
	ts.wg.Wait()
	close(ts.resultQueue)
}

// --- 具体任务实现示例 ---

// ImageCompressTask 图片压缩任务
type ImageCompressTask struct {
	id       string
	filePath string
	quality  int
}

func NewImageCompressTask(id, filePath string, quality int) *ImageCompressTask {
	return &ImageCompressTask{
		id:       id,
		filePath: filePath,
		quality:  quality,
	}
}

func (t *ImageCompressTask) Execute(ctx context.Context) error {
	// 这里是实际的图片压缩逻辑
	// 可以使用 imaging 库或调用外部工具
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// 模拟压缩操作
		// 实际实现中，这里会调用图片处理库
		log.Printf("Compressing image: %s with quality %d", t.filePath, t.quality)
		time.Sleep(100 * time.Millisecond) // 模拟耗时操作
		return nil
	}
}

func (t *ImageCompressTask) ID() string    { return t.id }
func (t *ImageCompressTask) Type() string  { return "image_compress" }
func (t *ImageCompressTask) Priority() int { return 5 }

// ExifReadTask EXIF读取任务
type ExifReadTask struct {
	id       string
	filePath string
	exifData map[string]interface{}
}

func NewExifReadTask(id, filePath string) *ExifReadTask {
	return &ExifReadTask{
		id:       id,
		filePath: filePath,
		exifData: make(map[string]interface{}),
	}
}

func (t *ExifReadTask) Execute(ctx context.Context) error {
	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
		// 使用 exiftool 命令行工具
		cmd := exec.CommandContext(ctx, "exiftool", "-j", t.filePath)
		output, err := cmd.Output()
		if err != nil {
			return fmt.Errorf("exiftool failed: %w", err)
		}

		// 这里应该解析 JSON 输出到 t.exifData
		log.Printf("Read EXIF from: %s, data size: %d", t.filePath, len(output))
		return nil
	}
}

func (t *ExifReadTask) ID() string    { return t.id }
func (t *ExifReadTask) Type() string  { return "exif_read" }
func (t *ExifReadTask) Priority() int { return 3 }

// PriorityTaskScheduler 支持优先级的任务调度器
type PriorityTaskScheduler struct {
	*TaskScheduler
	priorityQueue *PriorityQueue
	queueMutex    sync.Mutex
}

// PriorityQueue 优先级队列实现
type PriorityQueue struct {
	tasks []Task
}

func (pq *PriorityQueue) Push(task Task) {
	pq.tasks = append(pq.tasks, task)
	// 简单的插入排序，实际可以使用堆
	for i := len(pq.tasks) - 1; i > 0; i-- {
		if pq.tasks[i].Priority() > pq.tasks[i-1].Priority() {
			pq.tasks[i], pq.tasks[i-1] = pq.tasks[i-1], pq.tasks[i]
		} else {
			break
		}
	}
}

func (pq *PriorityQueue) Pop() Task {
	if len(pq.tasks) == 0 {
		return nil
	}
	task := pq.tasks[0]
	pq.tasks = pq.tasks[1:]
	return task
}

func (pq *PriorityQueue) Len() int {
	return len(pq.tasks)
}

// NewPriorityTaskScheduler 创建支持优先级的调度器
func NewPriorityTaskScheduler(workers int) *PriorityTaskScheduler {
	pts := &PriorityTaskScheduler{
		TaskScheduler: NewTaskScheduler(workers, 0), // 不使用内置队列
		priorityQueue: &PriorityQueue{tasks: make([]Task, 0)},
	}

	// 启动优先级调度协程
	go pts.priorityDispatcher()

	return pts
}

// priorityDispatcher 优先级调度器
func (pts *PriorityTaskScheduler) priorityDispatcher() {
	ticker := time.NewTicker(10 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-pts.ctx.Done():
			return
		case <-ticker.C:
			pts.queueMutex.Lock()
			if pts.priorityQueue.Len() > 0 && len(pts.taskQueue) < cap(pts.taskQueue) {
				task := pts.priorityQueue.Pop()
				select {
				case pts.taskQueue <- task:
					// 成功发送
				default:
					// 队列满了，放回优先级队列
					pts.priorityQueue.Push(task)
				}
			}
			pts.queueMutex.Unlock()
		}
	}
}

// SubmitWithPriority 提交带优先级的任务
func (pts *PriorityTaskScheduler) SubmitWithPriority(task Task) error {
	select {
	case <-pts.ctx.Done():
		return fmt.Errorf("scheduler is shutting down")
	default:
		pts.queueMutex.Lock()
		pts.priorityQueue.Push(task)
		atomic.AddInt64(&pts.totalTasks, 1)
		pts.queueMutex.Unlock()
		return nil
	}
}
