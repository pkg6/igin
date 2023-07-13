package middleware

import (
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

const (
	// ContextKeyHeaderAllow is set by Router for getting value for `Allow` header in later stages of handler call chain.
	// Allow header is mandatory for status 405 (method not found) and useful for OPTIONS method requests.
	// It is added to context only when Router does not find matching method handler for request.
	ContextKeyHeaderAllow = "gin_header_allow"
)

type CORSConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
	// AllowOrigins determines the value of the Access-Control-Allow-Origin
	// response header.  This header defines a list of origins that may access the
	// resource.  The wildcard characters '*' and '?' are supported and are
	// converted to regex fragments '.*' and '.' accordingly.
	//
	// Security: use extreme caution when handling the origin, and carefully
	// validate any logic. Remember that attackers may register hostile domain names.
	// See https://blog.portswigger.net/2016/10/exploiting-cors-misconfigurations-for.html
	//
	// Optional. Default value []string{"*"}.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Origin
	AllowOrigins []string

	// AllowOriginFunc is a custom function to validate the origin. It takes the
	// origin as an argument and returns true if allowed or false otherwise. If
	// an error is returned, it is returned by the handler. If this option is
	// set, AllowOrigins is ignored.
	//
	// Security: use extreme caution when handling the origin, and carefully
	// validate any logic. Remember that attackers may register hostile domain names.
	// See https://blog.portswigger.net/2016/10/exploiting-cors-misconfigurations-for.html
	//
	// Optional.
	AllowOriginFunc func(origin string) (bool, error)

	// AllowMethods determines the value of the Access-Control-Allow-Methods
	// response header.  This header specified the list of methods allowed when
	// accessing the resource.  This is used in response to a preflight request.
	//
	// Optional. Default value DefaultCORSConfig.AllowMethods.
	// If `allowMethods` is left empty, this middleware will fill for preflight
	// request `Access-Control-Allow-Methods` header value
	// from `Allow` header that Router set into context.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Methods
	AllowMethods []string

	// AllowHeaders determines the value of the Access-Control-Allow-Headers
	// response header.  This header is used in response to a preflight request to
	// indicate which HTTP headers can be used when making the actual request.
	//
	// Optional. Default value []string{}.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Headers
	AllowHeaders []string

	// AllowCredentials determines the value of the
	// Access-Control-Allow-Credentials response header.  This header indicates
	// whether or not the response to the request can be exposed when the
	// credentials mode (Request.credentials) is true. When used as part of a
	// response to a preflight request, this indicates whether or not the actual
	// request can be made using credentials.  See also
	// [MDN: Access-Control-Allow-Credentials].
	//
	// Optional. Default value false, in which case the header is not set.
	//
	// Security: avoid using `AllowCredentials = true` with `AllowOrigins = *`.
	// See "Exploiting CORS misconfigurations for Bitcoins and bounties",
	// https://blog.portswigger.net/2016/10/exploiting-cors-misconfigurations-for.html
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Allow-Credentials
	AllowCredentials bool

	// UnsafeWildcardOriginWithAllowCredentials UNSAFE/INSECURE: allows wildcard '*' origin to be used with AllowCredentials
	// flag. In that case we consider any origin allowed and send it back to the client with `Access-Control-Allow-Origin` header.
	//
	// This is INSECURE and potentially leads to [cross-origin](https://portswigger.net/research/exploiting-cors-misconfigurations-for-bitcoins-and-bounties)
	//
	// Optional. Default value is false.
	UnsafeWildcardOriginWithAllowCredentials bool

	// ExposeHeaders determines the value of Access-Control-Expose-Headers, which
	// defines a list of headers that clients are allowed to access.
	//
	// Optional. Default value []string{}, in which case the header is not set.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Expose-Header
	ExposeHeaders []string

	// MaxAge determines the value of the Access-Control-Max-Age response header.
	// This header indicates how long (in seconds) the results of a preflight
	// request can be cached.
	//
	// Optional. Default value 0.  The header is set only if MaxAge > 0.
	//
	// See also: https://developer.mozilla.org/en-US/docs/Web/HTTP/Headers/Access-Control-Max-Age
	MaxAge int

	ErrorHandler ErrorHandler
}

var (
	// DefaultCORSConfig is the default CORS middleware config.
	defaultCORSConfig = CORSConfig{
		Skipper:      DefaultSkipper,
		AllowOrigins: []string{"*"},
		AllowMethods: []string{
			http.MethodGet,
			http.MethodHead,
			http.MethodPost,
			http.MethodPut,
			http.MethodPatch,
			http.MethodDelete,
			http.MethodConnect,
			http.MethodOptions,
			http.MethodTrace,
		},
		AllowHeaders: []string{
			igin.HeaderOrigin,
			igin.HeaderContentType,
			igin.HeaderAccept,
			igin.HeaderAuthorization,
			igin.HeaderCookie,
			igin.HeaderXCSRFToken,
		},
		ErrorHandler: DefaultErrorHandler,
	}
)

func CORSNext() gin.HandlerFunc {
	return CORSNextWithConfig(defaultCORSConfig)
}

