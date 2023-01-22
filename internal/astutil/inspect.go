package astutil

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/ast/inspector"
	"golang.org/x/tools/go/packages"
)

func FindService(name string, pkgs []*packages.Package) (*packages.Package, *Service, error) {
	for _, pkg := range pkgs {
		var svc *Service
		var err error
		var imports map[string]string

		inspector.New(pkg.Syntax).Nodes([]ast.Node{
			new(ast.File),
			new(ast.ImportSpec),
			new(ast.TypeSpec),
		}, func(n ast.Node, push bool) (proceed bool) {
			if svc != nil || err != nil {
				// We already found our type, and tried to introspect
				// it. We don't need to see anything else.
				return false
			}

			switch ntyp := n.(type) {
			case *ast.File:
				// We're on a new file, so we need to start
				// a new import mapping.
				imports = make(map[string]string)
				return true
			case *ast.ImportSpec:
				if ntyp.Name != nil {
					// We have an import with an aliased name.
					// Let's store a mapping of alias -> path,
					// so we can map the name if our service
					// uses the import.
					imports[ntyp.Name.Name] = ntyp.Path.Value[1 : len(ntyp.Path.Value)-1]
				}
				return false
			case *ast.TypeSpec:
				if ntyp.Name == nil || ntyp.Name.Name != name {
					// This isn't our service declaration, so
					// we can skip introspecting it.
					return false
				}

				svc = new(Service)
				svc.Name = ntyp.Name.Name
				err = inspectTypeSpec(pkg, svc, ntyp)
				return false
			}
			return false
		})

		if svc == nil {
			// We didn't find the service in this package,
			// so let's continue to the next one.
			continue
		}

		if err != nil {
			return nil, nil, fmt.Errorf("parse service: %w", err)
		} else if err := resolvePkgPaths(svc, imports); err != nil {
			return nil, nil, err
		} else {
			return pkg, svc, nil
		}
	}
	return nil, nil, fmt.Errorf("service not found: %s", name)
}

func resolvePkgPaths(svc *Service, imports map[string]string) error {
	if len(imports) == 0 {
		return nil
	}

	for _, mth := range svc.Methods {
		if mth.Request != nil && mth.Request.Pkg != nil {
			if path, ok := imports[*mth.Request.Pkg]; ok {
				mth.Request.Pkg = &path
			}
		}
		if mth.Response != nil && mth.Response.Pkg != nil {
			if path, ok := imports[*mth.Response.Pkg]; ok {
				mth.Response.Pkg = &path
			}
		}
	}
	return nil
}

func inspectTypeSpec(pkg *packages.Package, svc *Service, d *ast.TypeSpec) error {
	if svc.Methods == nil {
		svc.Methods = make([]Method, 0)
	}

	switch i := d.Type.(type) {
	case *ast.InterfaceType:
		for _, embd := range i.Methods.List {
			switch etyp := embd.Type.(type) {
			case *ast.Ident:
				// We have a nested interface definition, so
				// we need to also add any methods from that.
				switch ityp := etyp.Obj.Decl.(type) {
				case *ast.TypeSpec:
					return inspectTypeSpec(pkg, svc, ityp)
				}
			case *ast.FuncType:
				// We have a method declaration, so let's parse
				// it and add it to our slice of methods.
				for _, fident := range embd.Names {
					if mth, berr := inspectMethod(pkg, fident); berr != nil {
						return fmt.Errorf("%s.%w", svc.Name, berr)
					} else {
						svc.Methods = append(svc.Methods, *mth)
					}
				}
			}
		}
	}

	// Deduplicate methods by name, so we don't accidentally
	// generate duplicate implementations. The go type system
	// will expose errors if the signatures aren't identical, so
	// we don't need to worry about name conflicts.
	dedup := make(map[string]struct{})
	for idx, mth := range svc.Methods {
		if _, ok := dedup[mth.Name]; ok {
			svc.Methods = append(svc.Methods[:idx], svc.Methods[idx+1:]...)
		} else {
			dedup[mth.Name] = struct{}{}
		}
	}

	return nil
}

func inspectMethod(pkg *packages.Package, id *ast.Ident) (*Method, error) {
	mth := &Method{
		Name: id.Name,
	}
	switch f := id.Obj.Decl.(type) {
	case *ast.Field:
		switch ft := f.Type.(type) {
		case *ast.FuncType:
			if err := ValidateServiceMethod(ft); err != nil {
				return nil, fmt.Errorf("%s: %w", mth.Name, err)
			}

			// Validation passed, which means we have:
			// - 1 param
			// - 1 or 2 results:
			//   - if 1 result, it must be an error
			//   - if 2 results, the second must be an error
			var err error

			if len(ft.Params.List) == 1 {
				// We have a parameter, so we can set it on our method
				// definition
				mth.Request, err = parseType(nil, ft.Params.List[0].Type)
				if err != nil {
					return nil, fmt.Errorf("%s: request: %w", mth.Name, err)
				}
			}

			switch len(ft.Results.List) {
			case 2:
				// 2 results is the only case where we need to parse out
				// a response type. Otherwise, it's just an error.
				mth.Response, err = parseType(nil, ft.Results.List[0].Type)
				if err != nil {
					return nil, fmt.Errorf("%s: response: %w", mth.Name, err)
				}
			default:
			}
		}
	}
	return mth, nil
}

// parseType collects information needed to render an implementation
// of an RPC method. We can guarantee that we only need the first name, due
// to previous validation of the method. If the *Type passed in is nil,
// one will be created.
func parseType(typ *Type, expr ast.Expr) (*Type, error) {
	if typ == nil {
		typ = new(Type)
	}

	switch xtyp := expr.(type) {
	case *ast.StarExpr:
		// We have a pointer type
		typ.Pointer = true
		return parseType(typ, xtyp.X)
	case *ast.SelectorExpr:
		// We have a type imported from another package.
		switch styp := xtyp.X.(type) {
		case *ast.Ident:
			typ.Pkg = &styp.Name
		default:
			return nil, fmt.Errorf("unexpected package expr type: %t", styp)
		}
		return parseType(typ, xtyp.Sel)
	case *ast.ArrayType:
		// We have an array/slice of items
		typ.Array = true
		return parseType(typ, xtyp.Elt)
	case *ast.Ident:
		// We have a type name
		typ.Type = xtyp.Name
		return typ, nil
	default:
		return nil, fmt.Errorf("unexpected param type: %+T", xtyp)
	}
}
