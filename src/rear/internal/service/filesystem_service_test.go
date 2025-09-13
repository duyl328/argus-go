package service

import (
	"testing"
)

// 这个测试文件主要用于展示FileSystemService的使用方法和参数传递
// 注意：这些测试更多是展示API使用方法，而不是严格的单元测试

// TestFileSystemService_Browse_Examples 展示Browse方法的各种使用场景
func TestFileSystemService_Browse_Examples(t *testing.T) {
	service := NewFileSystemService()

	// ===== 场景1: 获取根级别硬盘列表 =====
	t.Run("获取硬盘列表", func(t *testing.T) {
		// 不传参数或传空字符串，获取所有硬盘/挂载点
		result, err := service.Browse("")
		if err != nil {
			t.Logf("获取硬盘列表失败: %v", err)
			return
		}

		t.Logf("=== 硬盘列表 ===")
		t.Logf("当前路径: %s", result.CurrentPath)
		t.Logf("发现 %d 个硬盘/挂载点:", result.Summary.DriveCount)

		for _, item := range result.Items {
			if item.Type == ItemTypeDrive && item.DriveInfo != nil {
				t.Logf("  - %s (%s)", item.Name, item.Path)
				t.Logf("    文件系统: %s", item.DriveInfo.FileSystem)
				t.Logf("    总容量: %s", service.FormatSize(item.DriveInfo.TotalSpace))
				t.Logf("    可用容量: %s", service.FormatSize(item.DriveInfo.FreeSpace))
				t.Logf("    使用率: %.1f%%", item.DriveInfo.UsagePercent)
				t.Logf("    是否可移动: %v", item.DriveInfo.IsRemovable)
			}
		}
	})

	// ===== 场景2: Windows系统路径浏览 =====
	windowsPaths := []string{
		"C:\\",
		"C:\\Windows",
		"C:\\Users",
		"C:\\Program Files",
		"D:\\",
	}

	for _, path := range windowsPaths {
		t.Run("浏览Windows路径: "+path, func(t *testing.T) {
			result, err := service.Browse(path)
			if err != nil {
				t.Logf("浏览路径 %s 失败: %v", path, err)
				return
			}

			t.Logf("=== 浏览 %s ===", path)
			t.Logf("父路径: %s", result.ParentPath)
			t.Logf("包含项目: 目录%d个, 文件%d个",
				result.Summary.DirectoryCount, result.Summary.FileCount)

			// 显示前几个项目作为示例
			maxShow := 5
			if len(result.Items) < maxShow {
				maxShow = len(result.Items)
			}

			for i := 0; i < maxShow; i++ {
				item := result.Items[i]
				typeStr := string(item.Type)
				if item.Type == ItemTypeDirectory {
					typeStr = "目录"
				} else if item.Type == ItemTypeFile {
					typeStr = "文件"
				}

				t.Logf("  [%s] %s (%s)",
					typeStr, item.Name, service.FormatSize(item.Size))
			}

			if len(result.Items) > maxShow {
				t.Logf("  ... 还有 %d 个项目", len(result.Items)-maxShow)
			}
		})
	}

	// ===== 场景3: Unix系统路径浏览（如果在Unix系统上运行） =====
	unixPaths := []string{
		"/",
		"/home",
		"/usr",
		"/var",
		"/tmp",
	}

	for _, path := range unixPaths {
		t.Run("浏览Unix路径: "+path, func(t *testing.T) {
			result, err := service.Browse(path)
			if err != nil {
				t.Logf("浏览路径 %s 失败 (可能不在Unix系统上): %v", path, err)
				return
			}

			t.Logf("=== 浏览 %s ===", path)
			t.Logf("父路径: %s", result.ParentPath)
			t.Logf("包含项目: 目录%d个, 文件%d个",
				result.Summary.DirectoryCount, result.Summary.FileCount)
		})
	}

	// ===== 场景4: 项目目录浏览 =====
	projectPaths := []string{
		".",  // 当前目录
		"..", // 父目录
		"internal",
		"pkg",
		"tests",
	}

	for _, path := range projectPaths {
		t.Run("浏览项目路径: "+path, func(t *testing.T) {
			result, err := service.Browse(path)
			if err != nil {
				t.Logf("浏览项目路径 %s 失败: %v", path, err)
				return
			}

			t.Logf("=== 浏览项目目录 %s ===", path)
			t.Logf("完整路径: %s", result.CurrentPath)
			t.Logf("包含: 目录%d个, 文件%d个",
				result.Summary.DirectoryCount, result.Summary.FileCount)

			// 显示Go文件
			for _, item := range result.Items {
				if item.Type == ItemTypeFile && item.FileInfo != nil &&
				   item.FileInfo.Extension == ".go" {
					t.Logf("  [Go文件] %s (%s)",
						item.Name, service.FormatSize(item.Size))
				}
			}
		})
	}
}

