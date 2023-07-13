package igin

import (
	"html/template"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

type IController interface {
	Prefix() string
	Routes(g gin.IRoutes)
}

type IPlugin interface {
	Register(group *gin.RouterGroup)
	RouterPath() string
}

type Engine struct {
	*gin.Engine
}

func Default() *Engine {
	return NewEngine(gin.Default())
}
func New() *Engine {
	return NewEngine(gin.New())
}
func NewEngine(engine *gin.Engine) *Engine {
	e := &Engine{}
	e.Engine = engine
	return e
}
func (e *Engine) BindingValidatorEngine(locale string, translators ...*translator) {
	if len(translators) > 0 {
		Translator = translators[0]
	} else {
		Translator = DefaultTranslator()
	}
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		_, _ = Translator.UtTranslator(v, locale)
	}
}
func (e *Engine) SetFuncMaps(funcMaps ...template.FuncMap) *Engine {
	funcMaps = append(funcMaps, defaultTemplateFuncMaps)
	for _, funcMap := range funcMaps {
		e.SetFuncMap(funcMap)
	}
	return e
}
func (e *Engine) PrefixController(prefix string, controllers ...IController) {
	eg := e.Engine.Group(prefix)
	var route gin.IRoutes
	for _, gc := range controllers {
		if eg == nil || gc == nil {
			continue
		}
		route = eg
		if len(gc.Prefix()) > 1 {
			route = eg.Group(gc.Prefix())
		}
		gc.Routes(route)
	}
}
func (e *Engine) Controller(controllers ...IController) {
	e.PrefixController("", controllers...)
}
func (e *Engine) Plugin(Plugins ...IPlugin) {
	group := e.Group("")
	for i := range Plugins {
		PluginGroup := group.Group(Plugins[i].RouterPath())
		Plugins[i].Register(PluginGroup)
	}
}
