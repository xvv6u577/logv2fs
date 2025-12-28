package postgres

import (
	"os"

	controller "github.com/xvv6u577/logv2fs/controllers/postgres"
	"github.com/xvv6u577/logv2fs/middleware"

	"github.com/gin-gonic/gin"
)

func getGINMode() string {
	return os.Getenv("GIN_MODE")
}

// AuthorizedRoutesPG PostgreSQL版本的授权路由
func AuthorizedRoutesPG(incomingRoutes *gin.Engine) {

	if getGINMode() != "test" {
		incomingRoutes.Use(middleware.Authentication())
	}

	// PostgreSQL版本的路由
	incomingRoutes.POST("/v1/signup", controller.SignUpPG())
	incomingRoutes.POST("/v1/edit/:name", controller.EditUserPG())
	incomingRoutes.GET("/v1/n778cf", controller.GetAllUsersPG())
	incomingRoutes.GET("/v1/user/:name", controller.GetUserByNamePG())
	incomingRoutes.GET("/v1/deluser/:name", controller.DeleteUserByUserNamePG())
	incomingRoutes.PUT("/v1/disableuser/:name", controller.DisableUserPG())
	incomingRoutes.PUT("/v1/enableuser/:name", controller.EnableUserPG())
	incomingRoutes.PUT("/v1/759b0v", controller.AddNodePG())
	incomingRoutes.GET("/v1/681p32", controller.GetDomainsExpiryInfoPG())
	incomingRoutes.PUT("/v1/g7302b", controller.UpdateExpiryCheckDomainsInfoPG())
	incomingRoutes.GET("/v1/c47kr8", controller.GetSingboxNodesPG())
	incomingRoutes.GET("/v1/t7k033", controller.GetActiveGlobalNodesPG())

	// 自定义日期管理相关路由 - PostgreSQL版本
	incomingRoutes.PUT("/v1/custom-date", controller.SaveCustomDatePG())
	incomingRoutes.GET("/v1/custom-dates", controller.GetCustomDatesPG())

	// 费用管理相关路由 - PostgreSQL版本
	incomingRoutes.POST("/v1/payment", controller.AddPaymentRecordPG())
	incomingRoutes.GET("/v1/payment/user/:email", controller.GetUserPaymentsPG())
	incomingRoutes.GET("/v1/payment/statistics", controller.GetPaymentStatisticsPG())
	incomingRoutes.GET("/v1/payment/records", controller.GetPaymentRecordsPG())
	incomingRoutes.DELETE("/v1/payment/:id", controller.DeletePaymentRecordPG())
	incomingRoutes.PUT("/v1/payment/:id", controller.UpdatePaymentRecordPG())
}
