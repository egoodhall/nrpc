package render

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/egoodhall/nrpc/go/internal/parse"
	"github.com/iancoleman/strcase"
	"github.com/sourcegraph/conc/pool"
)

const (
	pkgErrors = "errors"
	pkgNats   = "github.com/nats-io/nats.go"
	pkgMicro  = "github.com/nats-io/nats.go/micro"
	pkgNrpc   = "github.com/egoodhall/nrpc/go/pkg/nrpc"
	pkgProto  = "google.golang.org/protobuf/proto"
	pkgAnypb  = "google.golang.org/protobuf/types/known/anypb"
)

const (
	clientName       = "client"
	connName         = "conn"
	serverName       = "server"
	dataName         = "data"
	dataPbName       = "datapb"
	handleErrorName  = "handleError"
	optionsFieldName = "options"
	requestName      = "request"
	requestWrapName  = "requestwrap"
	responseErrName  = "err"
	responseMsgName  = "resmsg"
	responseName     = "response"
	responseOkName   = "ok"
	responseWrapName = "responsewrap"
	serviceFieldName = "service"
)

type Generator struct {
	Vtproto bool
}

func (gen *Generator) GenerateFiles(files ...parse.File) error {
	genpool := pool.New().WithErrors().WithFirstError()
	for _, file := range files {
		genpool.Go(func() error {
			return gen.generate(file)
		})
	}
	return genpool.Wait()
}

func (gen *Generator) generate(filedef parse.File) error {
	file := jen.NewFile(filedef.Package)

	// Declare import mappings
	file.HeaderComment("Code generated by nrpc; DO NOT EDIT.")
	file.ImportName(pkgNats, "nats")
	file.ImportName(pkgMicro, "micro")
	file.ImportName(pkgNrpc, "nrpc")
	file.ImportName(pkgProto, "proto")
	file.ImportName(pkgAnypb, "anypb")

	for _, service := range filedef.Services {
		gen.serviceInterface(file, service)
		gen.clientConstructor(file, service)
		gen.clientImpl(file, service)
		gen.serverConstructor(file, service)
		gen.serverImpl(file, service)
	}

	return file.Render(filedef)
}

func (gen *Generator) serviceInterface(file *jen.File, service parse.Service) {
	file.Type().Id(service.Name).InterfaceFunc(func(g *jen.Group) {
		for _, method := range service.Methods {
			g.Id(method.Name).Params(gen.typeName(method.Input, true)).Params(gen.typeName(method.Output, true), jen.Error())
		}
	}).Line()
}

func (gen *Generator) clientConstructor(file *jen.File, service parse.Service) {
	file.Func().Id("New"+service.Name+"Client").Params(
		jen.Id(connName).Op("*").Qual(pkgNats, "Conn"),
		jen.Id(optionsFieldName).Op("...").Qual(pkgNrpc, "ClientOption"),
	).Params(jen.Id(service.Name), jen.Error()).BlockFunc(func(g *jen.Group) {
		g.List(jen.Id("clientOptions"), jen.Err()).Op(":=").Qual(pkgNrpc, "NewClientOptions").Params(jen.Id(optionsFieldName).Op("..."))
		g.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		)
		g.Line()
		g.Return(jen.Op("&").Add(gen.clientTypeName(service, false)).Values(jen.Dict{
			jen.Id(optionsFieldName): jen.Id("clientOptions"),
			jen.Id(connName):         jen.Id(connName),
		}), jen.Nil())
	})
}

func (gen *Generator) clientImpl(file *jen.File, service parse.Service) {
	// Client struct definition
	file.Type().Add(gen.clientTypeName(service, false)).Struct(
		jen.Id(optionsFieldName).Op("*").Qual(pkgNrpc, "ClientOptions"),
		jen.Id(connName).Op("*").Qual(pkgNats, "Conn"),
	)

	// Client method definitions
	for _, method := range service.Methods {
		file.Add(gen.clientMethodSignature(service, method)).BlockFunc(func(g *jen.Group) {
			gen.clientMethodBody(g, service, method)
		})
		file.Line()
	}
}

func (gen *Generator) clientMethodSignature(service parse.Service, method parse.Method) *jen.Statement {
	return jen.Func().
		Params(jen.Id(clientName).Add(gen.clientTypeName(service, true))).
		Id(method.Name).
		Params(jen.Id(requestName).Add(gen.typeName(method.Input, true))).
		Params(gen.typeName(method.Output, true), jen.Error())
}

