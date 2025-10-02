package service

import (
	"fmt"
	"path/filepath"
	"rear/pkg/system"
	"rear/pkg/utils"
	"runtime"
	"sort"
	"strings"
	"time"
)

// FileSystemItemType 文件系统项目类型
type FileSystemItemType string

const (
	ItemTypeDrive      FileSystemItemType = "drive"      // 驱动器/挂载点
	ItemTypeDirectory  FileSystemItemType = "directory"  // 目录
	ItemTypeFile       FileSystemItemType = "file"       // 文件
)

// FileSystemItem 文件系统项目统一结构
type FileSystemItem struct {
	ID           string             `json:"id"`                     // 唯一标识
	Name         string             `json:"name"`                   // 显示名称
	Path         string             `json:"path"`                   // 完整路径
	Type         FileSystemItemType `json:"type"`                   // 项目类型
	Size         int64              `json:"size"`                   // 大小（字节）
	ModTime      time.Time          `json:"mod_time"`               // 修改时间
	IsAccessible bool               `json:"is_accessible"`          // 是否可访问

	// 驱动器特有属性
	DriveInfo *DriveInfo `json:"drive_info,omitempty"`

	// 文件特有属性
	FileInfo *FileInfo `json:"file_info,omitempty"`

	// 目录特有属性
	DirectoryInfo *DirectoryInfo `json:"directory_info,omitempty"`
}

// DriveInfo 驱动器信息
type DriveInfo struct {
	Label         string `json:"label"`          // 卷标
	FileSystem    string `json:"file_system"`    // 文件系统类型
	TotalSpace    int64  `json:"total_space"`    // 总空间
	FreeSpace     int64  `json:"free_space"`     // 可用空间
	UsedSpace     int64  `json:"used_space"`     // 已用空间
	UsagePercent  float64 `json:"usage_percent"` // 使用百分比
	IsRemovable   bool   `json:"is_removable"`   // 是否可移动
	DriveType     string `json:"drive_type"`     // 驱动器类型
}

// FileInfo 文件信息
type FileInfo struct {
	Extension string `json:"extension"` // 文件扩展名
	MimeType  string `json:"mime_type"` // MIME类型
}

// DirectoryInfo 目录信息
type DirectoryInfo struct {
	ItemCount int `json:"item_count"` // 子项目数量
}

// FileSystemResponse 文件系统响应
type FileSystemResponse struct {
	CurrentPath string           `json:"current_path"` // 当前路径
	ParentPath  string           `json:"parent_path"`  // 父路径
	Items       []FileSystemItem `json:"items"`        // 项目列表
	Summary     *PathSummary     `json:"summary"`      // 路径摘要信息
}

// PathSummary 路径摘要信息
type PathSummary struct {
	TotalItems      int `json:"total_items"`      // 总项目数
	DirectoryCount  int `json:"directory_count"`  // 目录数量
	FileCount       int `json:"file_count"`       // 文件数量
	DriveCount      int `json:"drive_count"`      // 驱动器数量
}

// FileSystemService 文件系统服务
type FileSystemService struct {
	deviceManager *system.DeviceManager
}

// NewFileSystemService 创建文件系统服务实例
func NewFileSystemService() *FileSystemService {
	return &FileSystemService{
		deviceManager: system.NewDeviceManager(),
	}
}

// Browse 浏览文件系统
// path为空或"/"时返回驱动器/挂载点列表，否则返回指定路径的内容
func (fs *FileSystemService) Browse(path string) (*FileSystemResponse, error) {
	// 标准化路径
	if path == "" || path == "/" {
		return fs.getRootLevel()
	}

	// 清理路径
	path = filepath.Clean(path)

	// 检查路径是否存在
	if !utils.FileUtils.Exists(path) {
		return nil, fmt.Errorf("路径不存在: %s", path)
	}

	// 检查是否为目录
	if !utils.FileUtils.IsDir(path) {
		return nil, fmt.Errorf("路径不是目录: %s", path)
	}

	return fs.getDirectoryContents(path)
}

