package mongodb

import (
	"os"

	controller "github.com/xvv6u577/logv2fs/controllers/mongodb"
	"github.com/xvv6u577/logv2fs/middleware"

	"github.com/gin-gonic/gin"
)

var (
	GIN_MODE = os.Getenv("GIN_MODE")
)

// AuthorizedRoutes MongoDB版本的授权路由
func AuthorizedRoutes(incomingRoutes *gin.Engine) {

	if GIN_MODE != "test" {
		incomingRoutes.Use(middleware.Authentication())
	}

	// MongoDB版本的路由
	incomingRoutes.POST("/v1/signup", controller.SignUp())
	incomingRoutes.POST("/v1/edit/:name", controller.EditUser())
	incomingRoutes.GET("/v1/n778cf", controller.GetAllUsers())
	incomingRoutes.GET("/v1/user/:name", controller.GetUserByName())
	incomingRoutes.GET("/v1/deluser/:name", controller.DeleteUserByUserName())
	incomingRoutes.PUT("/v1/disableuser/:name", controller.DisableUser())
	incomingRoutes.PUT("/v1/enableuser/:name", controller.EnableUser())
	incomingRoutes.PUT("/v1/759b0v", controller.AddNode())
	incomingRoutes.GET("/v1/681p32", controller.GetDomainsExpiryInfo())
	incomingRoutes.PUT("/v1/g7302b", controller.UpdateExpiryCheckDomainsInfo())
	incomingRoutes.GET("/v1/c47kr8", controller.GetSingboxNodes())
	incomingRoutes.GET("/v1/t7k033", controller.GetActiveGlobalNodes())

	// 自定义日期管理相关路由 - MongoDB版本
	incomingRoutes.PUT("/v1/custom-date", controller.SaveCustomDate())
	incomingRoutes.GET("/v1/custom-dates", controller.GetCustomDates())

	// 费用管理相关路由 - MongoDB版本
	incomingRoutes.POST("/v1/payment", controller.AddPaymentRecord())
	incomingRoutes.GET("/v1/payment/user/:email", controller.GetUserPayments())
	incomingRoutes.GET("/v1/payment/statistics", controller.GetPaymentStatistics())
	incomingRoutes.GET("/v1/payment/records", controller.GetPaymentRecords())
	incomingRoutes.DELETE("/v1/payment/:id", controller.DeletePaymentRecord())
	incomingRoutes.PUT("/v1/payment/:id", controller.UpdatePaymentRecord())
}
