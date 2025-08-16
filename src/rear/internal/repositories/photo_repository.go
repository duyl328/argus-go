package repositories

import (
	"path/filepath"
	"rear/internal/db"
	"rear/internal/model/tables"
	"strings"

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

// Create 创建照片记录
func (r *PhotoRepository) Create(photo *tables.Photo) error {
	return r.db.Create(photo).Error
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

// Update 更新照片记录
func (r *PhotoRepository) Update(photo *tables.Photo) error {
	return r.db.Save(photo).Error
}

// Delete 删除照片记录
func (r *PhotoRepository) Delete(hash string) error {
	return r.db.Where("hash = ?", hash).Delete(&tables.Photo{}).Error
}

// Exists 检查照片记录是否存在
func (r *PhotoRepository) Exists(hash string) (bool, error) {
	var count int64
	err := r.db.Model(&tables.Photo{}).Where("hash = ?", hash).Count(&count).Error
	return count > 0, err
}

// CreateOrUpdate 创建或更新照片记录
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
		ImgPath: imgPath,
		ImgName: imgName,
		Hash:    hash,
		Format:  format,
		Rating:  0,
		ViewCount: 0,
	}

	// 如果有EXIF数据，提取基础信息
	// 这里可以根据实际的EXIF数据结构来填充
	// if exifData != nil {
	//     // 从EXIF数据中提取Width, Height, FileSize等
	// }

	return photo
}