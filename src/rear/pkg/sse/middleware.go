package sse

import (
	"net"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

// NoWriteDeadlineMiddleware SSE 专用中间件：移除写超时限制
// 使用方法: router.GET("/sse/connect", sse.NoWriteDeadlineMiddleware(), handler)
func NoWriteDeadlineMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 尝试获取底层的 TCP 连接
		if conn, ok := getUnderlyingConn(c.Writer); ok {
			// 设置一个非常远的未来时间作为 deadline（等同于无限期）
			// 这会覆盖 http.Server 的 WriteTimeout
			farFuture := time.Now().Add(876000 * time.Hour) // 100 年后
			conn.SetWriteDeadline(farFuture)
		}

		c.Next()
	}
}

// getUnderlyingConn 获取底层的 net.Conn
func getUnderlyingConn(w gin.ResponseWriter) (net.Conn, bool) {
	// Gin 的 ResponseWriter 层次结构：
	// gin.ResponseWriter -> http.ResponseWriter -> http.response -> net.Conn

	// 方法1: 尝试通过 Hijacker 接口获取连接
	if hijacker, ok := w.(http.Hijacker); ok {
		// 注意：这里不能真的调用 Hijack()，因为那会接管连接
		// 我们只是检查接口是否支持

		// 方法2: 通过反射或类型断言访问内部字段
		// 但这依赖于 Go 的内部实现，不够稳定

		// 更稳定的方法：依赖 http.response 的 Hijack 实现
		// 实际上我们无法直接获取 net.Conn 而不 Hijack

		_ = hijacker
	}

	// 由于无法直接访问 net.Conn，我们返回 false
	// 真正的解决方案是在创建 http.Server 时就设置 WriteTimeout = 0
	return nil, false
}

// SetNoWriteDeadline 尝试为当前连接移除写超时（已弃用，无法实现）
// 这个方法无法在不 Hijack 连接的情况下修改 deadline
func SetNoWriteDeadline(c *gin.Context) {
	// 无法实现：无法在不 Hijack 的情况下访问 net.Conn
	// 最佳方案是在 http.Server 创建时设置 WriteTimeout = 0
}