// CORSForceNext 暴力跨域设置
func CORSForceNext(inputExposeHeaders ...string) gin.HandlerFunc {
	exposeHeaders := []string{
		igin.HeaderAccessControlAllowHeaders,
		igin.HeaderXCSRFToken,
		igin.HeaderAuthorization,
		igin.HeaderXRequestID,
		igin.HeaderUserAgent,
		"Keep-Alive",
		igin.HeaderContentType,
	}
	for _, header := range inputExposeHeaders {
		exposeHeaders = append(exposeHeaders, header)
	}
	return func(context *gin.Context) {
		origin := context.Request.Header.Get(igin.HeaderOrigin)
		context.Header(igin.HeaderAccessControlAllowOrigin, origin)
		context.Header(igin.HeaderAccessControlAllowHeaders, strings.Join(exposeHeaders, ","))
		context.Header(igin.HeaderAccessControlAllowMethods, strings.Join(defaultCORSConfig.AllowMethods, ","))
		context.Header(igin.HeaderAccessControlExposeHeaders, strings.Join(defaultCORSConfig.AllowHeaders, ","))
		context.Header(igin.HeaderAccessControlAllowCredentials, "true")
		if context.Request.Method == http.MethodOptions {
			context.AbortWithStatus(http.StatusAccepted)
		}
		context.Next()
	}
}

func CORSNextWithConfig(config CORSConfig) gin.HandlerFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultCORSConfig.Skipper
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultCORSConfig.ErrorHandler
	}
	if len(config.AllowOrigins) == 0 {
		config.AllowOrigins = defaultCORSConfig.AllowOrigins
	}
	hasCustomAllowMethods := true
	if len(config.AllowMethods) == 0 {
		hasCustomAllowMethods = false
		config.AllowMethods = defaultCORSConfig.AllowMethods
	}
	var allowOriginPatterns []string
	for _, origin := range config.AllowOrigins {
		pattern := regexp.QuoteMeta(origin)
		pattern = strings.ReplaceAll(pattern, "\\*", ".*")
		pattern = strings.ReplaceAll(pattern, "\\?", ".")
		pattern = "^" + pattern + "$"
		allowOriginPatterns = append(allowOriginPatterns, pattern)
	}
	allowMethods := strings.Join(config.AllowMethods, ",")
	allowHeaders := strings.Join(config.AllowHeaders, ",")
	exposeHeaders := strings.Join(config.ExposeHeaders, ",")
	maxAge := strconv.Itoa(config.MaxAge)
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		req := c.Request
		origin := c.Request.Header.Get(igin.HeaderOrigin)
		allowOrigin := ""
		c.Header(igin.HeaderVary, igin.HeaderOrigin)
		preflight := req.Method == http.MethodOptions
		routerAllowMethods := ""
		if preflight {
			tmpAllowMethods, exists := c.Get(ContextKeyHeaderAllow)
			if exists {
				tmpAllowMethodStr, ok := tmpAllowMethods.(string)
				if ok && tmpAllowMethods != "" {
					routerAllowMethods = tmpAllowMethodStr
					c.Header(igin.HeaderAllow, routerAllowMethods)
				}
			}
		}
		if origin == "" {
			if !preflight {
				c.Next()
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		if config.AllowOriginFunc != nil {
			allowed, err := config.AllowOriginFunc(origin)
			if err != nil {
				config.ErrorHandler(c, err)
				return
			}
			if allowed {
				allowOrigin = origin
			}
		} else {
			// Check allowed origins
			for _, o := range config.AllowOrigins {
				if o == "*" && config.AllowCredentials && config.UnsafeWildcardOriginWithAllowCredentials {
					allowOrigin = origin
					break
				}
				if o == "*" || o == origin {
					allowOrigin = o
					break
				}
				if MatchSubdomain(origin, o) {
					allowOrigin = origin
					break
				}
			}
			checkPatterns := false
			if allowOrigin == "" {
				// to avoid regex cost by invalid (long) domains (253 is domain name max limit)
				if len(origin) <= (253+3+5) && strings.Contains(origin, "://") {
					checkPatterns = true
				}
			}
			if checkPatterns {
				for _, re := range allowOriginPatterns {
					if match, _ := regexp.MatchString(re, origin); match {
						allowOrigin = origin
						break
					}
				}
			}
		}
		// Origin not allowed
		if allowOrigin == "" {
			if !preflight {
				c.Next()
				return
			}
			c.AbortWithStatus(http.StatusNoContent)
			return
		}
		c.Header(igin.HeaderAccessControlAllowOrigin, allowOrigin)
		if config.AllowCredentials {
			c.Header(igin.HeaderAccessControlAllowCredentials, "true")
		}
		// Simple request
		if !preflight {
			if exposeHeaders != "" {
				c.Header(igin.HeaderAccessControlExposeHeaders, exposeHeaders)
			}
			c.Next()
			return
		}
		// Preflight
		c.Header(igin.HeaderVary, igin.HeaderAccessControlRequestMethod)
		c.Header(igin.HeaderVary, igin.HeaderAccessControlRequestHeaders)
		if !hasCustomAllowMethods && routerAllowMethods != "" {
			c.Header(igin.HeaderAccessControlAllowMethods, routerAllowMethods)
		} else {
			c.Header(igin.HeaderAccessControlAllowMethods, allowMethods)
		}
		if allowHeaders != "" {
			c.Header(igin.HeaderAccessControlAllowHeaders, allowHeaders)
		} else {
			h := req.Header.Get(igin.HeaderAccessControlRequestHeaders)
			if h != "" {
				c.Header(igin.HeaderAccessControlAllowHeaders, h)
			}
		}
		if config.MaxAge > 0 {
			c.Header(igin.HeaderAccessControlMaxAge, maxAge)
		}
		c.AbortWithStatus(http.StatusNoContent)
	}
}
