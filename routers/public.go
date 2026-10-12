package routers

import (
	"github.com/gin-contrib/static"
	controller "github.com/xvv6u577/logv2fs/controllers"
	"github.com/xvv6u577/logv2fs/middleware"

	"github.com/gin-gonic/gin"
)

// PublicRoutes MongoDB版本的公共路由
func PublicRoutes(incomingRoutes *gin.Engine) {

	frontendRoutes := []string{
		"/login",
		"/user",
		"/domain",
		"/mypanel",
		"/logout",
		"/macos",
		"/windows",
		"/iphone",
		"/android",
		"/nodes",
		"/addnode",
		"/domain-monitor",
		"/",
	}
	for _, route := range frontendRoutes {
		incomingRoutes.Use(static.Serve(route, static.LocalFile("./frontend/build/", true)))
	}

	// MongoDB版本的路由
	// login —— 叠加按 IP 维度的登录限流，防暴力破解；
	// 账号维度的限流在 controller 内通过 LoginAccountAllow 触发。
	incomingRoutes.POST("/v1/login", middleware.NoStore(), middleware.LoginRateLimit(), controller.Login())

	// shadowrocket config
	incomingRoutes.GET("/static/:name", controller.GetSubscripionURL())

	// singbox config
	incomingRoutes.GET("/singbox/:name", controller.ReturnSingboxJson())

	// verge config
	incomingRoutes.GET("/verge/:name", controller.ReturnVergeYAML())

	// AnyTLS 证书监控目标 —— 供 cloudflare-logsbfs Worker 定时同步。
	// 不走 JWT（Worker 无用户会话），改用独立只读 token：环境变量 CERT_TARGETS_TOKEN，
	// 请求头 X-Cert-Token 或 Authorization: Bearer。未配置 token 时接口返回 503。
	incomingRoutes.GET("/v1/cert-targets", middleware.NoStore(), middleware.CertTargetsToken(), controller.GetCertTargets())
}
