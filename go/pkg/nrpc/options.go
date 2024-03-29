package nrpc

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/nats-io/nats.go/micro"
)

var (
	ErrServerAlreadyStarted = errors.New("server already started")
)

// An option that can be used on both the server
// and client.
type Option interface {
	ServerOption
	ClientOption
}

func newOpt(so func(*ServerOptions) error, co func(*ClientOptions) error) Option {
	return &opt{so, co}
}

// A server configuration value
type ServerOption interface {
	setServer(*ServerOptions) error
}

// A client configuration value
type ClientOption interface {
	setClient(*ClientOptions) error
}

// Client configuration values
type ClientOptions struct {
	// Timeout sets the amount of time that a client will
	// wait for a response from the server
	Timeout time.Duration
	// Namespace will be added to the beginning of all NATS
	// subjects used by the client, effectively allowing
	// multiple servers to be run and accessed manually.
	Namespace string
}

func (opt *ClientOptions) ApplyNamespace(to string) string {
	if opt.Namespace != "" {
		return fmt.Sprintf("%s.%s", opt.Namespace, to)
	}
	return to
}

func NewClientOptions(options ...ClientOption) (*ClientOptions, error) {
	clientOptions := &ClientOptions{
		Timeout: 10 * time.Second,
	}
	for _, option := range options {
		if err := option.setClient(clientOptions); err != nil {
			return nil, err
		}
	}
	return clientOptions, nil
}

type ServerOptions struct {
	// Namespace will be added to the beginning of all NATS
	// subjects the server listens on, effectively allowing
	// multiple servers to be run.
	Namespace string
	// QueueGroup can be used to ensure that requests are only
	// sent to a single server when running multiple in parallel.
	QueueGroup string
	// A handler for errors that occur during service calls.
	ErrorHandler func(error)
	// The maximum number of pending messages that each endpoint
	// in the server supports. This size is per-endpoint.
	BufferSize int
}

func (opt *ServerOptions) ApplyNamespace(to micro.Group) micro.Group {
	if opt.Namespace != "" {
		for _, segment := range strings.Split(opt.Namespace, ".") {
			if segment == "" {
				continue
			}
			to = to.AddGroup(segment)
		}
	}
	return to
}

func NewServerOptions(options ...ServerOption) (*ServerOptions, error) {
	serverOptions := &ServerOptions{
		BufferSize: 64,
	}
	for _, option := range options {
		if err := option.setServer(serverOptions); err != nil {
			return nil, err
		}
	}
	return serverOptions, nil
}

type opt struct {
	so func(*ServerOptions) error
	co func(*ClientOptions) error
}

func (o *opt) setServer(so *ServerOptions) error {
	return o.so(so)
}

func (o *opt) setClient(co *ClientOptions) error {
	return o.co(co)
}

type serverOptFunc func(o *ServerOptions) error

func (sof serverOptFunc) setServer(o *ServerOptions) error {
	return sof(o)
}

type clientOptFunc func(o *ClientOptions) error

func (cof clientOptFunc) setClient(o *ClientOptions) error {
	return cof(o)
}

// Set the namespace used for all NATS subjects
func Namespace(ns string) Option {
	return newOpt(
		func(o *ServerOptions) error {
			if ns != "" {
				o.Namespace = strings.Trim(ns, ".")
			}
			return nil
		},
		func(o *ClientOptions) error {
			if ns != "" {
				o.Namespace = strings.Trim(ns, ".")
			}
			return nil
		},
	)
}

// Generate a namespace by taking a SHA-256 hash of
// the inputs. This is useful for generating a namespace
// from values that may have illegal characters for a
// NATS subject name.
func HashNamespace(inputs ...string) Option {
	hash := sha256.New()
	for _, seg := range inputs {
		hash.Write([]byte(seg))
	}
	return Namespace(hex.EncodeToString(hash.Sum(nil)))
}

// Set the maximum number of buffered messages
// for each server endpoint
func BufferSize(bs int) ServerOption {
	return serverOptFunc(func(o *ServerOptions) error {
		if bs >= 0 {
			o.BufferSize = bs
		}
		return nil
	})
}

// Set the error handler for the server.
func ErrorHandler(eh func(error)) ServerOption {
	return serverOptFunc(func(o *ServerOptions) error {
		o.ErrorHandler = eh
		return nil
	})
}

// Set the NATS queue group name for the server
func QueueGroup(qg string) ServerOption {
	return serverOptFunc(func(o *ServerOptions) error {
		o.QueueGroup = qg
		return nil
	})
}

// Set the maximum amount of time the client will wait
// for a response from a server
func Timeout(to time.Duration) ClientOption {
	return clientOptFunc(func(o *ClientOptions) error {
		if to > 0 {
			o.Timeout = to
		}
		return nil
	})
}
