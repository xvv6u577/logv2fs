package routes

import (
	controller "github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/gin-contrib/static"

	"github.com/gin-gonic/gin"
)

// login, home, logout, routers for react frontend app
func AuthRoutes(incomingRoutes *gin.Engine) {

	frontendPath := "./frontend-yarn/build/"
	// incomingRoutes.Use(static.Serve("/login", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/home", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/mypanel", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/logout", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/macos", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/windows", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/iphone", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/android", static.LocalFile(frontendPath, true)))
	// incomingRoutes.Use(static.Serve("/", static.LocalFile(frontendPath, true)))
	routes := []string{"/login", "/home", "/mypanel", "/logout", "/macos", "/windows", "/iphone", "/android", "/"}

	for _, route := range routes {
		incomingRoutes.Use(static.Serve(route, static.LocalFile(frontendPath, true)))
	}

	// http://127.0.0.1:8079/v1/login
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
