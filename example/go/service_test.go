package example_test

import (
	"testing"

	example "github.com/egoodhall/nrpc/example/go"
	"github.com/egoodhall/nrpc/go/pkg/nrpc"
	"github.com/nats-io/nats-server/v2/server"
	"github.com/nats-io/nats-server/v2/test"
	"github.com/nats-io/nats.go"
)

func TestMain(m *testing.M) {
	// We can use a test server without needing
	// to listen on any TCP ports
	test.DefaultTestOptions.DontListen = true
}

func newConn(t *testing.T, srv *server.Server) *nats.Conn {
	conn, err := nats.Connect("", nats.InProcessServer(srv))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(conn.Close)
	return conn
}

func TestServerAndClientWorkTogether(t *testing.T) {
	nsrv := test.RunDefaultServer()

	client, err := example.NewEchoServiceClient(newConn(t, nsrv))
	if err != nil {
		t.Fatal(err)
	}

	server, err := example.NewEchoServiceServer(newConn(t, nsrv), new(example.ServiceImpl))
	if err != nil {
		t.Fatal(err)
	}

	// Start the server. This will start a pool of goroutines
	// listening for messages, and responding to them
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	if response, err := client.Echo(&example.EchoRequest{Message: "Test!"}); err != nil {
		t.Fatal(err)
	} else if response.Message != "Test!" {
		t.Fatalf("Response does not match: '%s' != 'Test!'", response)
	}
}

func TestHashNamespacing(t *testing.T) {
	nsrv := test.RunDefaultServer()

	hashNsOpt := nrpc.HashNamespace("test", "a", "b", "c")

	client, err := example.NewEchoServiceClient(newConn(t, nsrv), hashNsOpt)
	if err != nil {
		t.Fatal(err)
	}

	server, err := example.NewEchoServiceServer(newConn(t, nsrv), new(example.ServiceImpl), hashNsOpt)
	if err != nil {
		t.Fatal(err)
	}

	// Start the server. This will start a pool of goroutines
	// listening for messages, and responding to them
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	if response, err := client.Echo(&example.EchoRequest{Message: "Test!"}); err != nil {
		t.Fatal(err)
	} else if response.Message != "Test!" {
		t.Fatalf("Response does not match: '%s' != 'Test!'", response)
	}
}
