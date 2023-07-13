package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin/middleware/logger"
)

func main() {
	//gin.SetMode(gin.ReleaseMode)
	g := gin.New()
	//curl -X GET -d 'username=admin' -d 'password=admin' http://127.0.0.1:8080
	g.Use(logger.Next())
	//g.Use(gin.Logger())
	g.GET("/", func(context *gin.Context) {
		data, _ := context.GetRawData()
		context.JSON(http.StatusOK, string(data))
	})
	g.Run()
}
