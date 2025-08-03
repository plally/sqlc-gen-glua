package main

import (
	"github.com/sqlc-dev/plugin-sdk-go/codegen"

	"github.com/plally/sqlc-gen-glua/internal/glua"
)

func main() {
	codegen.Run(glua.Generate)
}
