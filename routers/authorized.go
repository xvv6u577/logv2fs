package routes

import (
	"os"

	controller "github.com/xvv6u577/logv2fs/controllers"
	"github.com/xvv6u577/logv2fs/middleware"

	"github.com/gin-gonic/gin"
)

var (
	GIN_MODE = os.Getenv("GIN_MODE")
)

func AuthorizedRoutes(incomingRoutes *gin.Engine) {

	if GIN_MODE != "test" {
		incomingRoutes.Use(middleware.Authentication())
	}

	incomingRoutes.POST("/v1/signup", controller.SignUp())
	incomingRoutes.POST("/v1/edit/:name", controller.EditUser())

	incomingRoutes.GET("/v1/n778cf", controller.GetAllUsers())

	incomingRoutes.GET("/v1/user/:name", controller.GetUserByName())

	incomingRoutes.GET("/v1/offlineuser/:name", controller.TakeItOfflineByUserName())
	incomingRoutes.GET("/v1/onlineuser/:name", controller.TakeItOnlineByUserName())
	incomingRoutes.GET("/v1/deluser/:name", controller.DeleteUserByUserName())

	incomingRoutes.GET("/v1/cl6217", controller.WriteToDB())

	incomingRoutes.PUT("/v1/759b0v", controller.AddNode())
	incomingRoutes.GET("/v1/0l54vs", controller.DisableNodePerUser())
	incomingRoutes.GET("/v1/9mu6g1", controller.EnableNodePerUser())

	incomingRoutes.GET("/v1/681p32", controller.GetDomainInfo())
	incomingRoutes.PUT("/v1/g7302b", controller.UpdateDomainInfo())

	incomingRoutes.GET("/v1/c47kr8", controller.GetNodePartial())
	incomingRoutes.GET("/v1/t7k033", controller.GetActiveGlobalNodesInfo())

	incomingRoutes.GET("/v1/7b3f3f", controller.GetLegacyClashNodes())
	incomingRoutes.PUT("/v1/v242j6", controller.AddLegacyNodes())
}
