/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/middleware"
	routersMongodb "github.com/xvv6u577/logv2fs/routers/mongodb"
	"github.com/xvv6u577/logv2fs/websocket"
)

// NewHTTPServerCmd 返回 httpserver 命令
func NewHTTPServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "httpserver",
		Short: "Web API server for singbox with MongoDB",
		Long:  `Web API server for singbox with MongoDB.`,
		Run: func(cmd *cobra.Command, args []string) {
			// 在 Run 函数内读取环境变量，确保 .env 文件已经被加载
			SERVER_ADDRESS := os.Getenv("SERVER_ADDRESS")
			SERVER_PORT := os.Getenv("SERVER_PORT")
			GIN_MODE := os.Getenv("GIN_MODE")

			if _, err := os.Stat("./logs"); os.IsNotExist(err) {
				os.Mkdir("./logs", 0755)
			}
			logFile, err := os.OpenFile("./logs/httpserver.log", os.O_CREATE|os.O_APPEND|os.O_RDWR, 0644)
			if err != nil {
				log.Fatalln(err)
			}
			log.SetOutput(logFile)

			if GIN_MODE == "release" {
				gin.SetMode(gin.ReleaseMode)
			}

			router := gin.New()
			router.Use(middleware.CORS())
			router.Use(gin.Logger())

			// 初始化数据库监听器（MongoDB版本）
			websocket.InitMongoDBListener()

			// 添加 WebSocket 路由
			router.GET("/ws", func(c *gin.Context) {
				websocket.HandleWebSocket(c.Writer, c.Request)
			})

			routersMongodb.PublicRoutes(router)
			routersMongodb.AuthorizedRoutes(router)

			srv := &http.Server{
				Addr:    fmt.Sprintf("%s:%s", SERVER_ADDRESS, SERVER_PORT),
				Handler: router,
			}
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Panic("Start API Server Error: ", err)
			}

		},
	}
}
