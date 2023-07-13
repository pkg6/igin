package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/middleware/firebase"
)

//客户端sdk： https://firebase.google.com/docs/auth/?hl=zh-cn
//golang sdk: https://github.com/firebase/firebase-admin-go
//服务端初始化文档： https://firebase.google.com/docs/admin/setup?hl=zh-cn
//服务端解析IdToken文档：https://firebase.google.com/docs/auth/admin/verify-id-tokens?hl=zh-cn
//后端开发阶段，可以用自定义token调试：https://firebase.google.com/docs/auth/admin/create-custom-tokens?hl=zh-cn
//授权流程：客户端第三方授权登录，换取idtoken -> 服务端获取idtoken，通过上述文档解析IdToken，拿到用户的授权信息
//自定义token换取idToken https://firebase.google.com/docs/reference/rest/auth?hl=zh-cn
func main() {
	r := igin.New()
	projectId := os.Getenv("FIREBASE_PROJECT_ID")
	credentialsFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")

	authClient, err := firebase.NewAuthClient(
		projectId,
		credentialsFile,
		os.Getenv("FIREBASE_APIKEY"),
	)
	if err != nil {
		log.Fatal(err)
	}
	//curl -X POST  http://127.0.0.1:8080/login
	r.POST("/login", func(c *gin.Context) {
		token, err := authClient.CustomToken(authClient.Ctx, "test")
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"code":    http.StatusBadGateway,
				"message": err.Error(),
			})
			c.Abort()
			return
		}
		idToken, err := authClient.TokenToIDToken(token)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{
				"code":    http.StatusBadGateway,
				"message": err.Error(),
			})
			c.Abort()
			return
		}
		c.JSON(http.StatusOK, idToken)
	})

	if gin.Mode() == gin.ReleaseMode {
		r.Use(firebase.NextWithAuthClient(authClient))
		//r.Use(firebase.NextWithAuthClientSuccessHandler(authClient, func(c *gin.Context) {
		//	if token, exists := c.Get(firebase.ContextKey); exists {
		//		// do something
		//		fmt.Println(token)
		//	}
		//}))
	} else {
		r.Use(firebase.Next(projectId, credentialsFile))
		//r.Use(firebase.NextSuccessHandler(projectId, credentialsFile, func(c *gin.Context) {
		//	if token, exists := c.Get(firebase.ContextKey); exists {
		//		// do something
		//		fmt.Println(token)
		//	}
		//}))
	}

	//curl http://127.0.0.1:8080/info -H "Authorization: Bearer idToken"
	r.GET("/info", func(c *gin.Context) {
		token, err := firebase.ContextToken(c)
		if err != nil {
			c.JSON(http.StatusNotFound, err)
			return
		}
		c.JSON(http.StatusOK, token)
	})
	r.Run()
}
