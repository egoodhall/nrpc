package example

import "time"

//go:generate go run ../cmd/nrpc -s ExampleService
type ExampleService interface {
	Time() (time.Time, error)
	Echo(message string) (string, error)
	Restart() error
}
