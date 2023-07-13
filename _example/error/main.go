package main

import (
	"github.com/pkg6/igin/xerror"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/middleware"
)

func main() {
	g := igin.Default()
	g.Use(func(context *gin.Context) {
		//igin.AddStatusError(context, errors.NewHTTPError(201, "test error"))
		igin.AddStatusError(context, xerror.NewCodeMsg(201, "test error"))
		//igin.AddStatusError(context, fmt.Errorf("test error"))
		//igin.AddStatusError(context, fmt.Errorf("test error"), 202)
	})
	g.Use(middleware.ErrorsNext())
	g.GET("/", func(context *gin.Context) {
		context.String(200, "ok")
	})
	g.Run()
}
