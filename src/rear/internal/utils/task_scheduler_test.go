package utils

import (
	"context"
	"fmt"
	"log"
	"time"
)

// 示例：批量处理图片任务
func processImages(scheduler *TaskScheduler) {
	// 假设有10000张图片需要处理
	imagePaths := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		imagePaths[i] = fmt.Sprintf("/path/to/image_%d.jpg", i)
	}

	// 批量提交压缩任务
	log.Println("提交图片压缩任务...")
	for i, path := range imagePaths {
		task := NewImageCompressTask(
			fmt.Sprintf("compress_%d", i),
			path,
			85, // 压缩质量
		)

		if err := scheduler.Submit(task); err != nil {
			log.Printf("Failed to submit task: %v", err)
			// 可以实现重试逻辑
			time.Sleep(100 * time.Millisecond)
			scheduler.Submit(task) // 重试一次
		}
	}

	// 同时提交EXIF读取任务
	log.Println("提交EXIF读取任务...")
	for i, path := range imagePaths[:1000] { // 只处理前1000张
		task := NewExifReadTask(
			fmt.Sprintf("exif_%d", i),
			path,
		)
		scheduler.Submit(task)
	}
}

// 高级任务管理器 - 支持任务组和依赖
type TaskManager struct {
	scheduler  *PriorityTaskScheduler
	taskGroups map[string]*TaskGroup
}

type TaskGroup struct {
	ID        string
	Tasks     []Task
	Status    string
	Progress  float64
	StartTime time.Time
	EndTime   time.Time
}

func NewTaskManager(workers int) *TaskManager {
	return &TaskManager{
		scheduler:  NewPriorityTaskScheduler(workers),
		taskGroups: make(map[string]*TaskGroup),
	}
}

// CreateImageProcessingGroup 创建图片处理任务组
func (tm *TaskManager) CreateImageProcessingGroup(groupID string, imagePaths []string) error {
	group := &TaskGroup{
		ID:        groupID,
		Tasks:     make([]Task, 0),
		Status:    "pending",
		StartTime: time.Now(),
	}

	// 为每张图片创建压缩和EXIF读取任务
	for i, path := range imagePaths {
		// 压缩任务
		compressTask := NewImageCompressTask(
			fmt.Sprintf("%s_compress_%d", groupID, i),
			path,
			85,
		)
		group.Tasks = append(group.Tasks, compressTask)

		// EXIF任务
		exifTask := NewExifReadTask(
			fmt.Sprintf("%s_exif_%d", groupID, i),
			path,
		)
		group.Tasks = append(group.Tasks, exifTask)
	}

	tm.taskGroups[groupID] = group

	// 提交所有任务
	for _, task := range group.Tasks {
		if err := tm.scheduler.SubmitWithPriority(task); err != nil {
			return err
		}
	}

	group.Status = "running"
	return nil
}

// 动态Worker调整器
type DynamicWorkerManager struct {
	scheduler      *TaskScheduler
	minWorkers     int
	maxWorkers     int
	checkInterval  time.Duration
	scaleThreshold float64 // 队列使用率阈值
}

func NewDynamicWorkerManager(scheduler *TaskScheduler, min, max int) *DynamicWorkerManager {
	return &DynamicWorkerManager{
		scheduler:      scheduler,
		minWorkers:     min,
		maxWorkers:     max,
		checkInterval:  5 * time.Second,
		scaleThreshold: 0.8,
	}
}

func (dwm *DynamicWorkerManager) Start(ctx context.Context) {
	ticker := time.NewTicker(dwm.checkInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			dwm.adjustWorkers()
		}
	}
}

func (dwm *DynamicWorkerManager) adjustWorkers() {
	stats := dwm.scheduler.GetStats()
	pending := stats["pending_tasks"].(int)
	//active := stats["active_workers"].(int)

	// 计算队列使用率
	queueUsage := float64(pending) / float64(cap(dwm.scheduler.taskQueue))

	if queueUsage > dwm.scaleThreshold && dwm.scheduler.workers < dwm.maxWorkers {
		// 增加worker
		log.Printf("Scaling up workers: queue usage %.2f%%", queueUsage*100)
		// 这里需要实现动态增加worker的逻辑
	} else if queueUsage < 0.2 && dwm.scheduler.workers > dwm.minWorkers {
		// 减少worker
		log.Printf("Scaling down workers: queue usage %.2f%%", queueUsage*100)
		// 这里需要实现动态减少worker的逻辑
	}
}

// 任务监控器
type TaskMonitor struct {
	scheduler *TaskScheduler
}

func NewTaskMonitor(scheduler *TaskScheduler) *TaskMonitor {
	return &TaskMonitor{scheduler: scheduler}
}

func (tm *TaskMonitor) Start(ctx context.Context) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tm.printStats()
		}
	}
}

