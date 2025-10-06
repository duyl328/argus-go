package utils

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

// FileSystemEvent 文件系统变化事件
type FileSystemEvent struct {
	Type        string    `json:"type"`         // create, modify, delete, rename
	Path        string    `json:"path"`         // 变化的文件/文件夹路径
	Name        string    `json:"name"`         // 文件/文件夹名称
	Timestamp   time.Time `json:"timestamp"`    // 事件时间
	IsDir       bool      `json:"is_dir"`       // 是否为目录
	WatchedPath string    `json:"watched_path"` // 被监听的根路径（用于订阅匹配）
}

// FileSystemWatcher 文件系统监听器
type FileSystemWatcher struct {
	watcher       *fsnotify.Watcher
	watchedPaths  map[string]int // 路径 -> 订阅数
	pathsMutex    sync.RWMutex
	eventCallback func(*FileSystemEvent) // 事件回调
	stopChan      chan struct{}
	running       bool
	runMutex      sync.Mutex
}

// NewFileSystemWatcher 创建文件系统监听器
func NewFileSystemWatcher(eventCallback func(*FileSystemEvent)) (*FileSystemWatcher, error) {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return nil, fmt.Errorf("创建文件系统监听器失败: %w", err)
	}

	fsw := &FileSystemWatcher{
		watcher:       watcher,
		watchedPaths:  make(map[string]int),
		eventCallback: eventCallback,
		stopChan:      make(chan struct{}),
		running:       false,
	}

	return fsw, nil
}

// Start 启动监听
func (fsw *FileSystemWatcher) Start() {
	fsw.runMutex.Lock()
	if fsw.running {
		fsw.runMutex.Unlock()
		return
	}
	fsw.running = true
	fsw.runMutex.Unlock()

	go fsw.watchLoop()
}

// watchLoop 监听循环
func (fsw *FileSystemWatcher) watchLoop() {
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("文件系统监听循环 panic: %v\n", r)
		}
	}()

	for {
		select {
		case <-fsw.stopChan:
			return

		case event, ok := <-fsw.watcher.Events:
			if !ok {
				return
			}
			fsw.handleEvent(event)

		case err, ok := <-fsw.watcher.Errors:
			if !ok {
				return
			}
			fmt.Printf("文件系统监听错误: %v\n", err)
		}
	}
}

// handleEvent 处理文件系统事件
func (fsw *FileSystemWatcher) handleEvent(event fsnotify.Event) {
	// 防御性编程：捕获 panic
	defer func() {
		if r := recover(); r != nil {
			fmt.Printf("文件系统事件处理 panic: %v, 事件: %+v\n", r, event)
		}
	}()

	if fsw.eventCallback == nil {
		return
	}

	// 查找事件路径对应的被监听根路径
	watchedPath := fsw.findWatchedPath(event.Name)

	// 将 fsnotify.Event 转换为自定义事件
	fsEvent := &FileSystemEvent{
		Path:        event.Name,
		Name:        filepath.Base(event.Name),
		Timestamp:   time.Now(),
		IsDir:       false, // 默认为文件
		WatchedPath: watchedPath,
	}

	// 判断事件类型
	switch {
	case event.Op&fsnotify.Create == fsnotify.Create:
		fsEvent.Type = "create"
	case event.Op&fsnotify.Write == fsnotify.Write:
		fsEvent.Type = "modify"
	case event.Op&fsnotify.Remove == fsnotify.Remove:
		fsEvent.Type = "delete"
	case event.Op&fsnotify.Rename == fsnotify.Rename:
		fsEvent.Type = "rename"
	case event.Op&fsnotify.Chmod == fsnotify.Chmod:
		// 忽略权限变化
		return
	default:
		// 未知事件类型，记录并忽略
		fmt.Printf("未知文件系统事件类型: %v, 路径: %s\n", event.Op, event.Name)
		return
	}

	// 判断是否为目录
	// 注意：删除和重命名事件时文件可能已不存在，需要安全处理
	if fsEvent.Type != "delete" && fsEvent.Type != "rename" {
		if info, err := os.Stat(event.Name); err == nil {
			fsEvent.IsDir = info.IsDir()
		}
	}

	// 调用回调函数
	fsw.eventCallback(fsEvent)
}

