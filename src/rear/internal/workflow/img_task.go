package workflow

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"rear/internal/api"
	"rear/internal/config"
	"rear/pkg/logger"
	"rear/pkg/utils"
	"runtime"
	"sync"
	"time"
)

// --- 状态定义 ---
type TaskStatus string

const (
	StatusPending TaskStatus = "pending" // 等待处理
	StatusRunning TaskStatus = "running" // 正在处理
	StatusPaused  TaskStatus = "paused"  // 已暂停
	StatusFailed  TaskStatus = "failed"  // 处理失败
	StatusDone    TaskStatus = "done"    // 处理完成
)

// TaskStep 任务处理步骤
type TaskStep string

const (
	StepInitializing         TaskStep = "initializing"          // 初始化
	StepValidating           TaskStep = "validating"            // 验证文件格式
	StepReadingFile          TaskStep = "reading_file"          // 读取文件
	StepCalculatingHash      TaskStep = "calculating_hash"      // 计算哈希值
	StepExtractingExif       TaskStep = "extracting_exif"       // 提取EXIF信息
	StepConvertingFormat     TaskStep = "converting_format"     // 转换格式
	StepGeneratingThumbnails TaskStep = "generating_thumbnails" // 生成缩略图
	StepSavingToDatabase     TaskStep = "saving_to_database"    // 保存到数据库
	StepIntelligentAnalysis  TaskStep = "intelligent_analysis"  // 智能分析
	StepCompleted            TaskStep = "completed"             // 完成
)

// TaskStepInfo 步骤信息
type TaskStepInfo struct {
	Step        TaskStep `json:"step"`            // 当前步骤
	StepName    string   `json:"step_name"`       // 步骤名称
	Description string   `json:"description"`     // 步骤描述
	Progress    float64  `json:"progress"`        // 步骤内进度 (0.0-1.0)
	Started     bool     `json:"started"`         // 是否已开始
	Completed   bool     `json:"completed"`       // 是否已完成
	Error       string   `json:"error,omitempty"` // 错误信息
}

// GetStepInfo 获取步骤信息
func GetStepInfo(step TaskStep) TaskStepInfo {
	stepInfoMap := map[TaskStep]TaskStepInfo{
		StepInitializing: {
			Step:        StepInitializing,
			StepName:    "初始化",
			Description: "正在初始化任务...",
		},
		StepValidating: {
			Step:        StepValidating,
			StepName:    "验证格式",
			Description: "正在验证图像格式...",
		},
		StepReadingFile: {
			Step:        StepReadingFile,
			StepName:    "读取文件",
			Description: "正在读取图像文件...",
		},
		StepCalculatingHash: {
			Step:        StepCalculatingHash,
			StepName:    "计算哈希",
			Description: "正在计算文件哈希值...",
		},
		StepExtractingExif: {
			Step:        StepExtractingExif,
			StepName:    "提取EXIF",
			Description: "正在提取图像元数据...",
		},
		StepConvertingFormat: {
			Step:        StepConvertingFormat,
			StepName:    "格式转换",
			Description: "正在转换图像格式...",
		},
		StepGeneratingThumbnails: {
			Step:        StepGeneratingThumbnails,
			StepName:    "生成缩略图",
			Description: "正在生成缩略图...",
		},
		StepSavingToDatabase: {
			Step:        StepSavingToDatabase,
			StepName:    "保存数据",
			Description: "正在保存到数据库...",
		},
		StepIntelligentAnalysis: {
			Step:        StepIntelligentAnalysis,
			StepName:    "智能分析",
			Description: "正在进行智能分析...",
		},
		StepCompleted: {
			Step:        StepCompleted,
			StepName:    "已完成",
			Description: "任务已完成",
			Completed:   true,
		},
	}

	info, exists := stepInfoMap[step]
	if !exists {
		return TaskStepInfo{
			Step:        step,
			StepName:    string(step),
			Description: "未知步骤",
		}
	}
	return info
}

// --- PictureTask ---
type PictureTask struct {
	ID       string
	Path     string
	Hash     string
	ImageBuf []byte
	ExifData map[string]string

	Status        TaskStatus     `json:"status"`                   // 任务状态
	Progress      float64        `json:"progress"`                 // 总体进度 (0.0-1.0)
	CurrentStep   TaskStep       `json:"current_step"`             // 当前步骤
	StepInfo      TaskStepInfo   `json:"step_info"`                // 当前步骤详细信息
	AllSteps      []TaskStepInfo `json:"all_steps"`                // 所有步骤信息
	Error         error          `json:"error,omitempty"`          // 错误信息
	StartTime     time.Time      `json:"start_time"`               // 开始时间
	UpdateTime    time.Time      `json:"update_time"`              // 最后更新时间
	CompletedTime *time.Time     `json:"completed_time,omitempty"` // 完成时间

	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	pauseCh  chan struct{}
	resumeCh chan struct{}
}

