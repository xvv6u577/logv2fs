package routes

import (
	controller "github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/caster8013/logv2rayfullstack/middleware"

	"github.com/gin-gonic/gin"
)

//UserRoutes function
func UserRoutes(incomingRoutes *gin.Engine) {
	incomingRoutes.Use(middleware.Authentication())
	incomingRoutes.GET("/v1/users", controller.GetUsers())
	incomingRoutes.GET("/v1/alluser", controller.GetAllUsers())

	incomingRoutes.GET("/v1/users/:user_id", controller.GetUser())
	incomingRoutes.GET("/v1/user/:name", controller.GetUserByName())

	incomingRoutes.GET("/v1/takeuseroffline/:name", controller.TakeItOfflineByUserName())
	incomingRoutes.GET("/v1/takeuseronline/:name", controller.TakeItOnlineByUserName())
	incomingRoutes.GET("/v1/deluser/:name", controller.DeleteUserByUserName())

	incomingRoutes.GET("/v1/traffic/all", controller.GetAllUserTraffic())  // v2api
	incomingRoutes.GET("/v1/traffic/:name", controller.GetTrafficByUser()) // v2api
}