// findWatchedPath 查找事件路径对应的被监听根路径
func (fsw *FileSystemWatcher) findWatchedPath(eventPath string) string {
	fsw.pathsMutex.RLock()
	defer fsw.pathsMutex.RUnlock()

	// 清理事件路径
	eventPath = filepath.Clean(eventPath)

	// 遍历所有被监听的路径，找到最长匹配的根路径
	var bestMatch string
	maxLen := 0

	for watchedPath := range fsw.watchedPaths {
		cleanWatched := filepath.Clean(watchedPath)

		// 检查事件路径是否在该监听路径下
		rel, err := filepath.Rel(cleanWatched, eventPath)
		if err == nil && !filepath.IsAbs(rel) {
			// 找到匹配的监听路径，选择最长的（最具体的）
			if len(cleanWatched) > maxLen {
				bestMatch = watchedPath
				maxLen = len(cleanWatched)
			}
		}
	}

	return bestMatch
}

// Watch 添加监听路径
func (fsw *FileSystemWatcher) Watch(path string) error {
	fsw.pathsMutex.Lock()
	defer fsw.pathsMutex.Unlock()

	// 规范化路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	// 如果已经在监听，增加订阅计数
	if count, exists := fsw.watchedPaths[absPath]; exists {
		fsw.watchedPaths[absPath] = count + 1
		fmt.Printf("文件夹监听订阅数增加: %s (订阅数: %d)\n", absPath, count+1)
		return nil
	}

	// 添加到 fsnotify 监听
	if err := fsw.watcher.Add(absPath); err != nil {
		return fmt.Errorf("添加监听失败: %w", err)
	}

	fsw.watchedPaths[absPath] = 1
	fmt.Printf("开始监听文件夹: %s\n", absPath)

	return nil
}

// Unwatch 移除监听路径
func (fsw *FileSystemWatcher) Unwatch(path string) error {
	fsw.pathsMutex.Lock()
	defer fsw.pathsMutex.Unlock()

	// 检查监听器是否已经停止
	fsw.runMutex.Lock()
	isRunning := fsw.running
	fsw.runMutex.Unlock()

	// 规范化路径
	absPath, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("获取绝对路径失败: %w", err)
	}

	count, exists := fsw.watchedPaths[absPath]
	if !exists {
		return nil // 路径未被监听
	}

	// 减少订阅计数
	if count > 1 {
		fsw.watchedPaths[absPath] = count - 1
		fmt.Printf("文件夹监听订阅数减少: %s (订阅数: %d)\n", absPath, count-1)
		return nil
	}

	// 订阅数为 0，移除监听
	// 只有在监听器运行时才调用 Remove
	if isRunning {
		if err := fsw.watcher.Remove(absPath); err != nil {
			// 忽略 "can't remove non-existent watcher" 等错误
			fmt.Printf("警告: 移除监听时出现错误 (忽略): %v\n", err)
		}
	}

	delete(fsw.watchedPaths, absPath)
	fmt.Printf("停止监听文件夹: %s\n", absPath)

	return nil
}

// GetWatchedPaths 获取所有监听的路径
func (fsw *FileSystemWatcher) GetWatchedPaths() []string {
	fsw.pathsMutex.RLock()
	defer fsw.pathsMutex.RUnlock()

	paths := make([]string, 0, len(fsw.watchedPaths))
	for path := range fsw.watchedPaths {
		paths = append(paths, path)
	}
	return paths
}

// Stop 停止监听
func (fsw *FileSystemWatcher) Stop() {
	fsw.runMutex.Lock()
	defer fsw.runMutex.Unlock()

	if !fsw.running {
		return
	}

	close(fsw.stopChan)
	fsw.watcher.Close()
	fsw.running = false
}

// Close 关闭监听器（Stop 的别名，用于资源清理）
func (fsw *FileSystemWatcher) Close() {
	fsw.Stop()
}

// ConvertEventToJSON 将事件转换为 JSON
func (fsw *FileSystemWatcher) ConvertEventToJSON(event *FileSystemEvent) (string, error) {
	data, err := json.Marshal(event)
	if err != nil {
		return "", fmt.Errorf("序列化事件失败: %w", err)
	}
	return string(data), nil
}