func (tm *TaskMonitor) printStats() {
	stats := tm.scheduler.GetStats()

	fmt.Println("\n========== Task Scheduler Stats ==========")
	fmt.Printf("Total Tasks: %d\n", stats["total_tasks"])
	fmt.Printf("Completed: %d\n", stats["completed_tasks"])
	fmt.Printf("Failed: %d\n", stats["failed_tasks"])
	fmt.Printf("Pending: %d\n", stats["pending_tasks"])
	fmt.Printf("Active Workers: %d\n", stats["active_workers"])

	if typeStats, ok := stats["type_stats"].(map[string]*TaskTypeStats); ok {
		fmt.Println("\nTask Type Statistics:")
		for taskType, stat := range typeStats {
			fmt.Printf("  %s: Total=%d, Completed=%d, Failed=%d, AvgTime=%v\n",
				taskType, stat.Total, stat.Completed, stat.Failed, stat.AvgTime)
		}
	}
	fmt.Println("==========================================\n")
}

// 主函数示例
func main() {
	// 创建调度器 - 50个worker，队列容量10000
	scheduler := NewTaskScheduler(50, 10000)
	defer scheduler.Shutdown()

	// 创建任务管理器
	taskManager := NewTaskManager(50)

	// 创建监控器
	monitor := NewTaskMonitor(scheduler)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go monitor.Start(ctx)

	// 场景1：处理第一批10000张图片
	log.Println("开始处理第一批图片...")
	imagePaths1 := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		imagePaths1[i] = fmt.Sprintf("/batch1/image_%d.jpg", i)
	}
	taskManager.CreateImageProcessingGroup("batch1", imagePaths1)

	// 模拟过一段时间后又来了新任务
	time.Sleep(5 * time.Second)

	// 场景2：又来了10000张图片
	log.Println("开始处理第二批图片...")
	imagePaths2 := make([]string, 10000)
	for i := 0; i < 10000; i++ {
		imagePaths2[i] = fmt.Sprintf("/batch2/image_%d.jpg", i)
	}
	taskManager.CreateImageProcessingGroup("batch2", imagePaths2)

	// 场景3：批量解析EXIF（优先级更高）
	log.Println("开始批量解析EXIF...")
	exifPaths := make([]string, 1000)
	for i := 0; i < 1000; i++ {
		exifPaths[i] = fmt.Sprintf("/priority/image_%d.jpg", i)
	}
	// 这些任务会有更高的优先级
	for i, path := range exifPaths {
		task := &HighPriorityExifTask{
			ExifReadTask: *NewExifReadTask(fmt.Sprintf("priority_exif_%d", i), path),
		}
		taskManager.scheduler.SubmitWithPriority(task)
	}

	// 等待一段时间查看执行情况
	time.Sleep(30 * time.Second)

	// 打印最终统计
	monitor.printStats()
}

// HighPriorityExifTask 高优先级EXIF任务
type HighPriorityExifTask struct {
	ExifReadTask
}

func (t *HighPriorityExifTask) Priority() int {
	return 10 // 更高的优先级
}

// 任务重试装饰器
type RetryableTask struct {
	task       Task
	maxRetries int
	retries    int
}

func NewRetryableTask(task Task, maxRetries int) *RetryableTask {
	return &RetryableTask{
		task:       task,
		maxRetries: maxRetries,
		retries:    0,
	}
}

func (rt *RetryableTask) Execute(ctx context.Context) error {
	var lastErr error

	for rt.retries <= rt.maxRetries {
		err := rt.task.Execute(ctx)
		if err == nil {
			return nil
		}

		lastErr = err
		rt.retries++

		if rt.retries <= rt.maxRetries {
			// 指数退避
			backoff := time.Duration(rt.retries) * time.Second
			log.Printf("Task %s failed, retrying in %v (attempt %d/%d)",
				rt.task.ID(), backoff, rt.retries, rt.maxRetries)

			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(backoff):
				// 继续重试
			}
		}
	}

	return fmt.Errorf("task failed after %d retries: %w", rt.maxRetries, lastErr)
}

func (rt *RetryableTask) ID() string    { return rt.task.ID() }
func (rt *RetryableTask) Type() string  { return rt.task.Type() }
func (rt *RetryableTask) Priority() int { return rt.task.Priority() }

// 批处理优化 - 将多个小任务合并为一个大任务
type BatchImageCompressTask struct {
	id        string
	filePaths []string
	quality   int
}

func NewBatchImageCompressTask(id string, filePaths []string, quality int) *BatchImageCompressTask {
	return &BatchImageCompressTask{
		id:        id,
		filePaths: filePaths,
		quality:   quality,
	}
}

func (t *BatchImageCompressTask) Execute(ctx context.Context) error {
	// 批量处理多张图片，减少任务切换开销
	for _, path := range t.filePaths {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			// 处理单张图片
			log.Printf("Batch compressing: %s", path)
			// 实际的压缩逻辑
		}
	}
	return nil
}

func (t *BatchImageCompressTask) ID() string    { return t.id }
func (t *BatchImageCompressTask) Type() string  { return "batch_image_compress" }
func (t *BatchImageCompressTask) Priority() int { return 5 }