// getRootLevel 获取根级别内容（驱动器列表）
func (fs *FileSystemService) getRootLevel() (*FileSystemResponse, error) {
	// 扫描存储设备
	if err := fs.deviceManager.ScanDevices(); err != nil {
		return nil, fmt.Errorf("扫描存储设备失败: %w", err)
	}

	devices := fs.deviceManager.GetReadyDevices()
	items := make([]FileSystemItem, 0, len(devices))

	for _, device := range devices {
		// 计算使用百分比
		usagePercent := 0.0
		if device.TotalSpace > 0 {
			usagePercent = float64(device.UsedSpace) / float64(device.TotalSpace) * 100
		}

		item := FileSystemItem{
			ID:           device.ID,
			Name:         fs.formatDriveName(device),
			Path:         device.MountPoint,
			Type:         ItemTypeDrive,
			Size:         device.TotalSpace,
			ModTime:      device.CreatedAt,
			IsAccessible: device.IsReady,
			DriveInfo: &DriveInfo{
				Label:         device.Label,
				FileSystem:    device.FileSystem,
				TotalSpace:    device.TotalSpace,
				FreeSpace:     device.FreeSpace,
				UsedSpace:     device.UsedSpace,
				UsagePercent:  usagePercent,
				IsRemovable:   device.IsRemovable,
				DriveType:     string(device.Type),
			},
		}

		items = append(items, item)
	}

	// 按挂载点排序
	sort.Slice(items, func(i, j int) bool {
		return items[i].Path < items[j].Path
	})

	return &FileSystemResponse{
		CurrentPath: "/",
		ParentPath:  "",
		Items:       items,
		Summary: &PathSummary{
			TotalItems: len(items),
			DriveCount: len(items),
		},
	}, nil
}

// getDirectoryContents 获取目录内容
func (fs *FileSystemService) getDirectoryContents(path string) (*FileSystemResponse, error) {
	// 获取目录下的所有文件夹
	dirs, err := utils.FileUtils.GetAllDirs(path, false)
	if err != nil {
		return nil, fmt.Errorf("获取目录列表失败: %w", err)
	}

	// 获取目录下的所有文件
	files, err := utils.FileUtils.GetAllFiles(path, false)
	if err != nil {
		return nil, fmt.Errorf("获取文件列表失败: %w", err)
	}

	items := make([]FileSystemItem, 0, len(dirs)+len(files))

	// 添加目录
	for _, dir := range dirs {
		item := FileSystemItem{
			ID:           fmt.Sprintf("dir_%s", strings.ReplaceAll(dir.Path, string(filepath.Separator), "_")),
			Name:         dir.Name,
			Path:         dir.Path,
			Type:         ItemTypeDirectory,
			Size:         0, // 目录大小为0或需要特殊计算
			ModTime:      dir.ModTime,
			IsAccessible: true, // 假设可访问，可以通过实际检查优化
			DirectoryInfo: &DirectoryInfo{
				ItemCount: 0, // 可以选择性地计算子项目数量
			},
		}
		items = append(items, item)
	}

	// 添加文件
	for _, file := range files {
		item := FileSystemItem{
			ID:           fmt.Sprintf("file_%s", strings.ReplaceAll(file.Path, string(filepath.Separator), "_")),
			Name:         file.Name,
			Path:         file.Path,
			Type:         ItemTypeFile,
			Size:         file.Size,
			ModTime:      file.ModTime,
			IsAccessible: true,
			FileInfo: &FileInfo{
				Extension: file.Ext,
				MimeType:  "", // 可以通过文件扩展名推断MIME类型
			},
		}
		items = append(items, item)
	}

	// 排序：目录在前，然后按名称排序
	sort.Slice(items, func(i, j int) bool {
		if items[i].Type != items[j].Type {
			return items[i].Type == ItemTypeDirectory
		}
		return strings.ToLower(items[i].Name) < strings.ToLower(items[j].Name)
	})

	// 计算父路径
	parentPath := filepath.Dir(path)
	if parentPath == path {
		parentPath = "/" // 根目录的父路径
	}

	return &FileSystemResponse{
		CurrentPath: path,
		ParentPath:  parentPath,
		Items:       items,
		Summary: &PathSummary{
			TotalItems:     len(items),
			DirectoryCount: len(dirs),
			FileCount:      len(files),
		},
	}, nil
}

