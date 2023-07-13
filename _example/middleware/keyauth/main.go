package main

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin/middleware"
)

func main() {
	//curl -X GET -H "Authorization: Bearer <token>" http://127.0.0.1:8080
	g := gin.Default()
	g.Use(middleware.KeyAuthNext(func(auth string, c *gin.Context) (bool, error) {
		fmt.Println(auth)
		return true, nil
	}))
	g.GET("/", func(c *gin.Context) {
		c.String(http.StatusOK, "ok")
	})
	g.Run()
}
