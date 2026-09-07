package anytls

import (
	"context"
	"net"
	"strings"
	"sync"

	"github.com/sagernet/sing-box/adapter"
	"github.com/sagernet/sing-box/adapter/inbound"
	"github.com/sagernet/sing-box/common/listener"
	"github.com/sagernet/sing-box/common/tls"
	"github.com/sagernet/sing-box/common/uot"
	C "github.com/sagernet/sing-box/constant"
	"github.com/sagernet/sing-box/log"
	"github.com/sagernet/sing-box/option"
	"github.com/sagernet/sing/common"
	"github.com/sagernet/sing/common/auth"
	E "github.com/sagernet/sing/common/exceptions"
	"github.com/sagernet/sing/common/logger"
	M "github.com/sagernet/sing/common/metadata"
	N "github.com/sagernet/sing/common/network"

	anytls "github.com/anytls/sing-anytls"
	"github.com/anytls/sing-anytls/padding"
)

func RegisterInbound(registry *inbound.Registry) {
	inbound.Register[option.AnyTLSInboundOptions](registry, C.TypeAnyTLS, NewInbound)
}

type Inbound struct {
	inbound.Adapter
	tlsConfig     tls.ServerConfig
	router        adapter.ConnectionRouterEx
	logger        logger.ContextLogger
	listener      *listener.Listener
	service       *anytls.Service
	paddingScheme []byte
	users         []option.AnyTLSUser
	usersAccess   sync.RWMutex
}

func NewInbound(ctx context.Context, router adapter.Router, logger log.ContextLogger, tag string, options option.AnyTLSInboundOptions) (adapter.Inbound, error) {
	inbound := &Inbound{
		Adapter: inbound.NewAdapter(C.TypeAnyTLS, tag),
		router:  uot.NewRouter(router, logger),
		logger:  logger,
		users:   options.Users,
	}

	if options.TLS != nil && options.TLS.Enabled {
		tlsConfig, err := tls.NewServer(ctx, logger, common.PtrValueOrDefault(options.TLS))
		if err != nil {
			return nil, err
		}
		inbound.tlsConfig = tlsConfig
	}

	paddingScheme := padding.DefaultPaddingScheme
	if len(options.PaddingScheme) > 0 {
		paddingScheme = []byte(strings.Join(options.PaddingScheme, "\n"))
	}
	inbound.paddingScheme = paddingScheme

	if err := inbound.rebuildServiceLocked(); err != nil {
		return nil, err
	}
	inbound.listener = listener.New(listener.Options{
		Context:           ctx,
		Logger:            logger,
		Network:           []string{N.NetworkTCP},
		Listen:            options.ListenOptions,
		ConnectionHandler: inbound,
	})
	return inbound, nil
}

func (h *Inbound) Start(stage adapter.StartStage) error {
	if stage != adapter.StartStateStart {
		return nil
	}
	if h.tlsConfig != nil {
		err := h.tlsConfig.Start()
		if err != nil {
			return err
		}
	}
	return h.listener.Start()
}

func (h *Inbound) Close() error {
	return common.Close(h.listener, h.tlsConfig)
}

func (h *Inbound) NewConnectionEx(ctx context.Context, conn net.Conn, metadata adapter.InboundContext, onClose N.CloseHandlerFunc) {
	if h.tlsConfig != nil {
		tlsConn, err := tls.ServerHandshake(ctx, conn, h.tlsConfig)
		if err != nil {
			N.CloseOnHandshakeFailure(conn, onClose, err)
			h.logger.ErrorContext(ctx, E.Cause(err, "process connection from ", metadata.Source, ": TLS handshake"))
			return
		}
		conn = tlsConn
	}
	h.usersAccess.RLock()
	service := h.service
	h.usersAccess.RUnlock()
	err := service.NewConnection(adapter.WithContext(ctx, &metadata), conn, metadata.Source, onClose)
	if err != nil {
		N.CloseOnHandshakeFailure(conn, onClose, err)
		h.logger.ErrorContext(ctx, E.Cause(err, "process connection from ", metadata.Source))
	}
}

// AddUser adds or restores an AnyTLS user. patch by logv2fs
//
// sing-anytls Service 没有 UpdateUsers；运行时增删通过重建 Service 完成。
// 已建立的 session 仍走旧 Service，新连接使用新用户表。
func (h *Inbound) AddUser(user option.AnyTLSUser) error {
	h.usersAccess.Lock()
	defer h.usersAccess.Unlock()
	for index := range h.users {
		if h.users[index].Name == user.Name && h.users[index].Password == "" {
			h.users[index] = user
			return h.rebuildServiceLocked()
		}
	}
	h.users = append(h.users, user)
	return h.rebuildServiceLocked()
}

// RemoveUser disables matching AnyTLS users while preserving slice indexes. patch by logv2fs
func (h *Inbound) RemoveUser(name string) error {
	h.usersAccess.Lock()
	defer h.usersAccess.Unlock()
	changed := false
	for index := range h.users {
		if h.users[index].Name == name && h.users[index].Password != "" {
			h.users[index].Password = ""
			changed = true
		}
	}
	if !changed {
		return nil
	}
	return h.rebuildServiceLocked()
}

func (h *Inbound) rebuildServiceLocked() error {
	active := make([]anytls.User, 0, len(h.users))
	for _, user := range h.users {
		if user.Password == "" {
			continue
		}
		active = append(active, anytls.User(user))
	}
	service, err := anytls.NewService(anytls.ServiceConfig{
		Users:         active,
		PaddingScheme: h.paddingScheme,
		Handler:       (*inboundHandler)(h),
		Logger:        h.logger,
	})
	if err != nil {
		return err
	}
	h.service = service
	return nil
}

type inboundHandler Inbound

func (h *inboundHandler) NewConnectionEx(ctx context.Context, conn net.Conn, source M.Socksaddr, destination M.Socksaddr, onClose N.CloseHandlerFunc) {
	var metadata adapter.InboundContext
	metadata.Inbound = h.Tag()
	metadata.InboundType = h.Type()
	//nolint:staticcheck
	metadata.InboundDetour = h.listener.ListenOptions().Detour
	//nolint:staticcheck
	metadata.InboundOptions = h.listener.ListenOptions().InboundOptions
	metadata.Source = source
	metadata.Destination = destination.Unwrap()
	if userName, _ := auth.UserFromContext[string](ctx); userName != "" {
		metadata.User = userName
		h.logger.InfoContext(ctx, "[", userName, "] inbound connection to ", metadata.Destination)
	} else {
		h.logger.InfoContext(ctx, "inbound connection to ", metadata.Destination)
	}
	h.router.RouteConnectionEx(ctx, conn, metadata, onClose)
}
