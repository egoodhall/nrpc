package main

import (
	"errors"
	"flag"

	"github.com/egoodhall/nrpc/go/internal/parse"
	"github.com/egoodhall/nrpc/go/internal/render"
	"google.golang.org/protobuf/compiler/protogen"
)

var (
	ErrClientStreamingUnsupported        = errors.New("client streaming isn't supported")
	ErrServerStreamingUnsupported        = errors.New("server streaming isn't supported")
	ErrBidirectionalStreamingUnsupported = errors.New("bidirectional streaming isn't supported")
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
