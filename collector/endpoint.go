package collector

import (
	"context"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
)

// Endpoint definition
type Endpoint struct {
	SubscribeEndpoint      endpoint.Endpoint
	ListenToEventsEndpoint endpoint.Endpoint
}

// NewEndpoints returns a Set that wraps the provided service, and wires in all of the
// expected endpoint middlewares via the various parameters.
func NewEndpoints(s Service, l log.Logger) Endpoint {
	var subscribeEndpoint endpoint.Endpoint
	{
		subscribeEndpoint = MakeSubscribeEndpoint(s)
		// createAccountEndpoint = opentracing.TraceServer(t, "CreateAccount")(createAccountEndpoint)
	}
	var listenToEventsEndpoint endpoint.Endpoint
	{
		listenToEventsEndpoint = MakeListenToEventsEndpoint(s)
	}

	return Endpoint{
		SubscribeEndpoint:      subscribeEndpoint,
		ListenToEventsEndpoint: listenToEventsEndpoint,
	}
}

// MakeSubscribeEndpoint constructs a Subscribe endpoint wrapping the service.
func MakeSubscribeEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		req := request.(SubscribeToStreamer)
		err = s.Subscribe(ctx, req.mode, req.topic, req.secret)
		response = true
		return
	}
}

// MakeListenToEventsEndpoint constructs a ListenToEvenrts endpoint wrapping the service.
func MakeListenToEventsEndpoint(s Service) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (response interface{}, err error) {
		// req := request.(CreateAccountRequest)
		// id, err := s.Subscribe(ctx, req.UserAccountID, req.Blockchain, req.Name, req.MnemonicVaultKey)
		// response = &CreateAccountResponse{AccountID: id}
		return
	}
}

// SubscribeToStreamer describes interface how to subscribe to streamer
type SubscribeToStreamer struct {
	mode   string
	topic  string
	secret string
}
