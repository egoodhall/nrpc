package cli

import (
	"github.com/alecthomas/kong"
	"github.com/emm035/nrpc/internal/astutil"
	"github.com/emm035/nrpc/internal/render"
	"golang.org/x/tools/go/packages"
)

type RpcCmd struct {
	Config   kong.ConfigFlag `name:"config" short:"c" help:"A file to load flags from"`
	Services []string        `name:"service" arg:"" required:"" help:"The service(s) to implement as a NATS RPC server/client"`
	Package  string          `name:"package" short:"p" required:"" default:"." help:"The package to find the service(s) in"`
	Encoding string          `name:"encoding" short:"e" default:"gob" enum:"gob,json" help:"The encoding to use for RPC messages"`
	Client   bool            `name:"client" default:"true" negatable:"" help:"Generate a client implementation of the service"`
	Server   bool            `name:"server" default:"true" negatable:"" help:"Generate a server for the service"`
}

func (cmd *RpcCmd) Run() error {
	pkgs, err := packages.Load(&packages.Config{
		Mode: packages.NeedSyntax | packages.NeedFiles | packages.NeedName | packages.NeedTypes | packages.NeedTypesInfo | packages.NeedImports,
	}, cmd.Package)
	if err != nil {
		return err
	}

	seenpkgs := make(map[string]struct{})
	for _, service := range cmd.Services {
		pkg, svc, err := astutil.FindService(service, pkgs)
		if err != nil {
			return err
		}

		if _, seen := seenpkgs[pkg.PkgPath]; seen {
			continue
		}

		seenpkgs[pkg.PkgPath] = struct{}{}

		if err := render.CommonDecls(pkg); err != nil {
			return err
		}

		if err := (&render.Renderer{
			Client:   cmd.Client,
			Server:   cmd.Server,
			Encoding: cmd.Encoding,
		}).Render(pkg, *svc); err != nil {
			return err
		}
	}

	return nil
}
