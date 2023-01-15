package main

import (
	"fmt"
	"time"

	"github.com/emm035/nrpc/example"
	"github.com/nats-io/nats.go"
)

func main() {
	nc, err := nats.Connect(nats.DefaultURL)
	if err != nil {
		panic(err)
	}

	client, err := example.NewExampleServiceClient(nc, example.ClientTimeout(1*time.Second))
	if err != nil {
		panic(err)
	}

	if res, err := client.Echo("Hello world"); err != nil {
		panic(err)
	} else {
		fmt.Println(res)
	}

	if res, err := client.Time(); err != nil {
		panic(err)
	} else {
		fmt.Println(res)
	}

	if err := client.Restart(); err != nil {
		fmt.Println(err.Error())
	}
}
