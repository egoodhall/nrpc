package main

import (
	"context"
	"flag"
	"log/slog"
	"os"
	"os/signal"

	example "github.com/egoodhall/nrpc/example/go"
	"github.com/egoodhall/nrpc/go/pkg/nrpc"
	"github.com/nats-io/nats.go"
)

var (
	hashNamespace bool
	namespace     bool
	errHandler    bool
)

func main() {
	flag.BoolVar(&namespace, "ns", false, "")
	flag.BoolVar(&hashNamespace, "hns", false, "")
	flag.BoolVar(&hashNamespace, "err", false, "")
	flag.Parse()

	// Parse options from cmd line flags
	options := make([]nrpc.ServerOption, 0)
	if hashNamespace {
		options = append(options, nrpc.HashNamespace("a", "b", "c"))
	} else if namespace {
		options = append(options, nrpc.Namespace("name.space"))
	}
	if errHandler {
		options = append(options, nrpc.ErrorHandler(func(err error) {
			slog.Error("An error occurred", "error", err)
		}))
	}

	// Connect to NATS
	conn, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		slog.Error("Couldn't connect to NATS", "error", err)
		os.Exit(1)
	}

	// Create our server
	srv, err := example.NewEchoServiceServer(conn, new(example.ServiceImpl), options...)
	if err != nil {
		slog.Error("Couldn't start RPC server", "error", err)
		os.Exit(1)
	}
	defer srv.Stop()
	slog.Info("Server started", "info", srv.Info())

	// Wait for an interrupt to stop the server
	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	<-ctx.Done()
}
