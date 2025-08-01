package utils

import (
	"bufio"
	"crypto/md5"
	"crypto/sha256"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// FileInfo 文件信息结构体
type FileInfo struct {
	Name    string    `json:"name"`
	Path    string    `json:"path"`
	Size    int64     `json:"size"`
	ModTime time.Time `json:"mod_time"`
	IsDir   bool      `json:"is_dir"`
	Ext     string    `json:"ext"`
}

// FilteredFiles 包含按类型分类的文件结果
type FilteredFiles struct {
	SupportedFiles []FileInfo // 支持的文件类型
	OtherFiles     []FileInfo // 其他文件类型
}

// fileUtilsStruct 用于封装文件工具方法
type fileUtilsStruct struct{}

// FileUtils 是对外暴露的文件工具对象
var FileUtils = fileUtilsStruct{}

// Exists 1. 检查文件或目录是否存在
func (fileUtilsStruct) Exists(path string) bool {
	_, err := os.Stat(path)
	return !os.IsNotExist(err)
}

// 2. 检查是否为目录
func (fileUtilsStruct) IsDir(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// 3. 检查是否为文件
func (fileUtilsStruct) IsFile(path string) bool {
	info, err := os.Stat(path)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// 4. 获取指定目录下的所有文件夹
func (fileUtilsStruct) GetDirectories(dirPath string) ([]string, error) {
	var dirs []string

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	for _, entry := range entries {
		if entry.IsDir() {
			dirs = append(dirs, entry.Name())
		}
	}

	return dirs, nil
}

// 5. 获取指定目录下指定格式的文件
func (fileUtilsStruct) GetFilesByExtension(dirPath string, extensions ...string) ([]string, error) {
	var files []string

	// 标准化扩展名（确保以.开头）
	for i, ext := range extensions {
		if !strings.HasPrefix(ext, ".") {
			extensions[i] = "." + ext
		}
		extensions[i] = strings.ToLower(extensions[i])
	}

	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("读取目录失败: %w", err)
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			fileExt := strings.ToLower(filepath.Ext(entry.Name()))
			for _, ext := range extensions {
				if fileExt == ext {
					files = append(files, entry.Name())
					break
				}
			}
		}
	}

	return files, nil
}

// 6. 找到指定目录下第一个符合格式的文件
func (f fileUtilsStruct) GetFirstFileByExtension(dirPath string, extensions ...string) (string, error) {
	files, err := f.GetFilesByExtension(dirPath, extensions...)
	if err != nil {
		return "", err
	}

	if len(files) == 0 {
		return "", fmt.Errorf("未找到符合扩展名的文件")
	}

	return files[0], nil
}

// 7. 递归获取目录下所有文件
func (fileUtilsStruct) GetAllFiles(dirPath string, recursive bool) ([]FileInfo, error) {
	var files []FileInfo

	if recursive {
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				files = append(files, FileInfo{
					Name:    info.Name(),
					Path:    path,
					Size:    info.Size(),
					ModTime: info.ModTime(),
					IsDir:   false,
					Ext:     filepath.Ext(info.Name()),
				})
			}
			return nil
		})
		return files, err
	} else {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("读取目录失败: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err != nil {
					continue
				}

				files = append(files, FileInfo{
					Name:    info.Name(),
					Path:    filepath.Join(dirPath, info.Name()),
					Size:    info.Size(),
					ModTime: info.ModTime(),
					IsDir:   false,
					Ext:     filepath.Ext(info.Name()),
				})
			}
		}
		return files, nil
	}
}

