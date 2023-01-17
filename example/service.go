package example

import "time"

//go:generate go run ../cmd/nrpc -e json ExampleService
type ExampleService interface {
	Time() (time.Time, error)
	Echo(message string) (string, error)
	Restart() error
}
