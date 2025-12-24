package postgres

import (
	"github.com/gin-contrib/static"
	controller "github.com/xvv6u577/logv2fs/controllers/postgres"

	"github.com/gin-gonic/gin"
)

// PublicRoutesPG PostgreSQL版本的公共路由
func PublicRoutesPG(incomingRoutes *gin.Engine) {

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
		"/paymentrecords",    // 添加缴费记录页面
		"/paymentstatistics", // 费用统计页面
		"/",
	}
	for _, route := range frontendRoutes {
		incomingRoutes.Use(static.Serve(route, static.LocalFile("./frontend/build/", true)))
	}

	// PostgreSQL版本的路由
	// login
	incomingRoutes.POST("/v1/login", controller.LoginPG())

	// shadowrocket config
	incomingRoutes.GET("/static/:name", controller.GetSubscripionURLPG())

	// singbox config
	incomingRoutes.GET("/singbox/:name", controller.ReturnSingboxJsonPG())

	// verge config
	incomingRoutes.GET("/verge/:name", controller.ReturnVergeYAMLPG())
}
