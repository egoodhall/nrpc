package astutil

import (
	"fmt"
	"go/ast"

	"golang.org/x/tools/go/packages"
)

func FindService(name string, pkgs []*packages.Package) (*packages.Package, *Service, error) {
	for _, pkg := range pkgs {
		// We should iterate over type definitions in the
		// package, looking for an interface that matches
		// the name that we have been given.
		for ident, def := range pkg.TypesInfo.Defs {
			if def != nil && def.Name() == name {
				var err error
				svc := new(Service)

				// Inspect the node, looking for an interface declaration
				// with methods, or other nested interfaces.
				ast.Inspect(ident, func(n ast.Node) bool {
					switch t := n.(type) {
					case *ast.Ident:
						if t.Name == name {
							svc.Name = t.Name
							svc.Methods = make([]Method, 0)
							if t.Obj.Kind == ast.Typ {
								switch d := t.Obj.Decl.(type) {
								case *ast.TypeSpec:
									// We need to recursively inspect the type spec,
									// since it could have nested interfaces in it.
									methods, tserr := inspectTypeSpec(*svc, d)
									if tserr != nil {
										err = tserr
										return false
									}

									// Deduplicate methods by name, so we don't accidentally
									// generate duplicate code. The go type system will expose
									// errors if the function signatures aren't identical, so
									// we don't need to worry about name conflicts.
									dedup := make(map[string]struct{})
									for _, mth := range methods {
										if _, ok := dedup[mth.Name]; !ok {
											dedup[mth.Name] = struct{}{}
											svc.Methods = append(svc.Methods, mth)
										}
									}
								}
							}
						}
					}
					return false
				})
				if err != nil {
					return nil, nil, fmt.Errorf("parse service: %w", err)
				} else {
					return pkg, svc, nil
				}
			}
		}
	}
	return nil, nil, fmt.Errorf("service not found: %s", name)
}

func inspectTypeSpec(svc Service, d *ast.TypeSpec) ([]Method, error) {
	methods := make([]Method, 0)
	switch i := d.Type.(type) {
	case *ast.InterfaceType:
		for _, m := range i.Methods.List {
			switch mtyp := m.Type.(type) {
			case *ast.FuncType:
				// We have a method declaration, so let's parse
				// it and add it to our slice of methods.
				for _, fident := range m.Names {
					if mth, berr := inspectMethod(fident); berr != nil {
						return nil, fmt.Errorf("%s.%w", svc.Name, berr)
					} else {
						methods = append(methods, *mth)
					}
				}
			case *ast.Ident:
				// We have a nested interface definition, so
				// we need to also add any methods from that.
				switch ityp := mtyp.Obj.Decl.(type) {
				case *ast.TypeSpec:
					mth, err := inspectTypeSpec(svc, ityp)
					if err != nil {
						return nil, err
					}
					methods = append(methods, mth...)
				}
			}
		}
	}
	return methods, nil
}

func inspectMethod(id *ast.Ident) (*Method, error) {
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
				mth.Request, err = parseNamedField(ft.Params.List[0])
				if err != nil {
					return nil, fmt.Errorf("%s: request: %w", mth.Name, err)
				}
			}

			switch len(ft.Results.List) {
			case 2:
				// 2 results is the only case where we need to parse out
				// a response type. Otherwise, it's just an error.
				mth.Response, err = parseNamedField(ft.Results.List[0])
				if err != nil {
					return nil, fmt.Errorf("%s: response: %w", mth.Name, err)
				}
			default:
			}
		}
	}
	return mth, nil
}

// parseNamedField collects information needed to render an implementation
// of an RPC method. We can guarantee that we only need the first name, due
// to previous validation of the method.
func parseNamedField(field *ast.Field) (*NamedField, error) {
	namedField := new(NamedField)

	if len(field.Names) > 0 {
		namedField.Name = &field.Names[0].Name
	}

	switch ptyp := field.Type.(type) {
	case *ast.Ident:
		// We have a type
		namedField.Type = ptyp.Name
		return namedField, nil
	case *ast.StarExpr:
		// We have a pointer type
		namedField.Pointer = true
		typ, err := getExprName(ptyp.X)
		if err != nil {
			return nil, err
		}
		namedField.Type = typ
		return namedField, nil
	case *ast.SelectorExpr:
		// We have a type imported from another
		// package.
		namedField.Type = ptyp.Sel.Name
		typ, err := getExprName(ptyp.X)
		if err != nil {
			return nil, err
		}
		namedField.Pkg = &typ
		return namedField, nil
	default:
		return nil, fmt.Errorf("unexpected param type: %+T", ptyp)
	}
}

func getExprName(x ast.Expr) (string, error) {
	switch xtyp := x.(type) {
	case *ast.Ident:
		return xtyp.Name, nil
	default:
		return "", fmt.Errorf("unexpected param expression type: %+T", xtyp)
	}
}
