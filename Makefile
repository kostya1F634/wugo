.PHONY: all build setup bin link
all: build

build: 
	go build -o ./bin/wugo wugo.go

setup:
	go mod tidy

bin: bin-dir setup build

bin-dir:
	mkdir -p bin

link:
	sudo ln -sv "$$(pwd)/bin/wugo" /usr/local/bin/wugo