// formatDriveName 格式化驱动器名称
func (fs *FileSystemService) formatDriveName(device system.StorageDevice) string {
	switch runtime.GOOS {
	case "windows":
		// Windows: "C: (Windows)" 或 "D: (本地磁盘)"
		if device.Label != "" {
			return fmt.Sprintf("%s (%s)", device.Name, device.Label)
		}
		return fmt.Sprintf("%s (本地磁盘)", device.Name)
	default:
		// Unix: "/" 或 "/home" 等
		if device.Label != "" {
			return fmt.Sprintf("%s (%s)", device.MountPoint, device.Label)
		}
		return device.MountPoint
	}
}

// GetDiskUsage 获取磁盘使用情况
func (fs *FileSystemService) GetDiskUsage(path string) (*DriveInfo, error) {
	// 获取对应的设备信息
	device := fs.deviceManager.GetDevice(path)
	if device == nil {
		return nil, fmt.Errorf("未找到对应的设备: %s", path)
	}

	// 刷新设备信息
	if err := fs.deviceManager.RefreshDevice(path); err != nil {
		return nil, fmt.Errorf("刷新设备信息失败: %w", err)
	}

	usagePercent := 0.0
	if device.TotalSpace > 0 {
		usagePercent = float64(device.UsedSpace) / float64(device.TotalSpace) * 100
	}

	return &DriveInfo{
		Label:         device.Label,
		FileSystem:    device.FileSystem,
		TotalSpace:    device.TotalSpace,
		FreeSpace:     device.FreeSpace,
		UsedSpace:     device.UsedSpace,
		UsagePercent:  usagePercent,
		IsRemovable:   device.IsRemovable,
		DriveType:     string(device.Type),
	}, nil
}

// FormatSize 格式化文件大小
func (fs *FileSystemService) FormatSize(bytes int64) string {
	return utils.FileUtils.FormatFileSize(bytes)
}

// CreateDirectory 创建目录
func (fs *FileSystemService) CreateDirectory(path string) (map[string]interface{}, error) {
	// 检查路径是否已存在
	if utils.FileUtils.Exists(path) {
		return nil, fmt.Errorf("路径已存在: %s", path)
	}

	// 创建目录
	if err := utils.FileUtils.CreateDir(path); err != nil {
		return nil, fmt.Errorf("创建目录失败: %w", err)
	}

	return map[string]interface{}{
		"path":       path,
		"created_at": time.Now().Format(time.RFC3339),
	}, nil
}

// DeleteItem 删除文件或目录
func (fs *FileSystemService) DeleteItem(path string, operationID string) (map[string]interface{}, error) {
	// 检查路径是否存在
	if !utils.FileUtils.Exists(path) {
		return nil, fmt.Errorf("路径不存在: %s", path)
	}

	// 删除文件或目录
	if err := utils.FileUtils.Delete(path); err != nil {
		return nil, fmt.Errorf("删除失败: %w", err)
	}

	return map[string]interface{}{
		"path":         path,
		"deleted_at":   time.Now().Format(time.RFC3339),
		"operation_id": operationID,
	}, nil
}

// MoveItem 移动/重命名文件或目录
func (fs *FileSystemService) MoveItem(source, destination, operationID string, overwrite bool) (map[string]interface{}, error) {
	// 检查源路径是否存在
	if !utils.FileUtils.Exists(source) {
		return nil, fmt.Errorf("源路径不存在: %s", source)
	}

	// 检查目标路径是否已存在
	if !overwrite && utils.FileUtils.Exists(destination) {
		return nil, fmt.Errorf("目标路径已存在: %s", destination)
	}

	// 移动文件或目录
	if err := utils.FileUtils.MoveFile(source, destination); err != nil {
		return nil, fmt.Errorf("移动失败: %w", err)
	}

	return map[string]interface{}{
		"source":       source,
		"destination":  destination,
		"operation_id": operationID,
	}, nil
}

