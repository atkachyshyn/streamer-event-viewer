package manager

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/transport"
	amqptransport "github.com/go-kit/kit/transport/amqp"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/streadway/amqp"
	"golang.org/x/oauth2"

	"github.com/atkachyshyn/streamer-event-viewer/shared"
)

var (
	errBadRequest = errors.New("bad request type")
)

// NewHTTPHandler returns an HTTP handler that makes a set of endpoints
// available on predefined paths.
func NewHTTPHandler(e Endpoint, h AuthorizationHandler, l log.Logger) http.Handler {
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(l)),
		// httptransport.ServerErrorEncoder(errorEncoder),
	}

	subscribeHandler := httptransport.NewServer(
		e.SubscribeEndpoint,
		decodeSubscribeRequest,
		encodeSubscribeResponse,
		opts..., // append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetAccount", logger)))...,
	)

	m := http.NewServeMux()
	m.Handle("/login", h.LoginHandler)
	m.Handle("/redirect", h.OAuth2CallbackHandler)
	m.Handle("/subscribe", accessControl(subscribeHandler))
	return m
}

func accessControl(h http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Origin, Content-Type")

		if r.Method == "OPTIONS" {
			return
		}

		h.ServeHTTP(w, r)
	})
}

// Publisher publishes messages into queue
type Publisher interface {
	PublishSubscribeToStreamerMsg(ctx context.Context, msg SubscribeMsg) (response interface{}, err error)
}

type publisher struct {
	endpoint endpoint.Endpoint
	logger   log.Logger
}

// NewPublisher returns new publisher
func NewPublisher(conn *amqp.Connection, logger log.Logger) Publisher {
	ch, err := conn.Channel()
	if err != nil {
		level.Error(logger).Log(err)
	}

	responseQueue := "manager_publisher"

	publisherQueue, err := ch.QueueDeclare(
		responseQueue,
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   //args
	)

	subscribeToStreamerPublisher := amqptransport.NewPublisher(
		ch,
		&publisherQueue,
		encodeSubscribeToStreamerAMQPRequest,
		decodeSubscribeToStreamerAMQPResponse,
		amqptransport.PublisherBefore(
			// queue name specified by subscriber
			amqptransport.SetPublishKey("collector_subscribe"),
		),
	)

	subscribeToStreamerEndpoint := subscribeToStreamerPublisher.Endpoint()

	return &publisher{
		endpoint: subscribeToStreamerEndpoint,
		logger:   logger,
	}
}

func (p *publisher) PublishSubscribeToStreamerMsg(ctx context.Context, msg SubscribeMsg) (response interface{}, err error) {
	return p.endpoint(ctx, msg)
}

func encodeSubscribeToStreamerAMQPRequest(ctx context.Context, publishing *amqp.Publishing, request interface{}) error {
	req, ok := request.(*SubscribeToStreamer)
	// fmt.Println("encodeSubscribeToStreamerAMQPRequest", req, ok)
	if !ok {
		return errBadRequest
	}
	b, err := json.Marshal(req)
	if err != nil {
		return err
	}
	publishing.Body = b
	return nil
}

func decodeSubscribeToStreamerAMQPResponse(ctx context.Context, delivery *amqp.Delivery) (interface{}, error) {
	var response bool
	err := json.Unmarshal(delivery.Body, &response)
	if err != nil {
		return nil, err
	}
	return response, nil
}

func decodeSubscribeRequest(ctx context.Context, r *http.Request) (interface{}, error) {
	session, err := shared.CookieStore.Get(r, oauthSessionName)
	if err != nil {
		fmt.Println("decodeSubscribeRequest", err.Error())
	}

	fmt.Println("decodeSubscribeRequest", session.Values)

	token := session.Values[oauthTokenKey].(*oauth2.Token)

	fmt.Println("decodeSubscribeRequest", "token", token)

	var req SubscribeRequest
	err = json.NewDecoder(r.Body).Decode(&req)
	req.authorization = token.TokenType + " " + token.AccessToken
	return req, err
}

func encodeSubscribeResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		// errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}
