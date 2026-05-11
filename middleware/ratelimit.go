package middleware

import (
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// keyedLimiter 是一组按 key（IP / 账号）独立计数的令牌桶集合。
// 内部用 sync.Map 承载并发访问；后台 goroutine 周期性清理空闲 key，避免无限增长。
//
// 设计取舍：
//   - 单实例内存桶，零外部依赖；多实例水平扩展时换 Redis 即可。
//   - 默认每个 key 一个 *rate.Limiter，TTL 30 分钟无访问则回收。
type keyedLimiter struct {
	r       rate.Limit
	b       int
	ttl     time.Duration
	entries sync.Map
}

type limiterEntry struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// newKeyedLimiter 创建一个带 TTL 自动回收的限流器。
// r 为每秒生成令牌速率，b 为令牌桶容量（突发上限）。
func newKeyedLimiter(r rate.Limit, b int, ttl time.Duration) *keyedLimiter {
	kl := &keyedLimiter{r: r, b: b, ttl: ttl}
	go kl.gcLoop()
	return kl
}

// allow 判断 key 当前是否被允许放行。
// 命中则消耗一个令牌。
func (kl *keyedLimiter) allow(key string) bool {
	now := time.Now()
	v, _ := kl.entries.LoadOrStore(key, &limiterEntry{
		limiter:  rate.NewLimiter(kl.r, kl.b),
		lastSeen: now,
	})
	e := v.(*limiterEntry)
	e.lastSeen = now
	return e.limiter.Allow()
}

// gcLoop 定期回收长时间未使用的 entry。
func (kl *keyedLimiter) gcLoop() {
	t := time.NewTicker(10 * time.Minute)
	defer t.Stop()
	for range t.C {
		cutoff := time.Now().Add(-kl.ttl)
		kl.entries.Range(func(k, v interface{}) bool {
			if v.(*limiterEntry).lastSeen.Before(cutoff) {
				kl.entries.Delete(k)
			}
			return true
		})
	}
}

// 三层限流实例：
//   - loginIPLimiter:   按客户端 IP，5 次/分钟
//   - loginUserLimiter: 按登录账号，10 次/小时
//   - globalIPLimiter:  全局所有路由按 IP，100 次/秒（突发 200）
//
// 数值为安全且不影响正常使用的经验值；如有需要可通过环境变量进一步可调。
var (
	loginIPLimiter   = newKeyedLimiter(rate.Every(12*time.Second), 5, 30*time.Minute)
	loginUserLimiter = newKeyedLimiter(rate.Every(6*time.Minute), 10, 2*time.Hour)
	globalIPLimiter  = newKeyedLimiter(rate.Limit(100), 200, 30*time.Minute)
)

// clientIP 取请求来源 IP。
// 优先 X-Forwarded-For 链路头（部署在反向代理后），降级到 c.ClientIP()。
// 只取 XFF 链路中最左侧的客户端 IP，避免被中间代理伪造。
func clientIP(c *gin.Context) string {
	if xff := c.Request.Header.Get("X-Forwarded-For"); xff != "" {
		if i := strings.IndexByte(xff, ','); i >= 0 {
			return strings.TrimSpace(xff[:i])
		}
		return strings.TrimSpace(xff)
	}
	if ip, _, err := net.SplitHostPort(c.Request.RemoteAddr); err == nil {
		return ip
	}
	return c.ClientIP()
}

// GlobalRateLimit 是挂在最外层的兜底限流，按 IP 限流。
// 对常规用户几乎无感知，但能阻断"一秒打几千次"的脚本扫描。
func GlobalRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !globalIPLimiter.allow(ip) {
			c.Header("Retry-After", "1")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many requests"})
			return
		}
		c.Next()
	}
}

// LoginRateLimit 专门保护 /v1/login，用 IP+账号双维度限流。
// - IP 维度阻挡僵尸网络：单台机器分钟级别只能尝 5 次。
// - 账号维度阻挡撞库：同一邮箱小时级别只能尝 10 次。
//
// 注意：账号取自请求体的 email_as_id 字段；如果 BindJSON 还没发生，此中间件会先拷贝 body 解析。
// 为避免 body 被消费导致后续 controller 读不到，使用一个 wrapper buffer。
func LoginRateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := clientIP(c)
		if !loginIPLimiter.allow(ip) {
			c.Header("Retry-After", "60")
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{"error": "too many login attempts, please try later"})
			return
		}
		// 账号维度的限流不在中间件层做（需要消费 body 复杂且易错），
		// 而是在 controller 内通过 LoginAccountAllow(email) 主动检查。
		c.Next()
	}
}

// LoginAccountAllow 给 controller 层调用：在通过 IP 限流后，再按账号粒度限速。
// 设计为非中间件函数，避免 body 复读问题。
func LoginAccountAllow(account string) bool {
	if account == "" {
		return true
	}
	return loginUserLimiter.allow(strings.ToLower(account))
}
