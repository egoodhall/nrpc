package main

import (
	"github.com/alecthomas/kong"
	"github.com/emm035/nrpc/cmd/nrpc/cli"
)

var help = `A generator for implementing NATS-backed
RPC server/client implementations of go services.`

type Cli struct {
	Rpc cli.RpcCmd `name:"rpc" cmd:"" help:"Generate an RPC server/client"`
}

func main() {
	ctx := kong.Parse(new(Cli), kong.Description(help))
	ctx.FatalIfErrorf(ctx.Run())
}
