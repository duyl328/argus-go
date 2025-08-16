package tables

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"rear/internal/model"
)

// PhotoExif 表结构 - 保存图片EXIF信息
// 使用Hash作为主键，直接对应exif_model.go中的结构
type PhotoExif struct {
	BaseModel
	// Hash作为主键，与Photo表的Hash对应，但不使用外键约束
	Hash string `gorm:"uniqueIndex;size:64" json:"hash"`

	// ========================= 基础文件信息 ==============================
	// 对应 BaseImageInfo 结构
	FileName      string `gorm:"column:file_name;size:255" json:"fileName"`
	FileSize      int64  `gorm:"column:file_size" json:"fileSize"`
	ImageWidth    int    `gorm:"column:image_width" json:"imageWidth"`
	ImageHeight   int    `gorm:"column:image_height" json:"imageHeight"`
	ImageSize     string `gorm:"column:image_size;size:50" json:"imageSize"`
	MIMEType      string `gorm:"column:mime_type;size:100" json:"mimeType"`
	FileType      string `gorm:"column:file_type;size:50" json:"fileType"`
	FileTypeExt   string `gorm:"column:file_type_ext;size:10" json:"fileTypeExt"`
	ModifyDate    string `gorm:"column:modify_date;size:50" json:"modifyDate"`
	CreateDate    string `gorm:"column:create_date;size:50" json:"createDate"`
	ColorSpace    string `gorm:"column:color_space;size:50" json:"colorSpace"`
	BitsPerSample int    `gorm:"column:bits_per_sample" json:"bitsPerSample"`
	Resolution    string `gorm:"column:resolution;size:50" json:"resolution"`
	Quality       int    `gorm:"column:quality" json:"quality"`

	// ========================= EXIF摄影信息 ==============================
	// 对应 ExifInfo 结构
	Model        string  `gorm:"column:model;size:100;index" json:"model"`
	Make         string  `gorm:"column:make;size:100;index" json:"make"`
	ISO          int     `gorm:"column:iso;index" json:"iso"`
	GPSLatitude  float64 `gorm:"column:gps_latitude;index" json:"gpsLatitude"`
	GPSLongitude float64 `gorm:"column:gps_longitude;index" json:"gpsLongitude"`
	ExposureTime float64 `gorm:"column:exposure_time" json:"exposureTime"`
	Aperture     float64 `gorm:"column:aperture" json:"aperture"`
	FNumber      float64 `gorm:"column:f_number" json:"fNumber"`
	FocalLength  float64 `gorm:"column:focal_length" json:"focalLength"`
	LensID       string  `gorm:"column:lens_id;size:200" json:"lensId"`
	Title        string  `gorm:"column:title;size:500" json:"title"`
	Description  string  `gorm:"column:description;type:text" json:"description"`
	DateTimeOrig string  `gorm:"column:datetime_original;size:50;index" json:"datetimeOriginal"`

	// ========================= 扩展信息 ==============================
	// 其他EXIF字段 (JSON格式存储)
	OtherFields JSONMap `gorm:"column:other_fields;type:text" json:"otherFields,omitempty"`
}

// JSONMap 用于存储其他EXIF字段的JSON数据
type JSONMap map[string]interface{}

// Scan 实现 sql.Scanner 接口，用于从数据库读取
func (j *JSONMap) Scan(value interface{}) error {
	if value == nil {
		*j = make(JSONMap)
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("cannot scan %T into JSONMap", value)
	}

	if len(bytes) == 0 {
		*j = make(JSONMap)
		return nil
	}

	return json.Unmarshal(bytes, j)
}

// Value 实现 driver.Valuer 接口，用于写入数据库
func (j JSONMap) Value() (driver.Value, error) {
	if j == nil {
		return nil, nil
	}
	return json.Marshal(j)
}

// TableName 指定表名
func (PhotoExif) TableName() string {
	return "photo_exif"
}