func NewPictureTask(path string) *PictureTask {
	ctx, cancel := context.WithCancel(context.Background())

	// 初始化所有步骤信息
	allSteps := []TaskStepInfo{
		GetStepInfo(StepInitializing),
		GetStepInfo(StepValidating),
		GetStepInfo(StepReadingFile),
		GetStepInfo(StepCalculatingHash),
		GetStepInfo(StepExtractingExif),
		GetStepInfo(StepConvertingFormat),
		GetStepInfo(StepGeneratingThumbnails),
		GetStepInfo(StepSavingToDatabase),
		GetStepInfo(StepIntelligentAnalysis),
		GetStepInfo(StepCompleted),
	}

	return &PictureTask{
		ID:          uuid.New().String(),
		Path:        path,
		Status:      StatusPending,
		Progress:    0.0,
		CurrentStep: StepInitializing,
		StepInfo:    GetStepInfo(StepInitializing),
		AllSteps:    allSteps,
		StartTime:   time.Now(),
		UpdateTime:  time.Now(),
		ctx:         ctx,
		cancel:      cancel,
		pauseCh:     make(chan struct{}, 1),
		resumeCh:    make(chan struct{}, 1),
	}
}

func (pt *PictureTask) waitIfPaused() {
	select {
	case <-pt.pauseCh:
		pt.setStatus(StatusPaused)
		<-pt.resumeCh
		pt.setStatus(StatusRunning)
	default:
	}
}

func (pt *PictureTask) Pause() {
	select {
	case pt.pauseCh <- struct{}{}:
	default:
	}
}

func (pt *PictureTask) Resume() {
	select {
	case pt.resumeCh <- struct{}{}:
	default:
	}
}

// updateStep 更新任务步骤和进度
func (pt *PictureTask) updateStep(step TaskStep, stepProgress float64, overallProgress float64) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.CurrentStep = step
	pt.StepInfo = GetStepInfo(step)
	pt.StepInfo.Started = true
	pt.StepInfo.Progress = stepProgress
	pt.Progress = overallProgress
	pt.UpdateTime = time.Now()

	// 更新所有步骤状态
	for i, stepInfo := range pt.AllSteps {
		if stepInfo.Step == step {
			pt.AllSteps[i] = pt.StepInfo
			break
		}
	}

	logger.Debug("任务步骤更新",
		zap.String("task_id", pt.ID),
		zap.String("step_name", pt.StepInfo.StepName),
		zap.Float64("overall_progress", overallProgress))
}

// completeStep 完成当前步骤
func (pt *PictureTask) completeStep() {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.StepInfo.Completed = true
	pt.StepInfo.Progress = 1.0

	// 更新所有步骤状态
	for i, stepInfo := range pt.AllSteps {
		if stepInfo.Step == pt.CurrentStep {
			pt.AllSteps[i] = pt.StepInfo
			break
		}
	}
}

// setStepError 设置步骤错误
func (pt *PictureTask) setStepError(err error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()

	pt.StepInfo.Error = err.Error()
	pt.Error = err
	pt.Status = StatusFailed

	// 更新所有步骤状态
	for i, stepInfo := range pt.AllSteps {
		if stepInfo.Step == pt.CurrentStep {
			pt.AllSteps[i] = pt.StepInfo
			break
		}
	}
}

