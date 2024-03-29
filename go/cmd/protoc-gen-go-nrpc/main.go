package main

import (
	"flag"

	"github.com/egoodhall/nrpc/go/internal/parse"
	"github.com/egoodhall/nrpc/go/internal/render"
	"google.golang.org/protobuf/compiler/protogen"
)

func main() {
	gen := new(Plugin)

	flag.BoolVar(&gen.vtproto, "vtproto", true, "Use vtproto for (un)marshaling protobufs")

	opts := &protogen.Options{
		ParamFunc: flag.Set,
	}

	opts.Run(gen.Generate)
}

type Plugin struct {
	vtproto bool
}

func (plugin *Plugin) Generate(p *protogen.Plugin) error {
	files := parse.ProtoServices(p)
	return (&render.Generator{
		Vtproto: plugin.vtproto,
	}).GenerateFiles(files...)
}
