package abstract

import (
	"context"

	"github.com/TixiaOTA/gokit/types"
)

// Publisher message broker want to publish their message
type Publisher interface {
	PublishMessage(ctx context.Context, req types.PublisherArgument) error
}
