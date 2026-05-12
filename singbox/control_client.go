package singbox

// control_client.go：httpserver 进程侧调用 singbox 控制端口的小客户端。
//
// 设计要点：
//   - 提供 package 级单例 DefaultControlClient，由 init() 根据环境变量决定是否启用。
//   - 未配置（token 为空）时 client 为 nil，所有方法转为 no-op，HTTP 业务接口不报错。
//   - 所有方法都设计为 soft-fail：网络失败只打日志，不冒泡。理由：MongoDB 是
//     事实源，sing-box 是缓存；缓存丢一次不影响业务正确性，singbox 重启会全量重建。
//   - 默认超时 3 秒；loopback 调用通常 < 5ms，3 秒已极宽松。

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sync"
	"time"
)

// ControlClient 把 ApplyAddUser/ApplyRemoveUser 的语义封装成跨进程 HTTP 调用。
type ControlClient struct {
	baseURL string
	token   string
	http    *http.Client
}

// NewControlClient 显式构造 client。一般业务代码无需调用，请用 DefaultControlClient。
func NewControlClient(addr, token string) *ControlClient {
	return &ControlClient{
		baseURL: "http://" + addr,
		token:   token,
		http: &http.Client{
			Timeout: 3 * time.Second,
		},
	}
}

var (
	defaultControlClientOnce sync.Once
	defaultControlClient     *ControlClient
)

// DefaultControlClient 返回懒加载初始化的全局 client。
// 当 SINGBOX_CONTROL_TOKEN 为空时返回 nil，调用方需自行判空（或使用 Safe* 包装）。
func DefaultControlClient() *ControlClient {
	defaultControlClientOnce.Do(func() {
		token := os.Getenv("SINGBOX_CONTROL_TOKEN")
		if token == "" {
			log.Printf("[singbox.control_client] SINGBOX_CONTROL_TOKEN not set, control client disabled")
			return
		}
		addr := os.Getenv("SINGBOX_CONTROL_LISTEN")
		if addr == "" {
			addr = DefaultControlListen
		}
		defaultControlClient = NewControlClient(addr, token)
		log.Printf("[singbox.control_client] enabled, target=%s", addr)
	})
	return defaultControlClient
}

// AddUser 同步调用控制端添加用户。任意错误只返回，不阻塞调用方。
func (c *ControlClient) AddUser(req AddUserRequest) error {
	return c.doJSON(http.MethodPost, "/control/users", req)
}

// RemoveUser 同步调用控制端删除用户。
func (c *ControlClient) RemoveUser(emailAsId string) error {
	path := "/control/users/" + url.PathEscape(emailAsId)
	return c.doJSON(http.MethodDelete, path, nil)
}

// DisableUser 与 RemoveUser 在 sing-box 层完全等价；保留独立方法是为了将来扩展
// （例如 disable 后保留连接但限速）。
func (c *ControlClient) DisableUser(emailAsId string) error {
	path := "/control/users/" + url.PathEscape(emailAsId) + "/disable"
	return c.doJSON(http.MethodPut, path, nil)
}

// EnableUser 复活一个先前 disable / delete 过的用户。body 中需带 uuid / user_id。
func (c *ControlClient) EnableUser(req AddUserRequest) error {
	if req.EmailAsId == "" {
		return fmt.Errorf("email_as_id is required")
	}
	path := "/control/users/" + url.PathEscape(req.EmailAsId) + "/enable"
	return c.doJSON(http.MethodPut, path, req)
}

// HealthCheck 用于启动期快速验证控制端可达。
func (c *ControlClient) HealthCheck(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL+"/control/healthz", nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Control-Token", c.token)
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("healthz: %s", resp.Status)
	}
	return nil
}

// doJSON 通用 JSON 请求工具。body 为 nil 时不写请求体。
func (c *ControlClient) doJSON(method, path string, body any) error {
	var buf io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return fmt.Errorf("marshal body: %w", err)
		}
		buf = bytes.NewReader(raw)
	}
	req, err := http.NewRequest(method, c.baseURL+path, buf)
	if err != nil {
		return err
	}
	req.Header.Set("X-Control-Token", c.token)
	if body != nil {
		req.Header.Set("Content-Type", "application/json; charset=utf-8")
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(io.LimitReader(resp.Body, 1024))
		return fmt.Errorf("control %s %s: %s body=%s",
			method, path, resp.Status, string(respBody))
	}
	return nil
}

// SafeAddUser / SafeRemoveUser / SafeEnableUser 是给业务路径调用的"防御性"包装：
// client 为 nil 直接 no-op，错误只打日志不冒泡，确保 HTTP 业务永远不被同步阻塞。
//
// 调用方建议在 goroutine 里再包一层，让 HTTP 响应立刻返回。

// SafeAddUser 软失败地新增用户。
func SafeAddUser(req AddUserRequest) {
	c := DefaultControlClient()
	if c == nil {
		return
	}
	if err := c.AddUser(req); err != nil {
		log.Printf("[singbox.control_client] AddUser %s failed: %v", req.EmailAsId, err)
	}
}

// SafeRemoveUser 软失败地删除用户。
func SafeRemoveUser(emailAsId string) {
	c := DefaultControlClient()
	if c == nil {
		return
	}
	if err := c.RemoveUser(emailAsId); err != nil {
		log.Printf("[singbox.control_client] RemoveUser %s failed: %v", emailAsId, err)
	}
}

// SafeDisableUser 软失败地禁用用户。
func SafeDisableUser(emailAsId string) {
	c := DefaultControlClient()
	if c == nil {
		return
	}
	if err := c.DisableUser(emailAsId); err != nil {
		log.Printf("[singbox.control_client] DisableUser %s failed: %v", emailAsId, err)
	}
}

// SafeEnableUser 软失败地启用用户。
func SafeEnableUser(req AddUserRequest) {
	c := DefaultControlClient()
	if c == nil {
		return
	}
	if err := c.EnableUser(req); err != nil {
		log.Printf("[singbox.control_client] EnableUser %s failed: %v", req.EmailAsId, err)
	}
}
