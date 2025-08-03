package glua

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

func genModelsFile(req *plugin.GenerateRequest, opts Options) ([]*plugin.File, error) {
	catalog := req.Catalog
	if catalog == nil {
		return nil, fmt.Errorf("catalog is nil")
	}

	var builder strings.Builder
	for _, schema := range catalog.Schemas {
		for _, table := range schema.Tables {
			err := genTable(&builder, table)
			if err != nil {
				return nil, fmt.Errorf("error generating table %s: %w", table.GetRel().GetName(), err)
			}
		}
	}

	return []*plugin.File{
		{Contents: []byte(builder.String()), Name: "models.lua"},
	}, nil
}

func genTable(builder *strings.Builder, table *plugin.Table) error {
	name := table.GetRel().GetName()
	typeName := strings.Title(strings.ReplaceAll(name, "_", " "))
	typeName = strings.ReplaceAll(typeName, " ", "")

	fields, err := columnsToLuaTypeFields(table.GetColumns(), SnakeToCamel)
	if err != nil {
		return fmt.Errorf("error converting columns to lua type fields for table %s: %w", name, err)
	}

	data := luaTypeData{
		TypeName: typeName,
		Fields:   fields,
	}

	return rootTemplate.ExecuteTemplate(builder, "luaType", data)
}

var luaTypeMappings = map[string]string{
	"TEXT":    "string",
	"INTEGER": "number",
	"integer": "number",
}

func columnsToLuaTypeFields(columns []*plugin.Column, nameModifier func(string) string) ([]luaTypeField, error) {
	if nameModifier == nil {
		nameModifier = func(name string) string {
			return name
		}
	}

	out := make([]luaTypeField, 0, len(columns))
	for _, column := range columns {
		columnType, ok := luaTypeMappings[column.GetType().GetName()]
		if !ok {
			return nil, fmt.Errorf("unsupported column type: %s", column.GetType().GetName())
		}

		if column.GetIsSqlcSlice() {
			columnType = columnType + "[]"
		}

		if !column.GetNotNull() {
			columnType += "?"
		}
		out = append(out, luaTypeField{
			Name: nameModifier(column.GetName()),
			Type: columnType,
		})
	}

	return out, nil
}
