package nrpc

import "github.com/nats-io/nats.go/micro"

type Server interface {
	// Info returns the service info.
	Info() micro.Info

	// Stats returns statistics for the service endpoint and all monitoring endpoints.
	Stats() micro.Stats

	// Reset resets all statistics (for all endpoints) on a service instance.
	Reset()

	// Stop drains the endpoint subscriptions and marks the service as stopped.
	Stop() error

	// Stopped informs whether [Stop] was executed on the service.
	Stopped() bool
}

type ServerFunc func() error

func (sf ServerFunc) Stop() error {
	return sf()
}
