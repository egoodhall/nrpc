package parse

import "io"

type Files map[string][]Service

type File struct {
	Package  string
	Services []Service
	io.Writer
}

type Service struct {
	Name     string
	RawName  string
	Comments string
	Methods  []Method
}

type Method struct {
	Name     string
	RawName  string
	Comments string
	Input    Type
	Output   Type
}

type Type struct {
	Name    string
	Package string
}
