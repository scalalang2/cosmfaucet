BUF_VERSION:=1.5.0

.PHONEY: build
build:
	@echo "compiling source code"
	@go build -o ./bin/ -v ./cmd/...

.PHONEY: test
test:
	@echo "running tests"
	@go test -v ./...

.PHONEY: generate
generate:
	docker run -v $$(pwd):/src -w /src --rm bufbuild/buf:$(BUF_VERSION) generate

.PHONEY: lint
lint:
	docker run -v $$(pwd):/src -w /src --rm bufbuild/buf:$(BUF_VERSION) lint
	docker run -v $$(pwd):/src -w /src --rm bufbuild/buf:$(BUF_VERSION) breaking --against 'https://github.com/johanbrandhorst/grpc-gateway-boilerplate.git#branch=master'