package main

import (
	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/pkg6/igin/middleware/session"
)

func main() {
	r := gin.Default()
	r.Use(session.Next())
	r.GET("/set", func(c *gin.Context) {
		session.Set(c, "user", session.Value{"id": 1, "name": "admin"})
		session.Set(c, "user1", session.Value{"id": 2, "name": "admin1"}, &sessions.Options{})
		session.Set(c, "user2", session.Value{"id": 3, "name": "admin2"}, &sessions.Options{
			Path:   "/",       //所有页面都可以访问会话数据
			MaxAge: 86400 * 7, //会话有效期，单位秒
		})
		c.String(200, "登录成功!")
	})

	r.GET("/del", func(c *gin.Context) {
		session.Delete(c, "user")
		session.Delete(c, "user1")
		session.Delete(c, "user2")
		sess, _ := session.Get(c, "user")
		sess1, _ := session.Get(c, "user1")
		sess2, _ := session.Get(c, "user2")
		id := sess["id"]
		id1 := sess1["id"]
		id2 := sess2["id"]
		c.JSON(200, gin.H{
			"user":  id,
			"user1": id1,
			"user2": id2,
		})
	})
	r.GET("/get", func(c *gin.Context) {
		sess, _ := session.Get(c, "user")
		sess1, _ := session.Get(c, "user1")
		sess2, _ := session.Get(c, "user2")
		id := sess["id"]
		id1 := sess1["id"]
		id2 := sess2["id"]
		c.JSON(200, gin.H{
			"user":  id,
			"user1": id1,
			"user2": id2,
		})
	})
	r.Run()
}