// FromParsedExif 从ParsedExif转换为PhotoExif表结构
func (pe *PhotoExif) FromParsedExif(hash string, parsedExif *model.ParsedExif) {
	if parsedExif == nil {
		return
	}

	pe.Hash = hash

	// 转换基础文件信息
	base := parsedExif.BaseInfo
	pe.FileName = base.FileName
	pe.FileSize = base.FileSize
	pe.ImageWidth = base.ImageWidth
	pe.ImageHeight = base.ImageHeight
	pe.ImageSize = base.ImageSize
	pe.MIMEType = base.MIMEType
	pe.FileType = base.FileType
	pe.FileTypeExt = base.FileTypeExt
	pe.ModifyDate = base.ModifyDate
	pe.CreateDate = base.CreateDate
	pe.ColorSpace = base.ColorSpace
	pe.BitsPerSample = base.BitsPerSample
	pe.Resolution = base.Resolution
	pe.Quality = base.Quality

	// 转换EXIF摄影信息
	exif := parsedExif.Exif
	pe.Model = exif.Model
	pe.Make = exif.Make
	pe.ISO = exif.ISO
	pe.GPSLatitude = exif.GPSLatitude
	pe.GPSLongitude = exif.GPSLongitude
	pe.ExposureTime = exif.ExposureTime
	pe.Aperture = exif.Aperture
	pe.FNumber = exif.FNumber
	pe.FocalLength = exif.FocalLength
	pe.LensID = exif.LensID
	pe.Title = exif.Title
	pe.Description = exif.Description
	pe.DateTimeOrig = exif.DateTimeOrig

	// 转换其他字段
	if len(parsedExif.OtherFields) > 0 {
		pe.OtherFields = JSONMap(parsedExif.OtherFields)
	}
}

// ToParsedExif 将PhotoExif转换为ParsedExif结构
func (pe *PhotoExif) ToParsedExif() *model.ParsedExif {
	// 创建基础信息
	baseInfo := model.BaseImageInfo{
		FileName:      pe.FileName,
		FileSize:      pe.FileSize,
		ImageWidth:    pe.ImageWidth,
		ImageHeight:   pe.ImageHeight,
		ImageSize:     pe.ImageSize,
		MIMEType:      pe.MIMEType,
		FileType:      pe.FileType,
		FileTypeExt:   pe.FileTypeExt,
		ModifyDate:    pe.ModifyDate,
		CreateDate:    pe.CreateDate,
		ColorSpace:    pe.ColorSpace,
		BitsPerSample: pe.BitsPerSample,
		Resolution:    pe.Resolution,
		Quality:       pe.Quality,
	}

	// 创建EXIF信息
	exifInfo := model.ExifInfo{
		Model:        pe.Model,
		Make:         pe.Make,
		ISO:          pe.ISO,
		GPSLatitude:  pe.GPSLatitude,
		GPSLongitude: pe.GPSLongitude,
		ExposureTime: pe.ExposureTime,
		Aperture:     pe.Aperture,
		FNumber:      pe.FNumber,
		FocalLength:  pe.FocalLength,
		LensID:       pe.LensID,
		Title:        pe.Title,
		Description:  pe.Description,
		DateTimeOrig: pe.DateTimeOrig,
	}

	// 转换其他字段
	otherFields := make(map[string]interface{})
	if pe.OtherFields != nil {
		otherFields = map[string]interface{}(pe.OtherFields)
	}

	return &model.ParsedExif{
		BaseInfo:    baseInfo,
		Exif:        exifInfo,
		OtherFields: otherFields,
	}
}

// NewPhotoExifFromParsed 从ParsedExif创建PhotoExif实例
func NewPhotoExifFromParsed(hash string, parsedExif *model.ParsedExif) *PhotoExif {
	photoExif := &PhotoExif{}
	photoExif.FromParsedExif(hash, parsedExif)
	return photoExif
}

// NewPhotoExifFromRawData 从原始EXIF数据创建PhotoExif实例
func NewPhotoExifFromRawData(hash string, rawData map[string]interface{}) *PhotoExif {
	parsedExif := model.SplitExifData(rawData)
	return NewPhotoExifFromParsed(hash, parsedExif)
}

// HasGPS 判断是否有GPS信息
func (pe *PhotoExif) HasGPS() bool {
	return pe.GPSLatitude != 0 || pe.GPSLongitude != 0
}

// HasCameraInfo 判断是否有相机信息
func (pe *PhotoExif) HasCameraInfo() bool {
	return pe.Make != "" || pe.Model != ""
}

// HasShootingParams 判断是否有拍摄参数
func (pe *PhotoExif) HasShootingParams() bool {
	return pe.ISO > 0 || pe.ExposureTime > 0 || pe.FNumber > 0 || pe.FocalLength > 0
}

// GetImageRatio 计算图片宽高比
func (pe *PhotoExif) GetImageRatio() float64 {
	if pe.ImageHeight == 0 {
		return 0
	}
	return float64(pe.ImageWidth) / float64(pe.ImageHeight)
}
