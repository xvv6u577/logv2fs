package routes

import (
	"os"

	controller "github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/caster8013/logv2rayfullstack/middleware"

	"github.com/gin-gonic/gin"
)

//UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {

	BOOT_MODE := os.Getenv("BOOT_MODE")
	if BOOT_MODE == "" {
		incomingRoutes.Use(middleware.Authentication())
	}

	// http://127.0.0.1:8079/v1/signup
	// body:
	// {
	// 	"email":"anotheruser",
	// 	"password":"anotheruser"
	// 	"path":"ray",
	// 	"status":"plain",
	// 	"role":"normal",
	// }
	// or
	// curl http://127.0.0.1:8079/v1/signup \
	// --include \
	// --header "Content-Type: application/json" \
	// --request "POST" \
	// --data '{"email": "email","password": "email","status":"plain","uuid": "98a131b0-69a5-41ef-9339-d6dbcabaa773", "path": "ray", "role":"normal"}'
	incomingRoutes.POST("/v1/signup", controller.SignUp())
	incomingRoutes.POST("/v1/edit/:name", controller.EditUser())

	incomingRoutes.GET("/v1/users", controller.GetUsers())
	incomingRoutes.GET("/v1/alluser", controller.GetAllUsers())

	incomingRoutes.GET("/v1/users/:user_id", controller.GetUser())
	incomingRoutes.GET("/v1/user/:name", controller.GetUserByName())

	incomingRoutes.GET("/v1/takeuseroffline/:name", controller.TakeItOfflineByUserName())
	incomingRoutes.GET("/v1/takeuseronline/:name", controller.TakeItOnlineByUserName())
	incomingRoutes.GET("/v1/deluser/:name", controller.DeleteUserByUserName())
	incomingRoutes.GET("/v1/writetodb", controller.WriteToDB())

	incomingRoutes.GET("/v1/traffic/all", controller.GetAllUserTraffic())  // v2api
	incomingRoutes.GET("/v1/traffic/:name", controller.GetTrafficByUser()) // v2api

	incomingRoutes.PUT("/v1/addnode", controller.AddNode())
}
