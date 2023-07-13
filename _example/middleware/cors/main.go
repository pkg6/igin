package main

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin/middleware"
	"golang.org/x/sync/errgroup"
	"log"
	"net/http"
	"time"
)

var (
	g errgroup.Group
)

func main() {
	g1 := gin.Default()
	g1.LoadHTMLGlob("_example/middleware/cors/*")
	g1.GET("/", func(c *gin.Context) {
		c.HTML(200, "index.html", nil)
	})
	g2 := gin.Default()
	g2.Use(middleware.CORSNext())
	g2.POST("/", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"code":    200,
			"message": "ok",
		})
	})
	s1 := &http.Server{
		Addr:         ":8080",
		Handler:      g1,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	s2 := &http.Server{
		Addr:         ":8081",
		Handler:      g2,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	g.Go(func() error {
		return s1.ListenAndServe()
	})

	g.Go(func() error {
		return s2.ListenAndServe()
	})

	if err := g.Wait(); err != nil {
		log.Fatal(err)
	}
}
