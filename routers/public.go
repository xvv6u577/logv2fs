package routes

import (
	"github.com/gin-contrib/static"
	controller "github.com/xvv6u577/logv2fs/controllers"
	"github.com/xvv6u577/logv2fs/database"

	"github.com/gin-gonic/gin"
)

// login, home, logout, routers for react frontend app
func PublicRoutes(incomingRoutes *gin.Engine) {

	frontendRoutes := []string{
		"/login",
		"/home",
		"/domain",
		"/mypanel",
		"/logout",
		"/macos",
		"/windows",
		"/iphone",
		"/android",
		"/nodes",
		"/addnode",
		"/",
	}
	for _, route := range frontendRoutes {
		incomingRoutes.Use(static.Serve(route, static.LocalFile("./frontend/build/", true)))
	}

	// 根据环境变量决定使用PostgreSQL还是MongoDB版本的控制器
	if database.IsUsingPostgres() {
		// PostgreSQL版本的路由
		// login
		incomingRoutes.POST("/v1/login", controller.LoginPG())

		// shadowrocket config
		incomingRoutes.GET("/static/:name", controller.GetSubscripionURLPG())

		// singbox config
		incomingRoutes.GET("/singbox/:name", controller.ReturnSingboxJsonPG())

		// verge config
		incomingRoutes.GET("/verge/:name", controller.ReturnVergeYAMLPG())
	} else {
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
}
