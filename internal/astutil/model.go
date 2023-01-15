package astutil

import (
	"fmt"

	"github.com/dave/jennifer/jen"
	"github.com/iancoleman/strcase"
)

type Service struct {
	Name    string
	Methods []Method
}

func (svc Service) ClientStructName() string {
	return strcase.ToLowerCamel(svc.Name + "Client")
}

func (svc Service) ServerStructName() string {
	return strcase.ToLowerCamel(svc.Name + "Server")
}

func (svc Service) FileName() string {
	return strcase.ToSnake(svc.Name) + ".gen.go"
}

type Method struct {
	Name     string
	Request  *NamedField
	Response *NamedField
}

func (mth Method) ResponseTypeName(svc Service) string {
	return fmt.Sprintf("%s%sResponse", strcase.ToLowerCamel(svc.Name), strcase.ToCamel(mth.Name))
}

func (mth Method) ErrorTypeName(svc Service) string {
	return fmt.Sprintf("%s%sError", strcase.ToCamel(svc.Name), strcase.ToCamel(mth.Name))
}

func (mth Method) RenderForClient(svc Service) *jen.Statement {
	return mth.renderClientSignature(svc).BlockFunc(func(g *jen.Group) {
		if mth.Request != nil {
			// We have a request type, so we need to encode it.
			g.Commentf("Encode request data from %s using gob", mth.Request.Type)
			g.Id("buf").Op(":=").New(jen.Qual("bytes", "Buffer"))
			g.IfFunc(func(g *jen.Group) {
				encode := jen.Err().Op(":=").Qual("encoding/gob", "NewEncoder").Params(jen.Id("buf")).Dot("Encode")
				if mth.Request.Pointer {
					g.Add(encode.Params(jen.Id(mth.Request.RenderName("request"))))
				} else {
					g.Add(encode.Params(jen.Op("&").Id(mth.Request.RenderName("request"))))
				}
				g.Err().Op("!=").Nil()
			}).BlockFunc(func(g *jen.Group) {
				if mth.Response == nil {
					g.Return(jen.Err())
				} else {
					if mth.Response.Pointer {
						g.Return(jen.Nil(), jen.Err())
					} else {
						g.Return(jen.Id("response"), jen.Err())
					}
				}
			})
			g.Line()
		}

		// Send the actual request
		g.Commentf("Send RPC message to %s.%s", svc.Name, mth.Name)
		g.List(jen.Id("msg"), jen.Err()).Op(":=").Id("client").Dot("conn").Dot("Request").ParamsFunc(func(g *jen.Group) {
			g.Add(jen.Id("client").Dot("options").Dot("Namespace").Op("+").Lit(fmt.Sprintf("%s.%s", svc.Name, mth.Name)))
			if mth.Request != nil {
				g.Add(jen.Id("buf").Dot("Bytes").Params())
			} else {
				g.Add(jen.Nil())
			}
			g.Id("client").Dot("options").Dot("Timeout")
		})

		g.If(jen.Err().Op("!=").Nil()).BlockFunc(func(g *jen.Group) {
			if mth.Response == nil {
				g.Add(jen.Return(jen.Err()))
			} else if mth.Response.Pointer {
				g.Add(jen.Return(jen.Nil(), jen.Err()))
			} else {
				g.Add(jen.Return(jen.Id("response"), jen.Err()))
			}
		})
		g.Line()

		// We have a response that we need to decode
		g.Commentf("Decode response into a wrapper object")
		g.Var().Id("reswrap").Id(mth.ResponseTypeName(svc))
		g.If(
			jen.Err().Op(":=").Qual("encoding/gob", "NewDecoder").Params(jen.Qual("bytes", "NewReader").Params(jen.Id("msg").Dot("Data"))).Dot("Decode").Params(jen.Op("&").Id("reswrap")),
			jen.Err().Op("!=").Nil(),
		).BlockFunc(func(g *jen.Group) {
			if mth.Response == nil {
				g.Return(jen.Err())
			} else if mth.Response.Pointer {
				g.Return(jen.Nil(), jen.Err())
			} else {
				g.Return(jen.Id("response"), jen.Err())
			}
		})

		// Return our decoded response values
		if mth.Response == nil {
			g.Add(jen.Return(jen.Id("reswrap").Dot("Err")))
		} else {
			g.Add(jen.Return(jen.Id("reswrap").Dot("Res"), jen.Id("reswrap").Dot("Err")))
		}
	})
}

func (mth Method) renderClientSignature(svc Service) *jen.Statement {
	return jen.Func().Params(jen.Id("client").Op("*").Id(svc.ClientStructName())).Id(mth.Name).
		// Parameters for our method
		ParamsFunc(func(g *jen.Group) {
			if mth.Request != nil {
				g.Add(mth.Request.Render("request"))
			}
		}).
		// Response types for our method
		ParamsFunc(func(g *jen.Group) {
			if mth.Response != nil {
				g.Add(mth.Response.Render("response"))
			}
			g.Err().Error()
		})
}

