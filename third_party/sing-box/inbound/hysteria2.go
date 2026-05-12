//go:build with_quic

package inbound

import (
	"context"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"sync"
	"time"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/common/tls"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing-quic/hysteria"
	"github.com/sagernet/sing-quic/hysteria2"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	N "github.com/sagernet/sing/common/network"
)

var _ adapter.Inbound = (*Hysteria2)(nil)

type Hysteria2 struct {
	myInboundAdapter
	tlsConfig    tls.ServerConfig
	service      *hysteria2.Service[int]
	userNameList []string

	// users / usersAccess 由 logv2fs patch 引入，用于支持运行时 AddUser / RemoveUser。
	// users 切片与 userNameList 保持索引对齐：
	//   - userNameList[i] 为 ""，且 users[i].Password 为 "" 表示该 slot 已被墓碑化。
	users       []option.Hysteria2User
	usersAccess sync.RWMutex
}

func NewHysteria2(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.Hysteria2InboundOptions) (*Hysteria2, error) {
	options.UDPFragmentDefault = true
	if options.TLS == nil || !options.TLS.Enabled {
		return nil, C.ErrTLSRequired
	}
	tlsConfig, err := tls.NewServer(ctx, logger, common.PtrValueOrDefault(options.TLS))
	if err != nil {
		return nil, err
	}
	var salamanderPassword string
	if options.Obfs != nil {
		if options.Obfs.Password == "" {
			return nil, E.New("missing obfs password")
		}
		switch options.Obfs.Type {
		case hysteria2.ObfsTypeSalamander:
			salamanderPassword = options.Obfs.Password
		default:
			return nil, E.New("unknown obfs type: ", options.Obfs.Type)
		}
	}
	var masqueradeHandler http.Handler
	if options.Masquerade != "" {
		masqueradeURL, err := url.Parse(options.Masquerade)
		if err != nil {
			return nil, E.Cause(err, "parse masquerade URL")
		}
		switch masqueradeURL.Scheme {
		case "file":
			masqueradeHandler = http.FileServer(http.Dir(masqueradeURL.Path))
		case "http", "https":
			masqueradeHandler = &httputil.ReverseProxy{
				Rewrite: func(r *httputil.ProxyRequest) {
					r.SetURL(masqueradeURL)
					r.Out.Host = r.In.Host
				},
				ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
					w.WriteHeader(http.StatusBadGateway)
				},
			}
		default:
			return nil, E.New("unknown masquerade URL scheme: ", masqueradeURL.Scheme)
		}
	}
	inbound := &Hysteria2{
		myInboundAdapter: myInboundAdapter{
			protocol:      C.TypeHysteria2,
			network:       []string{N.NetworkUDP},
			ctx:           ctx,
			router:        router,
			logger:        logger,
			tag:           tag,
			listenOptions: options.ListenOptions,
		},
		tlsConfig: tlsConfig,
	}
	var udpTimeout time.Duration
	if options.UDPTimeout != 0 {
		udpTimeout = time.Duration(options.UDPTimeout)
	} else {
		udpTimeout = C.UDPTimeout
	}
	service, err := hysteria2.NewService[int](hysteria2.ServiceOptions{
		Context:               ctx,
		Logger:                logger,
		BrutalDebug:           options.BrutalDebug,
		SendBPS:               uint64(options.UpMbps * hysteria.MbpsToBps),
		ReceiveBPS:            uint64(options.DownMbps * hysteria.MbpsToBps),
		SalamanderPassword:    salamanderPassword,
		TLSConfig:             tlsConfig,
		IgnoreClientBandwidth: options.IgnoreClientBandwidth,
		UDPTimeout:            udpTimeout,
		Handler:               adapter.NewUpstreamHandler(adapter.InboundContext{}, inbound.newConnection, inbound.newPacketConnection, nil),
		MasqueradeHandler:     masqueradeHandler,
	})
	if err != nil {
		return nil, err
	}
	userList := make([]int, 0, len(options.Users))
	userNameList := make([]string, 0, len(options.Users))
	userPasswordList := make([]string, 0, len(options.Users))
	for index, user := range options.Users {
		userList = append(userList, index)
		userNameList = append(userNameList, user.Name)
		userPasswordList = append(userPasswordList, user.Password)
	}
	service.UpdateUsers(userList, userPasswordList)
	inbound.service = service
	inbound.userNameList = userNameList
	// 保留一份完整 user 列表，便于运行时 RemoveUser 后用墓碑标记重建 service
	// （patch by logv2fs）。
	inbound.users = append([]option.Hysteria2User(nil), options.Users...)
	return inbound, nil
}

