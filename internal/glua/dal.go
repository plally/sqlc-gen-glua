package glua

import (
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type dalData struct {
	GlobalLuaTable string
}

func genDalFiles(req *plugin.GenerateRequest, opts Options) ([]*plugin.File, error) {
	data := dalData{
		GlobalLuaTable: opts.GlobalLuaTable,
	}

	var builder strings.Builder
	err := rootTemplate.ExecuteTemplate(&builder, "dal", data)
	if err != nil {
		return nil, err
	}

	return []*plugin.File{
		{Contents: []byte(builder.String()), Name: "dal.lua"},
	}, nil
}
