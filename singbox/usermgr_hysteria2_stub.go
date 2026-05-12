//go:build !with_quic

package singbox

// usermgr_hysteria2_stub.go：在没有 with_quic build tag 时（如 httpserver 二进制），
// hysteria2 inbound 类型本身就不存在，分发函数退化为 no-op。
//
// 业务影响：httpserver 进程虽然链接进了 singbox 包，但它本身不创建 box.Box，
// 仅用到了 control_client 部分；无需真正分发 hy2 调用。

import "github.com/sagernet/sing-box/adapter"

func applyAddHysteria2(_ adapter.Inbound, _ AddUserRequest) error    { return nil }
func applyRemoveHysteria2(_ adapter.Inbound, _ string) error         { return nil }
