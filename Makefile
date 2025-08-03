.PHONY: build test bin/sqlc-gen-go bin/sqlc-gen-go.wasm all bin

build:
	go build ./...

test: bin/sqlc-gen-go.wasm
	go test ./...

all: bin/sqlc-gen-go bin/sqlc-gen-glua.wasm

bin/sqlc-gen-glua: bin
	cd plugin && go build -o ../bin/sqlc-gen-glua ./main.go

bin/sqlc-gen-glua.wasm: bin/sqlc-gen-glua
	cd plugin && GOOS=wasip1 GOARCH=wasm go build -o ../bin/sqlc-gen-glua.wasm main.go

bin:
	mkdir -p bin
