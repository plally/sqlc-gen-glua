.PHONY: build test bin/sqlc-gen-go bin/sqlc-gen-go.wasm all

build:
	go build ./...

test: bin/sqlc-gen-go.wasm
	go test ./...

all: bin/sqlc-gen-go bin/sqlc-gen-go.wasm

bin/sqlc-gen-go: bin
	cd plugin && go build -o ../bin/sqlc-gen-glua ./main.go

bin/sqlc-gen-go.wasm: bin/sqlc-gen-go
	cd plugin && GOOS=wasip1 GOARCH=wasm go build -o ../bin/sqlc-gen-glua.wasm main.go

bin:
	mkdir -p bin
