package example_test

import (
	"testing"

	"github.com/emm035/nrpc/example"
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

	client, err := example.NewExampleServiceClient(newConn(t, nsrv))
	if err != nil {
		t.Fatal(err)
	}

	server, err := example.NewExampleServiceServer(new(example.ServiceImpl), newConn(t, nsrv))
	if err != nil {
		t.Fatal(err)
	}

	// Start the server. This will start a pool of goroutines
	// listening for messages, and responding to them
	if err := server.Start(); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := server.Stop(); err != nil {
			t.Fatal(err)
		}
	}()

	if err := client.Restart(); err != nil {
		t.Fatal(err)
	}

	if response, err := client.Echo("Test!"); err != nil {
		t.Fatal(err)
	} else if response != "Test!" {
		t.Fatalf("Response does not match: '%s' != 'Test!'", response)
	}
}