func (pt *PictureTask) Run() {
	pt.setStatus(StatusRunning)

	// 步骤1: 初始化ImageAPI
	pt.updateStep(StepInitializing, 0.0, 0.05)
	imageAPI, err := api.NewImageAPI(pt.Path)
	if err != nil {
		pt.setStepError(fmt.Errorf("初始化ImageAPI失败: %w", err))
		return
	}
	pt.completeStep()

	// 步骤2: 验证图像格式
	pt.updateStep(StepValidating, 0.0, 0.10)
	if !imageAPI.IsSupported() {
		pt.setStepError(fmt.Errorf("不支持的图像格式: %s", imageAPI.GetFormat()))
		return
	}
	pt.completeStep()

	// 步骤3: 读取文件和计算哈希
	pt.updateStep(StepReadingFile, 0.0, 0.15)
	pt.waitIfPaused()
	pt.updateStep(StepCalculatingHash, 0.0, 0.20)
	pt.Hash = imageAPI.GetHash()
	pt.completeStep()

	// 步骤4-6: 图像处理（EXIF提取、格式转换、缩略图生成）
	pt.updateStep(StepExtractingExif, 0.0, 0.30)
	pt.waitIfPaused()

	// 使用 img_api.go 的统一处理方法
	if err := imageAPI.ProcessImage(); err != nil {
		pt.setStepError(fmt.Errorf("图像处理失败: %w", err))
		return
	}

	// 更新进度到图像处理完成
	pt.updateStep(StepGeneratingThumbnails, 1.0, 0.85)
	pt.completeStep()

	// 步骤7: 保存到数据库
	pt.updateStep(StepSavingToDatabase, 0.0, 0.90)
	pt.waitIfPaused()
	if err := pt.saveToDatabase(imageAPI); err != nil {
		pt.setStepError(fmt.Errorf("保存到数据库失败: %w", err))
		return
	}
	pt.completeStep()

	// 步骤8: 智能分析（异步进行）
	pt.updateStep(StepIntelligentAnalysis, 0.0, 0.95)
	go pt.performIntelligentAnalysis(imageAPI)
	pt.completeStep()

	// 完成
	pt.updateStep(StepCompleted, 1.0, 1.0)
	pt.setDone()
}

func (pt *PictureTask) setError(err error) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.Status = StatusFailed
	pt.Error = err
}

func (pt *PictureTask) setDone() {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.Status = StatusDone
	pt.Progress = 1.0
	now := time.Now()
	pt.CompletedTime = &now
	pt.UpdateTime = now
}

func (pt *PictureTask) setStatus(s TaskStatus) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.Status = s
	pt.UpdateTime = time.Now()
}

// saveToDatabase 保存图像信息到数据库
func (pt *PictureTask) saveToDatabase(imageAPI *api.ImageAPI) error {
	// TODO: 实现数据库保存逻辑
	// 这里应该保存到 photos 表和 photo_exif 表

	exifData := imageAPI.GetExifData()
	hash := imageAPI.GetHash()

	logger.Info("保存图像信息到数据库",
		zap.String("hash", hash),
		zap.String("path", pt.Path))

	// 模拟数据库操作
	if exifData != nil {
		logger.Debug("EXIF信息",
			zap.String("make", exifData.Exif.Make),
			zap.String("model", exifData.Exif.Model),
			zap.Int("width", exifData.BaseInfo.ImageWidth),
			zap.Int("height", exifData.BaseInfo.ImageHeight))
	}

	return nil
}

// performIntelligentAnalysis 执行智能分析
func (pt *PictureTask) performIntelligentAnalysis(imageAPI *api.ImageAPI) {
	// TODO: 实现异步调用Python API进行智能分析
	// 包括：人脸识别、场景识别、物体识别、OCR文字识别

	//logger.Info("开始智能分析", zap.String("path", pt.Path))

	// 这里可以调用 Python API 或其他 AI 服务
	// 分析结果应该更新到数据库的 has_face, has_object, has_scene, has_text 字段

	//logger.Debug("智能分析完成", zap.String("path", pt.Path))
}

// --- ImgTaskManager ---
type ImgTaskManager struct {
	tasks        map[string]*PictureTask
	queue        chan *PictureTask
	workerPool   chan struct{}
	workerLimit  int
	mu           sync.RWMutex
	globalPause  chan struct{}
	globalResume chan struct{}
	poolMu       sync.Mutex
	doneCount    int
	autoAdjust   bool
}

func NewImgTaskManager(concurrency int) *ImgTaskManager {
	taskConfig := config.GetTaskConfig()

	if concurrency <= 0 {
		concurrency = taskConfig.Concurrency
	}

	tm := &ImgTaskManager{
		tasks:        make(map[string]*PictureTask),
		queue:        make(chan *PictureTask, taskConfig.QueueCapacity),
		workerPool:   make(chan struct{}, concurrency),
		workerLimit:  concurrency,
		globalPause:  make(chan struct{}, 1),
		globalResume: make(chan struct{}, 1),
		autoAdjust:   taskConfig.AutoAdjust,
	}

	// 启动工作协程和监控协程
	go tm.run()
	go tm.monitorSystem()

	logger.Info("任务管理器已启动",
		zap.Int("concurrency", concurrency),
		zap.Int("queue_capacity", cap(tm.queue)),
		zap.Bool("auto_adjust", tm.autoAdjust))

	return tm
}