// TestFileSystemService_GetDiskUsage_Examples 展示GetDiskUsage方法的使用
func TestFileSystemService_GetDiskUsage_Examples(t *testing.T) {
	service := NewFileSystemService()

	// Windows 驱动器测试
	windowsDrives := []string{"C:\\", "D:\\", "E:\\"}

	for _, drive := range windowsDrives {
		t.Run("Windows磁盘使用情况: "+drive, func(t *testing.T) {
			usage, err := service.GetDiskUsage(drive)
			if err != nil {
				t.Logf("获取 %s 磁盘使用情况失败: %v", drive, err)
				return
			}

			t.Logf("=== %s 磁盘使用情况 ===", drive)
			t.Logf("卷标: %s", usage.Label)
			t.Logf("文件系统: %s", usage.FileSystem)
			t.Logf("总空间: %s", service.FormatSize(usage.TotalSpace))
			t.Logf("已用空间: %s", service.FormatSize(usage.UsedSpace))
			t.Logf("可用空间: %s", service.FormatSize(usage.FreeSpace))
			t.Logf("使用率: %.1f%%", usage.UsagePercent)
			t.Logf("驱动器类型: %s", usage.DriveType)
			t.Logf("是否可移动: %v", usage.IsRemovable)
		})
	}

	// Unix 挂载点测试
	unixMounts := []string{"/", "/home", "/usr", "/var"}

	for _, mount := range unixMounts {
		t.Run("Unix磁盘使用情况: "+mount, func(t *testing.T) {
			usage, err := service.GetDiskUsage(mount)
			if err != nil {
				t.Logf("获取 %s 磁盘使用情况失败 (可能不在Unix系统上): %v", mount, err)
				return
			}

			t.Logf("=== %s 磁盘使用情况 ===", mount)
			t.Logf("文件系统: %s", usage.FileSystem)
			t.Logf("总空间: %s", service.FormatSize(usage.TotalSpace))
			t.Logf("使用率: %.1f%%", usage.UsagePercent)
		})
	}
}

// TestFileSystemService_ErrorHandling_Examples 展示错误处理场景
func TestFileSystemService_ErrorHandling_Examples(t *testing.T) {
	service := NewFileSystemService()

	// ===== 错误场景1: 不存在的路径 =====
	t.Run("访问不存在的路径", func(t *testing.T) {
		nonExistentPaths := []string{
			"X:\\NonExistent",  // Windows不存在的盘符
			"/nonexistent",     // Unix不存在的路径
			"./invalid/path",   // 当前目录下不存在的路径
		}

		for _, path := range nonExistentPaths {
			_, err := service.Browse(path)
			if err != nil {
				t.Logf("预期错误 - 路径 %s 不存在: %v", path, err)
			} else {
				t.Logf("意外 - 路径 %s 竟然存在", path)
			}
		}
	})

	// ===== 错误场景2: 访问文件而不是目录 =====
	t.Run("尝试浏览文件而不是目录", func(t *testing.T) {
		// 尝试浏览一些可能存在的文件
		possibleFiles := []string{
			"go.mod",
			"main.go",
			"config.yaml",
		}

		for _, file := range possibleFiles {
			_, err := service.Browse(file)
			if err != nil {
				t.Logf("预期错误 - %s 是文件不是目录: %v", file, err)
			}
		}
	})

	// ===== 错误场景3: 权限问题 =====
	t.Run("访问可能没有权限的目录", func(t *testing.T) {
		restrictedPaths := []string{
			"C:\\System Volume Information", // Windows受限目录
			"/root",                         // Unix根用户目录
		}

		for _, path := range restrictedPaths {
			_, err := service.Browse(path)
			if err != nil {
				t.Logf("预期错误 - 访问 %s 可能没有权限: %v", path, err)
			} else {
				t.Logf("成功访问 %s", path)
			}
		}
	})
}

