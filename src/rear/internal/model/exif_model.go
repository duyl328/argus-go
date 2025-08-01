package model

import (
	"fmt"
	"strconv"
)

/*
基础文件信息（文件级元数据）
这些信息即使图片没有 EXIF、XMP、IPTC 等元数据，也会存在，来自文件本身：
| 字段名                                            | 示例                          | 说明              |
| ---------------------------------------------- | --------------------------- | --------------- |
| `FileName`                                     | `image.jpg`                 | 文件名             |
| `FileSize`                                     | `2.3 MB`                    | 文件大小            |
| `FileModifyDate`                               | `2024:11:04 10:40:29+08:00` | 最后修改时间（文件系统时间戳） |
| `FileCreateDate`                               | `2025:06:29 20:23:38+08:00` | 创建时间（视系统而定）     |
| `FileAccessDate`                               | `2025:06:29 20:23:42+08:00` | 访问时间            |
| `Directory`                                    | `D:/photos/`                | 所在路径            |
| `MIMEType`                                     | `image/jpeg`                | 文件类型            |
| `FileType` / `FileTypeExtension`               | `JPEG` / `JPG`              | 扩展名和类型          |
| `ImageSize`                                    | `5184x3888`                 | 像素尺寸            |
| `Megapixels`                                   | `20.1`                      | 计算得出            |
| `EncodingProcess`                              | `Baseline DCT`              | JPEG 编码方式       |
| `ColorComponents`                              | `3`                         | 通道数（RGB 通常为 3）  |
| `YResolution`, `XResolution`, `ResolutionUnit` | `180`, `180`, `inch`        | DPI 设置          |
*/

type BaseImageInfo struct {
	FileName    string `json:"FileName"`
	FileSize    int64  `json:"FileSize"`
	ImageWidth  int    `json:"ImageWidth"`
	ImageHeight int    `json:"ImageHeight"`
	ImageSize   string `json:"ImageSize"`
	// 文件类型
	MIMEType string `json:"MIMEType"`
	// 文件扩展名
	FileType string `json:"FileType"`
	// 文件扩展名
	FileTypeExt string `json:"FileTypeExtension"`
	// 文件最后修改时间
	ModifyDate string `json:"FileModifyDate"`
	// 文件创建时间
	CreateDate string `json:"FileCreateDate"`

	ColorSpace    string `json:"ColorSpace"`    // 色彩空间 (sRGB, Adobe RGB等)
	BitsPerSample int    `json:"BitsPerSample"` // 每个颜色通道的位深度
	Resolution    string `json:"Resolution"`    // 分辨率信息

	Quality int `json:"Quality"` // JPEG质量
}

type ExifInfo struct {
	Model        string  `json:"Model"`
	Make         string  `json:"Make"`
	ISO          int     `json:"ISO"`
	GPSLatitude  float64 `json:"GPSLatitude"`
	GPSLongitude float64 `json:"GPSLongitude"`
	ExposureTime float64 `json:"ExposureTime"`
	Aperture     float64 `json:"Aperture"`
	FNumber      float64 `json:"FNumber"` // 光圈值
	FocalLength  float64 `json:"FocalLength"`
	LensID       string  `json:"LensID"`
	Title        string  `json:"Title"`
	Description  string  `json:"Description"`
	DateTimeOrig string  `json:"DateTimeOriginal"`
}

type ParsedExif struct {
	BaseInfo    BaseImageInfo
	Exif        ExifInfo
	OtherFields map[string]interface{}
}

// 安全类型转换函数
func safeIntConvert(v interface{}) int {
	switch val := v.(type) {
	case int:
		return val
	case int64:
		return int(val)
	case float64:
		return int(val)
	case string:
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return 0
}

func safeInt64Convert(v interface{}) int64 {
	switch val := v.(type) {
	case int:
		return int64(val)
	case int64:
		return val
	case float64:
		return int64(val)
	case string:
		if i, err := strconv.ParseInt(val, 10, 64); err == nil {
			return i
		}
	}
	return 0
}

func safeFloat64Convert(v interface{}) float64 {
	switch val := v.(type) {
	case float64:
		return val
	case float32:
		return float64(val)
	case int:
		return float64(val)
	case int64:
		return float64(val)
	case string:
		if f, err := strconv.ParseFloat(val, 64); err == nil {
			return f
		}
	}
	return 0.0
}

func safeStringConvert(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case int, int64, float64, float32:
		return fmt.Sprintf("%v", val)
	}
	return ""
}

