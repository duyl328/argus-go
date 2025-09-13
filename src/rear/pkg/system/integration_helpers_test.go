package system

import (
	"os"
	"strings"
	"testing"
)

func TestStorageAnalyzerCreation(t *testing.T) {
	t.Log("Testing StorageAnalyzer creation...")

	analyzer := NewStorageAnalyzer()
	if analyzer == nil {
		t.Fatal("NewStorageAnalyzer returned nil")
	}

	if analyzer.deviceManager == nil {
		t.Error("DeviceManager should be initialized")
	}

	t.Log("StorageAnalyzer created successfully")
}

func TestAnalyzeDirectoryStorage(t *testing.T) {
	t.Log("Testing AnalyzeDirectoryStorage functionality...")

	analyzer := NewStorageAnalyzer()

	// 测试当前目录
	dirInfo, err := analyzer.AnalyzeDirectoryStorage(".")
	if err != nil {
		t.Fatalf("AnalyzeDirectoryStorage failed: %v", err)
	}

	if dirInfo == nil {
		t.Fatal("DirectoryStorageInfo should not be nil")
	}

	// 验证基础信息
	if dirInfo.Path == "" {
		t.Error("Path should not be empty")
	}

	if dirInfo.Size < 0 {
		t.Error("Size should not be negative")
	}

	if dirInfo.FileCount < 0 {
		t.Error("FileCount should not be negative")
	}

	if dirInfo.SizePercent < 0 || dirInfo.SizePercent > 100 {
		t.Errorf("SizePercent should be between 0-100, got %.2f", dirInfo.SizePercent)
	}

	t.Logf("Directory analysis results:")
	t.Logf("  Path: %s", dirInfo.Path)
	t.Logf("  Size: %d bytes", dirInfo.Size)
	t.Logf("  File Count: %d", dirInfo.FileCount)
	t.Logf("  Size Percent: %.2f%%", dirInfo.SizePercent)
	if dirInfo.Device != nil {
		t.Logf("  Device: %s (%s)", dirInfo.Device.Name, dirInfo.Device.MountPoint)
	}

	// 测试不存在的目录
	_, err = analyzer.AnalyzeDirectoryStorage("/nonexistent/directory")
	if err == nil {
		t.Error("AnalyzeDirectoryStorage should fail for nonexistent directory")
	}

	// 测试文件而不是目录（如果存在的话）
	if fileExists("go.mod") {
		_, err = analyzer.AnalyzeDirectoryStorage("go.mod")
		if err == nil {
			t.Error("AnalyzeDirectoryStorage should fail for files")
		}
	}
}

func TestRecommendStorageLocation(t *testing.T) {
	t.Log("Testing RecommendStorageLocation functionality...")

	analyzer := NewStorageAnalyzer()

	// 测试小文件的推荐
	requiredSpace := int64(1024 * 1024) // 1MB
	recommendation, err := analyzer.RecommendStorageLocation(requiredSpace)
	if err != nil {
		t.Fatalf("RecommendStorageLocation failed: %v", err)
	}

	if recommendation == nil {
		t.Fatal("StorageRecommendation should not be nil")
	}

	// 验证推荐结果
	if recommendation.RequiredSpace != requiredSpace {
		t.Errorf("Expected required space %d, got %d", requiredSpace, recommendation.RequiredSpace)
	}

	if recommendation.RequiredSpaceStr == "" {
		t.Error("RequiredSpaceStr should not be empty")
	}

	if recommendation.BestChoice.Device == nil {
		t.Error("Best choice device should not be nil")
	}

	if recommendation.BestChoice.Score <= 0 {
		t.Error("Best choice score should be positive")
	}

	if len(recommendation.AllOptions) == 0 {
		t.Error("Should have at least one recommendation option")
	}

	t.Logf("Storage recommendation results:")
	t.Logf("  Required Space: %s", recommendation.RequiredSpaceStr)
	t.Logf("  Best Choice: %s (Score: %.1f)",
		recommendation.BestChoice.Device.Name, recommendation.BestChoice.Score)
	t.Logf("  Reason: %s", recommendation.BestChoice.Reason)
	t.Logf("  Available Options: %d", len(recommendation.AllOptions))

	// 测试格式化输出
	formatted := recommendation.FormatStorageRecommendation()
	if formatted == "" {
		t.Error("Formatted recommendation should not be empty")
	}

	expectedContent := []string{
		"存储推荐",
		"最佳选择",
		recommendation.BestChoice.Device.Name,
	}

	for _, content := range expectedContent {
		if !strings.Contains(formatted, content) {
			t.Errorf("Formatted output should contain '%s'", content)
		}
	}

	// 测试非常大的空间需求
	hugeSpace := int64(1024 * 1024 * 1024 * 1024) // 1TB
	_, err = analyzer.RecommendStorageLocation(hugeSpace)
	if err != nil {
		t.Logf("Large space recommendation failed (expected): %v", err)
	}
}

