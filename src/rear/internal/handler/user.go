package handler

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"rear/internal/model"
	"rear/internal/repositories"
)

// 全局变量
var (
	Users      []model.User // 模拟数据存储
	userNextID = 1
)

// DeleteUser 删除用户
func DeleteUser(c *gin.Context) {
	id := c.Param("id")

	for i, user := range Users {
		if fmt.Sprintf("%d", user.ID) == id {
			Users = append(Users[:i], Users[i+1:]...)
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusOK,
				Message: "User deleted successfully",
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, model.Response{
		Code:    http.StatusNotFound,
		Message: "User not found",
	})
}

// GetUserByID 根据ID获取用户
func GetUserByID(c *gin.Context) {
	id := c.Param("id")

	for _, user := range Users {
		if fmt.Sprintf("%d", user.ID) == id {
			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusOK,
				Message: "Success",
				Data:    user,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, model.Response{
		Code:    http.StatusNotFound,
		Message: "User not found",
	})
}

// InitSampleData 初始化示例数据
func InitSampleData() {
	Users = []model.User{
		{Username: "John Doe", Email: "john@example.com"},
		{Username: "Jane Smith", Email: "jane@example.com"},
	}
	userNextID = 3
}

// UpdateUser 更新用户
func UpdateUser(c *gin.Context) {
	id := c.Param("id")

	for i, user := range Users {
		if fmt.Sprintf("%d", user.ID) == id {
			var updatedUser model.User
			if err := c.ShouldBindJSON(&updatedUser); err != nil {
				c.JSON(http.StatusBadRequest, model.Response{
					Code:    http.StatusBadRequest,
					Message: "Invalid request body",
				})
				return
			}

			updatedUser.ID = user.ID
			Users[i] = updatedUser

			c.JSON(http.StatusOK, model.Response{
				Code:    http.StatusOK,
				Message: "User updated successfully",
				Data:    updatedUser,
			})
			return
		}
	}

	c.JSON(http.StatusNotFound, model.Response{
		Code:    http.StatusNotFound,
		Message: "User not found",
	})
}

// GetUsers 获取所有用户
func GetUsers(c *gin.Context) {
	// 使用示例
	userService := repositories.NewUserService()

	if users, err := userService.GetAllUsers(); err != nil {
		log.Printf("Failed to get all users: %v", err)
	} else {
		log.Printf("Total users: %d", len(users))
		c.JSON(http.StatusOK, model.Response{
			Code:    http.StatusOK,
			Message: "Success",
			Data:    users,
		})

	}

	//c.JSON(http.StatusOK, model.Response{
	//	Code:    http.StatusOK,
	//	Message: "Success",
	//	Data:    Users,
	//})
}

// CreateUser 创建用户
func CreateUser(c *gin.Context) {
	var user model.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Invalid request body",
		})
		return
	}

	// 简单验证
	if user.Username == "" || user.Email == "" {
		c.JSON(http.StatusBadRequest, model.Response{
			Code:    http.StatusBadRequest,
			Message: "Name and email are required",
		})
		return
	}

	userNextID++
	Users = append(Users, user)

	c.JSON(http.StatusCreated, model.Response{
		Code:    http.StatusCreated,
		Message: "User created successfully",
		Data:    user,
	})
}
