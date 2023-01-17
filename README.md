# NATS RPC

A code generator for NATS-based RPC services. NRPC services are defined as go interfaces. The NRPC
binary can then generate server and client implementations that will make the interface transparently
accessible over a NATS connection.

### Installation

```
go install github.com/emm035/nrpc@latest
```

### Usage

Use the `nrpc` command to generate implementations:

```
Usage: nrpc --package="." <service>

A generator for implementing NATS-backed RPC server/client implementations of go services.

Arguments:
  <service>    The service to implement as a NATS RPC server/client

Flags:
  -h, --help                  Show context-sensitive help.
  -c, --config=CONFIG-FLAG    A file to load flags from
  -p, --package="."           The package to find the service in
  -e, --encoding="gob"        The encoding to use for RPC messages
      --[no-]client           Generate a client implementation of the service
      --[no-]server           Generate a server for the service
```

### Example

An example service can be found in `example/service.go`. When `go generate` is used on the example
package, a client and server implementation will be generated for the `ExampleService` interface.
Usage of the generated code is shown in `cmd/exampleserver` and `cmd/exampleclient`.

### Restrictions

To keep code generation sane, service methods must only expose methods like the following:

```go
type ExampleService interface {
  // 0 arguments, returns error
  MethodOne() error
  // 1 argument, returns error
  MethodTwo(req MyRequest) error
  // 0 arguments, returns type and error
  MethodThree() (MyResponse, error)
  // 1 argument, returns type and error
  MethodFour(req MyRequest) (MyResponse, error)
}
```

Any other signatures will result in an error during generation.
