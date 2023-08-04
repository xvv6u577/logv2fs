package routes

import (
	"os"

	controller "github.com/caster8013/logv2rayfullstack/controllers"
	"github.com/caster8013/logv2rayfullstack/middleware"

	"github.com/gin-gonic/gin"
)

//UserRoutes function
func AuthorizedRoutes(incomingRoutes *gin.Engine) {

	GIN_MODE := os.Getenv("GIN_MODE")
	if GIN_MODE != "test" {
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

	incomingRoutes.GET("/v1/n778cf", controller.GetAllUsers())

	incomingRoutes.GET("/v1/user/:name", controller.GetUserByName())

	incomingRoutes.GET("/v1/offlineuser/:name", controller.TakeItOfflineByUserName())
	incomingRoutes.GET("/v1/onlineuser/:name", controller.TakeItOnlineByUserName())
	incomingRoutes.GET("/v1/deluser/:name", controller.DeleteUserByUserName())

	// affect single node!
	incomingRoutes.GET("/v1/cl6217", controller.WriteToDB())

	incomingRoutes.PUT("/v1/759b0v", controller.AddNode())
	incomingRoutes.GET("/v1/0l54vs", controller.DisableNodePerUser())
	incomingRoutes.GET("/v1/9mu6g1", controller.EnableNodePerUser())

	incomingRoutes.GET("/v1/681p32", controller.GetDomainInfo())
	incomingRoutes.PUT("/v1/g7302b", controller.UpdateDomainInfo())

	// /nodes
	incomingRoutes.GET("/v1/c47kr8", controller.GetNodePartial())
	incomingRoutes.GET("/v1/t7k033", controller.GetActiveGlobalNodesInfo())
}
