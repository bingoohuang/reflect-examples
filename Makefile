.PHONY: default test
all: default test

default:
	gofmt -s -w .&&go mod tidy&&go fmt ./...&&revive .&&goimports -w .&&golangci-lint run --enable-all&&go install -ldflags="-s -w" ./...

install:
	go install -ldflags="-s -w" ./...

test:
	go test ./...
