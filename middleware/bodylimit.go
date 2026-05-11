package middleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// 默认请求体大小：1MB 适用于绝大多数接口。
// 节点批量上传等大体积接口可在路由层单独覆盖。
const (
	DefaultBodyLimit = 1 << 20  // 1 MB
	LargeBodyLimit   = 5 << 20  // 5 MB
)

// BodyLimit 通过 http.MaxBytesReader 限制请求体的最大读取字节数。
// 当请求体超出阈值时，下游 BindJSON 会得到一个 EOF/超限错误，从而被 controller 截断。
// 这能防御"巨型 JSON 打挂内存"类 DoS。
//
// 用法：
//   r.Use(middleware.BodyLimit(middleware.DefaultBodyLimit))
//   // 或在特定路由组覆盖更大上限
//   group.Use(middleware.BodyLimit(middleware.LargeBodyLimit))
func BodyLimit(maxBytes int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body != nil {
			c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxBytes)
		}
		c.Next()
	}
}