func (gen *Generator) clientMethodBody(g *jen.Group, service parse.Service, method parse.Method) {
	g.List(jen.Id(dataName), jen.Err()).Op(":=").Qual(pkgProto, "Marshal").Params(jen.Id(requestName))
	g.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Nil(), jen.Err()),
	)
	g.Line()
	g.List(jen.Id(responseMsgName), jen.Err()).Op(":=").Id(clientName).Dot(connName).Dot("Request").Params(
		gen.natsSubject(clientName, service, method),
		jen.Id(dataName),
		jen.Id(clientName).Dot(optionsFieldName).Dot("Timeout"),
	)
	g.If(jen.Err().Op("!=").Nil()).Block(
		jen.Return(jen.Nil(), jen.Err()),
	)
	g.Line()
	g.If(
		jen.Err().Op(":=").Qual(pkgNrpc, "ParseError").Params(jen.Id(responseMsgName)),
		jen.Err().Op("!=").Nil(),
	).Block(
		jen.Return(jen.Nil(), jen.Err()),
	)
	g.Line()
	g.Id(responseName).Op(":=").New(gen.typeName(method.Output, false))
	g.If(
		jen.Err().Op(":=").Qual(pkgProto, "Unmarshal").Params(jen.Id(responseMsgName).Dot("Data"), jen.Id(responseName)),
		jen.Err().Op("!=").Nil(),
	).Block(
		jen.Return(jen.Nil(), jen.Err()),
	)
	g.Return(jen.Id(responseName), jen.Nil())
}

func (gen *Generator) serverConstructor(file *jen.File, service parse.Service) {
	file.Func().Id("New"+service.Name+"Server").Params(
		jen.Id(connName).Op("*").Qual(pkgNats, "Conn"),
		jen.Id(serviceFieldName).Id(service.Name),
		jen.Id(optionsFieldName).Op("...").Qual(pkgNrpc, "ServerOption"),
	).Params(jen.Qual(pkgNrpc, "Server"), jen.Error()).BlockFunc(func(g *jen.Group) {
		g.List(jen.Id("serverOptions"), jen.Err()).Op(":=").Qual(pkgNrpc, "NewServerOptions").Params(jen.Id(optionsFieldName).Op("..."))
		g.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		)
		g.Line()
		g.Id("compat").Op(":=").Op("&").Add(gen.serverTypeName(service, false)).Values(jen.Dict{
			jen.Id(optionsFieldName): jen.Id("serverOptions"),
			jen.Id(serviceFieldName): jen.Id(serviceFieldName),
		})
		g.List(jen.Id("svc"), jen.Err()).Op(":=").Qual(pkgMicro, "AddService").Params(
			jen.Id(connName),
			jen.Qual(pkgMicro, "Config").Values(jen.Dict{
				jen.Id("Name"):    jen.Lit(service.RawName),
				jen.Id("Version"): jen.Lit("0.0.0"),
			}),
		)

		g.If(jen.Err().Op("!=").Nil()).Block(
			jen.Return(jen.Nil(), jen.Err()),
		)

		g.Id("grp").Op(":=").Id("svc").Dot("AddGroup").Params(jen.Lit(service.RawName))

		for _, method := range service.Methods {
			g.Line()
			g.If(
				jen.Err().Op(":=").Id("grp").Dot("AddEndpoint").Params(
					jen.Lit(method.RawName),
					jen.Qual(pkgMicro, "HandlerFunc").Params(jen.Id("compat").Dot(gen.serverHandlerName(method))),
				),
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Id("svc").Dot("Stop").Params(),
				jen.Return(jen.Nil(), jen.Err()),
			)
		}

		g.Line()
		g.Return(jen.Id("svc"), jen.Nil())
	})
}

func (gen *Generator) serverImpl(file *jen.File, service parse.Service) {
	// Server struct definition
	file.Type().Add(gen.serverTypeName(service, false)).Struct(
		jen.Id(optionsFieldName).Op("*").Qual(pkgNrpc, "ServerOptions"),
		jen.Id(serviceFieldName).Id(service.Name),
	)

	// Server compat implementation
	file.Line()
	gen.serverErrorHandler(file, service)
	file.Line()

	// Server compat method definitions
	for _, method := range service.Methods {
		gen.serverHandler(file, service, method)
		file.Line()
	}
}

