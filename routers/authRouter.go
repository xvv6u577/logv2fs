package routes

import (
	"os"

	controller "github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/gin-contrib/static"

	"github.com/gin-gonic/gin"
)

// login, home, logout, routers for react frontend app
func AuthRoutes(incomingRoutes *gin.Engine) {

	BOOT_MODE := os.Getenv("GIN_MODE")
	if BOOT_MODE == "release" {
		FRONTEND_PATH := "./frontend/build/"
		incomingRoutes.Use(static.Serve("/login", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/home", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/mypanel", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/logout", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/macos", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/windows", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/iphone", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/android", static.LocalFile(FRONTEND_PATH, true)))
		incomingRoutes.Use(static.Serve("/", static.LocalFile(FRONTEND_PATH, true)))
	}

	// http://127.0.0.1:8079/v1/user/login
	// body:
	// {
	// 	"email":"testuser",
	// 	"password":"testuser"
	// }
	incomingRoutes.POST("/v1/login", controller.Login())
	incomingRoutes.GET("/static/:name", controller.GetSubscripionURL())
	incomingRoutes.GET("/static/:name/ip", controller.GetSubscripionURL())
	incomingRoutes.GET("/config/:name", controller.GetUserSimpleInfo())
	incomingRoutes.Use(static.Serve("/clash/", static.LocalFile("./yaml/results/", false)))

	// incomingRoutes.NoRoute(func(c *gin.Context) {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "page not found."})
	// })
}