func SplitExifData(data map[string]interface{}) *ParsedExif {
	var base BaseImageInfo
	var exif ExifInfo

	// 手动提取BaseImageInfo字段
	if v, ok := data["FileName"]; ok {
		base.FileName = safeStringConvert(v)
	}
	if v, ok := data["FileSize"]; ok {
		base.FileSize = safeInt64Convert(v)
	}
	if v, ok := data["ImageWidth"]; ok {
		base.ImageWidth = safeIntConvert(v)
	}
	if v, ok := data["ImageHeight"]; ok {
		base.ImageHeight = safeIntConvert(v)
	}
	if v, ok := data["ImageSize"]; ok {
		base.ImageSize = safeStringConvert(v)
	}
	if v, ok := data["MIMEType"]; ok {
		base.MIMEType = safeStringConvert(v)
	}
	if v, ok := data["FileType"]; ok {
		base.FileType = safeStringConvert(v)
	}
	if v, ok := data["FileTypeExtension"]; ok {
		base.FileTypeExt = safeStringConvert(v)
	}
	if v, ok := data["FileModifyDate"]; ok {
		base.ModifyDate = safeStringConvert(v)
	}
	if v, ok := data["FileCreateDate"]; ok {
		base.CreateDate = safeStringConvert(v)
	}
	if v, ok := data["ColorSpace"]; ok {
		base.ColorSpace = safeStringConvert(v)
	}
	if v, ok := data["BitsPerSample"]; ok {
		base.BitsPerSample = safeIntConvert(v)
	}
	// 尝试从XResolution构建分辨率信息
	if v, ok := data["XResolution"]; ok {
		base.Resolution = safeStringConvert(v)
	}
	if v, ok := data["JPEGQuality"]; ok {
		base.Quality = safeIntConvert(v)
	}

	// 手动提取ExifInfo字段
	if v, ok := data["Model"]; ok {
		exif.Model = safeStringConvert(v)
	}
	if v, ok := data["Make"]; ok {
		exif.Make = safeStringConvert(v)
	}
	if v, ok := data["ISO"]; ok {
		exif.ISO = safeIntConvert(v)
	}
	if v, ok := data["GPSLatitude"]; ok {
		exif.GPSLatitude = safeFloat64Convert(v)
	}
	if v, ok := data["GPSLongitude"]; ok {
		exif.GPSLongitude = safeFloat64Convert(v)
	}
	if v, ok := data["ExposureTime"]; ok {
		exif.ExposureTime = safeFloat64Convert(v)
	}
	if v, ok := data["Aperture"]; ok {
		exif.Aperture = safeFloat64Convert(v)
	}
	if v, ok := data["FNumber"]; ok {
		exif.FNumber = safeFloat64Convert(v)
	}
	if v, ok := data["FocalLength"]; ok {
		exif.FocalLength = safeFloat64Convert(v)
	}
	if v, ok := data["LensID"]; ok {
		exif.LensID = safeStringConvert(v)
	}
	if v, ok := data["Title"]; ok {
		exif.Title = safeStringConvert(v)
	}
	if v, ok := data["Description"]; ok {
		exif.Description = safeStringConvert(v)
	}
	if v, ok := data["DateTimeOriginal"]; ok {
		exif.DateTimeOrig = safeStringConvert(v)
	}

	// 定义已处理的字段
	processedFields := map[string]bool{
		"FileName": true, "FileSize": true, "ImageWidth": true, "ImageHeight": true,
		"ImageSize": true, "MIMEType": true, "FileType": true, "FileTypeExtension": true,
		"FileModifyDate": true, "FileCreateDate": true, "ColorSpace": true, "BitsPerSample": true,
		"XResolution": true, "JPEGQuality": true,
		"Model": true, "Make": true, "ISO": true, "GPSLatitude": true, "GPSLongitude": true,
		"ExposureTime": true, "Aperture": true, "FNumber": true, "FocalLength": true,
		"LensID": true, "Title": true, "Description": true, "DateTimeOriginal": true,
	}

	// 提取剩余字段
	other := make(map[string]interface{})
	for k, v := range data {
		if !processedFields[k] {
			other[k] = v
		}
	}

	return &ParsedExif{
		BaseInfo:    base,
		Exif:        exif,
		OtherFields: other,
	}
}

// 辅助函数：判断字符串是否在切片中
func isIn(str string, slice ...string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
