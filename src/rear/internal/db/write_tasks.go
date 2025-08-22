package db

import (
	"fmt"
	"rear/internal/model/tables"
	"time"

	"gorm.io/gorm"
)

// PhotoCreateTask 创建照片任务
type PhotoCreateTask struct {
	Photo *tables.Photo
}

func (t *PhotoCreateTask) Execute(db *gorm.DB) error {
	return db.Create(t.Photo).Error
}

func (t *PhotoCreateTask) GetPriority() int {
	return 5 // 中等优先级
}

func (t *PhotoCreateTask) GetID() string {
	return fmt.Sprintf("photo_create_%s", t.Photo.Hash)
}

// PhotoUpdateTask 更新照片任务
type PhotoUpdateTask struct {
	Photo *tables.Photo
}

func (t *PhotoUpdateTask) Execute(db *gorm.DB) error {
	return db.Save(t.Photo).Error
}

func (t *PhotoUpdateTask) GetPriority() int {
	return 5 // 中等优先级
}

func (t *PhotoUpdateTask) GetID() string {
	return fmt.Sprintf("photo_update_%s", t.Photo.Hash)
}

// PhotoDeleteTask 删除照片任务
type PhotoDeleteTask struct {
	Hash string
}

func (t *PhotoDeleteTask) Execute(db *gorm.DB) error {
	return db.Where("hash = ?", t.Hash).Delete(&tables.Photo{}).Error
}

func (t *PhotoDeleteTask) GetPriority() int {
	return 3 // 高优先级
}

func (t *PhotoDeleteTask) GetID() string {
	return fmt.Sprintf("photo_delete_%s", t.Hash)
}

// ExifCreateTask 创建EXIF任务
type ExifCreateTask struct {
	Exif *tables.PhotoExif
}

func (t *ExifCreateTask) Execute(db *gorm.DB) error {
	return db.Create(t.Exif).Error
}

func (t *ExifCreateTask) GetPriority() int {
	return 5 // 中等优先级
}

func (t *ExifCreateTask) GetID() string {
	return fmt.Sprintf("exif_create_%s", t.Exif.Hash)
}

// ExifUpdateTask 更新EXIF任务
type ExifUpdateTask struct {
	Exif *tables.PhotoExif
}

func (t *ExifUpdateTask) Execute(db *gorm.DB) error {
	return db.Save(t.Exif).Error
}

func (t *ExifUpdateTask) GetPriority() int {
	return 5 // 中等优先级
}

func (t *ExifUpdateTask) GetID() string {
	return fmt.Sprintf("exif_update_%s", t.Exif.Hash)
}

// ExifDeleteTask 删除EXIF任务
type ExifDeleteTask struct {
	Hash string
}

func (t *ExifDeleteTask) Execute(db *gorm.DB) error {
	return db.Where("hash = ?", t.Hash).Delete(&tables.PhotoExif{}).Error
}

func (t *ExifDeleteTask) GetPriority() int {
	return 3 // 高优先级
}

func (t *ExifDeleteTask) GetID() string {
	return fmt.Sprintf("exif_delete_%s", t.Hash)
}

// UpdateViewCountTask 更新访问计数任务
type UpdateViewCountTask struct {
	Hash string
}

func (t *UpdateViewCountTask) Execute(db *gorm.DB) error {
	return db.Model(&tables.Photo{}).
		Where("hash = ?", t.Hash).
		UpdateColumns(map[string]interface{}{
			"view_count":     gorm.Expr("view_count + 1"),
			"last_viewed_at": time.Now(),
		}).Error
}

func (t *UpdateViewCountTask) GetPriority() int {
	return 1 // 低优先级
}

func (t *UpdateViewCountTask) GetID() string {
	return fmt.Sprintf("update_view_count_%s", t.Hash)
}

// UserCreateTask 创建用户任务
type UserCreateTask struct {
	User *tables.User
}

func (t *UserCreateTask) Execute(db *gorm.DB) error {
	return db.Create(t.User).Error
}

func (t *UserCreateTask) GetPriority() int {
	return 5 // 中等优先级
}

func (t *UserCreateTask) GetID() string {
	return fmt.Sprintf("user_create_%d", t.User.ID)
}

// UserDeleteTask 删除用户任务
type UserDeleteTask struct {
	ID uint
}

func (t *UserDeleteTask) Execute(db *gorm.DB) error {
	return db.Delete(&tables.User{}, t.ID).Error
}

func (t *UserDeleteTask) GetPriority() int {
	return 3 // 高优先级
}

func (t *UserDeleteTask) GetID() string {
	return fmt.Sprintf("user_delete_%d", t.ID)
}

// LibraryUpdateTask 更新图书馆任务
type LibraryUpdateTask struct {
	ImgPath string
	Enable  bool
}

func (t *LibraryUpdateTask) Execute(db *gorm.DB) error {
	return db.Model(&tables.LibraryTable{}).
		Where("img_path = ?", t.ImgPath).
		Update("is_enable", t.Enable).Error
}

func (t *LibraryUpdateTask) GetPriority() int {
	return 5 // 中等优先级
}

func (t *LibraryUpdateTask) GetID() string {
	return fmt.Sprintf("library_update_%s", t.ImgPath)
}

// LibraryDeleteTask 删除图书馆任务
type LibraryDeleteTask struct {
	ImgPath string
}

func (t *LibraryDeleteTask) Execute(db *gorm.DB) error {
	return db.Where("img_path = ?", t.ImgPath).Delete(&tables.LibraryTable{}).Error
}

func (t *LibraryDeleteTask) GetPriority() int {
	return 3 // 高优先级
}

func (t *LibraryDeleteTask) GetID() string {
	return fmt.Sprintf("library_delete_%s", t.ImgPath)
}