func TestCleanupStorage(t *testing.T) {
	t.Log("Testing CleanupStorage functionality...")

	analyzer := NewStorageAnalyzer()

	// 首先扫描设备以获取可用的设备路径
	err := analyzer.deviceManager.ScanDevices()
	if err != nil {
		t.Fatalf("Device scan failed: %v", err)
	}

	devices := analyzer.deviceManager.GetDevices()
	if len(devices) == 0 {
		t.Skip("No devices found for cleanup testing")
	}

	// 测试第一个设备的清理建议
	firstDevice := devices[0]
	cleanupInfo, err := analyzer.CleanupStorage(firstDevice.MountPoint)
	if err != nil {
		t.Fatalf("CleanupStorage failed: %v", err)
	}

	if cleanupInfo == nil {
		t.Fatal("StorageCleanupInfo should not be nil")
	}

	// 验证清理信息
	if cleanupInfo.Device == nil {
		t.Error("Device should not be nil")
	}

	if len(cleanupInfo.Suggestions) == 0 {
		t.Error("Should have at least one cleanup suggestion")
	}

	if cleanupInfo.CanFreeSpace < 0 {
		t.Error("CanFreeSpace should not be negative")
	}

	t.Logf("Cleanup suggestions for %s:", firstDevice.Name)
	for i, suggestion := range cleanupInfo.Suggestions {
		t.Logf("  %d. %s", i+1, suggestion)
	}
	t.Logf("  Estimated free space: %d bytes", cleanupInfo.CanFreeSpace)

	// 测试不存在的设备路径
	_, err = analyzer.CleanupStorage("/nonexistent/device")
	if err == nil {
		t.Error("CleanupStorage should fail for nonexistent device")
	}
}

func TestFindDeviceForPath(t *testing.T) {
	t.Log("Testing findDeviceForPath functionality...")

	analyzer := NewStorageAnalyzer()

	// 扫描设备
	err := analyzer.deviceManager.ScanDevices()
	if err != nil {
		t.Fatalf("Device scan failed: %v", err)
	}

	// 测试当前目录
	currentDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get current directory: %v", err)
	}

	device := analyzer.findDeviceForPath(currentDir)
	if device != nil {
		t.Logf("Current directory %s is on device: %s (%s)",
			currentDir, device.Name, device.MountPoint)
	} else {
		t.Log("No device found for current directory")
	}

	// 测试根路径 (Windows: C:\, Unix: /)
	var rootPath string
	if strings.Contains(currentDir, "\\") {
		rootPath = "C:\\"
	} else {
		rootPath = "/"
	}

	rootDevice := analyzer.findDeviceForPath(rootPath)
	if rootDevice != nil {
		t.Logf("Root path %s is on device: %s (%s)",
			rootPath, rootDevice.Name, rootDevice.MountPoint)
	} else {
		t.Logf("No device found for root path %s", rootPath)
	}
}

func TestCalculateStorageScore(t *testing.T) {
	t.Log("Testing calculateStorageScore functionality...")

	// 创建测试设备
	testDevice := &StorageDevice{
		Name:       "TestDevice",
		Type:       DeviceTypeFixed,
		FileSystem: "NTFS",
		TotalSpace: 1024 * 1024 * 1024 * 1024, // 1TB
		FreeSpace:  512 * 1024 * 1024 * 1024,  // 512GB
		UsedSpace:  512 * 1024 * 1024 * 1024,  // 512GB
	}

	requiredSpace := int64(1024 * 1024 * 1024) // 1GB

	score := calculateStorageScore(testDevice, requiredSpace)
	if score <= 0 {
		t.Error("Storage score should be positive")
	}

	if score > 100 {
		t.Error("Storage score should not exceed 100")
	}

	t.Logf("Storage score for test device: %.1f/100", score)

	// 测试不同设备类型的评分
	deviceTypes := []DeviceType{
		DeviceTypeFixed,
		DeviceTypeRemovable,
		DeviceTypeNetwork,
		DeviceTypeRAM,
	}

	for _, deviceType := range deviceTypes {
		testDevice.Type = deviceType
		score := calculateStorageScore(testDevice, requiredSpace)
		t.Logf("Score for %s device: %.1f", deviceType, score)
	}
}

