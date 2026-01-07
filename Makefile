.PHONY: all build setup bin link test
all: bin

build:
	go build -o ./bin/wugo ./cmd/wugo

setup:
	go mod tidy
	go mod download

bin: bin-dir setup build

bin-dir:
	mkdir -p bin

test:
	mkdir -p .cache/go-build
	GOCACHE="$(CURDIR)/.cache/go-build" go test ./...
