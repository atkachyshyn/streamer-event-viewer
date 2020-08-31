package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"

	"github.com/atkachyshyn/streamer-event-viewer/collector"
	"github.com/go-kit/kit/log"
	"github.com/streadway/amqp"
)

func main() {
	var (
		listen = flag.String("listen", ":8080", "HTTP listen address")
	)
	flag.Parse()

	var logger log.Logger
	{
		logger = log.NewLogfmtLogger(os.Stderr)
		logger = log.With(logger, "listen", *listen, "caller", log.DefaultCaller)
	}

	// connect to AMQP
	amqpURL := flag.String(
		"url",
		"amqp://user:guest@localhost:5672",
		"URL to AMQP server",
	)

	conn, err := amqp.Dial(*amqpURL)
	if err != nil {
		fmt.Println(err)
		logger.Log(err)
	}
	defer conn.Close()

	var (
		service     = collector.NewService(logger)
		endpoints   = collector.NewEndpoints(service, logger)
		subscriber  = collector.NewSubscriber(conn, endpoints, logger)
		httpHandler = collector.NewHTTPHandler(endpoints, logger)
	)

	subscriber.ListenToSubscriptions()

	logger.Log("transport", "HTTP", "addr", *listen)
	httpListener, err := net.Listen("tcp", *listen)
	if err != nil {
		logger.Log("transport", "HTTP", "during", "Listen", "err", err)
		os.Exit(1)
	}

	// go func() {
	http.Serve(httpListener, httpHandler)
	// }()
}
