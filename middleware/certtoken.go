package middleware

import (
	"crypto/subtle"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
)

// CertTargetsTokenHeader 是证书监控 Worker 调用 /v1/cert-targets 时携带的请求头。
// 也兼容 "Authorization: Bearer <token>"。
const CertTargetsTokenHeader = "X-Cert-Token"

// CertTargetsToken 用独立的只读 token 保护 /v1/cert-targets。
//
//   - token 来自环境变量 CERT_TARGETS_TOKEN，每次请求读取，便于热更新 .env 后重启生效；
//   - 环境变量为空时一律拒绝（503），避免"未配置即裸奔"；
//   - 使用 crypto/subtle 做常量时间比较，防止时序侧信道。
func CertTargetsToken() gin.HandlerFunc {
	return func(c *gin.Context) {
		expected := strings.TrimSpace(os.Getenv("CERT_TARGETS_TOKEN"))
		if expected == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{"error": "cert targets endpoint is disabled"})
			return
		}

		provided := strings.TrimSpace(c.GetHeader(CertTargetsTokenHeader))
		if provided == "" {
			if auth := c.GetHeader("Authorization"); strings.HasPrefix(auth, "Bearer ") {
				provided = strings.TrimSpace(strings.TrimPrefix(auth, "Bearer "))
			}
		}

		if provided == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(expected)) != 1 {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			return
		}
		c.Next()
	}
}
