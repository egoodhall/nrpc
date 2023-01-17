package main

import (
	"github.com/alecthomas/kong"
	"github.com/emm035/nrpc/internal/astutil"
	"github.com/emm035/nrpc/internal/render"
	"golang.org/x/tools/go/packages"
)

var help = `A generator for implementing NATS-backed
RPC server/client implementations of go services.`

type Cli struct {
	Config   kong.ConfigFlag `name:"config" short:"c" help:"A file to load flags from"`
	Service  string          `name:"service" arg:"" required:"" help:"The service to implement as a NATS RPC server/client"`
	Package  string          `name:"package" short:"p" required:"" default:"." help:"The package to find the service in"`
	Encoding string          `name:"encoding" short:"e" default:"gob" enum:"gob,json" help:"The encoding to use for RPC messages"`
	Client   bool            `name:"client" default:"true" negatable:"" help:"Generate a client implementation of the service"`
	Server   bool            `name:"server" default:"true" negatable:"" help:"Generate a server for the service"`
}

func main() {
	ctx := kong.Parse(new(Cli), kong.Description(help))
	ctx.FatalIfErrorf(ctx.Run())
}

func (cli *Cli) Run() error {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo,
	}, cli.Package)
	if err != nil {
		return err
	}

	pkg, svc, err := astutil.FindService(cli.Service, pkgs)
	if err != nil {
		return err
	}

	return (&render.Renderer{
		Client:   cli.Client,
		Server:   cli.Server,
		Encoding: cli.Encoding,
	}).Render(pkg, *svc)
}
