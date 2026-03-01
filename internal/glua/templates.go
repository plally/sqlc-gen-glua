package glua

import (
	"embed"
	"strings"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

var rootTemplate = template.Must(template.ParseFS(templateFS, "templates/*.tmpl", "templates/drivers/*.tmpl"))

type luaTypeData struct {
	TypeName string
	Fields   []luaTypeField
}

type luaTypeField struct {
	Name string
	Type string
}

func mustReadFile(path string) []byte {
	b, err := templateFS.ReadFile(path)
	if err != nil {
		panic(err)
	}
	return b
}

func templateFile(path string, data any) ([]byte, error) {
	var builder strings.Builder
	err := rootTemplate.ExecuteTemplate(&builder, path, data)
	if err != nil {
		return nil, err
	}
	return []byte(builder.String()), nil
}
