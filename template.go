package igin

import (
	"html/template"
)

var (
	defaultTemplateFuncMaps = template.FuncMap{
		"raw": func(str string) template.HTML {
			return template.HTML(str)
		},
	}
)
