package astutil

import (
	"errors"
	"go/ast"
)

var (
	ErrTooManyArguments = errors.New("too many arguments")
	ErrNoErrorInResults = errors.New("no error in results")
	ErrTooManyResults   = errors.New("too many results")
)

func ValidateServiceMethod(ft *ast.FuncType) error {
	if len(ft.Params.List) > 1 {
		// We only support 1 argument at most
		return ErrTooManyArguments
	}

	switch len(ft.Results.List) {
	case 0:
		// We need to return an error from the method.
		return ErrNoErrorInResults
	case 1:
		// The only result in a 1-result method must
		// be of type error.
		if !isErrorType(ft.Results.List[0].Type) {
			return ErrNoErrorInResults
		}
	case 2:
		// The second result in a 2-result method must
		// be of type error.
		if !isErrorType(ft.Results.List[1].Type) {
			return ErrNoErrorInResults
		}
	default:
		// >2 results is unsupported.
		return ErrTooManyResults
	}
	return nil
}

func isErrorType(expr ast.Expr) bool {
	switch et := expr.(type) {
	case *ast.Ident:
		return et.Name == "error"
	default:
		return false
	}
}
