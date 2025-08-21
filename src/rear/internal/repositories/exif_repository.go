package repositories

import (
	"context"
	"gorm.io/gorm"
	"rear/internal/db"
	"rear/internal/model/tables"
)

// ExifRepository EXIF数据仓库
type ExifRepository struct {
	db *gorm.DB
}

// NewExifRepository 创建EXIF仓库实例
func NewExifRepository() *ExifRepository {
	return &ExifRepository{
		db: db.DB,
	}
}

// Create 创建EXIF记录（同步）
func (r *ExifRepository) Create(exif *tables.PhotoExif) error {
	return r.db.Create(exif).Error
}

// CreateAsync 异步创建EXIF记录
func (r *ExifRepository) CreateAsync(ctx context.Context, exif *tables.PhotoExif) error {
	task := &db.ExifCreateTask{Exif: exif}
	callback := db.TaskCallback{}
	return db.GetManger().SubmitWriteTask(task, callback)
}

// CreateAsyncSync 同步等待异步创建EXIF记录
func (r *ExifRepository) CreateAsyncSync(ctx context.Context, exif *tables.PhotoExif) error {
	task := &db.ExifCreateTask{Exif: exif}
	return db.GetManger().SubmitWriteTaskSync(ctx, task)
}

// GetByHash 根据Hash获取EXIF信息
func (r *ExifRepository) GetByHash(hash string) (*tables.PhotoExif, error) {
	var exif tables.PhotoExif
	err := r.db.Where("hash = ?", hash).First(&exif).Error
	if err != nil {
		return nil, err
	}
	return &exif, nil
}

// Update 更新EXIF记录（同步）
func (r *ExifRepository) Update(exif *tables.PhotoExif) error {
	return r.db.Save(exif).Error
}

// UpdateAsync 异步更新EXIF记录
func (r *ExifRepository) UpdateAsync(ctx context.Context, exif *tables.PhotoExif) error {
	task := &db.ExifUpdateTask{Exif: exif}
	callback := db.TaskCallback{}
	return db.GetManger().SubmitWriteTask(task, callback)
}

// UpdateAsyncSync 同步等待异步更新EXIF记录
func (r *ExifRepository) UpdateAsyncSync(ctx context.Context, exif *tables.PhotoExif) error {
	task := &db.ExifUpdateTask{Exif: exif}
	return db.GetManger().SubmitWriteTaskSync(ctx, task)
}

// Delete 删除EXIF记录（同步）
func (r *ExifRepository) Delete(hash string) error {
	return r.db.Where("hash = ?", hash).Delete(&tables.PhotoExif{}).Error
}

// DeleteAsync 异步删除EXIF记录
func (r *ExifRepository) DeleteAsync(ctx context.Context, hash string) error {
	task := &db.ExifDeleteTask{Hash: hash}
	callback := db.TaskCallback{}
	return db.GetManger().SubmitWriteTask(task, callback)
}

// DeleteAsyncSync 同步等待异步删除EXIF记录
func (r *ExifRepository) DeleteAsyncSync(ctx context.Context, hash string) error {
	task := &db.ExifDeleteTask{Hash: hash}
	return db.GetManger().SubmitWriteTaskSync(ctx, task)
}

// Exists 检查EXIF记录是否存在
func (r *ExifRepository) Exists(hash string) (bool, error) {
	var count int64
	err := r.db.Model(&tables.PhotoExif{}).Where("hash = ?", hash).Count(&count).Error
	return count > 0, err
}

// GetExifsByCamera 根据相机型号获取EXIF记录
func (r *ExifRepository) GetExifsByCamera(make, model string, limit int, offset int) ([]*tables.PhotoExif, error) {
	var exifs []*tables.PhotoExif
	query := r.db.Model(&tables.PhotoExif{})

	if make != "" {
		query = query.Where("make = ?", make)
	}
	if model != "" {
		query = query.Where("model = ?", model)
	}

	err := query.Limit(limit).Offset(offset).Find(&exifs).Error
	return exifs, err
}

