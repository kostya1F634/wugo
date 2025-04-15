all: build

build: 
	go build -o ./bin/wugo wugo.go

x86_build:
	go build -o ./bin/wugo wugo.go

