package workflow

import (
	"bytes"
	"context"
	"crypto/sha1"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"os"
	"os/exec"
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

	if !fileExists(pt.Path) {
		pt.setError(errors.New("file does not exist"))
		return
	}

	pt.waitIfPaused()
	data, err := os.ReadFile(pt.Path)
	if err != nil {
		pt.setError(err)
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
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

func calcHash(data []byte) string {
	h := sha1.Sum(data)
	return hex.EncodeToString(h[:])
}

func compressImage(data []byte) ([]byte, error) {
	time.Sleep(100 * time.Millisecond) // 模拟耗时
	return data, nil
}

func readExifViaCmd(filePath string) (map[string]string, error) {
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
	fmt.Printf("Saving %s to DB with %d exif fields\n", hash, len(exif))
	return nil
}
