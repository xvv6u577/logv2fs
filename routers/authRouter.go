package routes

import (
	controller "github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/gin-contrib/static"

	"github.com/gin-gonic/gin"
)

//UserRoutes function
func AuthRoutes(incomingRoutes *gin.Engine) {

	incomingRoutes.Use(static.Serve("/", static.LocalFile("./frontend/build/", false)))
	incomingRoutes.Use(static.Serve("/login", static.LocalFile("./frontend/build/", false)))
	incomingRoutes.Use(static.Serve("/home", static.LocalFile("./frontend/build/", false)))
	incomingRoutes.Use(static.Serve("/logout", static.LocalFile("./frontend/build/", false)))

	// http://127.0.0.1:8079/v1/user/signup
	// body:
	// {
	// 	"email":"anotheruser",
	// 	"password":"anotheruser"
	// 	"path":"ray",
	// 	"status":"plain",
	// 	"role":"normal",
	// }
	// or
	// curl http://127.0.0.1:8079/v1/user/signup \
	// --include \
	// --header "Content-Type: application/json" \
	// --request "POST" \
	// --data '{"email": "email","password": "email","status":"plain","uuid": "98a131b0-69a5-41ef-9339-d6dbcabaa773", "path": "ray", "role":"normal"}'
	// incomingRoutes.POST("/v1/signup", controller.SignUp())

	// http://127.0.0.1:8079/v1/user/login
	// body:
	// {
	// 	"email":"testuser",
	// 	"password":"testuser"
	// }
	incomingRoutes.POST("/v1/login", controller.Login())

	// incomingRoutes.NoRoute(func(c *gin.Context) {
	// 	c.JSON(http.StatusNotFound, gin.H{"error": "page not found."})
	// })
}
