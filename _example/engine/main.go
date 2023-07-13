package main

import (
	"github.com/pkg6/igin/xhttp"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

type DemoController struct {
}

func (a DemoController) Prefix() string {
	return "/demo"
}
func (a DemoController) Routes(g gin.IRoutes) {
	g.GET("demo", a.demo)
	g.POST("post", a.request)
}

func (a DemoController) demo(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"code": http.StatusOK, "message": "ok", "data": ""})
}

type request struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

//curl --location --request POST 'http://127.0.0.1:8080/demo/post' --header 'Content-Type: application/json' --data-raw '{"username": "github","password": "123456"}'
//curl --location --request POST 'http://127.0.0.1:8080/demo/post' --header 'Content-Type: application/json' --data-raw '{"password": "123456"}'
func (a DemoController) request(ctx *gin.Context) {
	var r request
	err := ctx.Bind(&r)
	if err != nil {
		igin.JsonError(ctx, err)
		return
	}
	//验证 存储操作省略....
	igin.JsonSuccess(ctx, r)
}

type PrefixController struct {
}

func (a PrefixController) Prefix() string {
	return ""
}
func (a PrefixController) Routes(g gin.IRoutes) {
	g.GET("demo", func(context *gin.Context) {
		xhttp.JsonBaseResponse(context, "prefix ok")
	})
}

type demoPlugin struct {
}

func (d demoPlugin) Register(group *gin.RouterGroup) {
	group.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "ok")
	})
}

func (d demoPlugin) RouterPath() string {
	return "demo2"
}

func main() {
	g := igin.Default()
	g.BindingValidatorEngine(igin.TranslatorLocaleZH)
	g.GET("/", func(context *gin.Context) {
		context.String(http.StatusOK, "ok")
	})
	g.Controller(&DemoController{})
	g.PrefixController("/prefix", &PrefixController{})
	g.Plugin(&demoPlugin{})
	g.Run()
}
