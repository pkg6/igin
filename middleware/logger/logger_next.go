package logger

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/mattn/go-isatty"
	"github.com/pkg6/igin/middleware"
)

const (
	green   = "\033[97;42m"
	white   = "\033[90;47m"
	yellow  = "\033[90;43m"
	red     = "\033[97;41m"
	blue    = "\033[97;44m"
	magenta = "\033[97;45m"
	cyan    = "\033[97;46m"
	reset   = "\033[0m"
)

// LoggerConfig defines the config for Logger middleware.
type LoggerConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper middleware.Skipper
	// Optional. Default value is gin.defaultLogFormatter
	Formatter LogFormatter
	// Output is a writer where logs are written.
	// Optional. Default value is gin.DefaultWriter.
	Output io.Writer
	// SkipPaths is an url path array which logs are not written.
	// Optional.
	SkipPaths []string

	SkipBody middleware.Skipper
}

var (
	defaultConfig = LoggerConfig{
		Skipper: middleware.DefaultSkipper,
		Output:  gin.DefaultWriter,
		SkipBody: func(c *gin.Context) bool {
			return true
		},
		Formatter: func(param LogFormatterParams) string {
			var statusColor, methodColor, resetColor string
			if param.IsOutputColor() {
				statusColor = param.StatusCodeColor()
				methodColor = param.MethodColor()
				resetColor = param.ResetColor()
			}
			if param.Latency > time.Minute {
				param.Latency = param.Latency.Truncate(time.Second)
			}
			log := fmt.Sprintf("[IGIN-%s] %v |%s %3d %s| %13v | %15s |%s %-7s %s %#v",
				param.GinMode,
				param.TimeStamp.Format("2006/01/02-15:04:05"),
				statusColor, param.StatusCode, resetColor,
				param.Latency,
				param.ClientIP,
				methodColor, param.Method, resetColor,
				param.Path,
			)
			log += "\n"
			if param.GinMode != gin.ReleaseMode {
				log += fmt.Sprintf("Header: %v", param.Request.Header)
				log += "\n"
				log += fmt.Sprintf("BodyRaw: %v", param.Body)
				log += "\n"
			}
			log += param.ErrorMessage
			return log
		},
	}
)

// Next instances a Logger middleware that will write the logs to gin.DefaultWriter.
// By default, gin.DefaultWriter = os.Stdout.
func Next() gin.HandlerFunc {
	return NextWithConfig(defaultConfig)
}

// NextWithConfig instance a Logger middleware with config.
func NextWithConfig(config LoggerConfig) gin.HandlerFunc {
	if config.Skipper == nil {
		config.Skipper = defaultConfig.Skipper
	}
	if config.SkipBody == nil {
		config.SkipBody = defaultConfig.SkipBody
	}
	if config.Formatter == nil {
		config.Formatter = defaultConfig.Formatter
	}
	if config.Output == nil {
		config.Output = defaultConfig.Output
	}
	isTerm := true
	if w, ok := config.Output.(*os.File); !ok || os.Getenv("TERM") == "dumb" ||
		(!isatty.IsTerminal(w.Fd()) && !isatty.IsCygwinTerminal(w.Fd())) {
		isTerm = false
	}
	var skip map[string]struct{}
	if length := len(config.SkipPaths); length > 0 {
		skip = make(map[string]struct{}, length)
		for _, path := range config.SkipPaths {
			skip[path] = struct{}{}
		}
	}
	return func(c *gin.Context) {
		if config.Skipper(c) {
			config.Skipper = defaultConfig.Skipper
		}
		// Start timer
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		var body []byte
		if config.SkipBody(c) {
			body, _ = c.GetRawData()
			// 将原body塞回去
			c.Request.Body = io.NopCloser(bytes.NewBuffer(body))
		}
		// Process request
		c.Next()
		// Log only when path is not being skipped
		if _, ok := skip[path]; !ok {
			if raw != "" {
				path = path + "?" + raw
			}
			param := LogFormatterParams{
				Request:      c.Request,
				GinMode:      gin.Mode(),
				isTerm:       isTerm,
				Keys:         c.Keys,
				TimeStamp:    time.Now(),
				ClientIP:     c.ClientIP(),
				Method:       c.Request.Method,
				Path:         path,
				StatusCode:   c.Writer.Status(),
				ErrorMessage: c.Errors.ByType(gin.ErrorTypePrivate).String(),
				BodySize:     c.Writer.Size(),
				Body:         string(body),
			}
			param.Latency = param.TimeStamp.Sub(start)
			fmt.Fprint(config.Output, config.Formatter(param))
		}
	}
}

// LogFormatter gives the signature of the formatter function passed to LoggerWithFormatter
type LogFormatter func(params LogFormatterParams) string

// LogFormatterParams is the structure any formatter will be handed when time to log comes
type LogFormatterParams struct {
	Request *http.Request
	//GinMode gin Mode gin.Mode()
	GinMode string
	// TimeStamp shows the time after the server returns a response.
	TimeStamp time.Time
	// StatusCode is HTTP response code.
	StatusCode int
	// Latency is how much time the server cost to process a certain request.
	Latency time.Duration
	// ClientIP equals Context's ClientIP method.
	ClientIP string
	// Method is the HTTP method given to the request.
	Method string
	// Path is a path the client requests.
	Path string
	// ErrorMessage is set if error has occurred in processing the request.
	ErrorMessage string
	// isTerm shows whether gin's output descriptor refers to a terminal.
	isTerm bool
	// BodySize is the size of the Response Body
	BodySize int
	// Body is the size of the Request Body
	Body string
	// Keys are the keys set on the request's context.
	Keys map[string]any
}

// StatusCodeColor is the ANSI color for appropriately logging http status code to a terminal.
func (p *LogFormatterParams) StatusCodeColor() string {
	code := p.StatusCode
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}

// MethodColor is the ANSI color for appropriately logging http method to a terminal.
func (p *LogFormatterParams) MethodColor() string {
	switch p.Method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

// ResetColor resets all escape attributes.
func (p *LogFormatterParams) ResetColor() string {
	return reset
}

// IsOutputColor indicates whether can colors be outputted to the log.
func (p *LogFormatterParams) IsOutputColor() bool {
	return p.isTerm
}
