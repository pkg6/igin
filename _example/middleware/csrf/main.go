package main

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/middleware"
)

func main() {
	g := igin.Default()
	//g.Use(middleware.CSRFNextWithConfig(middleware.CSRFConfig{
	//	Skipper:        middleware.DefaultSkipper,
	//	TokenLength:    32,
	//	TokenLookup:    middleware.ExtractorMethodForm + ":" + igin.HeaderXCSRFToken,
	//	ContextKey:     middleware.DefaultCSRFCookieName,
	//	CookieName:     "_" + middleware.DefaultCSRFCookieName,
	//	CookieMaxAge:   86400,
	//	CookieSameSite: http.SameSiteDefaultMode,
	//}))
	g.Use(middleware.CSRFNext())
	g.LoadHTMLGlob("_example/middleware/csrf/*")
	g.GET("/", func(c *gin.Context) {
		csrfToken := c.GetString(middleware.CSRFContextKey)
		c.HTML(http.StatusOK, "csrf.html", gin.H{
			"name":  igin.HeaderXCSRFToken,
			"token": csrfToken,
			"_csrf": middleware.CSRFFormHTML(c, "text"),
		})
	})
	g.POST("/action", func(c *gin.Context) {
		ctx := igin.Context(c)
		c.JSON(http.StatusOK, gin.H{
			"csrf": ctx.PostForm(),
		})
	})
	g.Run()
}
