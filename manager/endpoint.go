package manager

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Endpoint definition
type Endpoint struct {
	SubscribeEndpoint endpoint.Endpoint
}

// NewEndpoints returns a Set that wraps the provided service, and wires in all of the
// expected endpoint middlewares via the various parameters.
func NewEndpoints(s Service, l log.Logger) Endpoint {
	var subscribeEndpoint endpoint.Endpoint
	{
		subscribeEndpoint = MakeSubscribeEndpoint(s)
	}

	return Endpoint{
		SubscribeEndpoint: subscribeEndpoint,
	}
}

// SubscribeToStreamer describes interface how to subscribe to streamer
type SubscribeToStreamer struct {
	mode   string
	topic  string
	secret string
}

// MakeSubscribeEndpoint constructs a Subscribe endpoint wrapping the service.
func MakeSubscribeEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SubscribeRequest)
		err = s.Subscribe(ctx, req.streamer, req.authorization)
		response = true
		return
	}
}

// SubscribeRequest request to subscribe
type SubscribeRequest struct {
	streamer      string
	authorization string
}

// SubscribeMsg describes message how to subscribe to topic
type SubscribeMsg struct {
	mode   string
	topic  string
	secret string
}
