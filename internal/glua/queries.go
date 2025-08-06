package glua

import (
	"fmt"
	"strings"

	"github.com/sqlc-dev/plugin-sdk-go/plugin"
)

type queryTemplateData struct {
	Query      string
	Name       string
	Params     []queryTemplateParam
	ParamsType string

	ReturnType     string
	ReturnMappings map[string]string

	ReturnsSlice bool

	GlobalLuaTable string
}

type queryTemplateParam struct {
	Name        string
	NotNull     bool
	IsSqlcSlice bool
}

var queriesFileHeader = `

local function repeatString(str, count)
    local result = {}
    for i = 1, count do
        result[i] = str
    end
    return table.concat(result, ",")
end
`

var queriesFileFooter = `
`

func genQueryFiles(req *plugin.GenerateRequest, opts Options) ([]*plugin.File, error) {
	builders := make(map[string]*strings.Builder)

	for _, query := range req.Queries {
		if builders[query.GetFilename()] == nil {
			builders[query.GetFilename()] = &strings.Builder{}
			builders[query.GetFilename()].WriteString(queriesFileHeader)
		}
		builder := builders[query.GetFilename()]
		err := GenQuery(builder, req, query, opts)
		if err != nil {
			return nil, err
		}
	}

	for _, builder := range builders {
		builder.WriteString(queriesFileFooter)
	}

	files := make([]*plugin.File, 0, len(builders))
	for name, builder := range builders {
		files = append(files, &plugin.File{
			Name:     name + ".lua",
			Contents: []byte(builder.String()),
		})
	}

	return files, nil
}

func GenQuery(builder *strings.Builder, req *plugin.GenerateRequest, q *plugin.Query, opts Options) error {
	query := q.GetText()
	if query == "" {
		return fmt.Errorf("query text is empty")
	}

	data := queryTemplateData{
		Query:          query,
		Name:           q.GetName(),
		Params:         make([]queryTemplateParam, 0, len(q.GetParams())),
		ParamsType:     fmt.Sprintf("%sParams", strings.Title(strings.ReplaceAll(q.GetName(), "_", " "))),
		ReturnMappings: make(map[string]string),

		GlobalLuaTable: opts.GlobalLuaTable,
	}

	tableTypeName := ""
	{
		columnTableName := ""
		if len(q.GetColumns()) > 0 {
			columnTableName = q.GetColumns()[0].GetTable().GetName()
			for _, column := range q.GetColumns() {
				if column.GetTable().GetName() != columnTableName {
					columnTableName = ""
					break
				}
			}
		}

		var foundTable *plugin.Table
		catalog := req.GetCatalog()
		for _, schema := range catalog.Schemas {
			for _, table := range schema.Tables {
				if table.GetRel().GetName() == columnTableName {
					foundTable = table
					break
				}
			}
		}

		if foundTable != nil {
			tableTypeName = SnakeToPascal(foundTable.GetRel().GetName())
			for i, column := range foundTable.GetColumns() {
				qColumn := q.GetColumns()[i]
				if column.GetName() != qColumn.GetName() {
					tableTypeName = ""
					break
				}
				if column.GetType().GetName() != qColumn.GetType().GetName() {
					tableTypeName = ""
					break
				}
			}
		}
	}

	returnTypeName := fmt.Sprintf("%sResult", SnakeToPascal(q.GetName()))
	if tableTypeName != "" {
		returnTypeName = tableTypeName
	}
	if q.GetCmd() == ":exec" {
		data.ReturnType = "nil"
	} else if q.GetCmd() == ":one" {
		data.ReturnType = returnTypeName
	} else if q.GetCmd() == ":many" {
		data.ReturnType = fmt.Sprintf("%s[]", returnTypeName)
		data.ReturnsSlice = true
	} else {
		return fmt.Errorf("unsupported query command: %s", q.GetCmd())
	}

	err := genParamsType(builder, data.ParamsType, q.GetParams(), opts)
	if err != nil {
		return fmt.Errorf("error generating params type for query %s: %w", q.GetName(), err)
	}

	if data.ReturnType != "nil" && tableTypeName == "" {
		err := genReturnType(builder, data.ReturnType, q.GetColumns(), opts)
		if err != nil {
			return fmt.Errorf("error generating return type for query %s: %w", q.GetName(), err)
		}
	}

	noMappings := true
	for _, column := range q.GetColumns() {
		oldName := column.GetName()

		newName := oldName
		if opts.Rename[oldName] != "" {
			newName = opts.Rename[oldName]
		} else {
			newName = SnakeToCamel(oldName)
		}
		if oldName != newName {
			noMappings = false
		}

		data.ReturnMappings[oldName] = newName
	}
	if noMappings {
		data.ReturnMappings = nil
	}

	// TODO should i use p.Number
	for _, p := range q.GetParams() {
		name := p.GetColumn().GetName()
		if opts.Rename[name] != "" {
			name = opts.Rename[name]
		} else {
			name = SnakeToCamel(name)
		}

		data.Params = append(data.Params, queryTemplateParam{
			Name:        name,
			IsSqlcSlice: p.GetColumn().GetIsSqlcSlice(),
			NotNull:     p.GetColumn().GetNotNull(),
		})
	}

	return rootTemplate.ExecuteTemplate(builder, "queryFunc", data)
}

func genReturnType(builder *strings.Builder, name string, columns []*plugin.Column, opts Options) error {

	fields, err := columnsToLuaTypeFields(columns, func(name string) string {
		if opts.Rename[name] != "" {
			return opts.Rename[name]
		}
		return SnakeToCamel(name)
	})
	if err != nil {
		return fmt.Errorf("error converting columns to lua type fields for table %s: %w", name, err)
	}

	data := luaTypeData{
		TypeName: name,
		Fields:   fields,
	}

	return rootTemplate.ExecuteTemplate(builder, "luaType", data)
}

func genParamsType(builder *strings.Builder, name string, params []*plugin.Parameter, opts Options) error {
	if len(params) == 0 {
		return nil
	}

	columns := make([]*plugin.Column, 0, len(params))
	for _, p := range params {
		columns = append(columns, p.GetColumn())
	}

	fields, err := columnsToLuaTypeFields(columns, func(name string) string {
		if opts.Rename[name] != "" {
			return opts.Rename[name]
		}
		return SnakeToCamel(name)
	})
	if err != nil {
		return fmt.Errorf("error converting columns to lua type fields for table %s: %w", name, err)
	}

	data := luaTypeData{
		TypeName: name,
		Fields:   fields,
	}

	return rootTemplate.ExecuteTemplate(builder, "luaType", data)
}

func SnakeToPascal(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		parts[i] = strings.Title(parts[i])
	}
	return strings.Join(parts, "")
}

func SnakeToCamel(s string) string {
	parts := strings.Split(s, "_")
	for i := range parts {
		if i > 0 {
			parts[i] = strings.Title(parts[i])
		}
	}
	return strings.Join(parts, "")
}
