package utils

import (
	"fmt"
	"log"
	"time"
)

func main() {
	// 1. 计算字符串Hash
	fmt.Println("=== 字符串Hash ===")
	text := "Hello, World!"
	fmt.Printf("MD5:    %s\n", MD5String(text))
	fmt.Printf("SHA256: %s\n", SHA256String(text))

	// 2. 计算单个文件Hash
	fmt.Println("\n=== 单个文件Hash ===")
	filename := "example.txt"
	if hash, err := SHA256File(filename); err != nil {
		log.Printf("计算文件Hash失败: %v", err)
	} else {
		fmt.Printf("文件 %s 的 SHA256: %s\n", filename, hash)
	}

	// 3. 带进度的文件Hash计算
	fmt.Println("\n=== 带进度的Hash计算 ===")
	progressCallback := func(processed, total int64) {
		percentage := float64(processed) / float64(total) * 100
		fmt.Printf("\r进度: %.2f%% (%d/%d bytes)", percentage, processed, total)
	}

	if hash, err := HashFileWithProgress(filename, SHA256, progressCallback); err != nil {
		log.Printf("计算失败: %v", err)
	} else {
		fmt.Printf("\n最终Hash: %s\n", hash)
	}

	// 4. 并发计算多个文件
	fmt.Println("\n=== 并发计算多个文件 ===")
	files := []string{"file1.txt", "file2.txt", "file3.txt"}
	results := HashMultipleFiles(files, SHA256)

	for _, result := range results {
		if result.Error != nil {
			fmt.Printf("文件 %s 计算失败: %v\n", result.Filename, result.Error)
		} else {
			fmt.Printf("文件 %s (%.2f KB): %s\n",
				result.Filename,
				float64(result.Size)/1024,
				result.Hash)
		}
	}

	// 5. 计算目录下所有文件
	// fmt.Println("\n=== 目录Hash计算 ===")
	// dirResults, err := HashDirectory("./testdir", SHA256, true)
	// if err != nil {
	// 	log.Printf("计算目录Hash失败: %v", err)
	// } else {
	// 	for _, result := range dirResults {
	// 		if result.Error == nil {
	// 			fmt.Printf("%s: %s\n", result.Filename, result.Hash)
	// 		}
	// 	}
	// }

	// 6. 查找重复文件
	fmt.Println("\n=== 查找重复文件 ===")
	allFiles := []string{"file1.txt", "file2.txt", "file3.txt", "file1_copy.txt"}
	duplicates := FindDuplicates(allFiles, SHA256)

	for hash, files := range duplicates {
		fmt.Printf("Hash %s 的重复文件:\n", hash[:16]+"...")
		for _, file := range files {
			fmt.Printf("  - %s\n", file)
		}
	}

	// 7. 比较两个文件
	fmt.Println("\n=== 文件比较 ===")
	if same, err := CompareFiles("file1.txt", "file2.txt", SHA256); err != nil {
		log.Printf("文件比较失败: %v", err)
	} else {
		if same {
			fmt.Println("两个文件内容相同")
		} else {
			fmt.Println("两个文件内容不同")
		}
	}

	// 8. 验证文件完整性
	fmt.Println("\n=== 文件完整性验证 ===")
	expectedHash := "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855" // 空文件的SHA256
	if valid, err := VerifyFileIntegrity("empty.txt", expectedHash, SHA256); err != nil {
		log.Printf("验证失败: %v", err)
	} else {
		if valid {
			fmt.Println("文件完整性验证通过")
		} else {
			fmt.Println("文件完整性验证失败")
		}
	}

	// 9. 多算法同时计算
	fmt.Println("\n=== 多算法Hash计算 ===")
	hashTypes := []HashType{
		MD5,
		SHA1,
		SHA256,
		SHA512,
	}

	multiHashes, err := HashFileMultipleAlgorithms(filename, hashTypes)
	if err != nil {
		log.Printf("多算法计算失败: %v", err)
	} else {
		for algorithm, hash := range multiHashes {
			fmt.Printf("%s: %s\n", algorithm, hash)
		}
	}

	// 10. 获取文件详细信息
	// fmt.Println("\n=== 文件详细信息 ===")
	// if fileInfo, err := GetFileInfo(filename, SHA256); err != nil {
	// 	log.Printf("获取文件信息失败: %v", err)
	// } else {
	// 	fmt.Printf("文件名: %s\n", fileInfo.Name)
	// 	fmt.Printf("路径: %s\n", fileInfo.Path)
	// 	fmt.Printf("大小: %d bytes\n", fileInfo.Size)
	// 	fmt.Printf("Hash类型: %s\n", fileInfo.HashType)
	// 	fmt.Printf("Hash值: %s\n", fileInfo.Hash)
	// }

	// 11. 性能测试示例
	fmt.Println("\n=== 性能测试 ===")
	start := time.Now()
	_, err = SHA256File("largefile.txt")
	if err != nil {
		log.Printf("大文件Hash计算失败: %v", err)
	} else {
		duration := time.Since(start)
		fmt.Printf("大文件Hash计算耗时: %v\n", duration)
	}
}

// 创建测试文件的辅助函数
func createTestFiles() {
	// 这个函数可以用来创建一些测试文件
	testData := []struct {
		filename string
		content  string
	}{
		{"file1.txt", "Hello World"},
		{"file2.txt", "Go Programming"},
		{"file3.txt", "Hash Utils Test"},
		{"file1_copy.txt", "Hello World"}, // 与file1.txt内容相同
		{"empty.txt", ""},
	}

	for _, test := range testData {
		// 创建文件的代码...
		fmt.Printf("需要创建测试文件: %s\n", test.filename)
	}
}
