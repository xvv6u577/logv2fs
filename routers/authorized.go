package routers

import (
	"os"

	controller "github.com/xvv6u577/logv2fs/controllers"
	"github.com/xvv6u577/logv2fs/middleware"

	"github.com/gin-gonic/gin"
)

func getGINMode() string {
	return os.Getenv("GIN_MODE")
}

// AuthorizedRoutes 安装所有需要鉴权的路由。
//
// 安全模型分两层：
//  1. 全部路由先经过 middleware.Authentication() 校验 JWT。
//  2. 仅管理员可访问的路由再叠加 middleware.AdminOnly() 做角色检查；
//     "本人或管理员"类（如查询自身用户信息）使用 SelfOrAdmin。
//
// controller 内部原有的 helper.CheckUserType("admin") 保留作为
// defense-in-depth：即便有人在路由层漏挂中间件，业务函数本身仍能拦下。
func AuthorizedRoutes(incomingRoutes *gin.Engine) {

	if getGINMode() != "test" {
		incomingRoutes.Use(middleware.Authentication())
	}
	incomingRoutes.Use(middleware.NoStore())

	// WebSocket 票据签发：任何已登录用户都可以为自己拿一张，30 秒一次性 ticket。
	incomingRoutes.POST("/v1/ws-ticket", controller.IssueWebSocketTicket())

	// =============== 管理员专属路由 ===============
	admin := incomingRoutes.Group("/v1")
	if getGINMode() != "test" {
		admin.Use(middleware.AdminOnly())
	}
	{
		admin.POST("/signup", controller.SignUp())
		admin.POST("/edit/:name", controller.EditUser())
		admin.GET("/users", controller.GetAllUsers())
		admin.DELETE("/user/:name", controller.DeleteUserByUserName())
		admin.PUT("/disableuser/:name", controller.DisableUser())
		admin.PUT("/enableuser/:name", controller.EnableUser())
		admin.PUT("/upsert-nodes", controller.UpsertNodes())
		admin.GET("/singbox-nodes", controller.GetSingboxNodes())

		admin.PUT("/custom-date", controller.SaveCustomDate())
		admin.GET("/custom-dates", controller.GetCustomDates())

		admin.POST("/payment", controller.AddPaymentRecord())
		admin.GET("/payment/statistics", controller.GetPaymentStatistics())
		admin.GET("/payment/records", controller.GetPaymentRecords())
		admin.DELETE("/payment/:id", controller.DeletePaymentRecord())
		admin.PUT("/payment/:id", controller.UpdatePaymentRecord())
	}

	// =============== 本人或管理员可访问 ===============
	// 路径中的 :name / :email 是用户自身标识，普通用户只能访问与自己匹配的资源。
	incomingRoutes.GET("/v1/user/:name",
		middleware.SelfOrAdmin("name"),
		controller.GetUserByName())
	incomingRoutes.GET("/v1/payment/user/:email",
		middleware.SelfOrAdmin("email"),
		controller.GetUserPayments())

	// =============== 任意已登录用户 ===============
	incomingRoutes.GET("/v1/subscription-nodes", controller.GetSubscriptionNodes())
}
