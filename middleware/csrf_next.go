package middleware

import (
	"crypto/subtle"
	"fmt"
	"github.com/pkg6/igin/xerror"
	"github.com/pkg6/igin/xstring"
	"html/template"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

type (
	// CSRFConfig defines the config for CSRF middleware.
	CSRFConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper
		// TokenLength is the length of the generated token.
		TokenLength int
		// Optional. Default value 32.
		// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:X-CSRF-Token".
		// Possible values:
		// - "header:<name>" or "header:<name>:<cut-prefix>"
		// - "query:<name>"
		// - "form:<name>"
		// Multiple sources example:
		// - "header:X-CSRF-Token,query:csrf"
		TokenLookup string
		// Context key to store generated CSRF token into context.
		// Optional. Default value "csrf".
		ContextKey string
		// Name of the CSRF cookie. This cookie will store CSRF token.
		// Optional. Default value "csrf".
		CookieName string
		// Domain of the CSRF cookie.
		// Optional. Default value none.
		CookieDomain string
		// Path of the CSRF cookie.
		// Optional. Default value none.
		CookiePath string
		// Max age (in seconds) of the CSRF cookie.
		// Optional. Default value 86400 (24hr).
		CookieMaxAge int
		// Indicates if CSRF cookie is secure.
		// Optional. Default value false.
		CookieSecure bool
		// Indicates if CSRF cookie is HTTP only.
		// Optional. Default value false.
		CookieHTTPOnly bool
		// Indicates SameSite mode of the CSRF cookie.
		// Optional. Default value SameSiteDefaultMode.
		CookieSameSite http.SameSite
		// ErrorHandler defines a function which is executed for returning custom errors.
		ErrorHandler ErrorHandler
	}
)

var (
	CSRFContextKey = "csrf"
	// DefaultCSRFConfig is the default CSRF middleware config.
	defaultCSRFConfig = CSRFConfig{
		Skipper:        DefaultSkipper,
		TokenLength:    32,
		TokenLookup:    ExtractorMethodForm + ":" + igin.HeaderXCSRFToken,
		ContextKey:     CSRFContextKey,
		CookieName:     "_" + CSRFContextKey,
		CookieMaxAge:   86400,
		CookieSameSite: http.SameSiteDefaultMode,
		ErrorHandler:   DefaultErrorHandler,
	}
	ErrCSRFInvalid = xerror.NewHTTPError(http.StatusForbidden, "invalid csrf token")
)

// CSRFNext returns a Cross-Site Request Forgery (CSRF) middleware.
// See: https://en.wikipedia.org/wiki/Cross-site_request_forgery
func CSRFNext() gin.HandlerFunc {
	return CSRFNextWithConfig(defaultCSRFConfig)
}

// CSRFFormHTML html builder
func CSRFFormHTML(c *gin.Context, inputTypes ...string) template.HTML {
	inputType := "hidden"
	if len(inputTypes) > 0 {
		inputType = inputTypes[0]
	}
	return template.HTML(fmt.Sprintf("<input type=\"%s\" name=\"%s\" value=\"%s\">", inputType, igin.HeaderXCSRFToken, c.GetString(CSRFContextKey)))
}

// CSRFNextWithConfig returns a CSRF middleware with config.
// See `CSRF()`.
func CSRFNextWithConfig(config CSRFConfig) gin.HandlerFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultCSRFConfig.Skipper
	}
	if config.TokenLength == 0 {
		config.TokenLength = defaultCSRFConfig.TokenLength
	}
	if config.TokenLookup == "" {
		config.TokenLookup = defaultCSRFConfig.TokenLookup
	}
	if config.ContextKey == "" {
		config.ContextKey = defaultCSRFConfig.ContextKey
	}
	if config.CookieName == "" {
		config.CookieName = defaultCSRFConfig.CookieName
	}
	if config.CookieMaxAge == 0 {
		config.CookieMaxAge = defaultCSRFConfig.CookieMaxAge
	}
	if config.CookieSameSite == http.SameSiteNoneMode {
		config.CookieSecure = true
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultCSRFConfig.ErrorHandler
	}
	extractors, cErr := CreateExtractors(config.TokenLookup, "")
	if cErr != nil {
		panic(cErr)
	}
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		token := ""
		if k, err := c.Cookie(config.CookieName); err != nil {
			token = xstring.RandomN(config.TokenLength)
		} else {
			token = k
		}
		switch c.Request.Method {
		case http.MethodGet, http.MethodHead, http.MethodOptions, http.MethodTrace:
		default:
			// Validate token only for requests which are not defined as 'safe' by RFC7231
			var lastExtractorErr error
			var lastTokenErr error
		outer:
			for _, extractor := range extractors {
				clientTokens, err := extractor(c)
				if err != nil {
					lastExtractorErr = err
					continue
				}
				for _, clientToken := range clientTokens {
					if validateCSRFToken(token, clientToken) {
						lastTokenErr = nil
						lastExtractorErr = nil
						break outer
					}
					lastTokenErr = ErrCSRFInvalid
				}
			}
			var finalErr error
			if lastTokenErr != nil {
				finalErr = lastTokenErr
			} else if lastExtractorErr != nil {
				// ugly part to preserve backwards compatible errors. someone could rely on them
				if lastExtractorErr == ErrQueryExtractorValueMissing {
					lastExtractorErr = xerror.NewHTTPError(http.StatusBadRequest, "missing csrf token in the query string")
				} else if lastExtractorErr == ErrFormExtractorValueMissing {
					lastExtractorErr = xerror.NewHTTPError(http.StatusBadRequest, "missing csrf token in the form parameter")
				} else if lastExtractorErr == ErrHeaderExtractorValueMissing {
					lastExtractorErr = xerror.NewHTTPError(http.StatusBadRequest, "missing csrf token in request header")
				} else {
					lastExtractorErr = xerror.NewHTTPError(http.StatusBadRequest, lastExtractorErr.Error())
				}
				finalErr = lastExtractorErr
			}
			if finalErr != nil {
				config.ErrorHandler(c, finalErr)
				return
			}
		}
		// Set CSRF cookie
		cookie := new(http.Cookie)
		cookie.Name = config.CookieName
		cookie.Value = token
		if config.CookiePath != "" {
			cookie.Path = config.CookiePath
		}
		if config.CookieDomain != "" {
			cookie.Domain = config.CookieDomain
		}
		if config.CookieSameSite != http.SameSiteDefaultMode {
			cookie.SameSite = config.CookieSameSite
		}
		cookie.Expires = time.Now().Add(time.Duration(config.CookieMaxAge) * time.Second)
		cookie.Secure = config.CookieSecure
		cookie.HttpOnly = config.CookieHTTPOnly
		http.SetCookie(c.Writer, cookie)
		c.Set(config.ContextKey, token)
		c.Header(igin.HeaderVary, igin.HeaderCookie)
		c.Next()
	}

}

func validateCSRFToken(token, clientToken string) bool {
	return subtle.ConstantTimeCompare([]byte(token), []byte(clientToken)) == 1
}
