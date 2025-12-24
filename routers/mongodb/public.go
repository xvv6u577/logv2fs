package mongodb

import (
	"github.com/gin-contrib/static"
	controller "github.com/xvv6u577/logv2fs/controllers/mongodb"

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
		"/paymentrecords",    // 添加缴费记录页面
		"/paymentstatistics", // 费用统计页面
		"/",
	}
	for _, route := range frontendRoutes {
		incomingRoutes.Use(static.Serve(route, static.LocalFile("./frontend/build/", true)))
	}

	// MongoDB版本的路由
	// login
	incomingRoutes.POST("/v1/login", controller.Login())

	// shadowrocket config
	incomingRoutes.GET("/static/:name", controller.GetSubscripionURL())

	// singbox config
	incomingRoutes.GET("/singbox/:name", controller.ReturnSingboxJson())

	// verge config
	incomingRoutes.GET("/verge/:name", controller.ReturnVergeYAML())
}
