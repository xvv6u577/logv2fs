package singbox

// usermgr.go：把数据库里的 model.UserTrafficLogs 翻译成对 *box.Box 上每个
// VLESS / VMess / Trojan / Hysteria2 inbound 的 AddUser / RemoveUser 调用，
// 同时维护 V2RayAPI StatsService 的流量统计白名单。
//
// 这是"业务模型 -> sing-box 运行时"的薄翻译层；本文件不感知 HTTP、不感知数据库，
// 输入只来源于上层（control server 解析后的请求体）。
//
// 注意：Hysteria2 的分发被拆到 usermgr_hysteria2.go / usermgr_hysteria2_stub.go，
// 这两个文件以 build tag 区分；本文件只处理与 build tag 无关的 VLESS / VMess / Trojan。

import (
	"fmt"
	"log"
	"strings"

	box "github.com/sagernet/sing-box"
	"github.com/sagernet/sing-box/inbound"
	"github.com/sagernet/sing-box/option"
)

const (
	// vlessUserSuffix / hy2UserSuffix 沿用 UpdateOptionsFromMongoDB 已有的命名约定。
	// 同一个业务用户在不同协议 inbound 上会注册带不同后缀的 sing-box "用户"，
	// 这样 V2RayAPI Stats 才能按 (email, 协议) 维度区分流量计数。
	vlessUserSuffix = "-reality"
	hy2UserSuffix   = "-hysteria2"
	vmessUserSuffix = "-vmess"
	trojanUserSuffix = "-trojan"

	// vlessDefaultFlow 与 singbox/singbox_mongodb.go 中初始化时使用的 Flow 保持一致。
	vlessDefaultFlow = "xtls-rprx-vision"
)

// AddUserRequest 是控制端口/调用方传入的业务用户最小信息集合。
// 字段命名沿用 model.UserTrafficLogs 的 JSON 标签，便于跨进程直接序列化。
type AddUserRequest struct {
	EmailAsId string `json:"email_as_id"`
	UUID      string `json:"uuid"`
	UserId    string `json:"user_id"` // Hysteria2/Trojan 密码（业务约定为 MongoDB ObjectID hex）
}

