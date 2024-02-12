package main

import (
	"flag"
	"log/slog"
	"os"
	"time"

	example "github.com/egoodhall/nrpc/example/go"
	"github.com/egoodhall/nrpc/go/pkg/nrpc"
	"github.com/nats-io/nats.go"
)

var (
	message = "Hello world!"
)
var (
	hashNamespace bool
	namespace     bool
	timeout       bool
)

func main() {
	flag.BoolVar(&namespace, "ns", false, "")
	flag.BoolVar(&hashNamespace, "hns", false, "")
	flag.BoolVar(&timeout, "tmo", false, "")
	flag.Parse()

	// Parse options from cmd line flags
	options := make([]nrpc.ClientOption, 0)
	if hashNamespace {
		options = append(options, nrpc.HashNamespace("a", "b", "c"))
	} else if namespace {
		options = append(options, nrpc.Namespace("name.space"))
	}
	if timeout {
		options = append(options, nrpc.Timeout(10*time.Second))
	}

	// Connect to NATS
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		slog.Error("Couldn't connect to NATS", "error", err)
		os.Exit(1)
	}

	// Create a client
	client, err := example.NewEchoServiceClient(conn, options...)
	if err != nil {
		slog.Error("Couldn't start RPC server", "error", err)
		os.Exit(1)
	}

	// Send a request to the service
	slog.Info("Sending echo request", "message", message)
	response, err := client.Echo(&example.EchoRequest{Message: message})
	if err != nil {
		slog.Error("Received an error", "error", err)
		os.Exit(1)
	}
	slog.Info("Received response", "message", response.Message)
}