func (mth Method) RenderForServer(svc Service) *jen.Statement {
	return mth.renderServerSignature(svc).BlockFunc(func(g *jen.Group) {
		g.Defer().Func().Params().Block(
			jen.If(jen.Id("val").Op(":=").Recover(), jen.Id("val").Op("!=").Nil()).Block(
				jen.If(jen.List(jen.Err(), jen.Id("ok")).Op(":=").Id("val").Assert(jen.Error()), jen.Id("ok")).Block(
					jen.Id("server").Dot("errs").Op("<-").Err(),
				),
			),
		).Params()
		if mth.Request != nil {
			g.Var().Id("request").Id(mth.Request.Type)
		}
		g.Var().Id("response").Id(mth.ResponseTypeName(svc))

		stmt := jen.Add()
		if mth.Request != nil {
			// We have a request type, so we need to decode it
			// before we can call the service method
			stmt = jen.If(
				jen.Err().Op(":=").Qual("encoding/gob", "NewDecoder").Params(jen.Qual("bytes", "NewReader").Params(jen.Id("msg").Dot("Data"))).Dot("Decode").Params(jen.Op("&").Id("request")),
				jen.Err().Op("!=").Nil(),
			).Block(
				jen.Id("errw").Op(":=").Op("&").Id(mth.ErrorTypeName(svc)).Values(jen.Err().Dot("Error").Params()),
				jen.Id("response").Dot("Err").Op("=").Id("errw"),
				jen.Id("server").Dot("errs").Op("<-").Id("errw"),
			).Else()
		}

		// Now we can make the service call
		stmt = stmt.If(
			jen.ListFunc(func(g *jen.Group) {
				if mth.Response != nil {
					g.Id("res")
				}
				g.Err()
			}).Op(":=").Id("server").Dot("service").Dot(mth.Name).ParamsFunc(func(g *jen.Group) {
				if mth.Request != nil {
					g.Id("request")
				}
			}),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Id("errw").Op(":=").Op("&").Id(mth.ErrorTypeName(svc)).Values(jen.Err().Dot("Error").Params()),
			jen.Id("response").Dot("Err").Op("=").Id("errw"),
			jen.Id("server").Dot("errs").Op("<-").Id("errw"),
		)

		// If we are expecting a response, we should set it
		// on the response wrapper if there wasn't an error
		// returned from the service method.
		if mth.Response != nil {
			stmt = stmt.Else().Block(
				jen.Id("response").Dot("Res").Op("=").Id("res"),
			)
		}
		g.Add(stmt)
		g.Line()

		// Next, we need to encode our response wrapper, and
		// send our response. This should be the same behavior
		// regardless of whether our method gives a response.
		g.Id("buf").Op(":=").New(jen.Qual("bytes", "Buffer"))
		g.If(
			jen.Err().Op(":=").Qual("encoding/gob", "NewEncoder").Params(jen.Id("buf")).Dot("Encode").Params(jen.Id("response")),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Id("server").Dot("errs").Op("<-").Op("&").Id(mth.ErrorTypeName(svc)).Values(jen.Err().Dot("Error").Params()),
		).Else().If(
			jen.Err().Op(":=").Id("msg").Dot("Respond").Params(jen.Id("buf").Dot("Bytes").Params()),
			jen.Err().Op("!=").Nil(),
		).Block(
			jen.Id("server").Dot("errs").Op("<-").Op("&").Id(mth.ErrorTypeName(svc)).Values(jen.Err().Dot("Error").Params()),
		)
	})
}

func (mth Method) renderServerSignature(svc Service) *jen.Statement {
	return jen.Func().
		Params(jen.Id("server").Op("*").Id(svc.ServerStructName())).
		Id(strcase.ToLowerCamel(mth.Name)).
		Params(jen.Id("msg").Op("*").Qual("github.com/nats-io/nats.go", "Msg"))
}

type NamedField struct {
	Pointer bool
	Name    *string
	Pkg     *string
	Type    string
}

func (nf NamedField) RenderName(defaultName string) string {
	if nf.Name != nil {
		return *nf.Name
	}
	return defaultName
}

func (nf NamedField) Render(defaultName string) *jen.Statement {
	stmt := jen.Add()
	if defaultName != "" {
		stmt = stmt.Id(nf.RenderName(defaultName))
	}
	if nf.Pointer {
		stmt = stmt.Op("*")
	}
	if nf.Pkg != nil {
		return stmt.Qual(*nf.Pkg, nf.Type)
	}
	return stmt.Id(nf.Type)
}
