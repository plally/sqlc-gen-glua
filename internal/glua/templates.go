package glua

import (
	"embed"
	"text/template"
)

//go:embed templates/*
var templateFS embed.FS

var rootTemplate = template.Must(template.ParseFS(templateFS, "templates/*.tmpl"))

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