// 全局任务管理器实例
var (
	globalTaskManager *ImgTaskManager
	taskManagerOnce   sync.Once
)

// InitGlobalTaskManager 初始化全局任务管理器
func InitGlobalTaskManager() {
	taskManagerOnce.Do(func() {
		taskConfig := config.GetTaskConfig()

		if !taskConfig.AutoStart {
			logger.Info("任务管理器自动启动已禁用")
			return
		}

		globalTaskManager = NewImgTaskManager(taskConfig.Concurrency)
		logger.Info("全局任务管理器已初始化",
			zap.Int("concurrency", taskConfig.Concurrency),
			zap.Bool("auto_adjust", taskConfig.AutoAdjust))
	})
}

// GetGlobalTaskManager 获取全局任务管理器实例
func GetGlobalTaskManager() *ImgTaskManager {
	if globalTaskManager == nil {
		logger.Warn("全局任务管理器未初始化，正在初始化...")
		InitGlobalTaskManager()
	}
	return globalTaskManager
}

// IsTaskManagerInitialized 检查任务管理器是否已初始化
func IsTaskManagerInitialized() bool {
	return globalTaskManager != nil
}

// ShutdownTaskManager 关闭任务管理器
func ShutdownTaskManager() {
	if globalTaskManager != nil {
		logger.Info("正在关闭任务管理器...")
		// 暂停所有任务
		globalTaskManager.PauseAll()

		// 等待当前任务完成（最多等待30秒）
		timeout := time.After(30 * time.Second)
		ticker := time.NewTicker(1 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-timeout:
				logger.Warn("任务管理器关闭超时，强制退出")
				return
			case <-ticker.C:
				remaining := globalTaskManager.RemainingCount()
				if remaining == 0 {
					logger.Info("所有任务已完成，任务管理器已关闭")
					return
				}
				logger.Info("等待任务完成", zap.Int("remaining", remaining))
			}
		}
	}
}

func (tm *ImgTaskManager) SetConcurrency(n int) {
	tm.poolMu.Lock()
	defer tm.poolMu.Unlock()

	if n <= 0 || n == tm.workerLimit {
		return
	}

	newPool := make(chan struct{}, n)
	tm.mu.Lock()
	used := len(tm.workerPool)
	for i := 0; i < used && i < n; i++ {
		newPool <- struct{}{}
	}
	tm.workerPool = newPool
	tm.workerLimit = n
	tm.mu.Unlock()
}

func (tm *ImgTaskManager) AddTask(path string) string {
	task := NewPictureTask(path)
	tm.mu.Lock()
	tm.tasks[task.ID] = task
	tm.mu.Unlock()
	tm.queue <- task
	return task.ID
}

func (tm *ImgTaskManager) PauseAll() {
	select {
	case tm.globalPause <- struct{}{}:
	default:
	}
}

func (tm *ImgTaskManager) ResumeAll() {
	select {
	case tm.globalResume <- struct{}{}:
	default:
	}
}

func (tm *ImgTaskManager) applyGlobalPauseResume(task *PictureTask) {
	go func() {
		for {
			select {
			case <-tm.globalPause:
				task.Pause()
			case <-tm.globalResume:
				task.Resume()
			}
		}
	}()
}

func (tm *ImgTaskManager) run() {
	for task := range tm.queue {
		tm.workerPool <- struct{}{}
		tm.applyGlobalPauseResume(task)
		go func(t *PictureTask) {
			defer func() { <-tm.workerPool }()
			t.Run()
			tm.mu.Lock()
			if t.Status == StatusDone {
				tm.doneCount++
			}
			tm.mu.Unlock()
		}(task)
	}
}

func (tm *ImgTaskManager) GetStatus(id string) TaskStatus {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	if task, ok := tm.tasks[id]; ok {
		return task.Status
	}
	return "not_found"
}
func (tm *ImgTaskManager) DoneCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return tm.doneCount
}

func (tm *ImgTaskManager) RemainingCount() int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()
	return len(tm.tasks) - tm.doneCount
}

