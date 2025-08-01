package repositories

import (
	"errors"
	"rear/internal/db"
	"rear/internal/model"

	"gorm.io/gorm"
)

type UserService struct{}

func NewUserService() *UserService {
	return &UserService{}
}

// 创建用户（写操作）
func (s *UserService) CreateUser(user *model.User) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Create(user).Error
	})
}

// 更新用户（写操作）
func (s *UserService) UpdateUser(id uint, updates map[string]interface{}) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Model(&model.User{}).
			Where("id = ?", id).
			Updates(updates).Error
	})
}

/*
Model() 方法的作用
1. 指定操作的表和模型
2. 表名推断

✅ 正确写法1：使用 Model()
goreturn db.GetDB().Model(&model.User{}).
    Where("id = ?", id).
    Update("name", name).Error
✅ 正确写法2：直接使用结构体实例
goreturn db.GetDB().Where("id = ?", id).
    Updates(&model.User{Name: name}).Error
✅ 正确写法3：使用 Table() 指定表名
goreturn db.GetDB().Table("users").
    Where("id = ?", id).
    Update("name", name).Error
*/

// 更新用户名称（写操作）
func (s *UserService) UpdateUserName(id uint, name string) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Model(&model.User{}).
			Where("id = ?", id).
			Update("name", name).Error
	})
}

// 删除用户（写操作）
func (s *UserService) DeleteUser(id uint) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Delete(&model.User{}, id).Error
	})
}

// 软删除用户（写操作）
func (s *UserService) SoftDeleteUser(id uint) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Delete(&model.User{}, id).Error
	})
}

// 批量创建用户（写操作）
func (s *UserService) CreateUsers(users []model.User) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Create(&users).Error
	})
}

// ============ 读操作（可以并发）============

// 根据 ID 获取用户
func (s *UserService) GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := ExecuteRead(func() error {
		return db.GetDB().First(&user, id).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// 根据用户名获取用户
func (s *UserService) GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := ExecuteRead(func() error {
		return db.GetDB().Where("username = ?", username).First(&user).Error
	})

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}

	return &user, nil
}

// 获取所有用户
func (s *UserService) GetAllUsers() ([]model.User, error) {
	var users []model.User
	err := ExecuteRead(func() error {
		return db.GetDB().Find(&users).Error
	})

	return users, err
}

// 分页获取用户
func (s *UserService) GetUsersPaginated(offset, limit int) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	err := ExecuteRead(func() error {
		// 获取总数
		if err := db.GetDB().Model(&model.User{}).Count(&total).Error; err != nil {
			return err
		}

		// 获取分页数据
		return db.GetDB().Offset(offset).Limit(limit).Find(&users).Error
	})

	return users, total, err
}

// 根据条件查询用户
func (s *UserService) GetUsersByCondition(condition map[string]interface{}) ([]model.User, error) {
	var users []model.User
	err := ExecuteRead(func() error {
		query := db.GetDB()
		for key, value := range condition {
			query = query.Where(key+" = ?", value)
		}
		return query.Find(&users).Error
	})

	return users, err
}

// 事务示例（写操作）
func (s *UserService) CreateUserWithTransaction(user *model.User, callback func(tx *gorm.DB) error) error {
	return ExecuteWrite(func() error {
		return db.GetDB().Transaction(func(tx *gorm.DB) error {
			// 创建用户
			if err := tx.Create(user).Error; err != nil {
				return err
			}

			// 执行其他操作
			if callback != nil {
				return callback(tx)
			}

			return nil
		})
	})
}