func (h *Hysteria2) newConnection(ctx context.Context, conn net.Conn, metadata adapter.InboundContext) error {
	ctx = log.ContextWithNewID(ctx)
	metadata = h.createMetadata(conn, metadata)
	userID, _ := auth.UserFromContext[int](ctx)
	h.usersAccess.RLock()
	userName := h.userNameList[userID]
	h.usersAccess.RUnlock()
	if userName != "" {
		metadata.User = userName
		h.logger.InfoContext(ctx, "[", userName, "] inbound connection to ", metadata.Destination)
	} else {
		h.logger.InfoContext(ctx, "inbound connection to ", metadata.Destination)
	}
	return h.router.RouteConnection(ctx, conn, metadata)
}

func (h *Hysteria2) newPacketConnection(ctx context.Context, conn N.PacketConn, metadata adapter.InboundContext) error {
	ctx = log.ContextWithNewID(ctx)
	metadata = h.createPacketMetadata(conn, metadata)
	userID, _ := auth.UserFromContext[int](ctx)
	h.usersAccess.RLock()
	userName := h.userNameList[userID]
	h.usersAccess.RUnlock()
	if userName != "" {
		metadata.User = userName
		h.logger.InfoContext(ctx, "[", userName, "] inbound packet connection to ", metadata.Destination)
	} else {
		h.logger.InfoContext(ctx, "inbound packet connection to ", metadata.Destination)
	}
	return h.router.RoutePacketConnection(ctx, conn, metadata)
}

// AddUser 运行时追加一个 Hysteria2 用户（patch by logv2fs）。语义见 VLESS.AddUser。
func (h *Hysteria2) AddUser(user option.Hysteria2User) error {
	h.usersAccess.Lock()
	defer h.usersAccess.Unlock()

	for i := range h.users {
		if h.users[i].Name == user.Name && h.users[i].Password == "" {
			h.users[i] = user
			h.userNameList[i] = user.Name
			h.rebuildServiceLocked()
			return nil
		}
	}
	h.users = append(h.users, user)
	h.userNameList = append(h.userNameList, user.Name)
	h.rebuildServiceLocked()
	return nil
}

// RemoveUser 运行时按 Name 移除一个 Hysteria2 用户（patch by logv2fs）。
// 以 Password 置空作为墓碑标记，rebuild 时跳过；同时清空 userNameList 对应 slot，
// 避免删除后再次有相同 userID 的旧连接误打日志。幂等。
func (h *Hysteria2) RemoveUser(name string) error {
	h.usersAccess.Lock()
	defer h.usersAccess.Unlock()

	hit := false
	for i := range h.users {
		if h.users[i].Name == name && h.users[i].Password != "" {
			h.users[i].Password = ""
			h.userNameList[i] = ""
			hit = true
		}
	}
	if !hit {
		return nil
	}
	h.rebuildServiceLocked()
	return nil
}

// rebuildServiceLocked 重建底层 hysteria2.Service 的用户表，调用方持有写锁。
func (h *Hysteria2) rebuildServiceLocked() {
	indices := make([]int, 0, len(h.users))
	passwords := make([]string, 0, len(h.users))
	for i, u := range h.users {
		if u.Password == "" {
			continue
		}
		indices = append(indices, i)
		passwords = append(passwords, u.Password)
	}
	h.service.UpdateUsers(indices, passwords)
}

func (h *Hysteria2) Start() error {
	if h.tlsConfig != nil {
		err := h.tlsConfig.Start()
		if err != nil {
			return err
		}
	}
	packetConn, err := h.myInboundAdapter.ListenUDP()
	if err != nil {
		return err
	}
	return h.service.Start(packetConn)
}

func (h *Hysteria2) Close() error {
	return common.Close(
		&h.myInboundAdapter,
		h.tlsConfig,
		common.PtrOrNil(h.service),
	)
}