func TestGenerateRecommendationReason(t *testing.T) {
	t.Log("Testing generateRecommendationReason functionality...")

	testDevice := &StorageDevice{
		Name:       "TestDevice",
		Type:       DeviceTypeFixed,
		TotalSpace: 1024 * 1024 * 1024 * 1024, // 1TB
		FreeSpace:  512 * 1024 * 1024 * 1024,  // 512GB
		UsedSpace:  512 * 1024 * 1024 * 1024,  // 512GB
	}

	requiredSpace := int64(1024 * 1024 * 1024) // 1GB

	reason := generateRecommendationReason(testDevice, requiredSpace)
	if reason == "" {
		t.Error("Recommendation reason should not be empty")
	}

	t.Logf("Recommendation reason: %s", reason)

	// 测试不同使用率的设备
	usageRates := []struct {
		used  int64
		total int64
		desc  string
	}{
		{1024 * 1024 * 1024 * 100, 1024 * 1024 * 1024 * 1024, "10% usage"},  // 10%
		{1024 * 1024 * 1024 * 500, 1024 * 1024 * 1024 * 1024, "50% usage"},  // 50%
		{1024 * 1024 * 1024 * 900, 1024 * 1024 * 1024 * 1024, "90% usage"},  // 90%
	}

	for _, usage := range usageRates {
		testDevice.UsedSpace = usage.used
		testDevice.TotalSpace = usage.total
		testDevice.FreeSpace = usage.total - usage.used

		reason := generateRecommendationReason(testDevice, requiredSpace)
		t.Logf("Reason for %s: %s", usage.desc, reason)
	}
}

func TestGlobalStorageFunctions(t *testing.T) {
	t.Log("Testing global storage functions...")

	// 测试全局分析函数
	dirInfo, err := AnalyzeDirectory(".")
	if err != nil {
		t.Fatalf("AnalyzeDirectory failed: %v", err)
	}

	if dirInfo == nil {
		t.Error("DirectoryStorageInfo should not be nil")
	}

	t.Logf("Global AnalyzeDirectory result: %d files, %d bytes",
		dirInfo.FileCount, dirInfo.Size)

	// 测试全局推荐函数
	recommendation, err := RecommendStorage(1024 * 1024) // 1MB
	if err != nil {
		t.Fatalf("RecommendStorage failed: %v", err)
	}

	if recommendation == nil {
		t.Error("StorageRecommendation should not be nil")
	}

	t.Logf("Global RecommendStorage result: %s",
		recommendation.BestChoice.Device.Name)

	// 测试全局清理函数
	if len(recommendation.AllOptions) > 0 {
		devicePath := recommendation.AllOptions[0].Device.MountPoint
		cleanupInfo, err := SuggestCleanup(devicePath)
		if err != nil {
			t.Logf("SuggestCleanup failed: %v", err)
		} else {
			t.Logf("Global SuggestCleanup result: %d suggestions",
				len(cleanupInfo.Suggestions))
		}
	}
}

func TestIntegrationHelpersEdgeCases(t *testing.T) {
	t.Log("Testing IntegrationHelpers edge cases...")

	analyzer := NewStorageAnalyzer()

	// 测试空路径
	_, err := analyzer.AnalyzeDirectoryStorage("")
	if err == nil {
		t.Error("AnalyzeDirectoryStorage should fail for empty path")
	}

	// 测试零空间需求
	_, err = analyzer.RecommendStorageLocation(0)
	if err != nil {
		t.Logf("Zero space recommendation failed (expected): %v", err)
	}

	// 测试负数空间需求
	_, err = analyzer.RecommendStorageLocation(-1024)
	if err != nil {
		t.Logf("Negative space recommendation failed (expected): %v", err)
	}

	// 测试空设备路径
	_, err = analyzer.CleanupStorage("")
	if err == nil {
		t.Error("CleanupStorage should fail for empty device path")
	}

	t.Log("Edge case testing completed")
}

// 辅助函数
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	return err == nil
}