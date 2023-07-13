package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/xstring"
)

type RequestIDConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
	// Generator defines a function to generate an ID.
	// Optional. Default value random.String(32).
	Generator func() string
	// RequestIDHandler defines a function which is executed for a request id.
	RequestIDHandler func(*gin.Context, string)
	// TargetHeader defines what header to look for to populate the id
	TargetHeader string
}

var defaultRequestIDConfig = RequestIDConfig{
	Skipper: DefaultSkipper,
	Generator: func() string {
		return xstring.RandomN(32)
	},
	TargetHeader: igin.HeaderXRequestID,
}

// RequestIdNext returns a X-Request-ID middleware.
func RequestIdNext() gin.HandlerFunc {
	return RequestIdNextWithConfig(defaultRequestIDConfig)
}

func RequestIdNextWithConfig(config RequestIDConfig) gin.HandlerFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultRequestIDConfig.Skipper
	}
	if config.Generator == nil {
		config.Generator = defaultRequestIDConfig.Generator
	}
	if config.TargetHeader == "" {
		config.TargetHeader = igin.HeaderXRequestID
	}
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		rid := c.Request.Header.Get(config.TargetHeader)
		if rid == "" {
			rid = config.Generator()
		}
		c.Header(config.TargetHeader, rid)
		if config.RequestIDHandler != nil {
			config.RequestIDHandler(c, rid)
		}
		c.Next()
	}
}
