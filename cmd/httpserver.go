/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"fmt"
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/middleware"
	routers "github.com/xvv6u577/logv2fs/routers"
	"github.com/xvv6u577/logv2fs/websocket"
)

// httpserverCmd represents the httpserver command
var httpserverCmd = &cobra.Command{
	Use:   "httpserver",
	Short: "Web API server for singbox",
	Long:  `Web API server for singbox.`,
	Run: func(cmd *cobra.Command, args []string) {

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

		// 初始化数据库监听器
		websocket.InitMongoDBListener()
		websocket.InitSupabaseListener()

		// 添加 WebSocket 路由
		router.GET("/ws", func(c *gin.Context) {
			websocket.HandleWebSocket(c.Writer, c.Request)
		})

		routers.PublicRoutes(router)
		routers.AuthorizedRoutes(router)

		srv := &http.Server{
			Addr:    fmt.Sprintf("%s:%s", SERVER_ADDRESS, SERVER_PORT),
			Handler: router,
		}
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Panic("Start API Server Error: ", err)
		}

	},
}

func init() {
	rootCmd.AddCommand(httpserverCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// httpserverCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// httpserverCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