func (tm *ImgTaskManager) monitorSystem() {
	taskConfig := config.GetTaskConfig()
	interval := time.Duration(taskConfig.MonitorInterval) * time.Second

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		if !tm.autoAdjust {
			continue
		}

		// 获取系统状态
		memPercent := getCurrentMemoryUsage()
		cpuPercent := getCurrentCPUUsage()
		goroutineCount := runtime.NumGoroutine()

		tm.mu.RLock()
		queueLen := len(tm.queue)
		activeWorkers := len(tm.workerPool)
		totalTasks := len(tm.tasks)
		tm.mu.RUnlock()

		logger.Debug("系统监控",
			zap.Int("cpu_percent", cpuPercent),
			zap.Int("memory_percent", memPercent),
			zap.Int("goroutines", goroutineCount),
			zap.Int("queue_length", queueLen),
			zap.Int("active_workers", activeWorkers),
			zap.Int("total_tasks", totalTasks),
			zap.Int("worker_limit", tm.workerLimit))

		// 根据系统负载调整并发数
		newConcurrency := tm.calculateOptimalConcurrency(cpuPercent, memPercent, queueLen)
		if newConcurrency != tm.workerLimit {
			logger.Info("调整并发数",
				zap.Int("old", tm.workerLimit),
				zap.Int("new", newConcurrency),
				zap.String("reason", tm.getAdjustmentReason(cpuPercent, memPercent, queueLen)))
			tm.SetConcurrency(newConcurrency)
		}
	}
}

// calculateOptimalConcurrency 计算最优并发数
func (tm *ImgTaskManager) calculateOptimalConcurrency(cpuPercent, memPercent, queueLen int) int {
	taskConfig := config.GetTaskConfig()
	maxConcurrency := taskConfig.MaxConcurrency
	minConcurrency := taskConfig.MinConcurrency

	current := tm.workerLimit

	// 如果CPU使用率过高，减少并发
	if cpuPercent > 85 {
		return max(minConcurrency, current/2)
	}

	// 如果内存使用率过高，减少并发
	if memPercent > 90 {
		return max(minConcurrency, current-1)
	}

	// 如果队列积压严重且系统负载不高，增加并发
	if queueLen > 50 && cpuPercent < 60 && memPercent < 70 {
		return min(maxConcurrency, current+1)
	}

	// 如果系统负载很低且有任务队列，适度增加并发
	if queueLen > 0 && cpuPercent < 40 && memPercent < 50 {
		return min(maxConcurrency, current+1)
	}

	// 如果没有任务且并发数较高，适度减少
	if queueLen == 0 && current > minConcurrency {
		return max(minConcurrency, current-1)
	}

	return current
}

// getAdjustmentReason 获取调整原因
func (tm *ImgTaskManager) getAdjustmentReason(cpuPercent, memPercent, queueLen int) string {
	if cpuPercent > 85 {
		return "CPU使用率过高"
	}
	if memPercent > 90 {
		return "内存使用率过高"
	}
	if queueLen > 50 {
		return "任务队列积压"
	}
	if cpuPercent < 40 && memPercent < 50 && queueLen > 0 {
		return "系统负载低，有待处理任务"
	}
	if queueLen == 0 {
		return "无待处理任务"
	}
	return "系统优化"
}

// getCurrentMemoryUsage 获取当前内存使用率
func getCurrentMemoryUsage() int {
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// 简单估算内存使用率（实际项目中可使用 gopsutil 等库）
	usedMB := bToMb(m.Alloc)
	totalMB := bToMb(m.Sys)

	if totalMB == 0 {
		return 0
	}

	return int((usedMB * 100) / totalMB)
}

// bToMb 字节转MB
func bToMb(b uint64) uint64 {
	return b / 1024 / 1024
}

// getCurrentCPUUsage 获取当前CPU使用率
func getCurrentCPUUsage() int {
	// 简单估算CPU使用率（基于goroutine数量）
	// 实际项目中应该使用 gopsutil 等库获取真实CPU使用率
	goroutineCount := runtime.NumGoroutine()
	cpuCount := runtime.NumCPU()

	// 粗略估算：每个CPU核心处理50个goroutine为满负载
	usage := (goroutineCount * 100) / (cpuCount * 50)
	if usage > 100 {
		usage = 100
	}
	return usage
}

// 辅助函数
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// --- 工具函数 ---

// AddBatchTasks 批量添加任务
func (tm *ImgTaskManager) AddBatchTasks(paths []string) []string {
	var taskIDs []string
	for _, path := range paths {
		if utils.FileUtils.Exists(path) {
			taskID := tm.AddTask(path)
			taskIDs = append(taskIDs, taskID)
		} else {
			logger.Warn("文件不存在，跳过", zap.String("path", path))
		}
	}

	logger.Info("批量添加任务完成",
		zap.Int("total_paths", len(paths)),
		zap.Int("valid_tasks", len(taskIDs)))

	return taskIDs
}

