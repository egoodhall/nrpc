package parse

import (
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
)

func ProtoServices(p *protogen.Plugin) []File {
	servicesFiles := make([]File, 0)
	for _, protoFile := range p.Files {
		services := make([]Service, 0)
		// We need to check all files for service declarations. Structs
		// will be generated for each message already, so we don't need
		// to worry about them being present. We'll just use the types
		// as we're given them.
		for _, protoService := range protoFile.Services {
			service := Service{
				Name:     protoService.GoName,
				RawName:  string(protoService.Desc.Name()),
				Comments: strings.Trim(protoService.Comments.Leading.String(), "\t\n "),
				Methods:  make([]Method, 0),
			}

			for _, protoMethod := range protoService.Methods {
				input := parseType(protoFile, protoMethod.Input)

				output := parseType(protoFile, protoMethod.Output)

				service.Methods = append(service.Methods, Method{
					Name:     protoMethod.GoName,
					RawName:  string(protoMethod.Desc.Name()),
					Comments: strings.Trim(protoMethod.Comments.Leading.String(), "\t\n "),
					Input:    input,
					Output:   output,
				})
			}

			services = append(services, service)
		}

		if len(services) > 0 {
			servicesFiles = append(servicesFiles, File{
				Writer:   p.NewGeneratedFile(protoFile.GeneratedFilenamePrefix+".nrpc.go", protoFile.GoImportPath),
				Package:  string(protoFile.GoPackageName),
				Services: services,
			})
		}
	}
	return servicesFiles
}

func parseType(file *protogen.File, msg *protogen.Message) Type {
	var typ Type

	typ.Name = msg.GoIdent.GoName
	if msg.GoIdent.GoImportPath != file.GoImportPath {
		typ.Package = string(msg.GoIdent.GoImportPath)
	}

	return typ
}