// GetFilteredFiles 根据指定的文件类型递归获取文件，并按类型分类
func (fileUtilsStruct) GetFilteredFiles(dirPath string, recursive bool, supportedTypes []string) (*FilteredFiles, error) {
	var result FilteredFiles

	// 将支持的文件类型转换为map，便于快速查找（转为小写）
	supportedMap := make(map[string]bool)
	for _, ext := range supportedTypes {
		supportedMap[strings.ToLower(ext)] = true
	}

	if recursive {
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				fileInfo := FileInfo{
					Name:    info.Name(),
					Path:    path,
					Size:    info.Size(),
					ModTime: info.ModTime(),
					IsDir:   false,
					Ext:     filepath.Ext(info.Name()),
				}

				// 检查文件扩展名是否在支持列表中（不区分大小写）
				ext := strings.ToLower(filepath.Ext(info.Name()))
				if supportedMap[ext] {
					result.SupportedFiles = append(result.SupportedFiles, fileInfo)
				} else {
					result.OtherFiles = append(result.OtherFiles, fileInfo)
				}
			}
			return nil
		})
		return &result, err
	} else {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("读取目录失败: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				info, err := entry.Info()
				if err != nil {
					continue
				}

				fileInfo := FileInfo{
					Name:    info.Name(),
					Path:    filepath.Join(dirPath, info.Name()),
					Size:    info.Size(),
					ModTime: info.ModTime(),
					IsDir:   false,
					Ext:     filepath.Ext(info.Name()),
				}

				// 检查文件扩展名是否在支持列表中（不区分大小写）
				ext := strings.ToLower(filepath.Ext(info.Name()))
				if supportedMap[ext] {
					result.SupportedFiles = append(result.SupportedFiles, fileInfo)
				} else {
					result.OtherFiles = append(result.OtherFiles, fileInfo)
				}
			}
		}
		return &result, nil
	}
}

// 递归获取目录下所有文件夹
func (fileUtilsStruct) GetAllDirs(dirPath string, recursive bool) ([]FileInfo, error) {
	var dirs []FileInfo

	if recursive {
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			// 只收集目录，排除根目录本身
			if info.IsDir() && path != dirPath {
				dirs = append(dirs, FileInfo{
					Name:    info.Name(),
					Path:    path,
					Size:    0, // 目录大小通常为0或不适用
					ModTime: info.ModTime(),
					IsDir:   true,
					Ext:     "", // 目录没有扩展名
				})
			}
			return nil
		})
		return dirs, err
	} else {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("读取目录失败: %w", err)
		}

		for _, entry := range entries {
			if entry.IsDir() {
				info, err := entry.Info()
				if err != nil {
					continue
				}

				dirs = append(dirs, FileInfo{
					Name:    info.Name(),
					Path:    filepath.Join(dirPath, info.Name()),
					Size:    0, // 目录大小通常为0或不适用
					ModTime: info.ModTime(),
					IsDir:   true,
					Ext:     "", // 目录没有扩展名
				})
			}
		}
		return dirs, nil
	}
}

// 8. 删除文件或目录
func (f fileUtilsStruct) Delete(path string) error {
	if !f.Exists(path) {
		return fmt.Errorf("路径不存在: %s", path)
	}

	return os.RemoveAll(path)
}

// 9. 创建目录（递归创建）
func (fileUtilsStruct) CreateDir(dirPath string) error {
	return os.MkdirAll(dirPath, 0755)
}

// 10. 复制文件
func (f fileUtilsStruct) CopyFile(src, dst string) error {
	// 检查目标文件是否存在
	if _, err := os.Stat(dst); err == nil {
		// 目标文件存在，跳过
		return os.ErrExist
	}

	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("打开源文件失败: %w", err)
	}
	defer sourceFile.Close()

	// 获取源文件信息
	srcInfo, err := sourceFile.Stat()
	if err != nil {
		return err
	}

	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := f.CreateDir(dstDir); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("创建目标文件失败: %w", err)
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return fmt.Errorf("复制文件失败: %w", err)
	}

	// 设置文件权限
	return os.Chmod(dst, srcInfo.Mode())
}

