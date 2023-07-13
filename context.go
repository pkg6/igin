package igin

import (
	"net/http"
	"net/url"

	"github.com/gin-gonic/gin"
)

type GinContext struct {
	*gin.Context
}

// Context 重新初始化Context
func Context(c *gin.Context) *GinContext {
	ctx := &GinContext{}
	ctx.Context = c
	return ctx
}

// Host 请求的host
func (c *GinContext) Host() string {
	return c.Request.Host
}

// Path 请求的路径(不附带querystring)
func (c *GinContext) Path() string {
	return c.Request.URL.Path
}

// URI unescape后的uri
func (c *GinContext) URI() string {
	uri, _ := url.QueryUnescape(c.Request.URL.RequestURI())
	return uri
}

// Form 获取所有参数
func (c *GinContext) Form() url.Values {
	_ = c.Request.ParseForm()
	return c.Request.Form
}

// PostForm 获取 PostForm 参数
func (c *GinContext) PostForm() url.Values {
	_ = c.Request.ParseForm()
	return c.Request.PostForm
}

// Header 获取 Header 参数
func (c *GinContext) Header() http.Header {
	header := c.Request.Header
	clone := make(http.Header, len(header))
	for k, v := range header {
		value := make([]string, len(v))
		copy(value, v)
		clone[k] = value
	}
	return clone
}

// IsPost 是否为POST请求
func (c *GinContext) IsPost() bool {
	return c.Request.Method == http.MethodPost
}

// IsGet 是否为GET请求
func (c *GinContext) IsGet() bool {
	return c.Request.Method == http.MethodGet
}

// IsPut 是否为PUT请求
func (c *GinContext) IsPut() bool {
	return c.Request.Method == http.MethodPut
}

// IsDelete 是否为DELTE请求
func (c *GinContext) IsDelete() bool {
	return c.Request.Method == http.MethodDelete
}

// IsHead 是否为HEAD请求
func (c *GinContext) IsHead() bool {
	return c.Request.Method == http.MethodHead
}

// IsPatch 是否为PATCH请求
func (c *GinContext) IsPatch() bool {
	return c.Request.Method == http.MethodPatch
}

// IsOptions 是否为OPTIONS请求
func (c *GinContext) IsOptions() bool {
	return c.Request.Method == http.MethodOptions
}
