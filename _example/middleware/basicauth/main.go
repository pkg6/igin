package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin/middleware"
)

func main() {
	g := gin.Default()
	g.Use(middleware.BasicAuthNext(func(s1, s2 string, c *gin.Context) bool {
		if s1 == "name" && s2 == "pwd" {
			return true
		}
		return false
	}))
	g.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	g.Run()
}