// ApplyAddUser 把一个业务用户分别注册到 box.Box 持有的所有相关 inbound 与
// V2RayAPI StatsService 的统计白名单。
//
// 实现细节：
//   - 遍历 instance.Inbounds()，按 Type() 分派到具体 inbound 的 AddUser 方法。
//   - 任一 inbound 出错只记 warn 继续往下，最终错误聚合返回；保证整体尽量幂等。
//   - StatsService 的 AddUser 在协议分派之前先做，确保即使后续 inbound 出错，
//     重启 singbox 后初始化阶段也能把缺失的协议补齐。
func ApplyAddUser(instance *box.Box, req AddUserRequest) error {
	if instance == nil {
		return fmt.Errorf("sing-box instance is nil")
	}
	if req.EmailAsId == "" {
		return fmt.Errorf("email_as_id is required")
	}

	// 先把可能涉及的 stats 用户名一次性加进 V2RayAPI 白名单。
	// AddUser 是幂等的，重复调用无害。
	if stats := getStatsService(instance); stats != nil {
		stats.AddUser(req.EmailAsId + vlessUserSuffix)
		stats.AddUser(req.EmailAsId + hy2UserSuffix)
		stats.AddUser(req.EmailAsId + vmessUserSuffix)
		stats.AddUser(req.EmailAsId + trojanUserSuffix)
	}

	var errs []string
	for _, in := range instance.Inbounds() {
		switch v := in.(type) {
		case *inbound.VLESS:
			if req.UUID == "" {
				log.Printf("[singbox.usermgr] skip VLESS inbound %q for %s: UUID empty",
					in.Tag(), req.EmailAsId)
				continue
			}
			err := v.AddUser(option.VLESSUser{
				Name: req.EmailAsId + vlessUserSuffix,
				UUID: req.UUID,
				Flow: vlessDefaultFlow,
			})
			if err != nil {
				errs = append(errs, fmt.Sprintf("vless[%s]: %v", in.Tag(), err))
			}

		case *inbound.VMess:
			if req.UUID == "" {
				continue
			}
			err := v.AddUser(option.VMessUser{
				Name: req.EmailAsId + vmessUserSuffix,
				UUID: req.UUID,
			})
			if err != nil {
				errs = append(errs, fmt.Sprintf("vmess[%s]: %v", in.Tag(), err))
			}

		case *inbound.Trojan:
			if req.UserId == "" {
				continue
			}
			err := v.AddUser(option.TrojanUser{
				Name:     req.EmailAsId + trojanUserSuffix,
				Password: req.UserId,
			})
			if err != nil {
				errs = append(errs, fmt.Sprintf("trojan[%s]: %v", in.Tag(), err))
			}

		default:
			// Hysteria2 由 hy2 build tag 文件兜底处理，避免无 with_quic 时编译失败。
			if err := applyAddHysteria2(in, req); err != nil {
				errs = append(errs, fmt.Sprintf("hysteria2[%s]: %v", in.Tag(), err))
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("ApplyAddUser partial failure: %s", strings.Join(errs, "; "))
	}
	log.Printf("[singbox.usermgr] AddUser ok: %s", req.EmailAsId)
	return nil
}

// ApplyRemoveUser 把一个业务用户从所有 inbound 与 StatsService 移除。
//
// 注意：移除 stats 白名单后已积累的计数器不会被立即清除，但因为定时任务每 15
// 分钟带 Reset_=true 调一次 QueryStats，该用户的残留计数最终会被自然回收。
func ApplyRemoveUser(instance *box.Box, emailAsId string) error {
	if instance == nil {
		return fmt.Errorf("sing-box instance is nil")
	}
	if emailAsId == "" {
		return fmt.Errorf("email_as_id is required")
	}

	var errs []string
	for _, in := range instance.Inbounds() {
		switch v := in.(type) {
		case *inbound.VLESS:
			if err := v.RemoveUser(emailAsId + vlessUserSuffix); err != nil {
				errs = append(errs, fmt.Sprintf("vless[%s]: %v", in.Tag(), err))
			}
		case *inbound.VMess:
			if err := v.RemoveUser(emailAsId + vmessUserSuffix); err != nil {
				errs = append(errs, fmt.Sprintf("vmess[%s]: %v", in.Tag(), err))
			}
		case *inbound.Trojan:
			if err := v.RemoveUser(emailAsId + trojanUserSuffix); err != nil {
				errs = append(errs, fmt.Sprintf("trojan[%s]: %v", in.Tag(), err))
			}
		default:
			if err := applyRemoveHysteria2(in, emailAsId); err != nil {
				errs = append(errs, fmt.Sprintf("hysteria2[%s]: %v", in.Tag(), err))
			}
		}
	}

	if stats := getStatsService(instance); stats != nil {
		stats.RemoveUser(emailAsId + vlessUserSuffix)
		stats.RemoveUser(emailAsId + hy2UserSuffix)
		stats.RemoveUser(emailAsId + vmessUserSuffix)
		stats.RemoveUser(emailAsId + trojanUserSuffix)
	}

	if len(errs) > 0 {
		return fmt.Errorf("ApplyRemoveUser partial failure: %s", strings.Join(errs, "; "))
	}
	log.Printf("[singbox.usermgr] RemoveUser ok: %s", emailAsId)
	return nil
}

// getStatsService 安全拿到 V2RayAPI 的 StatsService；任意一层为 nil 都直接返回 nil
// 让上层逻辑选择跳过。logv2fs 已在 adapter.V2RayStatsService 接口上加上
// AddUser/RemoveUser，调用方无需做具体类型断言。
func getStatsService(instance *box.Box) interface {
	AddUser(name string)
	RemoveUser(name string)
} {
	router := instance.Router()
	if router == nil {
		return nil
	}
	v2rayServer := router.V2RayServer()
	if v2rayServer == nil {
		return nil
	}
	stats := v2rayServer.StatsService()
	if stats == nil {
		return nil
	}
	return stats
}
