package glua

import (
	"context"
	"encoding/json"
	"fmt"
	"slices"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type generatorState struct {
	modelTypes []luaTypeData
}

type Options struct {
	GlobalLuaTable string `json:"global_lua_table"`
}

func getOptions(req *plugin.GenerateRequest) (Options, error) {
	options := req.GetSettings().GetCodegen().GetOptions()
	if options == nil {
		return Options{}, nil
	}

	var opts Options
	err := json.Unmarshal(options, &opts)
	if err != nil {
		return Options{}, err
	}

	if opts.GlobalLuaTable == "" {
		return Options{}, fmt.Errorf("global_lua_table option is required %v", string(options))
	}

	return opts, nil
}
func Generate(ctx context.Context, req *plugin.GenerateRequest) (*plugin.GenerateResponse, error) {
	opts, err := getOptions(req)
	if err != nil {
		return nil, err
	}
	dalFiles, err := genDalFiles(req, opts)
	if err != nil {
		return nil, err
	}

	modelsFiles, err := genModelsFile(req, opts)
	if err != nil {
		return nil, err
	}

	queriesFiles, err := genQueryFiles(req, opts)
	if err != nil {
		return nil, err
	}

	reqJson, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, err
	}

	files := slices.Concat(queriesFiles, modelsFiles, dalFiles)
	files = append(files, &plugin.File{
		Name:     "request.json",
		Contents: reqJson,
	})

	resp := &plugin.GenerateResponse{
		Files: files,
	}
	return resp, err
}
