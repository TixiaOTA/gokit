package grpc

import (
	"context"
	"fmt"
	"log"
	"net"
	"time"

	"github.com/TixiaOTA/gokit/factory"
	"github.com/TixiaOTA/gokit/logger"
	"github.com/TixiaOTA/gokit/types"
	"github.com/TixiaOTA/gokit/utils/env"
	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

type rpc struct {
	opt          option
	serverEngine *grpc.Server
	listener     net.Listener
	service      factory.ServiceFactory
}

// New create a new gRPC server
func New(svc factory.ServiceFactory, opts ...OptionFunc) factory.ApplicationFactory {
	var (
		keepAliveEnforce = keepalive.EnforcementPolicy{
			MinTime:             env.GetDuration("GRPC_MIN_TIME", time.Duration(10)*time.Second),
			PermitWithoutStream: true,
		}
		keepAliveServer = keepalive.ServerParameters{
			MaxConnectionIdle:     env.GetDuration("GRPC_MAX_CONNECTION_IDLE_DURATION", time.Duration(10)*time.Second), // if a client idle for 10s, send a go away
			MaxConnectionAgeGrace: env.GetDuration("GRPC_MAX_CONNECTION_AGE_GRACE", time.Duration(10)*time.Second),     // allows 10s for pending RPCs to complete before forcibly closing connections
			Time:                  env.GetDuration("GRPC_TIME_PING_CLIENT", time.Duration(10)*time.Second),             // ping the client if it's idle for 10s to ensure the connection is still alive
			Timeout:               env.GetDuration("GRPC_TIMEOUT", time.Duration(10)*time.Second),                      // wait 10s for the ping ack before assuming the connection is dead
		}
		intercept = newInterceptor("", svc.Name()) // init intercept
	)

	// init instance
	srv := &rpc{
		service: svc,
		opt:     defaultOption(),
		serverEngine: grpc.NewServer(
			grpc.KeepaliveEnforcementPolicy(keepAliveEnforce),
			grpc.KeepaliveParams(keepAliveServer),
			grpc.UnaryInterceptor(
				intercept.chainUnaryServer(
					intercept.unaryServerTracerInterceptor,
				),
			),
		),
	}

	for _, opt := range opts {
		opt(&srv.opt)
	}

	tcpURI := srv.opt.tcpHost + ":" + srv.opt.tcpPort
	var err error
	srv.listener, err = net.Listen("tcp", tcpURI)
	if err != nil {
		panic(err)
	}

	intercept.opt = &srv.opt
	if h := srv.service.GRPCHandler(); h != nil {
		h.Register(srv.serverEngine)
	}

	for root, info := range srv.serverEngine.GetServiceInfo() {
		for _, method := range info.Methods {
			logger.Green(fmt.Sprintf("[GRPC-METHOD] /%s/%s \t\t[metadata]--> %v", root, method.Name, info.Metadata))
		}
	}

	logger.GreenBold(fmt.Sprintf("⇨ GRPC server run at %s\n", tcpURI))
	return srv
}

func (r *rpc) Serve() {
	err := r.serverEngine.Serve(r.listener)
	if err != nil {
		log.Fatal(err)
	}
}

func (r *rpc) Shutdown(_ context.Context) {
	defer logger.RedBold("Stopping GRPC Server")

	r.serverEngine.GracefulStop()
	_ = r.listener.Close()
}

func (r *rpc) Name() string {
	return types.GRPC.String()
}
