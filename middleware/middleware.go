package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

type (
	// Skipper defines a function to skip middleware. Returning true skips processing
	// the middleware.
	Skipper func(c *gin.Context) bool
	// SuccessHandler defines a function to Handler middleware
	SuccessHandler func(c *gin.Context)
	// ErrorHandler middleware handler error
	ErrorHandler func(c *gin.Context, err error, statusCodes ...int)
)

// DefaultSkipper returns false which processes the middleware.
func DefaultSkipper(ctx *gin.Context) bool {
	return false
}

// DefaultErrorHandler 默认错误返回响应
func DefaultErrorHandler(c *gin.Context, err error, statusCodes ...int) {
	igin.JsonError(c, err, statusCodes...)
}
