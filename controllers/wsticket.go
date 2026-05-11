package controllers

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/xvv6u577/logv2fs/websocket"
)

// IssueWebSocketTicket 是 POST /v1/ws-ticket 的处理器。
// 调用方必须先经过 JWT Authentication 中间件，因此 c 中已携带 email/uid/user_type。
//
// 返回结构:
//
//	{
//	  "ticket":  "<32B random>",
//	  "ttl_sec": 5
//	}
//
// 客户端拿到 ticket 后立刻用它建立 WebSocket（ws://host/ws?ticket=xxx）。
// ticket 有 5 秒 TTL，且仅可使用一次。
func IssueWebSocketTicket() gin.HandlerFunc {
	return func(c *gin.Context) {
		uid := c.GetString("uid")
		role := c.GetString("user_type")
		if uid == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "missing identity"})
			return
		}

		ticket, err := websocket.IssueTicket(uid, role)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to issue ticket"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"ticket":  ticket,
			"ttl_sec": 5,
		})
	}
}
