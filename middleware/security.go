package middleware

import (
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// SecurityHeaders 注入一组业界推荐的安全响应头。
// 这些头属于"零业务侵入"的快速加固，建议挂在最外层。
//
// 头部说明：
//   - X-Content-Type-Options: nosniff 防止浏览器对响应做 MIME 嗅探
//   - X-Frame-Options: DENY 禁止任意页面通过 <iframe> 嵌入本站，防点击劫持
//   - Referrer-Policy: strict-origin-when-cross-origin 跨域跳转时只带 origin，最小化 Referer 泄漏
//   - Permissions-Policy 关闭一组浏览器敏感能力，前端用不到
//   - Cross-Origin-Opener-Policy: same-origin 隔离浏览上下文，缓解 Spectre 类侧信道
//   - Cross-Origin-Resource-Policy: same-site 限制本站资源被跨站读取
//   - X-XSS-Protection: 0 现代浏览器已弃用过滤器，显式关闭避免误伤
//   - Strict-Transport-Security 仅在 HTTPS 部署时启用（由 ENABLE_HSTS=true 控制）
//
// 注：CSP 由前端 index.html 通过 <meta http-equiv> 注入，避免后端 hardcode。
func SecurityHeaders() gin.HandlerFunc {
	enableHSTS := strings.EqualFold(os.Getenv("ENABLE_HSTS"), "true")

	return func(c *gin.Context) {
		h := c.Writer.Header()
		h.Set("X-Content-Type-Options", "nosniff")
		h.Set("X-Frame-Options", "DENY")
		h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		h.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=(), payment=(), usb=()")
		h.Set("Cross-Origin-Opener-Policy", "same-origin")
		h.Set("Cross-Origin-Resource-Policy", "same-site")
		h.Set("X-XSS-Protection", "0")

		if enableHSTS {
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		}

		c.Next()
	}
}
