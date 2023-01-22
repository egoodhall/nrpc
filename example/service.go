package example

//go:generate go run ../cmd/nrpc rpc -e json ExampleService
type ExampleService interface {
	EchoBytes(message []byte) ([]byte, error)
	Echo(message string) (string, error)
	Restart() error
}
