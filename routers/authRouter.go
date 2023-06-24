package routes

import (
	controller "github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/gin-contrib/static"

	"github.com/gin-gonic/gin"
)

// login, home, logout, routers for react frontend app
func AuthRoutes(incomingRoutes *gin.Engine) {

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
		"/",
	}

	for _, route := range frontendRoutes {
		incomingRoutes.Use(static.Serve(route, static.LocalFile("./frontend/build/", true)))
	}

	// http://127.0.0.1:8079/v1/login
	// body:
	// {
	// 	"email":"testuser",
	// 	"password":"testuser"
	// }
	incomingRoutes.POST("/v1/login", controller.Login())
	incomingRoutes.GET("/static/:name", controller.GetSubscripionURL())
	incomingRoutes.GET("/config/:name", controller.GetUserSimpleInfo())
	incomingRoutes.Use(static.Serve("/clash/", static.LocalFile("./yaml/results/", false)))

	// incomingRoutes.NoRoute(func(c *gin.Context) {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "page not found."})
	// })
}
