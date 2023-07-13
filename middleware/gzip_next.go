package middleware

import (
	"compress/gzip"
	"fmt"
	"io"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

type (
	// GzipConfig defines the config for Gzip middleware.
	GzipConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper
		// Gzip compression level.
		// Optional. Default value -1.
		Level int
	}
)

var (
	// DefaultGzipConfig is the default Gzip middleware config.
	defaultGzipConfig = GzipConfig{
		Skipper: DefaultSkipper,
		Level:   -1,
	}
	gzipScheme = "gzip"
)

// GzipNext returns a middleware which compresses HTTP response using gzip compression
// scheme.
func GzipNext() gin.HandlerFunc {
	return GzipNextWithConfig(defaultGzipConfig)
}

func GzipNextWithConfig(config GzipConfig) gin.HandlerFunc {
	if config.Skipper == nil {
		config.Skipper = defaultGzipConfig.Skipper
	}
	pool := gzipCompressPool(config)
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		gz := pool.Get().(*gzip.Writer)
		gz.Reset(c.Writer)
		req := c.Request
		if !strings.Contains(req.Header.Get(igin.HeaderAcceptEncoding), gzipScheme) ||
			strings.Contains(req.Header.Get(igin.HeaderConnection), igin.HeaderUpgrade) ||
			strings.Contains(req.Header.Get(igin.HeaderAccept), igin.MIMEEventStream) {
			return
		}
		c.Header(igin.HeaderVary, igin.HeaderAcceptEncoding)
		c.Header(igin.HeaderContentEncoding, gzipScheme)
		grw := &gzipWriter{writer: gz, ResponseWriter: c.Writer}
		defer func() {
			gz.Close()
			gz.Reset(io.Discard)
			pool.Put(gz)
			c.Header(igin.HeaderContentLength, fmt.Sprint(c.Writer.Size()))
		}()
		c.Writer = grw
		c.Next()
	}
}

func gzipCompressPool(config GzipConfig) sync.Pool {
	return sync.Pool{
		New: func() interface{} {
			w, err := gzip.NewWriterLevel(io.Discard, config.Level)
			if err != nil {
				return err
			}
			return w
		},
	}
}

type gzipWriter struct {
	gin.ResponseWriter
	writer *gzip.Writer
}

func (g *gzipWriter) WriteString(s string) (int, error) {
	g.Header().Del(igin.HeaderContentLength)
	return g.writer.Write([]byte(s))
}

func (g *gzipWriter) Write(data []byte) (int, error) {
	g.Header().Del(igin.HeaderContentLength)
	return g.writer.Write(data)
}

// WriteHeader Fix: https://github.com/mholt/caddy/issues/38
func (g *gzipWriter) WriteHeader(code int) {
	g.Header().Del(igin.HeaderContentLength)
	g.ResponseWriter.WriteHeader(code)
}
