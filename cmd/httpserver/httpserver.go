/*
Copyright © 2024 NAME HERE <EMAIL ADDRESS>
*/
package httpserver

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/spf13/cobra"
	"github.com/xvv6u577/logv2fs/middleware"
	"github.com/xvv6u577/logv2fs/routers"
	"github.com/xvv6u577/logv2fs/websocket"
)

// HTTP 服务超时常量。
// 设计目标：抵御 Slowloris/慢速请求耗尽连接，对正常业务足够宽松。
//
// - readHeaderTimeout: 读取 HTTP 请求头的最长时间，5 秒能筛掉绝大多数慢速攻击。
// - readTimeout:       从开始读到 body 读完的总时长。
// - writeTimeout:      handler + 响应写出的总时长，覆盖大多数业务接口；
//                      流式接口（如 WebSocket 升级后）不受 writeTimeout 影响，因为升级后连接被劫持。
// - idleTimeout:       Keep-Alive 空闲时长。
// - maxHeaderBytes:    限制头部尺寸，防止请求头炸弹。
const (
	readHeaderTimeout = 5 * time.Second
	readTimeout       = 15 * time.Second
	writeTimeout      = 30 * time.Second
	idleTimeout       = 120 * time.Second
	maxHeaderBytes    = 1 << 20 // 1 MB
)

// NewHTTPServerCmd 返回 httpserver 命令
func NewHTTPServerCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "httpserver",
		Short: "Web API server for singbox with MongoDB",
		Long:  `Web API server for singbox with MongoDB.`,
		Run: func(cmd *cobra.Command, args []string) {
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

			// 中间件挂载顺序很关键：
			//  1. Recovery   保证 panic 不打挂进程
			//  2. Logger     在 Recovery 之后，记录可恢复的请求日志
			//  3. SecurityHeaders 给所有响应统一加固
			//  4. CORS       预检 OPTIONS 请求需要在限流之前放行
			//  5. GlobalRateLimit 兜底 IP 限流
			//  6. BodyLimit  限制请求体大小，防止内存炸弹
			router.Use(gin.Recovery())
			router.Use(gin.Logger())
			router.Use(middleware.SecurityHeaders())
			router.Use(middleware.CORS())
			router.Use(middleware.GlobalRateLimit())
			router.Use(middleware.BodyLimit(middleware.DefaultBodyLimit))

			// MaxMultipartMemory 仅影响 multipart/form-data，与 BodyLimit 互补。
			router.MaxMultipartMemory = 8 << 20

			// 添加 WebSocket 路由（在所有路由分组之前，避免被分组中间件影响）。
			// WebSocket 自身已通过 ticket 鉴权，不在此再叠加 JWT 中间件。
			router.GET("/ws", func(c *gin.Context) {
				websocket.HandleWebSocket(c.Writer, c.Request)
			})

			routers.PublicRoutes(router)
			routers.AuthorizedRoutes(router)

			srv := &http.Server{
				Addr:              fmt.Sprintf("%s:%s", SERVER_ADDRESS, SERVER_PORT),
				Handler:           router,
				ReadHeaderTimeout: readHeaderTimeout,
				ReadTimeout:       readTimeout,
				WriteTimeout:      writeTimeout,
				IdleTimeout:       idleTimeout,
				MaxHeaderBytes:    maxHeaderBytes,
			}
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Panic("Start API Server Error: ", err)
			}
		},
	}
}
