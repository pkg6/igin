package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin/middleware"
)

func main() {
	g := gin.Default()
	g.Use(middleware.RequestIdNext())
	g.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	g.Run()
}
