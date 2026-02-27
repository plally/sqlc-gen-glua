package glua

import (
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

	var files []*plugin.File
	dalContent, err := templateFile("dal", data)
	if err != nil {
		return nil, err
	}

	files = append(files, &plugin.File{
		Contents: dalContent,
		Name:     "dal.lua",
	})
	drivers := []string{"gmod", "libsql"}
	for _, driver := range drivers {
		driverContent, err := templateFile("drivers/"+driver, data)
		if err != nil {
			return nil, err
		}
		files = append(files, &plugin.File{
			Contents: driverContent,
			Name:     "drivers/" + driver + ".lua",
		})
	}

	return files, nil
}
