package middleware

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

type (
	// BasicAuthConfig defines the config for BasicAuth middleware.
	BasicAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper
		// Validator is a function to validate BasicAuth credentials.
		// Required.
		Validator BasicAuthValidator
		// Realm is a string to define realm attribute of BasicAuth.
		// Default value "Restricted".
		Realm string
	}
	// BasicAuthValidator defines a function to validate BasicAuth credentials.
	BasicAuthValidator func(s1, s2 string, c *gin.Context) bool
)

const (
	basic        = "basic"
	defaultRealm = "Restricted"
)

var (
	// DefaultBasicAuthConfig is the default BasicAuth middleware config.
	defaultBasicAuthConfig = BasicAuthConfig{
		Skipper: DefaultSkipper,
		Realm:   defaultRealm,
	}
)

// BasicAuthNext returns an BasicAuth middleware.
//
// For valid credentials it calls the next handler.
// For missing or invalid credentials, it sends "401 - Unauthorized" iresponse.
func BasicAuthNext(fn BasicAuthValidator) gin.HandlerFunc {
	c := defaultBasicAuthConfig
	c.Validator = fn
	return BasicAuthNextWithConfig(c)
}

func BasicAuthNextWithConfig(config BasicAuthConfig) gin.HandlerFunc {
	// Defaults
	if config.Validator == nil {
		panic("GIN: basic-auth middleware requires a validator function")
	}
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultBasicAuthConfig.Skipper
	}
	if config.Realm == "" {
		config.Realm = defaultRealm
	}
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		auth := c.Request.Header.Get(igin.HeaderAuthorization)
		s1, s2, sErr := basicAuthNextSearchCredential(auth)
		if sErr != nil || !config.Validator(s1, s2, c) {
			realm := defaultRealm
			if config.Realm != defaultRealm {
				realm = strconv.Quote(config.Realm)
			}
			// Credentials doesn't match, we return 401 and abort handlers chain.
			c.Header(igin.HeaderWWWAuthenticate, basic+" realm="+realm)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
	}
}

func basicAuthNextSearchCredential(tokenStr string) (s1, s2 string, err error) {
	l := len(basic)
	if len(tokenStr) > l+1 && strings.EqualFold(tokenStr[:l], basic) {
		b, err := base64.StdEncoding.DecodeString(tokenStr[l+1:])
		if err != nil {
			return "", "", err
		}
		cred := string(b)
		for i := 0; i < len(cred); i++ {
			if cred[i] == ':' {
				// Verify credentials
				return cred[:i], cred[i+1:], nil
			}
		}
	}
	return "", "", fmt.Errorf("token error")
}
