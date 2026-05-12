package singbox

// control_server.go：在 singbox 子进程里启动一个**仅监听 loopback** 的小型 HTTP
// 控制服务，httpserver 进程通过它把"运行时增/禁/删用户"指令传过来。
//
// 设计原则：
//   - 不引入 gin/echo 这类大框架，标准库 net/http 足够。
//   - 端口默认 127.0.0.1:8479，可由 SINGBOX_CONTROL_LISTEN 改写。
//   - 鉴权采用单个 shared secret token (Header: X-Control-Token)。
//   - 处理函数全部把工作委托给 ApplyAddUser / ApplyRemoveUser，本文件只负责协议。

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	box "github.com/sagernet/sing-box"
)

// DefaultControlListen / 控制服务默认监听地址。
// 出于安全考虑，未配置时仍走 loopback，绝不绑定 0.0.0.0。
const DefaultControlListen = "127.0.0.1:8479"

// controlReadHeaderTimeout / controlReadTimeout 控制请求处理的最大耗时，
// loopback 调用极快，给个 3 秒已经非常宽松；防止异常请求把进程拖住。
const (
	controlReadHeaderTimeout = 3 * time.Second
	controlReadTimeout       = 5 * time.Second
	controlWriteTimeout      = 5 * time.Second
	controlIdleTimeout       = 60 * time.Second
)

// RunControlServer 在 ctx 取消之前阻塞运行控制服务。
//
//   - addr：监听地址；空字符串自动取 DefaultControlListen。如显式给的不是 loopback
//     地址，会在日志里 WARN 但仍尊重配置（管理员可能故意走 docker bridge）。
//   - token：鉴权 token，空时拒绝启动，避免裸跑。
//   - instance：当前 box.Box 引用，所有路由的实际工作目标。
//
// 返回的 error：监听失败 / 立即退出错误。Graceful shutdown 不视作错误。
func RunControlServer(ctx context.Context, addr, token string, instance *box.Box) error {
	if token == "" {
		return errors.New("control server requires SINGBOX_CONTROL_TOKEN to be set")
	}
	if instance == nil {
		return errors.New("control server requires a non-nil sing-box instance")
	}
	if addr == "" {
		addr = DefaultControlListen
	}

	mux := http.NewServeMux()
	handler := &controlHandler{instance: instance}

	mux.HandleFunc("/control/healthz", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})
	mux.HandleFunc("/control/users", handler.usersCollection)
	mux.HandleFunc("/control/users/", handler.usersItem) // 注意尾斜杠：匹配 /users/{...}

	srv := &http.Server{
		Addr:              addr,
		Handler:           authMiddleware(token, loggingMiddleware(mux)),
		ReadHeaderTimeout: controlReadHeaderTimeout,
		ReadTimeout:       controlReadTimeout,
		WriteTimeout:      controlWriteTimeout,
		IdleTimeout:       controlIdleTimeout,
	}

	if !isLoopback(addr) {
		log.Printf("[singbox.control] WARN: listening on non-loopback address %q; ensure firewall is configured", addr)
	}

	// 提前建监听器，方便把启动错误带回调用方（go func 里再 panic 就难处理了）
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
	log.Printf("[singbox.control] listening on %s", ln.Addr())

	// 后台关闭：ctx 取消后给 server 一个 5 秒优雅退出窗口。
	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_ = srv.Shutdown(shutdownCtx)
	}()

	if err := srv.Serve(ln); err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

// controlHandler 把路由与具体业务方法关联起来；所有依赖都注入到字段里，
// 方便后续单测把 instance 换成 fake。
type controlHandler struct {
	instance *box.Box
}

// usersCollection 处理 /control/users 上的 POST（增加用户）。
func (h *controlHandler) usersCollection(w http.ResponseWriter, r *http.Request) {
	if r.URL.Path != "/control/users" {
		http.NotFound(w, r)
		return
	}
	switch r.Method {
	case http.MethodPost:
		var req AddUserRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
			return
		}
		if err := ApplyAddUser(h.instance, req); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
	default:
		w.Header().Set("Allow", http.MethodPost)
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
	}
}

// usersItem 处理:
//
//	DELETE /control/users/:email          (delete)
//	PUT    /control/users/:email/disable  (disable, 等价于 delete)
//	PUT    /control/users/:email/enable   (enable, body 带 uuid/user_id 重建用户)
//
// 路径手解析，不引第三方 mux。
func (h *controlHandler) usersItem(w http.ResponseWriter, r *http.Request) {
	rest := strings.TrimPrefix(r.URL.Path, "/control/users/")
	if rest == "" {
		http.NotFound(w, r)
		return
	}
	parts := strings.Split(rest, "/")
	email := parts[0]
	if email == "" {
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": "email is required"})
		return
	}

	// 形如 /control/users/{email}
	if len(parts) == 1 {
		if r.Method != http.MethodDelete {
			w.Header().Set("Allow", http.MethodDelete)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
			return
		}
		if err := ApplyRemoveUser(h.instance, email); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
		return
	}

	// 形如 /control/users/{email}/{action}
	if len(parts) == 2 {
		action := parts[1]
		if r.Method != http.MethodPut {
			w.Header().Set("Allow", http.MethodPut)
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "method not allowed"})
			return
		}
		switch action {
		case "disable":
			if err := ApplyRemoveUser(h.instance, email); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
			return
		case "enable":
			var req AddUserRequest
			if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "invalid json body"})
				return
			}
			// 路径中的 email 是事实来源，避免与 body 不一致
			req.EmailAsId = email
			if err := ApplyAddUser(h.instance, req); err != nil {
				writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusOK, map[string]any{"status": "ok"})
			return
		default:
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "unknown action: " + action})
			return
		}
	}

	http.NotFound(w, r)
}

// authMiddleware 校验 X-Control-Token；常量时间比较防止 timing 攻击。
func authMiddleware(token string, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got := r.Header.Get("X-Control-Token")
		if !constantTimeEqual(got, token) {
			writeJSON(w, http.StatusUnauthorized, map[string]any{"error": "invalid control token"})
			return
		}
		next.ServeHTTP(w, r)
	})
}

// loggingMiddleware 记录每次控制端请求，便于排查同步问题。
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rec := &statusRecorder{ResponseWriter: w, status: 200}
		next.ServeHTTP(rec, r)
		log.Printf("[singbox.control] %s %s -> %d (%s)",
			r.Method, r.URL.Path, rec.status, time.Since(start))
	})
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (s *statusRecorder) WriteHeader(code int) {
	s.status = code
	s.ResponseWriter.WriteHeader(code)
}

// constantTimeEqual 简易常量时间字符串比较，避免引入 subtle 包多套类型转换。
// 长度不等先短路是安全的：不同长度本身就不可能匹配，不构成 token 长度泄漏。
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

// isLoopback 粗略判断 host 是否是 loopback；用于启动日志提示，无安全语义。
func isLoopback(addr string) bool {
	host, _, err := net.SplitHostPort(addr)
	if err != nil {
		return false
	}
	if host == "" || host == "localhost" || host == "127.0.0.1" || host == "::1" {
		return true
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func writeJSON(w http.ResponseWriter, code int, body any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(body)
}
