
.PHONY: example protos

protos:
	@protoc \
	--go_out=pkg/nrpc --go_opt=module=github.com/emm035/nrpc/pkg/nrpc \
	--go-vtproto_out=pkg/nrpc --go-vtproto_opt=module=github.com/emm035/nrpc/pkg/nrpc \
	--go-vtproto_opt=features=marshal+unmarshal+size \
	proto/nrpc.proto

example:
	@protoc \
	--go-nrpc_out=. --go-nrpc_opt=module=github.com/emm035/nrpc \
	--go_out=. --go_opt=module=github.com/emm035/nrpc \
	example/example.proto
