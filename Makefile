
.PHONY: example protos

protos:
	@protoc \
	--go_out=go --go_opt=module=github.com/emm035/nrpc/go \
	--go-vtproto_out=go --go-vtproto_opt=module=github.com/emm035/nrpc/go \
	--go-vtproto_opt=features=marshal+unmarshal+size \
	proto/nrpc.proto

example:
	@protoc \
	--go-nrpc_out=go --go-nrpc_opt=module=github.com/emm035/nrpc/go \
	--go_out=go --go_opt=module=github.com/emm035/nrpc/go \
	go/example/example.proto

install:
	@go install ./go/cmd/protoc-gen-go-nrpc
