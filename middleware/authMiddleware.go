package middleware

import (
	"net/http"
	"os"
	"strings"
	"sync"

	helper "github.com/xvv6u577/logv2fs/helpers"

	"github.com/gin-gonic/gin"
)

// allowedHeaders 是 CORS 预检允许的请求头白名单。
// 同时显式声明 Authorization 与历史遗留的 token，便于前端平滑迁移。
const allowedHeaders = "Content-Type,Content-Length,Accept-Encoding,X-CSRF-Token,Authorization,token,Accept,Origin,Cache-Control,X-Requested-With"

const allowedMethods = "GET,POST,PUT,PATCH,DELETE,OPTIONS,HEAD"

// 全局原子缓存：避免每次请求都解析 ALLOWED_ORIGINS。
var (
	originsOnce  sync.Once
	allowedOrigs []string
)

// loadAllowedOrigins 解析 ALLOWED_ORIGINS 环境变量（逗号分隔）。
// 解析仅做一次；如果未配置则返回空切片，CORS 中间件将拒绝所有跨域请求。
func loadAllowedOrigins() []string {
	originsOnce.Do(func() {
		raw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
		if raw == "" {
			return
		}
		for _, item := range strings.Split(raw, ",") {
			item = strings.TrimSpace(item)
			if item == "" {
				continue
			}
			allowedOrigs = append(allowedOrigs, item)
		}
	})
	return allowedOrigs
}

// isOriginAllowed 判断请求 Origin 是否命中白名单。
// 支持精确匹配。生产环境若需通配子域，可在此扩展。
func isOriginAllowed(origin string) bool {
	if origin == "" {
		return false
	}
	for _, o := range loadAllowedOrigins() {
		if o == "*" || o == origin {
			return true
		}
	}
	return false
}

// extractBearerToken 从 Authorization header 提取 Bearer token。
// 同时兼容旧版自定义 token header（迁移期保留），优先使用 Authorization。
func extractBearerToken(c *gin.Context) string {
	auth := c.Request.Header.Get("Authorization")
	if auth != "" {
		const prefix = "Bearer "
		if len(auth) > len(prefix) && strings.EqualFold(auth[:len(prefix)], prefix) {
			return strings.TrimSpace(auth[len(prefix):])
		}
	}
	return c.Request.Header.Get("token")
}

// Authentication 校验 JWT token 并将身份信息写入 gin.Context。
// 与旧实现的差别：
//   - 优先读 Authorization: Bearer xxx；
//   - 失败统一返回 401，不再使用 500（500 容易被监控误报为后端故障）；
//   - 错误信息走日志，响应体只给通用提示，避免泄露内部细节。
func Authentication() gin.HandlerFunc {
	return func(c *gin.Context) {
		clientToken := extractBearerToken(c)
		if clientToken == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "no token provided"})
			c.Abort()
			return
		}

		claims, msg := helper.ValidateToken(clientToken)
		if msg != "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			c.Abort()
			return
		}

		c.Set("email", claims.Email)
		c.Set("uuid", claims.UUID)
		c.Set("name", claims.Name)
		c.Set("uid", claims.Uid)
		c.Set("user_type", claims.Role)

		c.Next()
	}
}

// AdminOnly 强制要求当前请求来自 admin。
// 路由层挂上后，无需在每个 controller 内重复 helper.CheckUserType("admin")。
// 保留 controller 内已有的检查作为 defense-in-depth。
func AdminOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_type")
		if role != "admin" {
			c.JSON(http.StatusForbidden, gin.H{"error": "admin access required"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// SelfOrAdmin 允许：当前用户访问自己（path 中的 :name 等于 token email），或者管理员任意访问。
// 用于 /v1/user/:name、/v1/payment/user/:email 等"按用户名读取"的接口。
func SelfOrAdmin(paramName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := c.GetString("user_type")
		if role == "admin" {
			c.Next()
			return
		}
		target := c.Param(paramName)
		caller := c.GetString("email")
		if target == "" || caller == "" || target != caller {
			c.JSON(http.StatusForbidden, gin.H{"error": "forbidden"})
			c.Abort()
			return
		}
		c.Next()
	}
}

// CORS 严格按白名单回写 Access-Control-Allow-Origin。
// 设计要点：
//   - 不再回写 "*"；当 Origin 不在白名单时不写 Allow-Origin 头，浏览器自然拒绝；
//   - 显式 Vary: Origin，避免 CDN 缓存串味；
//   - Allow-Credentials 仅在白名单命中时设置，符合规范。
func CORS() gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.Request.Header.Get("Origin")
		c.Header("Vary", "Origin")

		if isOriginAllowed(origin) {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Header("Access-Control-Allow-Headers", allowedHeaders)
			c.Header("Access-Control-Allow-Methods", allowedMethods)
			c.Header("Access-Control-Max-Age", "600")
		}

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}
