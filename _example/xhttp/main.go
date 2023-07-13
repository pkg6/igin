package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/xhttp"
)

func main() {
	i := igin.New()
	//{
	//  "code": 0,
	//  "msg": "ok",
	//  "data": "json ok"
	//}
	i.GET("/json", func(context *gin.Context) {
		xhttp.JsonBaseResponse(context, "json ok")
	})
	//{
	//  "code": -1,
	//  "msg": "json no ok"
	//}
	i.GET("/jsone", func(context *gin.Context) {
		xhttp.JsonBaseResponse(context, fmt.Errorf("json no ok"))
	})
	//<xml version="1.0" encoding="UTF-8">
	//<code>0</code>
	//<msg>ok</msg>
	//<data>xml ok</data>
	//</xml>
	i.GET("/xml", func(context *gin.Context) {
		xhttp.XmlBaseResponse(context, "xml ok")
	})
	//<xml version="1.0" encoding="UTF-8">
	//<code>-1</code>
	//<msg>xml no ok</msg>
	//</xml>
	i.GET("/xmle", func(context *gin.Context) {
		xhttp.XmlBaseResponse(context, fmt.Errorf("xml no ok"))
	})
	i.Run()
}
