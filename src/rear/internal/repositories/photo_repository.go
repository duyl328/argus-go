package repositories

import (
	"context"
	"errors"
	"math"
	"os"
	"path/filepath"
	"rear/internal/db"
	"rear/internal/model"
	"rear/internal/model/tables"
	"strings"
	"time"

	"gorm.io/gorm"
)

// PhotoRepository 照片数据仓库
type PhotoRepository struct {
	db *gorm.DB
}

// NewPhotoRepository 创建照片仓库实例
func NewPhotoRepository() *PhotoRepository {
	return &PhotoRepository{
		db: db.DB,
	}
}

// Create 创建照片记录（同步）
func (r *PhotoRepository) Create(photo *tables.Photo) error {
	return r.db.Create(photo).Error
}

// CreateAsync 异步创建照片记录
func (r *PhotoRepository) CreateAsync(ctx context.Context, photo *tables.Photo) error {
	task := &db.PhotoCreateTask{Photo: photo}
	callback := db.TaskCallback{}
	return db.GetManger().SubmitWriteTask(task, callback)
}

// CreateAsyncSync 同步等待异步创建照片记录
func (r *PhotoRepository) CreateAsyncSync(ctx context.Context, photo *tables.Photo) error {
	task := &db.PhotoCreateTask{Photo: photo}
	return db.GetManger().SubmitWriteTaskSync(ctx, task)
}

// GetByHash 根据Hash获取照片信息
func (r *PhotoRepository) GetByHash(hash string) (*tables.Photo, error) {
	var photo tables.Photo
	err := r.db.Where("hash = ?", hash).First(&photo).Error
	if err != nil {
		return nil, err
	}
	return &photo, nil
}

// Update 更新照片记录（同步）
func (r *PhotoRepository) Update(photo *tables.Photo) error {
	return r.db.Save(photo).Error
}

// UpdateAsync 异步更新照片记录
func (r *PhotoRepository) UpdateAsync(ctx context.Context, photo *tables.Photo) error {
	task := &db.PhotoUpdateTask{Photo: photo}
	callback := db.TaskCallback{}
	return db.GetManger().SubmitWriteTask(task, callback)
}

// UpdateAsyncSync 同步等待异步更新照片记录
func (r *PhotoRepository) UpdateAsyncSync(ctx context.Context, photo *tables.Photo) error {
	task := &db.PhotoUpdateTask{Photo: photo}
	return db.GetManger().SubmitWriteTaskSync(ctx, task)
}

// Delete 删除照片记录（同步）
func (r *PhotoRepository) Delete(hash string) error {
	return r.db.Where("hash = ?", hash).Delete(&tables.Photo{}).Error
}

// DeleteAsync 异步删除照片记录
func (r *PhotoRepository) DeleteAsync(ctx context.Context, hash string) error {
	task := &db.PhotoDeleteTask{Hash: hash}
	callback := db.TaskCallback{}
	return db.GetManger().SubmitWriteTask(task, callback)
}

// DeleteAsyncSync 同步等待异步删除照片记录
func (r *PhotoRepository) DeleteAsyncSync(ctx context.Context, hash string) error {
	task := &db.PhotoDeleteTask{Hash: hash}
	return db.GetManger().SubmitWriteTaskSync(ctx, task)
}

// Exists 检查照片记录是否存在
func (r *PhotoRepository) Exists(hash string) (bool, error) {
	var count int64
	err := r.db.Model(&tables.Photo{}).Where("hash = ?", hash).Count(&count).Error
	return count > 0, err
}

// CreateOrUpdate 创建或更新照片记录（同步）
func (r *PhotoRepository) CreateOrUpdate(photo *tables.Photo) error {
	// 先检查是否存在
	exists, err := r.Exists(photo.Hash)
	if err != nil {
		return err
	}

	if exists {
		return r.Update(photo)
	} else {
		return r.Create(photo)
	}
}

// CreateOrUpdateAsync 异步创建或更新照片记录
func (r *PhotoRepository) CreateOrUpdateAsync(ctx context.Context, photo *tables.Photo) error {
	// 先检查是否存在
	exists, err := r.Exists(photo.Hash)
	if err != nil {
		return err
	}

	if exists {
		return r.UpdateAsync(ctx, photo)
	} else {
		return r.CreateAsync(ctx, photo)
	}
}

// GetAll 获取所有照片记录
func (r *PhotoRepository) GetAll(limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Limit(limit).Offset(offset).Find(&photos).Error
	return photos, err
}

// GetByPath 根据路径获取照片信息
func (r *PhotoRepository) GetByPath(imgPath string) (*tables.Photo, error) {
	var photo tables.Photo
	err := r.db.Where("img_path = ?", imgPath).First(&photo).Error
	if err != nil {
		return nil, err
	}
	return &photo, nil
}

// BatchCreate 批量创建照片记录
func (r *PhotoRepository) BatchCreate(photos []*tables.Photo) error {
	if len(photos) == 0 {
		return nil
	}
	return r.db.CreateInBatches(photos, 100).Error
}

