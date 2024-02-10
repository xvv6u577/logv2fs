package routes

import (
	"github.com/gin-contrib/static"
	controller "github.com/xvv6u577/logv2fs/controllers"

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

	// login
	incomingRoutes.POST("/v1/login", controller.Login())

	// shadowrocket config
	incomingRoutes.GET("/static/:name", controller.GetSubscripionURL())

	// singbox config
	incomingRoutes.GET("/singbox/:name", controller.ReturnSingboxJson())

	// verge config
	incomingRoutes.GET("/verge/:name", controller.ReturnVergeYAML())

	// clash config
	// incomingRoutes.Use(static.Serve("/clash/", static.LocalFile("./config/results/", false)))
	incomingRoutes.GET("/clash/:filename", controller.ReturnClashYAML())

}