// GetTaskProgress 获取任务进度
func (tm *ImgTaskManager) GetTaskProgress(id string) (TaskStatus, float64, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, ok := tm.tasks[id]
	if !ok {
		return "", 0, fmt.Errorf("任务不存在: %s", id)
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	return task.Status, task.Progress, task.Error
}

// TaskDetail 任务详细信息
type TaskDetail struct {
	ID            string         `json:"id"`                       // 任务ID
	Path          string         `json:"path"`                     // 文件路径
	Hash          string         `json:"hash,omitempty"`           // 文件哈希
	Status        TaskStatus     `json:"status"`                   // 任务状态
	Progress      float64        `json:"progress"`                 // 总体进度
	CurrentStep   TaskStep       `json:"current_step"`             // 当前步骤
	StepInfo      TaskStepInfo   `json:"step_info"`                // 当前步骤信息
	AllSteps      []TaskStepInfo `json:"all_steps"`                // 所有步骤信息
	Error         string         `json:"error,omitempty"`          // 错误信息
	StartTime     time.Time      `json:"start_time"`               // 开始时间
	UpdateTime    time.Time      `json:"update_time"`              // 最后更新时间
	CompletedTime *time.Time     `json:"completed_time,omitempty"` // 完成时间
	Duration      string         `json:"duration"`                 // 持续时间
}

// GetTaskDetail 获取任务详细信息
func (tm *ImgTaskManager) GetTaskDetail(id string) (*TaskDetail, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	task, ok := tm.tasks[id]
	if !ok {
		return nil, fmt.Errorf("任务不存在: %s", id)
	}

	task.mu.Lock()
	defer task.mu.Unlock()

	// 计算持续时间
	var duration time.Duration
	if task.CompletedTime != nil {
		duration = task.CompletedTime.Sub(task.StartTime)
	} else {
		duration = time.Since(task.StartTime)
	}

	var errorStr string
	if task.Error != nil {
		errorStr = task.Error.Error()
	}

	return &TaskDetail{
		ID:            task.ID,
		Path:          task.Path,
		Hash:          task.Hash,
		Status:        task.Status,
		Progress:      task.Progress,
		CurrentStep:   task.CurrentStep,
		StepInfo:      task.StepInfo,
		AllSteps:      task.AllSteps,
		Error:         errorStr,
		StartTime:     task.StartTime,
		UpdateTime:    task.UpdateTime,
		CompletedTime: task.CompletedTime,
		Duration:      duration.String(),
	}, nil
}

// OverallProgress 整体进度信息
type OverallProgress struct {
	TotalTasks       int                `json:"total_tasks"`        // 总任务数
	CompletedTasks   int                `json:"completed_tasks"`    // 已完成任务数
	FailedTasks      int                `json:"failed_tasks"`       // 失败任务数
	RunningTasks     int                `json:"running_tasks"`      // 正在运行任务数
	PendingTasks     int                `json:"pending_tasks"`      // 等待中任务数
	PausedTasks      int                `json:"paused_tasks"`       // 已暂停任务数
	OverallProgress  float64            `json:"overall_progress"`   // 整体进度 (0.0-1.0)
	QueueLength      int                `json:"queue_length"`       // 队列长度
	ActiveWorkers    int                `json:"active_workers"`     // 活跃工作线程数
	WorkerLimit      int                `json:"worker_limit"`       // 工作线程限制
	AutoAdjust       bool               `json:"auto_adjust"`        // 是否自动调整
	RunningTasksList []TaskDetail       `json:"running_tasks_list"` // 正在运行的任务列表
	StatusCounts     map[TaskStatus]int `json:"status_counts"`      // 各状态任务数量
}

// GetOverallProgress 获取整体进度
func (tm *ImgTaskManager) GetOverallProgress() *OverallProgress {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	statusCounts := make(map[TaskStatus]int)
	var totalProgress float64
	var runningTasks []TaskDetail

	for _, task := range tm.tasks {
		task.mu.Lock()
		status := task.Status
		statusCounts[status]++
		totalProgress += task.Progress

		// 如果是正在运行的任务，添加到列表中
		if status == StatusRunning {
			var duration time.Duration
			if task.CompletedTime != nil {
				duration = task.CompletedTime.Sub(task.StartTime)
			} else {
				duration = time.Since(task.StartTime)
			}

			var errorStr string
			if task.Error != nil {
				errorStr = task.Error.Error()
			}

			runningTasks = append(runningTasks, TaskDetail{
				ID:            task.ID,
				Path:          task.Path,
				Hash:          task.Hash,
				Status:        task.Status,
				Progress:      task.Progress,
				CurrentStep:   task.CurrentStep,
				StepInfo:      task.StepInfo,
				AllSteps:      task.AllSteps,
				Error:         errorStr,
				StartTime:     task.StartTime,
				UpdateTime:    task.UpdateTime,
				CompletedTime: task.CompletedTime,
				Duration:      duration.String(),
			})
		}
		task.mu.Unlock()
	}

	totalTasks := len(tm.tasks)
	var overallProgress float64
	if totalTasks > 0 {
		overallProgress = totalProgress / float64(totalTasks)
	}

	return &OverallProgress{
		TotalTasks:       totalTasks,
		CompletedTasks:   statusCounts[StatusDone],
		FailedTasks:      statusCounts[StatusFailed],
		RunningTasks:     statusCounts[StatusRunning],
		PendingTasks:     statusCounts[StatusPending],
		PausedTasks:      statusCounts[StatusPaused],
		OverallProgress:  overallProgress,
		QueueLength:      len(tm.queue),
		ActiveWorkers:    len(tm.workerPool),
		WorkerLimit:      tm.workerLimit,
		AutoAdjust:       tm.autoAdjust,
		RunningTasksList: runningTasks,
		StatusCounts:     statusCounts,
	}
}

// GetActiveTasksDetails 获取所有活跃任务的详细信息
func (tm *ImgTaskManager) GetActiveTasksDetails() []*TaskDetail {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var activeTasks []*TaskDetail

	for _, task := range tm.tasks {
		task.mu.Lock()
		// 只返回未完成的任务
		if task.Status != StatusDone && task.Status != StatusFailed {
			var duration time.Duration
			if task.CompletedTime != nil {
				duration = task.CompletedTime.Sub(task.StartTime)
			} else {
				duration = time.Since(task.StartTime)
			}

			var errorStr string
			if task.Error != nil {
				errorStr = task.Error.Error()
			}

			activeTasks = append(activeTasks, &TaskDetail{
				ID:            task.ID,
				Path:          task.Path,
				Hash:          task.Hash,
				Status:        task.Status,
				Progress:      task.Progress,
				CurrentStep:   task.CurrentStep,
				StepInfo:      task.StepInfo,
				AllSteps:      task.AllSteps,
				Error:         errorStr,
				StartTime:     task.StartTime,
				UpdateTime:    task.UpdateTime,
				CompletedTime: task.CompletedTime,
				Duration:      duration.String(),
			})
		}
		task.mu.Unlock()
	}

	return activeTasks
}

// GetRecentCompletedTasks 获取最近完成的任务（限制数量）
func (tm *ImgTaskManager) GetRecentCompletedTasks(limit int) []*TaskDetail {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	var completedTasks []*TaskDetail

	for _, task := range tm.tasks {
		task.mu.Lock()
		if task.Status == StatusDone || task.Status == StatusFailed {
			var duration time.Duration
			if task.CompletedTime != nil {
				duration = task.CompletedTime.Sub(task.StartTime)
			} else {
				duration = time.Since(task.StartTime)
			}

			var errorStr string
			if task.Error != nil {
				errorStr = task.Error.Error()
			}

			completedTasks = append(completedTasks, &TaskDetail{
				ID:            task.ID,
				Path:          task.Path,
				Hash:          task.Hash,
				Status:        task.Status,
				Progress:      task.Progress,
				CurrentStep:   task.CurrentStep,
				StepInfo:      task.StepInfo,
				AllSteps:      task.AllSteps,
				Error:         errorStr,
				StartTime:     task.StartTime,
				UpdateTime:    task.UpdateTime,
				CompletedTime: task.CompletedTime,
				Duration:      duration.String(),
			})
		}
		task.mu.Unlock()
	}

	// 按完成时间倒序排序，返回最近的
	if len(completedTasks) > limit && limit > 0 {
		// 简单截取，实际应该按时间排序
		completedTasks = completedTasks[:limit]
	}

	return completedTasks
}

// GetAllTasksStatus 获取所有任务状态概览
func (tm *ImgTaskManager) GetAllTasksStatus() map[TaskStatus]int {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	statusCount := make(map[TaskStatus]int)

	for _, task := range tm.tasks {
		task.mu.Lock()
		statusCount[task.Status]++
		task.mu.Unlock()
	}

	return statusCount
}

// ClearCompletedTasks 清理已完成的任务
func (tm *ImgTaskManager) ClearCompletedTasks() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cleared := 0
	for id, task := range tm.tasks {
		if task.Status == StatusDone || task.Status == StatusFailed {
			delete(tm.tasks, id)
			cleared++
		}
	}

	logger.Info("清理已完成任务", zap.Int("cleared_count", cleared))
	return cleared
}

