
.PHONY: example example_go example_java install install-gravel

all: example

# protos:
# 	@mkdir -p java/target/generated-sources/protobuf
# 	@protoc \
# 	--go_out=go --go_opt=module=github.com/egoodhall/nrpc/go \
# 	--go-vtproto_out=go --go-vtproto_opt=module=github.com/egoodhall/nrpc/go \
# 	--go-vtproto_opt=features=marshal+unmarshal+size \
# 	--java_out=java/target/generated-sources/protobuf \
# 	proto/nrpc.proto

example: example_go example_java

example_go: install
	@protoc \
	--go_out=example/go --go_opt=module=github.com/egoodhall/nrpc/example/go \
	--go-nrpc_out=example/go --go-nrpc_opt=module=github.com/egoodhall/nrpc/example/go \
	example/example.proto

example_java:
	@mkdir -p example/java/target/generated-sources/protobuf
	@mvn clean verify

install: install-gravel
	@gravel install --root go

install-gravel:
	@go install github.com/egoodhall/gravel/cmd/gravel@latest
