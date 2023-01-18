package example

//go:generate go run ../cmd/nrpc -e json ExampleService
type ExampleService interface {
	Echo(message string) (string, error)
	Restart() error
}
