package abstract

import (
	"github.com/TixiaOTA/gokit/types"
	"github.com/gofiber/fiber/v2"
	"google.golang.org/grpc"
)

// RestHandler abstraction for REST Handler
type RestHandler interface {
	Router(r fiber.Router)
}

// GRPCHandler abstraction for gRPC Handler
type GRPCHandler interface {
	Register(srv *grpc.Server)
}

// BrokerHandler abstraction for worker handler
type BrokerHandler interface {
	Register(broker *types.BrokerHandlerGroup)
}
