package main

import (
	"flag"
	"os"

	"github.com/egoodhall/nrpc/go/internal/parse"
	"github.com/egoodhall/nrpc/go/pkg/render"
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
	for _, file := range parse.ProtoServices(p) {
		javaFiles, err := render.Java(file)
		if err != nil {
			return err
		}
		for name, content := range javaFiles {
			if err := os.WriteFile(name, content, 0640); err != nil {
				return err
			}
		}
	}

	return nil
}