// CopyDir 递归复制整个目录
func (f fileUtilsStruct) CopyDir(src, dst string) error {
	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源目录信息失败: %w", err)
	}

	// 确保源路径是一个目录
	if !srcInfo.IsDir() {
		return fmt.Errorf("源路径不是目录: %s", src)
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("读取源目录失败: %w", err)
	}

	// 遍历源目录中的每个条目
	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			// 如果是目录，递归复制
			if err := f.CopyDir(srcPath, dstPath); err != nil {
				// 子目录失败也继续复制其他内容
				fmt.Printf("[WARN] 复制子目录 %s 失败: %v\n", entry.Name(), err)
				continue
			}
		} else {
			err := f.CopyFile(srcPath, dstPath)
			if err != nil {
				// 判断是否是文件已存在或拒绝访问
				if os.IsPermission(err) || os.IsExist(err) {
					fmt.Printf("[WARN] 复制文件 %s 被跳过: %v\n", dstPath, err)
					continue
				}
				// 其他错误还是报错
				return fmt.Errorf("复制文件 %s 失败: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// CopyDirWithOptions 带选项的目录复制函数
type CopyOptions struct {
	Overwrite    bool // 是否覆盖已存在的文件
	SkipSymlinks bool // 是否跳过符号链接
}

func (f fileUtilsStruct) CopyDirWithOptions(src, dst string, opts *CopyOptions) error {
	if opts == nil {
		opts = &CopyOptions{Overwrite: true, SkipSymlinks: false}
	}

	// 获取源目录信息
	srcInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("获取源目录信息失败: %w", err)
	}

	if !srcInfo.IsDir() {
		return fmt.Errorf("源路径不是目录: %s", src)
	}

	// 检查目标目录是否已存在
	if dstInfo, err := os.Stat(dst); err == nil {
		if !dstInfo.IsDir() {
			return fmt.Errorf("目标路径已存在但不是目录: %s", dst)
		}
		if !opts.Overwrite {
			return fmt.Errorf("目标目录已存在: %s", dst)
		}
	} else if !os.IsNotExist(err) {
		return fmt.Errorf("检查目标目录失败: %w", err)
	}

	// 创建目标目录
	if err := os.MkdirAll(dst, srcInfo.Mode()); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	// 读取源目录内容
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("读取源目录失败: %w", err)
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		// 处理符号链接
		if entry.Type()&os.ModeSymlink != 0 {
			if opts.SkipSymlinks {
				continue
			}
			if err := f.copySymlink(srcPath, dstPath); err != nil {
				return fmt.Errorf("复制符号链接 %s 失败: %w", entry.Name(), err)
			}
			continue
		}

		if entry.IsDir() {
			// 递归复制子目录
			if err := f.CopyDirWithOptions(srcPath, dstPath, opts); err != nil {
				return fmt.Errorf("复制子目录 %s 失败: %w", entry.Name(), err)
			}
		} else {
			// 检查文件是否已存在
			if !opts.Overwrite {
				if _, err := os.Stat(dstPath); err == nil {
					return fmt.Errorf("目标文件已存在: %s", dstPath)
				}
			}

			// 复制文件
			if err := f.CopyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("复制文件 %s 失败: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

// copySymlink 复制符号链接
func (fileUtilsStruct) copySymlink(src, dst string) error {
	link, err := os.Readlink(src)
	if err != nil {
		return err
	}

	// 删除可能已存在的目标文件/链接
	os.Remove(dst)

	return os.Symlink(link, dst)
}

// 11. 移动/重命名文件
func (f fileUtilsStruct) MoveFile(src, dst string) error {
	// 确保目标目录存在
	dstDir := filepath.Dir(dst)
	if err := f.CreateDir(dstDir); err != nil {
		return fmt.Errorf("创建目标目录失败: %w", err)
	}

	return os.Rename(src, dst)
}

// 12. 读取文件全部内容
func (fileUtilsStruct) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %w", err)
	}
	return string(content), nil
}

// 13. 按行读取文件
func (fileUtilsStruct) ReadFileLines(filePath string) ([]string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	var lines []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		lines = append(lines, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("读取文件失败: %w", err)
	}

	return lines, nil
}

// 14. 写入文件内容
func (f fileUtilsStruct) WriteFile(filePath, content string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := f.CreateDir(dir); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	return os.WriteFile(filePath, []byte(content), 0644)
}

// 15. 追加写入文件
func (f fileUtilsStruct) AppendFile(filePath, content string) error {
	// 确保目录存在
	dir := filepath.Dir(filePath)
	if err := f.CreateDir(dir); err != nil {
		return fmt.Errorf("创建目录失败: %w", err)
	}

	file, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	_, err = file.WriteString(content)
	return err
}

// 16. 获取文件大小
func (fileUtilsStruct) GetFileSize(filePath string) (int64, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return 0, fmt.Errorf("获取文件信息失败: %w", err)
	}
	return info.Size(), nil
}

// 17. 获取文件修改时间
func (fileUtilsStruct) GetModTime(filePath string) (time.Time, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return time.Time{}, fmt.Errorf("获取文件信息失败: %w", err)
	}
	return info.ModTime(), nil
}

// 18. 获取文件扩展名
func (fileUtilsStruct) GetExtension(filePath string) string {
	return filepath.Ext(filePath)
}

// 19. 获取文件名（不含扩展名）
func (fileUtilsStruct) GetFileName(filePath string) string {
	base := filepath.Base(filePath)
	ext := filepath.Ext(base)
	return strings.TrimSuffix(base, ext)
}

// 20. 获取文件基本名称（含扩展名）
func (fileUtilsStruct) GetBaseName(filePath string) string {
	return filepath.Base(filePath)
}

// 21. 获取文件目录
func (fileUtilsStruct) GetDir(filePath string) string {
	return filepath.Dir(filePath)
}

// 22. 计算文件MD5哈希值
func (fileUtilsStruct) GetFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("计算哈希失败: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// 23. 计算文件SHA256哈希值
func (fileUtilsStruct) GetFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", fmt.Errorf("打开文件失败: %w", err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", fmt.Errorf("计算哈希失败: %w", err)
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

// 24. 清空目录内容（保留目录本身）
func (fileUtilsStruct) CleanDir(dirPath string) error {
	entries, err := os.ReadDir(dirPath)
	if err != nil {
		return fmt.Errorf("读取目录失败: %w", err)
	}

	for _, entry := range entries {
		path := filepath.Join(dirPath, entry.Name())
		if err := os.RemoveAll(path); err != nil {
			return fmt.Errorf("删除 %s 失败: %w", path, err)
		}
	}

	return nil
}

// 25. 获取目录大小
func (fileUtilsStruct) GetDirSize(dirPath string) (int64, error) {
	var size int64

	err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() {
			size += info.Size()
		}
		return nil
	})

	return size, err
}

// 26. 格式化文件大小
func (fileUtilsStruct) FormatFileSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// 27. 搜索文件（按文件名模式）
func (fileUtilsStruct) SearchFiles(dirPath, pattern string, recursive bool) ([]string, error) {
	var matches []string

	if recursive {
		err := filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}

			if !info.IsDir() {
				matched, err := filepath.Match(pattern, info.Name())
				if err != nil {
					return err
				}
				if matched {
					matches = append(matches, path)
				}
			}
			return nil
		})
		return matches, err
	} else {
		entries, err := os.ReadDir(dirPath)
		if err != nil {
			return nil, fmt.Errorf("读取目录失败: %w", err)
		}

		for _, entry := range entries {
			if !entry.IsDir() {
				matched, err := filepath.Match(pattern, entry.Name())
				if err != nil {
					return nil, err
				}
				if matched {
					matches = append(matches, filepath.Join(dirPath, entry.Name()))
				}
			}
		}
		return matches, nil
	}
}

// 28. 获取当前工作目录
func (fileUtilsStruct) GetCurrentDir() (string, error) {
	return os.Getwd()
}

// 29. 改变工作目录
func (fileUtilsStruct) ChangeDir(dirPath string) error {
	return os.Chdir(dirPath)
}