func (gen *Generator) serverErrorHandler(file *jen.File, service parse.Service) {
	file.Func().Params(
		jen.Id(serverName).Add(gen.serverTypeName(service, true)),
	).Add(jen.Id(handleErrorName)).Params(jen.Err().Error()).Block(
		jen.If(jen.Id(serverName).Dot(optionsFieldName).Dot("ErrorHandler").Op("!=").Nil()).Block(
			jen.Id(serverName).Dot(optionsFieldName).Dot("ErrorHandler").Params(jen.Err()),
		),
	)
}

func (gen *Generator) serverHandler(file *jen.File, service parse.Service, method parse.Method) {
	file.Func().Params(
		jen.Id(serverName).Add(gen.serverTypeName(service, true)),
	).Add(jen.Id(gen.serverHandlerName(method))).Params(
		jen.Id("msg").Qual(pkgMicro, "Request"),
	).Block(
		jen.Id(requestName).Op(":=").New(gen.typeName(method.Input, false)),
		jen.If(
			jen.Err().Op(":=").Qual(pkgProto, "Unmarshal").Params(
				jen.Id("msg").Dot("Data").Params(),
				jen.Id(requestName),
			),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Id("msg").Dot("Error").Params(
				jen.Qual("strconv", "Itoa").Params(jen.Lit(500)),
				jen.Err().Dot("Error").Params(),
				jen.Nil(),
			),
			jen.Id(serverName).Dot(handleErrorName).Params(jen.Err()),
			jen.Return(),
		),
		jen.Line(),
		jen.List(jen.Id(responseName), jen.Err()).Op(":=").
			Id(serverName).Dot(serviceFieldName).Dot(method.Name).Params(jen.Id(requestName)),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Id("e").Op(":=").New(jen.Qual(pkgNrpc, "Error")),
			jen.If(jen.Op("!").Qual(pkgErrors, "As").Params(jen.Err(), jen.Op("&").Id("e"))).Block(
				jen.Id("e").Op("=").Qual(pkgNrpc, "NewError").Params(jen.Lit(500), jen.Err().Dot("Error").Params()),
			),
			jen.Line(),
			jen.Id("msg").Dot("Error").Params(
				jen.Qual("strconv", "Itoa").Params(jen.Id("e").Dot("Code").Params()),
				jen.Id("e").Dot("Error").Params(),
				jen.Nil(),
			),
			jen.Id(serverName).Dot(handleErrorName).Params(jen.Err()),
		),
		jen.Line(),
		jen.List(jen.Id("data"), jen.Err()).Op(":=").Qual(pkgProto, "Marshal").Params(jen.Id(responseName)),
		jen.If(jen.Err().Op("!=").Nil()).Block(
			jen.Id("msg").Dot("Error").Params(
				jen.Qual("strconv", "Itoa").Params(jen.Lit(500)),
				jen.Err().Dot("Error").Params(),
				jen.Nil(),
			),
			jen.Id(serverName).Dot(handleErrorName).Params(jen.Err()),
			jen.Return(),
		),
		jen.Line(),
		jen.If(
			jen.Id("err").Op(":=").Id("msg").Dot("Respond").Params(jen.Id("data")),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Id(serverName).Dot(handleErrorName).Params(jen.Err()),
		),
	)
}

func (gen *Generator) serverHandlerName(method parse.Method) string {
	return "serve" + strcase.ToCamel(method.Name)
}

func (gen *Generator) natsSubject(container string, service parse.Service, method parse.Method) jen.Code {
	return jen.Id(container).Dot(optionsFieldName).Dot("ApplyNamespace").
		Params(jen.Lit(fmt.Sprintf("%s.%s", service.RawName, method.RawName)))
}

func (gen *Generator) typeName(typ parse.Type, pointer bool) jen.Code {
	stmt := jen.Add()
	if pointer {
		stmt = stmt.Op("*")
	}
	if typ.Package != "" {
		return stmt.Qual(typ.Package, typ.Name)
	} else {
		return stmt.Id(typ.Name)
	}
}

func (gen *Generator) clientTypeName(service parse.Service, pointer bool) jen.Code {
	stmt := jen.Add()
	if pointer {
		stmt = stmt.Op("*")
	}
	return stmt.Id(strcase.ToLowerCamel(service.Name) + "Client")
}

func (gen *Generator) serverTypeName(service parse.Service, pointer bool) jen.Code {
	stmt := jen.Add()
	if pointer {
		stmt = stmt.Op("*")
	}
	return stmt.Id(strcase.ToLowerCamel(service.Name) + "ServerCompat")
}