// SetAutoAdjust 设置是否启用自动调整
func (tm *ImgTaskManager) SetAutoAdjust(enable bool) {
	tm.autoAdjust = enable
	logger.Info("自动调整设置", zap.Bool("enabled", enable))
}

// GetStats 获取任务管理器统计信息
func (tm *ImgTaskManager) GetStats() map[string]interface{} {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	stats := map[string]interface{}{
		"total_tasks":     len(tm.tasks),
		"done_count":      tm.doneCount,
		"remaining_count": len(tm.tasks) - tm.doneCount,
		"queue_length":    len(tm.queue),
		"queue_capacity":  cap(tm.queue),
		"worker_limit":    tm.workerLimit,
		"active_workers":  len(tm.workerPool),
		"auto_adjust":     tm.autoAdjust,
		"goroutines":      runtime.NumGoroutine(),
		"cpu_cores":       runtime.NumCPU(),
	}

	return stats
}

// --- 全局便捷函数 ---

// AddImageTask 添加图像处理任务 (全局便捷函数)
func AddImageTask(imagePath string) string {
	taskManager := GetGlobalTaskManager()
	if taskManager == nil {
		logger.Error("任务管理器未初始化")
		return ""
	}
	return taskManager.AddTask(imagePath)
}

// AddImageTasks 批量添加图像处理任务 (全局便捷函数)
func AddImageTasks(imagePaths []string) []string {
	taskManager := GetGlobalTaskManager()
	if taskManager == nil {
		logger.Error("任务管理器未初始化")
		return nil
	}
	return taskManager.AddBatchTasks(imagePaths)
}

