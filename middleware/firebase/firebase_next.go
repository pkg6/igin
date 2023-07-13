package firebase

import (
	"firebase.google.com/go/auth"
	"fmt"
	"github.com/pkg6/igin/xerror"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/middleware"
)

type (
	// FirebaseConfig defines the config for BasicAuth middleware.
	FirebaseConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// middleware or handler.
		SuccessHandler middleware.SuccessHandler
		// ErrorHandler defines a function which is executed for an invalid token.
		// It may be used to define a custom JWT error.
		ErrorHandler middleware.ErrorHandler
		// Context key to store user information from the token into context.
		// Optional. Default value "user".
		ContextKey string
		// TokenLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
		// to extract token from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>" or "header:<name>:<cut-prefix>"
		// 			`<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
		//			value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
		//			want to cut is `<auth-scheme> ` note the space at the end.
		//			In case of JWT tokens `Authorization: Bearer <token>` prefix we cut is `Bearer `.
		// If prefix is left empty the whole value is returned.
		// - "query:<name>"
		// - "param:<name>"
		// - "cookie:<name>"
		// - "form:<name>"
		// Multiple sources example:
		// - "header:Authorization,cookie:myowncookie"
		TokenLookup string
		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string
		//NewAuthClient
		AuthClient *AuthClient
	}
)

var (
	ContextKey = "user"
	// DefaultJWTConfig is the default JWT auth middleware config.
	defaultConfig = FirebaseConfig{
		Skipper:        middleware.DefaultSkipper,
		SuccessHandler: nil,
		ErrorHandler:   middleware.DefaultErrorHandler,
		ContextKey:     ContextKey,
		TokenLookup:    middleware.ExtractorMethodHeader + ":" + igin.HeaderAuthorization,
		AuthScheme:     "Bearer",
	}
	ErrMissing = xerror.NewHTTPError(http.StatusBadRequest, "missing or malformed firebase")
	ErrInvalid = xerror.NewHTTPError(http.StatusUnauthorized, "invalid or expired firebase")
)

func ContextToken(c *gin.Context) (*auth.Token, error) {
	if value, exists := c.Get(ContextKey); exists {
		if token, ok := value.(*auth.Token); ok {
			return token, nil
		}
	}
	return nil, fmt.Errorf("user information does not exist")
}

// Next config
func Next(projectID, credentialsFile string) gin.HandlerFunc {
	firebase, err := NewAuthClient(projectID, credentialsFile, "")
	if err != nil {
		panic(err)
	}
	return NextWithAuthClient(firebase)
}

// NextSuccessHandler SuccessHandler config
func NextSuccessHandler(projectID, credentialsFile string, handler middleware.SuccessHandler) gin.HandlerFunc {
	firebase, err := NewAuthClient(projectID, credentialsFile, "")
	if err != nil {
		panic(err)
	}
	return NextWithAuthClientSuccessHandler(firebase, handler)
}

// NextWithAuthClient firebase AuthClient
func NextWithAuthClient(authClient *AuthClient) gin.HandlerFunc {
	c := defaultConfig
	c.AuthClient = authClient
	return NextWithConfig(c)
}

// NextWithAuthClientSuccessHandler  firebase AuthClient and middleware.SuccessHandler
func NextWithAuthClientSuccessHandler(authClient *AuthClient, handler middleware.SuccessHandler) gin.HandlerFunc {
	c := defaultConfig
	c.AuthClient = authClient
	c.SuccessHandler = handler
	return NextWithConfig(c)
}

func NextWithConfig(config FirebaseConfig) gin.HandlerFunc {
	if config.Skipper == nil {
		config.Skipper = defaultConfig.Skipper
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultConfig.ErrorHandler
	}
	if config.AuthClient == nil {
		panic("IGin: firebase middleware requires *firebase.Firebase")
	}
	if config.TokenLookup == "" {
		config.TokenLookup = defaultConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = defaultConfig.AuthScheme
	}
	extractors, cErr := middleware.CreateExtractors(config.TokenLookup, config.AuthScheme)
	if cErr != nil {
		panic(cErr)
	}
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		var lastExtractorErr error
		var lastTokenErr error
		for _, extractor := range extractors {
			auths, err := extractor(c)
			if err != nil {
				lastExtractorErr = ErrMissing
				continue
			}
			for _, auth := range auths {
				token, err := config.AuthClient.VerifyIDToken(config.AuthClient.Ctx, auth)
				if err != nil {
					lastTokenErr = err
					continue
				}
				// Store user information from token into context.
				c.Set(config.ContextKey, token)
				if config.SuccessHandler != nil {
					config.SuccessHandler(c)
				}
				c.Next()
				return
			}
		}
		if lastExtractorErr != nil {
			config.ErrorHandler(c, ErrMissing, ErrMissing.Code)
			return
		}
		if lastTokenErr != nil {
			config.ErrorHandler(c, ErrInvalid, ErrInvalid.Code)
			return
		}
	}
}
