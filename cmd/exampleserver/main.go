package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"time"

	"github.com/emm035/nrpc/example"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic((err))
	}

	server, err := example.NewExampleServiceServer(&svc{}, nc, example.ServerErrorHandler(func(err error) {
		fmt.Println(err.Error())
	}))
	if err != nil {
		panic(err)
	}

	if err := server.Start(); err != nil {
		panic(err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	<-ctx.Done()

	if err := server.Stop(); err != nil {
		panic(err)
	}
}

type svc struct {
}

func (service *svc) Echo(msg string) (string, error) {
	return msg, nil
}

func (service *svc) Restart() error {
	return errors.New("couldn't restart")
}

func (service *svc) Time() (time.Time, error) {
	return time.Now(), nil
}
