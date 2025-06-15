package handler

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"rear/internal/model"
	"time"
)

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "OK",
		Data: map[string]interface{}{
			"timestamp": time.Now().Unix(),
			"version":   "1.0.0",
		},
	})
}

func BasicResponse(c *gin.Context) {
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "OK",
		Data:    "hi",
	})
}
func BasicResponseV1(c *gin.Context) {
	c.JSON(http.StatusOK, model.Response{
		Code:    http.StatusOK,
		Message: "OK",
		Data:    "hi, this is v1",
	})
}