// GetImageTaskDetail 获取图像任务详情 (全局便捷函数)
func GetImageTaskDetail(taskID string) (*TaskDetail, error) {
	taskManager := GetGlobalTaskManager()
	if taskManager == nil {
		return nil, fmt.Errorf("任务管理器未初始化")
	}
	return taskManager.GetTaskDetail(taskID)
}

// GetImageTasksProgress 获取整体任务进度 (全局便捷函数)
func GetImageTasksProgress() *OverallProgress {
	taskManager := GetGlobalTaskManager()
	if taskManager == nil {
		return &OverallProgress{
			TotalTasks:      0,
			OverallProgress: 0,
			StatusCounts:    make(map[TaskStatus]int),
		}
	}
	return taskManager.GetOverallProgress()
}

// GetActiveImageTasks 获取活跃的图像任务 (全局便捷函数)
func GetActiveImageTasks() []*TaskDetail {
	taskManager := GetGlobalTaskManager()
	if taskManager == nil {
		return nil
	}
	return taskManager.GetActiveTasksDetails()
}

// PauseAllImageTasks 暂停所有图像任务 (全局便捷函数)
func PauseAllImageTasks() {
	taskManager := GetGlobalTaskManager()
	if taskManager != nil {
		taskManager.PauseAll()
		logger.Info("所有图像任务已暂停")
	}
}

// ResumeAllImageTasks 恢复所有图像任务 (全局便捷函数)
func ResumeAllImageTasks() {
	taskManager := GetGlobalTaskManager()
	if taskManager != nil {
		taskManager.ResumeAll()
		logger.Info("所有图像任务已恢复")
	}
}

// ClearCompletedImageTasks 清理已完成的图像任务 (全局便捷函数)
func ClearCompletedImageTasks() int {
	taskManager := GetGlobalTaskManager()
	if taskManager == nil {
		return 0
	}
	return taskManager.ClearCompletedTasks()
}

// GetImageTaskStats 获取任务统计信息 (全局便捷函数)
func GetImageTaskStats() map[string]interface{} {
	taskManager := GetGlobalTaskManager()
	if taskManager == nil {
		return map[string]interface{}{
			"initialized": false,
			"error":       "任务管理器未初始化",
		}
	}

	stats := taskManager.GetStats()
	stats["initialized"] = true
	return stats
}
