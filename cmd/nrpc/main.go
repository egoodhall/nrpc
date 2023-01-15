package main

import (
	"github.com/alecthomas/kong"
	"github.com/emm035/nrpc/internal/astutil"
	"golang.org/x/tools/go/packages"
)

type Cli struct {
	Config  kong.ConfigFlag `name:"config" short:"c"`
	Package string          `arg:"" name:"pkg" required:"" default:"."`
	Service string          `name:"service" short:"s"`
}

func main() {
	ctx := kong.Parse(new(Cli))
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

	return astutil.Render(pkg, *svc)
}
