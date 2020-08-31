package collector

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	"github.com/go-kit/kit/log"
	level "github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/transport"
	amqptransport "github.com/go-kit/kit/transport/amqp"
	httptransport "github.com/go-kit/kit/transport/http"
	"github.com/gorilla/mux"
	amqp "github.com/streadway/amqp"
)

// NewHTTPHandler returns an HTTP handler that makes a set of endpoints
// available on predefined paths.
func NewHTTPHandler(e Endpoint, logger log.Logger) http.Handler {
	opts := []httptransport.ServerOption{
		httptransport.ServerErrorHandler(transport.NewLogErrorHandler(logger)),
		// httptransport.ServerErrorEncoder(errorEncoder),
	}

	callbackHandler := httptransport.NewServer(
		e.ListenToEventsEndpoint,
		decodeListenToEventsRequest,
		encodeListenToEventsResponse,
		opts..., // append(options, httptransport.ServerBefore(opentracing.HTTPToContext(otTracer, "GetAccount", logger)))...,
	)

	r := mux.NewRouter()

	r.Handle("/callback", callbackHandler).Methods("GET")

	return r
}

// decodeListenToEventsRequest is a transport/http.DecodeRequestFunc that decodes a
// JSON-encoded GetKeyAccounts request from the HTTP request body. Primarily useful in a
// server.
func decodeListenToEventsRequest(_ context.Context, r *http.Request) (interface{}, error) {
	var req interface{}
	fmt.Println(r.Body)
	err := json.NewDecoder(r.Body).Decode(&req)
	fmt.Println(req)
	return req, err
}

// encodeListenToEventsResponse is a transport/http.EncodeResponseFunc that encodes
// the response as JSON to the response writer. Primarily useful in a server.
func encodeListenToEventsResponse(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	if f, ok := response.(endpoint.Failer); ok && f.Failed() != nil {
		// errorEncoder(ctx, f.Failed(), w)
		return nil
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	return json.NewEncoder(w).Encode(response)
}

// Subscriber listens to the messages from queue
type Subscriber interface {
	ListenToSubscriptions()
}

type subscriber struct {
	channel                       *amqp.Channel
	streamerSubscriber            *amqptransport.Subscriber
	subscribeToStreamerMsgChannel <-chan amqp.Delivery
	logger                        log.Logger
}

// NewSubscriber returns new subscriber
func NewSubscriber(
	conn *amqp.Connection,
	e Endpoint,
	logger log.Logger,
) Subscriber {
	ch, err := conn.Channel()
	if err != nil {
		level.Error(logger).Log(err)
	}

	subscribeToStreamerQueue, err := ch.QueueDeclare(
		"collector_subscribe",
		false, // durable
		false, // autoDelete
		false, // exclusive
		false, // noWait
		nil,   //args
	)

	subscribeToStreamerMsgs, err := ch.Consume(
		subscribeToStreamerQueue.Name,
		"",    // consumer
		false, // autoAck
		false, // exclusive
		false, // noLocal
		false, // noWait
		nil,   // args
	)

	subscribeToStreamerAMQPHandler := amqptransport.NewSubscriber(
		e.SubscribeEndpoint,
		decodeSubscribeToStreamerAMQPRequest,
		amqptransport.EncodeJSONResponse,
	)

	return &subscriber{
		channel:                       ch,
		streamerSubscriber:            subscribeToStreamerAMQPHandler,
		subscribeToStreamerMsgChannel: subscribeToStreamerMsgs,
		logger:                        logger,
	}
}

func (s *subscriber) ListenToSubscriptions() {
	s.logger.Log("method", "ListenToSubscriptions", "event", "Started listening to subscriptions")

	subscribeToStreamerListener := s.streamerSubscriber.ServeDelivery(s.channel)

	forever := make(chan bool)

	go func(logger log.Logger) {
		for true {
			select {
			case subscribeToStreamerDeliv := <-s.subscribeToStreamerMsgChannel:
				logger.Log("metod", "ListenToSubscriptions", "event", "Received create user account request")
				subscribeToStreamerListener(&subscribeToStreamerDeliv)
				subscribeToStreamerDeliv.Ack(false) // multiple = false
			}
		}
	}(s.logger)

	<-forever
}

func decodeSubscribeToStreamerAMQPRequest(ctx context.Context, delivery *amqp.Delivery) (interface{}, error) {
	var request SubscribeToStreamer
	err := json.Unmarshal(delivery.Body, &request)

	fmt.Println(err, request)

	if err != nil {
		return nil, err
	}
	return request, nil
}
