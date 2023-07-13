package main

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	jwt2 "github.com/pkg6/igin/middleware/jwt"
)

var jwtKey = "secretdasjdkasjdlkasjdlkasjd"

type JwtClaims struct {
	Name string
	jwt.StandardClaims
}

//func main() {
//	//curl -X POST -d 'username=admin' -d 'password=admin' http://127.0.0.1:8080/login
//	g := gin.Default()
//	g.POST("/login", func(c *gin.Context) {
//		username := c.PostForm("username")
//		password := c.PostForm("password")
//		if username == "admin" && password == "admin" {
//			//// Create token
//			//token := jwt.New(jwt.SigningMethodHS256)
//			//// Set claims
//			//claims := token.Claims.(jwt.MapClaims)
//			//claims["name"] = "admin"
//			//claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
//			// Generate encoded token and send it as response.
//			//t, _ := token.SignedString([]byte("secret"))
//
//			claims := &JwtClaims{
//				"admin", jwt.StandardClaims{
//					ExpiresAt: time.Now().Add(72 * time.Hour).Unix(),
//				},
//			}
//			token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
//			t, _ := token.SignedString([]byte(jwtKey))
//
//			c.JSON(http.StatusOK, map[string]string{
//				"token": t,
//			})
//			return
//		}
//		c.JSON(http.StatusUnauthorized, nil)
//	})
//
//	g.Use(jwt2.NextWithConfig(jwt2.JWTConfig{
//		Claims:     &JwtClaims{},
//		SigningKey: []byte(jwtKey),
//	}))
//	//curl http://127.0.0.1:8080/info -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJOYW1lIjoiYWRtaW4iLCJleHAiOjE2ODY1NTg1MjF9.89k6RMz2hYU80zYvDiZIjFRtYKDQ_cNMuElgFhlCPQU"
//	g.GET("/info", func(c *gin.Context) {
//		if value, exists := c.Get(jwt2.ContextKey); exists {
//			c.JSON(http.StatusOK, value)
//		}
//		c.JSON(http.StatusNotFound, nil)
//	})
//	g.Run()
//}

func main() {
	//curl -X POST -d 'username=admin' -d 'password=admin' http://127.0.0.1:8080/login
	g := gin.Default()
	g.POST("/login", func(c *gin.Context) {
		username := c.PostForm("username")
		password := c.PostForm("password")
		if username == "admin" && password == "admin" {
			// Create token
			token := jwt.New(jwt.SigningMethodHS256)
			// Set claims
			claims := token.Claims.(jwt.MapClaims)
			claims["name"] = "admin2"
			claims["exp"] = time.Now().Add(time.Hour * 72).Unix()
			//Generate encoded token and send it as response.
			t, _ := token.SignedString([]byte(jwtKey))
			c.JSON(http.StatusOK, map[string]string{
				"token": t,
			})
			return
		}
		c.JSON(http.StatusUnauthorized, nil)
	})
	g.Use(jwt2.Next(jwtKey))
	//curl http://127.0.0.1:8080/info -H "Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE2ODY5MDk3MTMsIm5hbWUiOiJhZG1pbjIifQ.VyoZz1XM0JLnefaqFxdaVw5zDlBFLJZSWAhxIXYI1Ww"
	g.GET("/info", func(c *gin.Context) {
		token, err := jwt2.ContextToken(c)
		if err != nil {
			c.JSON(http.StatusNotFound, err)
			return
		}
		c.JSON(http.StatusOK, token)
	})
	g.Run()
}