// CopyItem 复制文件或目录
func (fs *FileSystemService) CopyItem(source, destination, operationID string, overwrite bool) (map[string]interface{}, error) {
	// 检查源路径是否存在
	if !utils.FileUtils.Exists(source) {
		return nil, fmt.Errorf("源路径不存在: %s", source)
	}

	// 获取源文件/目录信息
	isDir := utils.FileUtils.IsDir(source)

	var err error
	var size int64

	if isDir {
		// 复制目录
		opts := &utils.CopyOptions{
			Overwrite:    overwrite,
			SkipSymlinks: false,
		}
		err = utils.FileUtils.CopyDirWithOptions(source, destination, opts)
		if err != nil {
			return nil, fmt.Errorf("复制目录失败: %w", err)
		}
		// 计算目录大小
		size, _ = utils.FileUtils.GetDirSize(destination)
	} else {
		// 检查目标文件是否已存在
		if !overwrite && utils.FileUtils.Exists(destination) {
			return nil, fmt.Errorf("目标文件已存在: %s", destination)
		}
		// 复制文件
		err = utils.FileUtils.CopyFile(source, destination)
		if err != nil {
			return nil, fmt.Errorf("复制文件失败: %w", err)
		}
		size, _ = utils.FileUtils.GetFileSize(destination)
	}

	return map[string]interface{}{
		"source":       source,
		"destination":  destination,
		"size":         size,
		"operation_id": operationID,
	}, nil
}

// SearchFiles 搜索文件
func (fs *FileSystemService) SearchFiles(path, pattern, fileType string, recursive bool) (map[string]interface{}, error) {
	// 检查路径是否存在
	if !utils.FileUtils.Exists(path) {
		return nil, fmt.Errorf("路径不存在: %s", path)
	}

	// 检查是否为目录
	if !utils.FileUtils.IsDir(path) {
		return nil, fmt.Errorf("路径不是目录: %s", path)
	}

	startTime := time.Now()

	// 搜索文件
	matches, err := utils.FileUtils.SearchFiles(path, pattern, recursive)
	if err != nil {
		return nil, fmt.Errorf("搜索文件失败: %w", err)
	}

	// 根据文件类型过滤
	results := make([]FileSystemItem, 0)
	for _, matchPath := range matches {
		// 获取文件信息
		info, err := utils.FileUtils.GetFileSize(matchPath)
		if err != nil {
			continue
		}

		modTime, _ := utils.FileUtils.GetModTime(matchPath)
		ext := utils.FileUtils.GetExtension(matchPath)
		name := utils.FileUtils.GetBaseName(matchPath)

		// 判断文件类型
		isPhoto := fs.isPhotoFile(ext)
		isVideo := fs.isVideoFile(ext)

		// 根据类型过滤
		if fileType == "photo" && !isPhoto {
			continue
		}
		if fileType == "video" && !isVideo {
			continue
		}

		item := FileSystemItem{
			ID:           fmt.Sprintf("file_%s", strings.ReplaceAll(matchPath, string(filepath.Separator), "_")),
			Name:         name,
			Path:         matchPath,
			Type:         ItemTypeFile,
			Size:         info,
			ModTime:      modTime,
			IsAccessible: true,
		}

		results = append(results, item)
	}

	duration := time.Since(startTime).Milliseconds()

	return map[string]interface{}{
		"results":             results,
		"total_count":         len(results),
		"search_duration_ms":  duration,
	}, nil
}

// 辅助函数：判断是否为照片文件
func (fs *FileSystemService) isPhotoFile(ext string) bool {
	photoExts := []string{".jpg", ".jpeg", ".png", ".gif", ".bmp", ".webp", ".heic", ".heif", ".raw", ".cr2", ".nef", ".arw"}
	ext = strings.ToLower(ext)
	for _, photoExt := range photoExts {
		if ext == photoExt {
			return true
		}
	}
	return false
}

// 辅助函数：判断是否为视频文件
func (fs *FileSystemService) isVideoFile(ext string) bool {
	videoExts := []string{".mp4", ".avi", ".mov", ".wmv", ".flv", ".mkv", ".webm", ".m4v", ".mpg", ".mpeg"}
	ext = strings.ToLower(ext)
	for _, videoExt := range videoExts {
		if ext == videoExt {
			return true
		}
	}
	return false
}