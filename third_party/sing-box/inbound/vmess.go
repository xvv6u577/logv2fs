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
	"github.com/sagernet/sing-vmess"
	"github.com/sagernet/sing-vmess/packetaddr"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	F "github.com/sagernet/sing/common/format"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"
	"github.com/sagernet/sing/common/ntp"
)

var (
	_ adapter.Inbound           = (*VMess)(nil)
	_ adapter.InjectableInbound = (*VMess)(nil)
)

type VMess struct {
	myInboundAdapter
	ctx       context.Context
	service   *vmess.Service[int]
	users     []option.VMessUser
	tlsConfig tls.ServerConfig
	transport adapter.V2RayServerTransport

	// usersAccess 保护 users 切片与底层 service 用户表的运行时变更
	// （patch by logv2fs，用于支持 AddUser / RemoveUser）。
	usersAccess sync.RWMutex
}

func NewVMess(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.VMessInboundOptions) (*VMess, error) {
	inbound := &VMess{
		myInboundAdapter: myInboundAdapter{
			protocol:      C.TypeVMess,
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
	var serviceOptions []vmess.ServiceOption
	if timeFunc := ntp.TimeFuncFromContext(ctx); timeFunc != nil {
		serviceOptions = append(serviceOptions, vmess.ServiceWithTimeFunc(timeFunc))
	}
	if options.Transport != nil && options.Transport.Type != "" {
		serviceOptions = append(serviceOptions, vmess.ServiceWithDisableHeaderProtection())
	}
	service := vmess.NewService[int](adapter.NewUpstreamContextHandler(inbound.newConnection, inbound.newPacketConnection, inbound), serviceOptions...)
	inbound.service = service
	err = service.UpdateUsers(common.MapIndexed(options.Users, func(index int, it option.VMessUser) int {
		return index
	}), common.Map(options.Users, func(it option.VMessUser) string {
		return it.UUID
	}), common.Map(options.Users, func(it option.VMessUser) int {
		return it.AlterId
	}))
	if err != nil {
		return nil, err
	}
	if options.TLS != nil {
		inbound.tlsConfig, err = tls.NewServer(ctx, logger, common.PtrValueOrDefault(options.TLS))
		if err != nil {
			return nil, err
		}
	}
	if options.Transport != nil {
		inbound.transport, err = v2ray.NewServerTransport(ctx, common.PtrValueOrDefault(options.Transport), inbound.tlsConfig, (*vmessTransportHandler)(inbound))
		if err != nil {
			return nil, E.Cause(err, "create server transport: ", options.Transport.Type)
		}
	}
	inbound.connHandler = inbound
	return inbound, nil
}

func (h *VMess) Start() error {
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

func (h *VMess) Close() error {
	return common.Close(
		h.service,
		&h.myInboundAdapter,
		h.tlsConfig,
		h.transport,
	)
}

func (h *VMess) newTransportConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	h.injectTCP(conn, metadata)
	return nil
}

func (h *VMess) NewConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	var err error
	if h.tlsConfig != nil && h.transport == nil {
		conn, err = tls.ServerHandshake(ctx, conn, h.tlsConfig)
		if err != nil {
			return err
		}
	}
	return h.service.NewConnection(adapter.WithContext(log.ContextWithNewID(ctx), &metadata), conn, adapter.UpstreamMetadata(metadata))
}

func (h *VMess) NewPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	return os.ErrInvalid
}

func (h *VMess) newConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
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
	h.logger.InfoContext(ctx, "[", user, "] inbound connection to ", metadata.Destination)
	return h.router.RouteConnection(ctx, conn, metadata)
}

func (h *VMess) newPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
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

// AddUser 运行时追加一个 VMess 用户（patch by logv2fs）。语义见 VLESS.AddUser。
func (h *VMess) AddUser(user option.VMessUser) error {
	h.usersAccess.Lock()
	defer h.usersAccess.Unlock()

	for i := range h.users {
		if h.users[i].Name == user.Name && h.users[i].UUID == "" {
			h.users[i] = user
			return h.rebuildServiceLocked()
		}
	}
	h.users = append(h.users, user)
	return h.rebuildServiceLocked()
}

// RemoveUser 运行时按 Name 移除一个 VMess 用户（patch by logv2fs）。
// 以 UUID 置空作为墓碑标记，保持索引稳定，幂等。
func (h *VMess) RemoveUser(name string) error {
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

// rebuildServiceLocked 重建底层 vmess.Service 的用户表，调用方持有写锁。
func (h *VMess) rebuildServiceLocked() error {
	indices := make([]int, 0, len(h.users))
	uuids := make([]string, 0, len(h.users))
	alterIDs := make([]int, 0, len(h.users))
	for i, u := range h.users {
		if u.UUID == "" {
			continue
		}
		indices = append(indices, i)
		uuids = append(uuids, u.UUID)
		alterIDs = append(alterIDs, u.AlterId)
	}
	return h.service.UpdateUsers(indices, uuids, alterIDs)
}

var _ adapter.V2RayServerTransportHandler = (*vmessTransportHandler)(nil)

type vmessTransportHandler VMess

func (t *vmessTransportHandler) NewConnection(ctx context.Context, conn net.Conn, metadata M.Metadata) error {
	return (*VMess)(t).newTransportConnection(ctx, conn, adapter.InboundContext{
		Source:      metadata.Source,
		Destination: metadata.Destination,
	})
}