// GetCount 获取总记录数
func (r *PhotoRepository) GetCount() (int64, error) {
	var count int64
	err := r.db.Model(&tables.Photo{}).Count(&count).Error
	return count, err
}

// GetPhotosByFormat 根据格式获取照片
func (r *PhotoRepository) GetPhotosByFormat(format string, limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("format = ?", format).
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// GetPhotosBySize 根据文件大小范围获取照片
func (r *PhotoRepository) GetPhotosBySize(minSize, maxSize int64, limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("file_size BETWEEN ? AND ?", minSize, maxSize).
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// GetPhotosByDimensions 根据尺寸范围获取照片
func (r *PhotoRepository) GetPhotosByDimensions(minWidth, maxWidth, minHeight, maxHeight int, limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("width BETWEEN ? AND ? AND height BETWEEN ? AND ?",
		minWidth, maxWidth, minHeight, maxHeight).
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// SearchPhotos 搜索照片
func (r *PhotoRepository) SearchPhotos(keyword string, limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("img_name LIKE ? OR img_path LIKE ? OR notes LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// GetPhotosByRating 根据评分获取照片
func (r *PhotoRepository) GetPhotosByRating(minRating, maxRating int, limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("rating BETWEEN ? AND ?", minRating, maxRating).
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// UpdateViewCount 更新访问次数
func (r *PhotoRepository) UpdateViewCount(hash string) error {
	return r.db.Model(&tables.Photo{}).
		Where("hash = ?", hash).
		UpdateColumns(map[string]interface{}{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_viewed_at": "NOW()",
		}).Error
}

// CreatePhotoFromImageAPI 从ImageAPI创建Photo记录
func (r *PhotoRepository) CreatePhotoFromImageAPI(hash, imgPath string, exifData interface{}) *tables.Photo {
	// 提取文件名
	imgName := filepath.Base(imgPath)

	// 提取文件格式
	format := strings.ToLower(filepath.Ext(imgPath))
	if format != "" && format[0] == '.' {
		format = format[1:] // 去掉点号
	}

	photo := &tables.Photo{
		ImgPath:   imgPath,
		ImgName:   imgName,
		Hash:      hash,
		Format:    format,
		Rating:    0,
		ViewCount: 0,
	}

	// 获取文件信息以获取修改时间和文件大小
	if fileInfo, err := os.Stat(imgPath); err == nil {
		fileModTime := fileInfo.ModTime()
		photo.LastModified = &fileModTime

		// 如果没有EXIF数据，至少设置文件大小
		if photo.FileSize == 0 {
			photo.FileSize = fileInfo.Size()
		}
	}

	// 如果有EXIF数据，提取基础信息和时间信息
	if exifData != nil {
		if parsedExif, ok := exifData.(*model.ParsedExif); ok && parsedExif != nil {
			// 设置基础图片信息
			photo.Width = parsedExif.BaseInfo.ImageWidth
			photo.Height = parsedExif.BaseInfo.ImageHeight
			photo.FileSize = parsedExif.BaseInfo.FileSize

			// 计算宽高比（保留2位小数）
			if photo.Height > 0 {
				ratio := float64(photo.Width) / float64(photo.Height)
				photo.AspectRatio = float32(math.Round(ratio*100) / 100)
			}

			// 处理拍摄时间 - 优先使用EXIF中的拍摄时间，如果没有则使用文件修改时间
			var takenAt *time.Time

			// 尝试解析EXIF中的拍摄时间
			if parsedExif.Exif.DateTimeOrig != "" {
				if parsedTime, err := parseExifDateTime(parsedExif.Exif.DateTimeOrig); err == nil {
					takenAt = &parsedTime
				}
			}

			// 如果EXIF中没有拍摄时间，使用文件修改时间
			if takenAt == nil && photo.LastModified != nil {
				takenAt = photo.LastModified
			}

			photo.TakenAt = takenAt

			// 处理文件创建时间
			if parsedExif.BaseInfo.CreateDate != "" {
				if parsedTime, err := parseExifDateTime(parsedExif.BaseInfo.CreateDate); err == nil {
					photo.FileCreatedAt = &parsedTime
				}
			}
		}
	}

	return photo
}

// parseExifDateTime 解析EXIF时间格式
// EXIF时间格式通常为: "2024:11:04 10:40:29" 或 "2024:11:04 10:40:29+08:00"
func parseExifDateTime(dateTimeStr string) (time.Time, error) {
	// 常见的EXIF时间格式
	formats := []string{
		"2006:01:02 15:04:05",       // 标准EXIF格式
		"2006:01:02 15:04:05-07:00", // 带时区
		"2006:01:02 15:04:05+07:00", // 带时区
		"2006-01-02 15:04:05",       // ISO格式
		"2006-01-02T15:04:05Z",      // ISO格式带Z
		"2006-01-02T15:04:05-07:00", // ISO格式带时区
	}

	// 替换EXIF格式中的冒号为横线（年月日部分）
	normalizedStr := strings.Replace(dateTimeStr, ":", "-", 2)

	// 尝试不同的格式
	for _, format := range formats {
		if t, err := time.Parse(format, dateTimeStr); err == nil {
			return t, nil
		}
		if t, err := time.Parse(format, normalizedStr); err == nil {
			return t, nil
		}
	}

	// 如果都解析失败，返回错误
	return time.Time{}, errors.New("unable to parse datetime string: " + dateTimeStr)
}

// GetPhotosByTakenDateRange 根据拍摄时间范围获取照片
func (r *PhotoRepository) GetPhotosByTakenDateRange(startDate, endDate time.Time, limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("taken_at BETWEEN ? AND ?", startDate, endDate).
		Order("taken_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// GetPhotosByModifiedDateRange 根据文件修改时间范围获取照片
func (r *PhotoRepository) GetPhotosByModifiedDateRange(startDate, endDate time.Time, limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("last_modified BETWEEN ? AND ?", startDate, endDate).
		Order("last_modified DESC").
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// GetRecentPhotosByTaken 获取最近拍摄的照片
func (r *PhotoRepository) GetRecentPhotosByTaken(limit int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("taken_at IS NOT NULL").
		Order("taken_at DESC").
		Limit(limit).
		Find(&photos).Error
	return photos, err
}

// GetOldestPhotosByTaken 获取最早拍摄的照片
func (r *PhotoRepository) GetOldestPhotosByTaken(limit int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("taken_at IS NOT NULL").
		Order("taken_at ASC").
		Limit(limit).
		Find(&photos).Error
	return photos, err
}

// GetPhotosWithoutTakenDate 获取没有拍摄时间的照片
func (r *PhotoRepository) GetPhotosWithoutTakenDate(limit int, offset int) ([]*tables.Photo, error) {
	var photos []*tables.Photo
	err := r.db.Where("taken_at IS NULL").
		Limit(limit).
		Offset(offset).
		Find(&photos).Error
	return photos, err
}

// PhotoTimelineItem 时间线统计项
type PhotoTimelineItem struct {
	Date  string `json:"date"`
	Count int64  `json:"count"`
}

// GetPhotoTimeline 获取照片时间线统计（按拍摄日期）
func (r *PhotoRepository) GetPhotoTimeline() ([]PhotoTimelineItem, error) {
	var results []PhotoTimelineItem

	// 使用DATE函数按日期分组统计
	err := r.db.Model(&tables.Photo{}).
		Select("DATE(taken_at) as date, COUNT(*) as count").
		Where("taken_at IS NOT NULL").
		Group("DATE(taken_at)").
		Order("date ASC").
		Scan(&results).Error

	return results, err
}

// GetPhotoTimelineByDateRange 获取指定时间范围的照片时间线统计
func (r *PhotoRepository) GetPhotoTimelineByDateRange(startDate, endDate time.Time) ([]PhotoTimelineItem, error) {
	var results []PhotoTimelineItem

	err := r.db.Model(&tables.Photo{}).
		Select("DATE(taken_at) as date, COUNT(*) as count").
		Where("taken_at IS NOT NULL AND DATE(taken_at) BETWEEN DATE(?) AND DATE(?)", startDate, endDate).
		Group("DATE(taken_at)").
		Order("date ASC").
		Scan(&results).Error

	return results, err
}

// PhotoListItem 照片列表项（用于前端渲染框架）
type PhotoListItem struct {
	Hash        string     `json:"hash"`
	Width       int        `json:"width"`
	Height      int        `json:"height"`
	AspectRatio float32    `json:"aspectRatio"`
	Format      string     `json:"format"`
	TakenAt     *time.Time `json:"takenAt,omitempty"`
	FileSize    int64      `json:"fileSize"`
}

// GetPhotoListItems 获取照片列表项（支持范围查询）
func (r *PhotoRepository) GetPhotoListItems(limit int, offset int) ([]PhotoListItem, error) {
	var items []PhotoListItem

	err := r.db.Model(&tables.Photo{}).
		Select("hash, width, height, aspect_ratio, format, taken_at, file_size").
		Order("taken_at DESC, created_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error

	return items, err
}

// GetPhotoListItemsByDateRange 根据拍摄时间范围获取照片列表项
func (r *PhotoRepository) GetPhotoListItemsByDateRange(startDate, endDate time.Time, limit int, offset int) ([]PhotoListItem, error) {
	var items []PhotoListItem

	err := r.db.Model(&tables.Photo{}).
		Select("hash, width, height, aspect_ratio, format, taken_at, file_size").
		Where("taken_at IS NOT NULL AND DATE(taken_at) BETWEEN DATE(?) AND DATE(?)", startDate, endDate).
		Order("taken_at DESC").
		Limit(limit).
		Offset(offset).
		Scan(&items).Error

	return items, err
}
