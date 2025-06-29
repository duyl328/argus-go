package workflow

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"fmt"
	"github.com/google/uuid"
	"github.com/h2non/filetype"
	"go.uber.org/zap"
	"os"
	"os/exec"
	"rear/internal/consts"
	"rear/internal/model"
	"rear/internal/utils/tools"
	"rear/pkg/logger"
	"rear/pkg/utils"
	"runtime"
	"sync"
	"time"
)

// --- 状态定义 ---
type TaskStatus string

const (
	StatusPending TaskStatus = "pending"
	StatusRunning TaskStatus = "running"
	StatusPaused  TaskStatus = "paused"
	StatusFailed  TaskStatus = "failed"
	StatusDone    TaskStatus = "done"
)

// --- PictureTask ---
type PictureTask struct {
	ID       string
	Path     string
	Hash     string
	ImageBuf []byte
	ExifData map[string]string

	Status   TaskStatus
	Progress float64
	Error    error

	ctx      context.Context
	cancel   context.CancelFunc
	mu       sync.Mutex
	pauseCh  chan struct{}
	resumeCh chan struct{}
}

func NewPictureTask(path string) *PictureTask {
	ctx, cancel := context.WithCancel(context.Background())
	return &PictureTask{
		ID:       uuid.New().String(),
		Path:     path,
		Status:   StatusPending,
		ctx:      ctx,
		cancel:   cancel,
		pauseCh:  make(chan struct{}, 1),
		resumeCh: make(chan struct{}, 1),
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

func (pt *PictureTask) Run() {
	pt.setStatus(StatusRunning)

	// 文件是否存在
	if !fileExists(pt.Path) {
		logger.Error(
			"指定文件不存在!",
			zap.String("path", pt.Path),
		)
		return
	}

	pt.waitIfPaused()
	// 读取文件
	buf, err := os.ReadFile(pt.Path)
	if err != nil {
		logger.Error(
			"文件读取失败!",
			zap.String("path", pt.Path),
		)
		return
	}

	pt.waitIfPaused()
	// 检测照片格式
	kind, err := filetype.Match(buf)
	if err != nil {
		logger.Error(
			"文件类型匹配失败!",
			zap.String("path", pt.Path),
		)
		return
	}

	pt.waitIfPaused()

	// 不同图像类型不同的处理方式
	fileType := kind.Extension
	logger.Info("探测到的文件类型.", zap.String("fileType", fileType))

	// 读取 hash
	hash, err := utils.HashUtils.HashFile(pt.Path, utils.SHA256)
	if err != nil {
		logger.Error(
			"Hash获取失败!",
			zap.String("path", pt.Path),
		)
		return
	}
	logger.Info("获取到 Hash", zap.String("hash", hash))

	// 获取基本信息，如果图像的很小则不进行压缩
	ctx := context.Background()
	exifData, err := tools.GetExifData(ctx, pt.Path)
	if err != nil {
		logger.Error(
			"EXIF数据获取失败!",
			zap.String("path", pt.Path),
			zap.Error(err),
		)
		return
	}

	// 分割 EXIF 数据
	splitExifData := model.SplitExifData(exifData)

	// 如果是非常规格式或 raw 则转换为 png ；如果是 PNG 则无损转换为 webp 或 jpg；如果是 webp 或 jpg 则进行压缩和其他处理
	if fileType == string(consts.FormatJPG) {
		//options := tools.DefaultOptions()
		//options.MaxSize = 800
		//options.Quality = 90
		//options.Format = "jpeg"
		//tools.LibVipsUtil.ProcessImage(options)
	} else if fileType == string(consts.FormatWEBP) {
	} else if fileType == string(consts.FormatPNG) {
	} else {

	}

	// 图像转换后，将转换后照片信息检索判断是否有必要进行压缩

	// 压缩图像 【判断图像的大小是否需要压缩】
	imgWidth := splitExifData.BaseInfo.ImageWidth
	imgHeight := splitExifData.BaseInfo.ImageHeight
	logger.Info("图像宽度",
		zap.Int("width", imgWidth),
		zap.Int("height", imgHeight),
	)

	// 如果图像的宽度或高度小于 800，则不进行压缩

	return

	// 分割 Hash 路径
	//utils.HashUtils.HashThumbPath(config.CONFIG.AppDir, hash)

	// 获取 exif
	// 保存到数据库

	pt.waitIfPaused()
	data, err := os.ReadFile(pt.Path)
	if err != nil {
		pt.setError(err)
		logger.Error(
			"文件读取失败!",
			zap.String("filename", pt.Path),
			zap.Error(err),
		)
		return
	}

	pt.ImageBuf = data
	pt.Hash = calcHash(data)

	pt.waitIfPaused()
	compressed, err := compressImage(data)
	if err != nil {
		pt.setError(err)
		return
	}
	pt.ImageBuf = compressed

	pt.waitIfPaused()
	exif, err := readExifViaCmd(pt.Path)
	if err != nil {
		pt.setError(err)
		return
	}
	pt.ExifData = exif

	pt.waitIfPaused()
	err = saveToDB(pt.Hash, exif)
	if err != nil {
		pt.setError(err)
		return
	}

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
}

func (pt *PictureTask) setStatus(s TaskStatus) {
	pt.mu.Lock()
	defer pt.mu.Unlock()
	pt.Status = s
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
	tm := &ImgTaskManager{
		tasks:        make(map[string]*PictureTask),
		queue:        make(chan *PictureTask, 100),
		workerPool:   make(chan struct{}, concurrency),
		workerLimit:  concurrency,
		globalPause:  make(chan struct{}, 1),
		globalResume: make(chan struct{}, 1),
		autoAdjust:   true,
	}
	go tm.run()
	go tm.monitorCPU()
	return tm
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

func (tm *ImgTaskManager) monitorCPU() {
	ticker := time.NewTicker(5 * time.Second)
	for range ticker.C {
		if !tm.autoAdjust {
			continue
		}
		cpuPercent := getCurrentCPUUsage()
		if cpuPercent > 80 {
			tm.SetConcurrency(tm.workerLimit / 2)
		} else if cpuPercent < 40 {
			tm.SetConcurrency(tm.workerLimit + 1)
		}
	}
}

// 模拟 CPU 使用率函数（替换为真实库，如 gopsutil）
func getCurrentCPUUsage() int {
	return runtime.NumGoroutine() * 5 // 简单估算
}

// --- 工具函数 ---

func fileExists(path string) bool {
	return utils.FileUtils.Exists(path)
}

func calcHash(data []byte) string {
	logger.Info("calcHash")
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}

func compressImage(data []byte) ([]byte, error) {
	logger.Info("compressImage")
	time.Sleep(100 * time.Millisecond) // 模拟耗时
	return data, nil
}

func readExifViaCmd(filePath string) (map[string]string, error) {
	logger.Info("readExifViaCmd")
	cmd := exec.Command("exiftool", filePath)
	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}
	result := make(map[string]string)
	lines := bytes.Split(out, []byte("\n"))
	for _, line := range lines {
		kv := bytes.SplitN(line, []byte(":"), 2)
		if len(kv) == 2 {
			key := string(bytes.TrimSpace(kv[0]))
			val := string(bytes.TrimSpace(kv[1]))
			result[key] = val
		}
	}
	return result, nil
}

func saveToDB(hash string, exif map[string]string) error {
	logger.Info("saveToDB")
	fmt.Printf("Saving %s to DB with %d exif fields\n", hash, len(exif))
	return nil
}
