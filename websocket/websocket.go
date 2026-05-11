package websocket

import (
	"encoding/json"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// Message 定义 WebSocket 消息格式
type Message struct {
	Type      string      `json:"type"`      // 消息类型
	Action    string      `json:"action"`    // 操作类型
	Data      interface{} `json:"data"`      // 消息数据
	Timestamp time.Time   `json:"timestamp"` // 时间戳
}

// Client 表示一个 WebSocket 客户端连接
type Client struct {
	ID      string          `json:"id"`
	Conn    *websocket.Conn `json:"-"`
	Send    chan []byte     `json:"-"`
	Hub     *Hub            `json:"-"`
	UserID  string          `json:"user_id,omitempty"`  // 关联的用户ID
	IsAdmin bool            `json:"is_admin,omitempty"` // 是否为管理员
}

// Hub 管理所有 WebSocket 连接
type Hub struct {
	clients    map[*Client]bool
	broadcast  chan []byte
	register   chan *Client
	unregister chan *Client
	mutex      sync.RWMutex
}

// NewHub 创建新的 Hub 实例
func NewHub() *Hub {
	return &Hub{
		clients:    make(map[*Client]bool),
		broadcast:  make(chan []byte),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

// Run 启动 Hub 的消息处理循环
func (h *Hub) Run() {
	for {
		select {
		case client := <-h.register:
			h.mutex.Lock()
			h.clients[client] = true
			h.mutex.Unlock()
			log.Printf("WebSocket 客户端已连接: %s", client.ID)

		case client := <-h.unregister:
			h.mutex.Lock()
			if _, ok := h.clients[client]; ok {
				delete(h.clients, client)
				close(client.Send)
			}
			h.mutex.Unlock()
			log.Printf("WebSocket 客户端已断开: %s", client.ID)

		case message := <-h.broadcast:
			h.mutex.RLock()
			for client := range h.clients {
				select {
				case client.Send <- message:
				default:
					close(client.Send)
					delete(h.clients, client)
				}
			}
			h.mutex.RUnlock()
		}
	}
}

// BroadcastMessage 广播消息给所有客户端
func (h *Hub) BroadcastMessage(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}
	h.broadcast <- data
}

// BroadcastToAdmins 只向管理员广播消息
func (h *Hub) BroadcastToAdmins(msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.IsAdmin {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

// BroadcastToUser 向特定用户广播消息
func (h *Hub) BroadcastToUser(userID string, msg Message) {
	data, err := json.Marshal(msg)
	if err != nil {
		log.Printf("序列化消息失败: %v", err)
		return
	}

	h.mutex.RLock()
	defer h.mutex.RUnlock()

	for client := range h.clients {
		if client.UserID == userID {
			select {
			case client.Send <- data:
			default:
				close(client.Send)
				delete(h.clients, client)
			}
		}
	}
}

// 全局 Hub 实例
var GlobalHub = NewHub()

// loadAllowedOrigins 解析 ALLOWED_ORIGINS 环境变量，返回允许跨域 WebSocket 的 Origin 集合。
// 与 HTTP 中间件保持一致，避免出现"HTTP 严格、WS 任意"的安全错位。
func loadAllowedOrigins() map[string]struct{} {
	out := map[string]struct{}{}
	raw := strings.TrimSpace(os.Getenv("ALLOWED_ORIGINS"))
	if raw == "" {
		return out
	}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item != "" {
			out[item] = struct{}{}
		}
	}
	return out
}

// WebSocket 升级器配置：
//   - 严格 CheckOrigin：只允许 ALLOWED_ORIGINS 白名单内的 Origin；
//   - 同源请求(Origin == Host)默认放行，方便前后端打包同域部署的常见场景。
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		origin := r.Header.Get("Origin")
		if origin == "" {
			// 非浏览器客户端（如自家命令行工具）不带 Origin，按需放行。
			return true
		}

		if u, err := url.Parse(origin); err == nil && u.Host == r.Host {
			return true
		}

		allowed := loadAllowedOrigins()
		if _, ok := allowed["*"]; ok {
			return true
		}
		_, ok := allowed[origin]
		return ok
	},
}

// HandleWebSocket 处理 WebSocket 连接握手。
// 安全模型：
//   - 不再相信 query 中的 user_id / is_admin；
//   - 客户端必须先调 POST /v1/ws-ticket 用 JWT 换取一次性 ticket；
//   - 这里仅用 ticket 做最终入场校验，命中后从 ticket 中获取真实身份。
func HandleWebSocket(w http.ResponseWriter, r *http.Request) {
	ticketStr := r.URL.Query().Get("ticket")
	tk := ConsumeTicket(ticketStr)
	if tk == nil {
		http.Error(w, "invalid or expired ticket", http.StatusUnauthorized)
		return
	}

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket 升级失败: %v", err)
		return
	}

	client := &Client{
		ID:      generateClientID(),
		Conn:    conn,
		Send:    make(chan []byte, 256),
		Hub:     GlobalHub,
		UserID:  tk.UserID,
		IsAdmin: tk.IsAdmin,
	}

	client.Hub.register <- client

	// 启动读写协程
	go client.writePump()
	go client.readPump()
}

// generateClientID 生成客户端ID
func generateClientID() string {
	return time.Now().Format("20060102150405") + "-" + randomString(8)
}

// randomString 生成随机字符串
func randomString(length int) string {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, length)
	for i := range b {
		b[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(b)
}

// readPump 处理从客户端读取消息
func (c *Client) readPump() {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(512)
	c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(60 * time.Second))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("WebSocket 读取错误: %v", err)
			}
			break
		}

		// 处理客户端消息（心跳等）
		var msg map[string]interface{}
		if err := json.Unmarshal(message, &msg); err != nil {
			log.Printf("解析客户端消息失败: %v", err)
			continue
		}

		// 处理心跳消息
		if msgType, ok := msg["type"].(string); ok && msgType == "ping" {
			response := map[string]interface{}{
				"type":      "pong",
				"timestamp": time.Now(),
			}
			responseData, _ := json.Marshal(response)
			c.Send <- responseData
		}
	}
}

// writePump 处理向客户端发送消息
func (c *Client) writePump() {
	ticker := time.NewTicker(54 * time.Second)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.Send:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if !ok {
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			if err := w.Close(); err != nil {
				return
			}
		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}

// 启动全局 Hub
func init() {
	go GlobalHub.Run()
}