// GetExifsWithGPS 获取包含GPS信息的EXIF记录
func (r *ExifRepository) GetExifsWithGPS(limit int, offset int) ([]*tables.PhotoExif, error) {
	var exifs []*tables.PhotoExif
	err := r.db.Where("gps_latitude != 0 OR gps_longitude != 0").
		Limit(limit).
		Offset(offset).
		Find(&exifs).Error
	return exifs, err
}

// GetExifsByDateRange 根据拍摄时间范围获取EXIF记录
func (r *ExifRepository) GetExifsByDateRange(startDate, endDate string, limit int, offset int) ([]*tables.PhotoExif, error) {
	var exifs []*tables.PhotoExif
	err := r.db.Where("datetime_original BETWEEN ? AND ?", startDate, endDate).
		Limit(limit).
		Offset(offset).
		Find(&exifs).Error
	return exifs, err
}

// GetCameraStatistics 获取相机统计信息
func (r *ExifRepository) GetCameraStatistics() ([]map[string]interface{}, error) {
	var results []map[string]interface{}

	err := r.db.Model(&tables.PhotoExif{}).
		Select("make, model, count(*) as count").
		Where("make != '' AND model != ''").
		Group("make, model").
		Order("count DESC").
		Scan(&results).Error

	return results, err
}

// BatchCreate 批量创建EXIF记录
func (r *ExifRepository) BatchCreate(exifs []*tables.PhotoExif) error {
	if len(exifs) == 0 {
		return nil
	}
	return r.db.CreateInBatches(exifs, 100).Error
}

// GetExifsByISO 根据ISO范围获取EXIF记录
func (r *ExifRepository) GetExifsByISO(minISO, maxISO int, limit int, offset int) ([]*tables.PhotoExif, error) {
	var exifs []*tables.PhotoExif
	err := r.db.Where("iso BETWEEN ? AND ?", minISO, maxISO).
		Limit(limit).
		Offset(offset).
		Find(&exifs).Error
	return exifs, err
}

// GetExifsByAperture 根据光圈范围获取EXIF记录
func (r *ExifRepository) GetExifsByAperture(minAperture, maxAperture float64, limit int, offset int) ([]*tables.PhotoExif, error) {
	var exifs []*tables.PhotoExif
	err := r.db.Where("f_number BETWEEN ? AND ?", minAperture, maxAperture).
		Limit(limit).
		Offset(offset).
		Find(&exifs).Error
	return exifs, err
}

// SearchExifs 搜索EXIF记录
func (r *ExifRepository) SearchExifs(keyword string, limit int, offset int) ([]*tables.PhotoExif, error) {
	var exifs []*tables.PhotoExif
	err := r.db.Where("title LIKE ? OR description LIKE ? OR make LIKE ? OR model LIKE ?",
		"%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%", "%"+keyword+"%").
		Limit(limit).
		Offset(offset).
		Find(&exifs).Error
	return exifs, err
}

// GetUniqueCameras 获取所有唯一的相机信息
func (r *ExifRepository) GetUniqueCameras() ([]map[string]string, error) {
	var cameras []map[string]string
	err := r.db.Model(&tables.PhotoExif{}).
		Select("DISTINCT make, model").
		Where("make != '' AND model != ''").
		Order("make, model").
		Scan(&cameras).Error
	return cameras, err
}

// GetCount 获取总记录数
func (r *ExifRepository) GetCount() (int64, error) {
	var count int64
	err := r.db.Model(&tables.PhotoExif{}).Count(&count).Error
	return count, err
}

// GetAll 获取所有EXIF记录
func (r *ExifRepository) GetAll(limit int, offset int) ([]*tables.PhotoExif, error) {
	var exifs []*tables.PhotoExif
	err := r.db.Limit(limit).Offset(offset).Find(&exifs).Error
	return exifs, err
}

// CreateOrUpdate 创建或更新EXIF记录
func (r *ExifRepository) CreateOrUpdate(exif *tables.PhotoExif) error {
	// 先检查是否存在
	exists, err := r.Exists(exif.Hash)
	if err != nil {
		return err
	}

	if exists {
		return r.Update(exif)
	} else {
		return r.Create(exif)
	}
}
