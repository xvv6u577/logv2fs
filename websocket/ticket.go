package websocket

import (
	"crypto/rand"
	"encoding/base64"
	"sync"
	"time"
)

// Ticket 是一次性 WebSocket 握手票据。
// 设计目的：避免在 WebSocket 连接的 query string 中暴露 JWT。
//
// 流程：
//  1. 客户端先用 JWT 调用 POST /v1/ws-ticket，后端校验 JWT 后签发短 TTL ticket。
//  2. 客户端用 ticket 通过 ws://host/ws?ticket=xxx 建立连接。
//  3. 服务端校验 ticket 命中 → 取出绑定的 user_id/role → 立即从 store 删除（一次性）。
//
// 安全特性：
//   - TTL 30 秒：覆盖 Cloudflare 边缘转发抖动，同时保持较小重放窗口。
//   - 一次性消费：成功换连接后立即失效。
//   - 32 字节随机 base64：不可被预测。
type Ticket struct {
	UserID  string
	Role    string
	IsAdmin bool
	Expires time.Time
}

var (
	ticketStore sync.Map // ticket(string) -> *Ticket

	// ticketTTL 控制 ticket 有效期。Cloudflare Worker 反代链路下 5 秒容易被边缘延迟吃掉。
	ticketTTL = 30 * time.Second
)

// IssueTicket 签发新 ticket 并写入内存 store。
// 调用方负责确保 userID/role 来自经过 JWT 校验的 claims。
func IssueTicket(userID, role string) (string, error) {
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	t := base64.RawURLEncoding.EncodeToString(buf)

	ticketStore.Store(t, &Ticket{
		UserID:  userID,
		Role:    role,
		IsAdmin: role == "admin",
		Expires: time.Now().Add(ticketTTL),
	})

	return t, nil
}

// ConsumeTicket 校验并消费 ticket（一次性）。
// 命中且未过期返回 ticket 内容；否则返回 nil。
func ConsumeTicket(t string) *Ticket {
	if t == "" {
		return nil
	}
	v, ok := ticketStore.LoadAndDelete(t)
	if !ok {
		return nil
	}
	tk := v.(*Ticket)
	if time.Now().After(tk.Expires) {
		return nil
	}
	return tk
}

// startTicketGC 后台周期清理过期 ticket，避免极端情况下因签而未用导致内存堆积。
func startTicketGC() {
	go func() {
		t := time.NewTicker(30 * time.Second)
		defer t.Stop()
		for range t.C {
			now := time.Now()
			ticketStore.Range(func(k, v interface{}) bool {
				if now.After(v.(*Ticket).Expires) {
					ticketStore.Delete(k)
				}
				return true
			})
		}
	}()
}

func init() {
	startTicketGC()
}
