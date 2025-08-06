package glua

import (
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type dalData struct {
	GlobalLuaTable string
	Filenames      []string
}

func genDalFiles(req *plugin.GenerateRequest, opts Options, filenames []string) ([]*plugin.File, error) {
	data := dalData{
		GlobalLuaTable: opts.GlobalLuaTable,
		Filenames:      filenames,
	}

	var builder strings.Builder
	err := rootTemplate.ExecuteTemplate(&builder, "dal", data)
	if err != nil {
		return nil, err
	}

	return []*plugin.File{
		{Contents: []byte(builder.String()), Name: "dal.lua"},
		{Contents: mustReadFile("templates/drivers/gmod.lua"), Name: "drivers/gmod.lua"},
		{Contents: mustReadFile("templates/drivers/libsql.lua"), Name: "drivers/libsql.lua"},
	}, nil
}