// TestFileSystemService_SpecialCases_Examples 展示特殊情况处理
func TestFileSystemService_SpecialCases_Examples(t *testing.T) {
	service := NewFileSystemService()

	// ===== 特殊情况1: 空路径和根路径 =====
	t.Run("空路径和根路径处理", func(t *testing.T) {
		testCases := []struct {
			name string
			path string
		}{
			{"空字符串", ""},
			{"根路径", "/"},
			{"当前目录", "."},
			{"父目录", ".."},
		}

		for _, tc := range testCases {
			t.Run(tc.name, func(t *testing.T) {
				result, err := service.Browse(tc.path)
				if err != nil {
					t.Logf("%s (%s) 处理失败: %v", tc.name, tc.path, err)
					return
				}

				t.Logf("=== %s (%s) ===", tc.name, tc.path)
				t.Logf("解析后路径: %s", result.CurrentPath)
				t.Logf("项目数量: %d", result.Summary.TotalItems)
			})
		}
	})

	// ===== 特殊情况2: 包含特殊字符的路径 =====
	t.Run("特殊字符路径处理", func(t *testing.T) {
		specialPaths := []string{
			"C:\\Program Files",      // 包含空格
			"C:\\Program Files (x86)", // 包含空格和括号
		}

		for _, path := range specialPaths {
			result, err := service.Browse(path)
			if err != nil {
				t.Logf("特殊路径 %s 处理失败: %v", path, err)
				continue
			}

			t.Logf("=== 特殊路径 %s ===", path)
			t.Logf("成功解析，包含 %d 个项目", result.Summary.TotalItems)
		}
	})
}

// TestFileSystemService_FormatSize_Examples 展示文件大小格式化
func TestFileSystemService_FormatSize_Examples(t *testing.T) {
	service := NewFileSystemService()

	testSizes := []int64{
		0,                    // 0 B
		512,                  // 512 B
		1024,                 // 1.0 KB
		1536,                 // 1.5 KB
		1024 * 1024,          // 1.0 MB
		1024 * 1024 * 1024,   // 1.0 GB
		1024 * 1024 * 1024 * 1024, // 1.0 TB
	}

	t.Run("文件大小格式化示例", func(t *testing.T) {
		t.Logf("=== 文件大小格式化示例 ===")
		for _, size := range testSizes {
			formatted := service.FormatSize(size)
			t.Logf("%d 字节 -> %s", size, formatted)
		}
	})
}

// BenchmarkFileSystemService_Browse 性能基准测试
func BenchmarkFileSystemService_Browse(b *testing.B) {
	service := NewFileSystemService()

	b.Run("获取硬盘列表", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = service.Browse("")
		}
	})

	b.Run("浏览当前目录", func(b *testing.B) {
		for i := 0; i < b.N; i++ {
			_, _ = service.Browse(".")
		}
	})
}

// ExampleFileSystemService_Browse 示例代码
func ExampleFileSystemService_Browse() {
	service := NewFileSystemService()

	// 获取所有硬盘列表
	drives, err := service.Browse("")
	if err != nil {
		panic(err)
	}

	// 打印硬盘信息
	for _, item := range drives.Items {
		if item.Type == ItemTypeDrive {
			// 输出示例: C: (本地磁盘) - 500.0 GB 总计, 100.0 GB 可用
		}
	}

	// 浏览C盘内容
	contents, err := service.Browse("C:\\")
	if err != nil {
		panic(err)
	}

	// 打印目录和文件
	for _, item := range contents.Items {
		switch item.Type {
		case ItemTypeDirectory:
			// 输出示例: [目录] Windows
		case ItemTypeFile:
			// 输出示例: [文件] config.txt (1.2 KB)
		}
	}
}

// ExampleFileSystemService_GetDiskUsage 磁盘使用情况示例
func ExampleFileSystemService_GetDiskUsage() {
	service := NewFileSystemService()

	// 获取C盘使用情况
	usage, err := service.GetDiskUsage("C:\\")
	if err != nil {
		panic(err)
	}

	// 输出磁盘信息
	// 卷标: Windows
	// 文件系统: NTFS
	// 总空间: 500.0 GB
	// 已用空间: 400.0 GB
	// 可用空间: 100.0 GB
	// 使用率: 80.0%
	_ = usage
}