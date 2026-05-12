package inbound

import (
	"context"
	"net"
	"os"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/mux"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/common/uot"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-box/transport/v2ray"
	"github.com/sagernet/sing-box/transport/vless"
	"github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing-vmess/packetaddr"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
)

var (
	_ adapter.Inbound           = (*VLESS)(nil)
	_ adapter.InjectableInbound = (*VLESS)(nil)
)

type VLESS struct {
	myInboundAdapter
	ctx       context.Context
	users     []option.VLESSUser
	service   *vless.Service[int]
	tlsConfig tls.ServerConfig
	transport adapter.V2RayServerTransport

	// usersAccess 保护 users 切片与底层 service 用户表的运行时变更
	// （patch by logv2fs，用于支持 AddUser / RemoveUser）。
	usersAccess sync.RWMutex
}

func NewVLESS(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.VLESSInboundOptions) (*VLESS, error) {
	inbound := &VLESS{
		myInboundAdapter: myInboundAdapter{
			protocol:      C.TypeVLESS,
			network:       []string{N.NetworkTCP},
			ctx:           ctx,
			router:        uot.NewRouter(router, logger),
			logger:        logger,
			tag:           tag,
			listenOptions: options.ListenOptions,
		},
		ctx:   ctx,
		users: options.Users,
	}
	var err error
	inbound.router, err = mux.NewRouterWithOptions(inbound.router, logger, common.PtrValueOrDefault(options.Multiplex))
	if err != nil {
		return nil, err
	}
	service := vless.NewService[int](logger, adapter.NewUpstreamContextHandler(inbound.newConnection, inbound.newPacketConnection, inbound))
	service.UpdateUsers(common.MapIndexed(inbound.users, func(index int, _ option.VLESSUser) int {
		return index
	}), common.Map(inbound.users, func(it option.VLESSUser) string {
		return it.UUID
	}), common.Map(inbound.users, func(it option.VLESSUser) string {
		return it.Flow
	}))
	inbound.service = service
	if options.TLS != nil {
		inbound.tlsConfig, err = tls.NewServer(ctx, logger, common.PtrValueOrDefault(options.TLS))
		if err != nil {
			return nil, err
		}
	}
	if options.Transport != nil {
		inbound.transport, err = v2ray.NewServerTransport(ctx, common.PtrValueOrDefault(options.Transport), inbound.tlsConfig, (*vlessTransportHandler)(inbound))
		if err != nil {
			return nil, E.Cause(err, "create server transport: ", options.Transport.Type)
		}
	}
	inbound.connHandler = inbound
	return inbound, nil
}

func (h *VLESS) Start() error {
	err := common.Start(
		h.service,
		h.tlsConfig,
	)
	if err != nil {
		return err
	}
	if h.transport == nil {
		return h.myInboundAdapter.Start()
	}
	if common.Contains(h.transport.Network(), N.NetworkTCP) {
		tcpListener, err := h.myInboundAdapter.ListenTCP()
		if err != nil {
			return err
		}
		go func() {
			sErr := h.transport.Serve(tcpListener)
			if sErr != nil && !E.IsClosed(sErr) {
				h.logger.Error("transport serve error: ", sErr)
			}
		}()
	}
	if common.Contains(h.transport.Network(), N.NetworkUDP) {
		udpConn, err := h.myInboundAdapter.ListenUDP()
		if err != nil {
			return err
		}
		go func() {
			sErr := h.transport.ServePacket(udpConn)
			if sErr != nil && !E.IsClosed(sErr) {
				h.logger.Error("transport serve error: ", sErr)
			}
		}()
	}
	return nil
}

func (h *VLESS) Close() error {
	return common.Close(
		h.service,
		&h.myInboundAdapter,
		h.tlsConfig,
		h.transport,
	)
}

func (h *VLESS) newTransportConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	h.injectTCP(conn, metadata)
	return nil
}

func (h *VLESS) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	var err error
	if h.tlsConfig != nil && h.transport == nil {
		conn, err = tls.ServerHandshake(ctx, conn, h.tlsConfig)
		if err != nil {
			return err
		}
	}
	return h.service.NewConnection(adapter.WithContext(log.ContextWithNewID(ctx), &metadata), conn, adapter.UpstreamMetadata(metadata))
}

func (h *VLESS) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	return os.ErrInvalid
}

