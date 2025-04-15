.PHONY: all build setup bin link
all: build

build: 
	go build -o ./bin/wugo wugo.go

setup:
	go mod tidy

bin: setup build

link:
	sudo ln -sv "$$(pwd)/bin/wugo" /usr/local/bin/wugo
