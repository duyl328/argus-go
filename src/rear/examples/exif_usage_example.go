package main

import (
	"fmt"
	"rear/internal/model"
	"rear/internal/model/tables"
)

// ExifUsageExample 展示如何使用EXIF表结构
func ExifUsageExample() {
	fmt.Println("=== EXIF 表使用示例 ===")

	// 1. 模拟从ExifTool获取的原始数据
	rawExifData := map[string]interface{}{
		"FileName":          "IMG_1234.jpg",
		"FileSize":          2048000,
		"ImageWidth":        3840,
		"ImageHeight":       2160,
		"ImageSize":         "3840x2160",
		"MIMEType":          "image/jpeg",
		"FileType":          "JPEG",
		"FileTypeExtension": "jpg",
		"FileModifyDate":    "2024:01:15 14:30:25",
		"FileCreateDate":    "2024:01:15 14:30:25",
		"ColorSpace":        "sRGB",
		"BitsPerSample":     8,
		"XResolution":       "300",
		"JPEGQuality":       95,
		"Make":              "Canon",
		"Model":             "EOS R5",
		"ISO":               400,
		"GPSLatitude":       39.9042,
		"GPSLongitude":      116.4074,
		"ExposureTime":      0.008,  // 1/125
		"Aperture":          5.6,
		"FNumber":           5.6,
		"FocalLength":       85.0,
		"LensID":            "Canon RF 85mm f/1.2L USM",
		"Title":             "Beautiful Sunset",
		"Description":       "A stunning sunset captured at Beijing",
		"DateTimeOriginal":  "2024:01:15 14:30:25",
		"CustomField1":      "Some custom data",
		"CustomField2":      123,
	}

	photoHash := "abc123def456hash789"

	// 2. 方法一：从原始数据直接创建PhotoExif
	fmt.Println("\n方法一：从原始EXIF数据创建PhotoExif")
	photoExif1 := tables.NewPhotoExifFromRawData(photoHash, rawExifData)
	fmt.Printf("创建的PhotoExif - 相机: %s %s, ISO: %d\n", 
		photoExif1.Make, photoExif1.Model, photoExif1.ISO)

	// 3. 方法二：先解析为ParsedExif，再转换为PhotoExif
	fmt.Println("\n方法二：通过ParsedExif转换")
	parsedExif := model.SplitExifData(rawExifData)
	photoExif2 := tables.NewPhotoExifFromParsed(photoHash, parsedExif)
	
	// 4. 展示PhotoExif的便捷方法
	fmt.Println("\n便捷方法示例:")
	fmt.Printf("是否有GPS信息: %t\n", photoExif2.HasGPS())
	fmt.Printf("是否有相机信息: %t\n", photoExif2.HasCameraInfo())
	fmt.Printf("是否有拍摄参数: %t\n", photoExif2.HasShootingParams())
	fmt.Printf("图片宽高比: %.2f\n", photoExif2.GetImageRatio())

	// 5. 转换回ParsedExif用于API返回
	fmt.Println("\n转换回ParsedExif:")
	convertedBack := photoExif2.ToParsedExif()
	fmt.Printf("转换后的相机信息: %s %s\n", 
		convertedBack.Exif.Make, convertedBack.Exif.Model)
	fmt.Printf("其他字段数量: %d\n", len(convertedBack.OtherFields))

	// 6. 展示如何使用Repository (伪代码)
	fmt.Println("\n数据库操作示例 (伪代码):")
	fmt.Printf(`
	// 保存到数据库
	exifRepo := repositories.NewExifRepository()
	err := exifRepo.CreateOrUpdate(photoExif2)
	
	// 从数据库查询
	savedExif, err := exifRepo.GetByHash("%s")
	
	// 查询特定相机的照片
	canonPhotos, err := exifRepo.GetExifsByCamera("Canon", "EOS R5", 10, 0)
	
	// 查询有GPS的照片
	gpsPhotos, err := exifRepo.GetExifsWithGPS(10, 0)
	
	// 根据ISO范围查询
	highISOPhotos, err := exifRepo.GetExifsByISO(800, 6400, 10, 0)
	`, photoHash)

	fmt.Println("\n=== 示例完成 ===")
}

// 展示实际使用场景
func RealisticsUsageScenario() {
	fmt.Println("\n=== 实际使用场景 ===")
	
	// 假设我们正在处理一批照片
	photos := []map[string]interface{}{
		{
			"hash": "photo1_hash",
			"exif": map[string]interface{}{
				"Make":     "Canon",
				"Model":    "EOS R6",
				"ISO":      800,
				"FNumber":  2.8,
				"FileName": "portrait.jpg",
			},
		},
		{
			"hash": "photo2_hash", 
			"exif": map[string]interface{}{
				"Make":         "Sony",
				"Model":        "Alpha 7R V",
				"ISO":          100,
				"FNumber":      8.0,
				"GPSLatitude":  40.7128,
				"GPSLongitude": -74.0060,
				"FileName":     "landscape.jpg",
			},
		},
	}

	var photoExifs []*tables.PhotoExif
	
	for _, photo := range photos {
		hash := photo["hash"].(string)
		exifData := photo["exif"].(map[string]interface{})
		
		// 创建PhotoExif
		photoExif := tables.NewPhotoExifFromRawData(hash, exifData)
		photoExifs = append(photoExifs, photoExif)
		
		fmt.Printf("处理照片: %s, 相机: %s %s\n", 
			photoExif.FileName, photoExif.Make, photoExif.Model)
	}
	
	fmt.Printf("批量处理完成，共处理 %d 张照片\n", len(photoExifs))
	
	// 模拟批量保存到数据库
	fmt.Println("准备批量保存到数据库...")
	fmt.Println("// exifRepo.BatchCreate(photoExifs)")
}

func main() {
	ExifUsageExample()
	RealisticsUsageScenario()
}