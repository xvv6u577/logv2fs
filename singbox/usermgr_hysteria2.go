//go:build with_quic

package singbox

// usermgr_hysteria2.go：Hysteria2 inbound 的 AddUser / RemoveUser 分发实现。
//
// 由于 sing-box 把 inbound.Hysteria2 类型用 //go:build with_quic 包起来了，
// 没有 with_quic 时该类型甚至不存在；因此我们必须把所有引用 inbound.Hysteria2
// 的代码也放到同样的 build tag 下，否则纯 httpserver build（默认不带 quic）
// 会编译失败。
//
// 注意：本文件仅处理 hysteria2 的类型断言；通用 nil 检查、参数校验、统计等已经
// 在 usermgr.go 的主流程里完成，本文件保持极简即可。

import (
	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/inbound"
	"github.com/sagernet/sing-box/option"
)

func applyAddHysteria2(in adapter.Inbound, req AddUserRequest) error {
	v, ok := in.(*inbound.Hysteria2)
	if !ok {
		return nil
	}
	if req.UserId == "" {
		return nil
	}
	return v.AddUser(option.Hysteria2User{
		Name:     req.EmailAsId + hy2UserSuffix,
		Password: req.UserId,
	})
}

func applyRemoveHysteria2(in adapter.Inbound, emailAsId string) error {
	v, ok := in.(*inbound.Hysteria2)
	if !ok {
		return nil
	}
	return v.RemoveUser(emailAsId + hy2UserSuffix)
}