func (h *VLESS) newConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	userIndex, loaded := auth.UserFromContext[int](ctx)
	if !loaded {
		return os.ErrInvalid
	}
	// usersAccess 加读锁，防止与运行时 AddUser/RemoveUser 竞争（patch by logv2fs）
	h.usersAccess.RLock()
	user := h.users[userIndex].Name
	h.usersAccess.RUnlock()
	if user == "" {
		user = F.ToString(userIndex)
	} else {
		metadata.User = user
	}
	h.logger.InfoContext(ctx, "[", user, "] inbound connection to ", metadata.Destination)
	return h.router.RouteConnection(ctx, conn, metadata)
}

func (h *VLESS) newPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	userIndex, loaded := auth.UserFromContext[int](ctx)
	if !loaded {
		return os.ErrInvalid
	}
	h.usersAccess.RLock()
	user := h.users[userIndex].Name
	h.usersAccess.RUnlock()
	if user == "" {
		user = F.ToString(userIndex)
	} else {
		metadata.User = user
	}
	if metadata.Destination.Fqdn == packetaddr.SeqPacketMagicAddress {
		metadata.Destination = M.Socksaddr{}
		conn = packetaddr.NewConn(conn.(vmess.PacketConn), metadata.Destination)
		h.logger.InfoContext(ctx, "[", user, "] inbound packet addr connection")
	} else {
		h.logger.InfoContext(ctx, "[", user, "] inbound packet connection to ", metadata.Destination)
	}
	return h.router.RoutePacketConnection(ctx, conn, metadata)
}

// AddUser 运行时追加一个 VLESS 用户（patch by logv2fs）。
//
// 设计要点：
//   - 采用 append-only 语义：新用户加在切片末尾，已存在用户索引永不变更，
//     防止已建立连接持有的 userIndex 在 service map 重建后越界或指向错误用户。
//   - 同名用户视作幂等：直接覆盖第一个未被墓碑化的同名 slot 的 UUID/Flow，
//     如全部同名 slot 都已墓碑化则新追加一个。
func (h *VLESS) AddUser(user option.VLESSUser) error {
	h.usersAccess.Lock()
	defer h.usersAccess.Unlock()

	// 复用历史墓碑 slot（同名且 UUID 被清空过的），避免无限增长
	for i := range h.users {
		if h.users[i].Name == user.Name && h.users[i].UUID == "" {
			h.users[i] = user
			return h.rebuildServiceLocked()
		}
	}
	h.users = append(h.users, user)
	return h.rebuildServiceLocked()
}

// RemoveUser 运行时按 Name 移除一个 VLESS 用户（patch by logv2fs）。
//
// 实现以 "墓碑" 方式工作：保留 slot 本身（维持索引稳定），仅把认证字段 UUID
// 清空，从而：1) 新发起的连接因 UUID 不在 userMap 而被拒；2) 旧连接的
// metadata.User 仍可正常打日志；3) service map 重建时跳过墓碑。
//
// 若找不到对应 Name 返回 nil（幂等），便于上层 retry。
func (h *VLESS) RemoveUser(name string) error {
	h.usersAccess.Lock()
	defer h.usersAccess.Unlock()

	hit := false
	for i := range h.users {
		if h.users[i].Name == name && h.users[i].UUID != "" {
			h.users[i].UUID = ""
			hit = true
		}
	}
	if !hit {
		return nil
	}
	return h.rebuildServiceLocked()
}

// rebuildServiceLocked 用当前 users 切片重建底层 vless.Service 的用户表。
// 调用方必须持有 usersAccess 写锁。
func (h *VLESS) rebuildServiceLocked() error {
	indices := make([]int, 0, len(h.users))
	uuids := make([]string, 0, len(h.users))
	flows := make([]string, 0, len(h.users))
	for i, u := range h.users {
		if u.UUID == "" { // tombstone
			continue
		}
		indices = append(indices, i)
		uuids = append(uuids, u.UUID)
		flows = append(flows, u.Flow)
	}
	h.service.UpdateUsers(indices, uuids, flows)
	return nil
}

var _ adapter.V2RayServerTransportHandler = (*vlessTransportHandler)(nil)

type vlessTransportHandler VLESS

func (t *vlessTransportHandler) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	return (*VLESS)(t).newTransportConnection(ctx, conn, adapter.InboundContext{
		Source:      metadata.Source,
		Destination: metadata.Destination,
	})
}
