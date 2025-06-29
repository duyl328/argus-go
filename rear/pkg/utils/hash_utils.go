package utils

import (
	"crypto/md5"
	"crypto/sha1"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

type hashUtilsStruct struct{}

var HashUtils = hashUtilsStruct{}

// HashType 定义支持的Hash算法类型
type HashType int

const (
	MD5 HashType = iota
	SHA1
	SHA256
	SHA512
)

// 缓冲区大小配置
const (
	SmallFileThreshold = 1024 * 1024 // 1MB，小文件阈值
	DefaultBufferSize  = 64 * 1024   // 64KB，默认缓冲区
	LargeBufferSize    = 1024 * 1024 // 1MB，大文件缓冲区
)

// HashResult 存储Hash计算结果
type HashResult struct {
	Filename string
	Hash     string
	Error    error
	Size     int64
}

// getHasher 根据类型返回对应的Hash实例
func (h hashUtilsStruct) getHasher(hashType HashType) hash.Hash {
	switch hashType {
	case MD5:
		return md5.New()
	case SHA1:
		return sha1.New()
	case SHA256:
		return sha256.New()
	case SHA512:
		return sha512.New()
	default:
		return sha256.New()
	}
}

// getHasherName 获取Hash算法名称
func (h hashUtilsStruct) getHasherName(hashType HashType) string {
	switch hashType {
	case MD5:
		return "MD5"
	case SHA1:
		return "SHA1"
	case SHA256:
		return "SHA256"
	case SHA512:
		return "SHA512"
	default:
		return "SHA256"
	}
}

// HashString 计算字符串的Hash值
func (h hashUtilsStruct) HashString(data string, hashType HashType) string {
	hasher := h.getHasher(hashType)
	hasher.Write([]byte(data))
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// HashBytes 计算字节数组的Hash值
func (h hashUtilsStruct) HashBytes(data []byte, hashType HashType) string {
	hasher := h.getHasher(hashType)
	hasher.Write(data)
	return fmt.Sprintf("%x", hasher.Sum(nil))
}

// HashFile 计算单个文件的Hash值（自动选择优化策略）
func (h hashUtilsStruct) HashFile(filename string, hashType HashType) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	// 获取文件信息
	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	hasher := h.getHasher(hashType)

	// 根据文件大小选择不同的处理策略
	if fileInfo.Size() <= SmallFileThreshold {
		// 小文件：直接读取
		return h.hashSmallFile(file, hasher)
	} else {
		// 大文件：使用缓冲区
		return h.hashLargeFile(file, hasher, fileInfo.Size())
	}
}

// hashSmallFile 处理小文件
func (h hashUtilsStruct) hashSmallFile(file *os.File, hasher hash.Hash) (string, error) {
	_, err := io.Copy(hasher, file)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// hashLargeFile 处理大文件
func (h hashUtilsStruct) hashLargeFile(file *os.File, hasher hash.Hash, fileSize int64) (string, error) {
	// 根据文件大小动态调整缓冲区
	bufferSize := DefaultBufferSize
	if fileSize > 100*1024*1024 { // 大于100MB使用更大缓冲区
		bufferSize = LargeBufferSize
	}

	buffer := make([]byte, bufferSize)

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// HashFileWithProgress 计算文件Hash值并提供进度回调
func (h hashUtilsStruct) HashFileWithProgress(filename string, hashType HashType, progressCallback func(processed, total int64)) (string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return "", err
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return "", err
	}

	hasher := h.getHasher(hashType)
	buffer := make([]byte, DefaultBufferSize)
	var processed int64

	for {
		n, err := file.Read(buffer)
		if n > 0 {
			hasher.Write(buffer[:n])
			processed += int64(n)
			if progressCallback != nil {
				progressCallback(processed, fileInfo.Size())
			}
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", err
		}
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// HashMultipleFiles 并发计算多个文件的Hash值
func (h hashUtilsStruct) HashMultipleFiles(filenames []string, hashType HashType) []HashResult {
	// 使用CPU核心数作为并发数
	numWorkers := runtime.NumCPU()
	jobs := make(chan string, len(filenames))
	results := make(chan HashResult, len(filenames))

	// 启动工作协程
	var wg sync.WaitGroup
	for i := 0; i < numWorkers; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			for filename := range jobs {
				hash, err := h.HashFile(filename, hashType)

				var size int64
				if err == nil {
					if fileInfo, statErr := os.Stat(filename); statErr == nil {
						size = fileInfo.Size()
					}
				}

				results <- HashResult{
					Filename: filename,
					Hash:     hash,
					Error:    err,
					Size:     size,
				}
			}
		}()
	}

	// 发送任务
	go func() {
		for _, filename := range filenames {
			jobs <- filename
		}
		close(jobs)
	}()

	// 等待所有工作完成
	go func() {
		wg.Wait()
		close(results)
	}()

	// 收集结果
	var hashResults []HashResult
	for result := range results {
		hashResults = append(hashResults, result)
	}

	return hashResults
}

// CompareFiles 比较两个文件是否相同（通过Hash值）
func (h hashUtilsStruct) CompareFiles(file1, file2 string, hashType HashType) (bool, error) {
	hash1, err := h.HashFile(file1, hashType)
	if err != nil {
		return false, err
	}

	hash2, err := h.HashFile(file2, hashType)
	if err != nil {
		return false, err
	}

	return hash1 == hash2, nil
}

// FindDuplicates 在文件列表中查找重复文件
func (h hashUtilsStruct) FindDuplicates(filenames []string, hashType HashType) map[string][]string {
	results := h.HashMultipleFiles(filenames, hashType)
	duplicates := make(map[string][]string)

	for _, result := range results {
		if result.Error == nil {
			duplicates[result.Hash] = append(duplicates[result.Hash], result.Filename)
		}
	}

	// 只保留有重复的Hash
	for hash, files := range duplicates {
		if len(files) <= 1 {
			delete(duplicates, hash)
		}
	}

	return duplicates
}

// VerifyFileIntegrity 验证文件完整性
func (h hashUtilsStruct) VerifyFileIntegrity(filename, expectedHash string, hashType HashType) (bool, error) {
	actualHash, err := h.HashFile(filename, hashType)
	if err != nil {
		return false, err
	}

	// 不区分大小写比较
	return strings.EqualFold(actualHash, expectedHash), nil
}

// HashFileMultipleAlgorithms 使用多种算法计算文件Hash
func (h hashUtilsStruct) HashFileMultipleAlgorithms(filename string, hashTypes []HashType) (map[string]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// 创建多个hasher
	hashers := make(map[HashType]hash.Hash)
	for _, hashType := range hashTypes {
		hashers[hashType] = h.getHasher(hashType)
	}

	// 创建MultiWriter
	var writers []io.Writer
	for _, hasher := range hashers {
		writers = append(writers, hasher)
	}
	multiWriter := io.MultiWriter(writers...)

	// 一次读取，多重计算
	_, err = io.Copy(multiWriter, file)
	if err != nil {
		return nil, err
	}

	// 收集结果
	results := make(map[string]string)
	for hashType, hasher := range hashers {
		hashName := h.getHasherName(hashType)
		results[hashName] = fmt.Sprintf("%x", hasher.Sum(nil))
	}

	return results, nil
}

// HashThumbPath 将 Hash 转换为具体路径
func (h hashUtilsStruct) HashThumbPath(
	basePath string,
	hash string,
	filename string, // 文件名，不带扩展名
	ext string, // 扩展名，不带点，如 "jpg", "webp"
) string {
	var parts []string
	start := 0
	segments := []int{2, 2, 2}
	for _, length := range segments {
		end := start + length
		if end > len(hash) {
			break
		}
		parts = append(parts, hash[start:end])
		start = end
	}
	parts = append(parts, hash) // 最后一级目录为完整 hash

	dir := filepath.Join(append([]string{basePath}, parts...)...)
	fullName := fmt.Sprintf("%s.%s", filename, ext)
	return filepath.Join(dir, fullName)
}

// 便捷函数
func (h hashUtilsStruct) MD5File(filename string) (string, error) {
	return h.HashFile(filename, MD5)
}

func (h hashUtilsStruct) SHA1File(filename string) (string, error) {
	return h.HashFile(filename, SHA1)
}

func (h hashUtilsStruct) SHA256File(filename string) (string, error) {
	return h.HashFile(filename, SHA256)
}

func (h hashUtilsStruct) SHA512File(filename string) (string, error) {
	return h.HashFile(filename, SHA512)
}

func (h hashUtilsStruct) MD5String(data string) string {
	return h.HashString(data, MD5)
}

func (h hashUtilsStruct) SHA256String(data string) string {
	return h.HashString(data, SHA256)
}
