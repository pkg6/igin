package jwt

import (
	"errors"
	"fmt"
	"github.com/pkg6/igin/xerror"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt"
	"github.com/pkg6/igin"
	"github.com/pkg6/igin/middleware"
)

type (
	// JWTConfig defines the config for JWT middleware.
	JWTConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// SuccessHandler defines a function which is executed for a valid token before middleware chain continues with next
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

		// Signing key to validate token.
		// This is one of the three options to provide a token validation key.
		// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
		// Required if neither user-defined KeyFunc nor SigningKeys is provided.
		SigningKey any
		// Signing method used to check the token's signing algorithm.
		// Optional. Default value HS256.
		SigningMethod string
		// Map of signing keys to validate token with kid field usage.
		// This is one of the three options to provide a token validation key.
		// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
		// Required if neither user-defined KeyFunc nor SigningKey is provided.
		SigningKeys map[string]any
		// Claims are extendable claims data defining token content. Used by default ParseTokenFunc implementation.
		// Not used if custom ParseTokenFunc is set.
		// Optional. Default value jwt.MapClaims
		Claims jwt.Claims
		// KeyFunc defines a user-defined function that supplies the public key for a token validation.
		// The function shall take care of verifying the signing algorithm and selecting the proper key.
		// A user-defined KeyFunc can be useful if tokens are issued by an external party.
		// Used by default ParseTokenFunc implementation.
		//
		// When a user-defined KeyFunc is provided, SigningKey, SigningKeys, and SigningMethod are ignored.
		// This is one of the three options to provide a token validation key.
		// The order of precedence is a user-defined KeyFunc, SigningKeys and SigningKey.
		// Required if neither SigningKeys nor SigningKey is provided.
		// Not used if custom ParseTokenFunc is set.
		// Default to an internal implementation verifying the signing algorithm and selecting the proper key.
		KeyFunc jwt.Keyfunc
		// ParseTokenFunc defines a user-defined function that parses token from given auth. Returns an error when token
		// parsing fails or parsed token is invalid.
		// Defaults to implementation using `github.com/golang-jwt/jwt` as JWT implementation library
		ParseTokenFunc func(auth string, c *gin.Context) (any, error)
	}
)

var (
	ContextKey = "user"
	// DefaultJWTConfig is the default JWT auth middleware config.
	defaultConfig = JWTConfig{
		Skipper:        middleware.DefaultSkipper,
		SuccessHandler: nil,
		SigningMethod:  "HS256",
		ErrorHandler:   middleware.DefaultErrorHandler,
		ContextKey:     ContextKey,
		TokenLookup:    middleware.ExtractorMethodHeader + ":" + igin.HeaderAuthorization,
		AuthScheme:     "Bearer",
		Claims:         jwt.MapClaims{},
	}
	ErrJWTMissing = xerror.NewHTTPError(http.StatusBadRequest, "missing or malformed jwt")
	ErrJWTInvalid = xerror.NewHTTPError(http.StatusUnauthorized, "invalid or expired jwt")
)

func ContextToken(c *gin.Context) (*jwt.Token, error) {
	if value, exists := c.Get(ContextKey); exists {
		if token, ok := value.(*jwt.Token); ok {
			return token, nil
		}
	}
	return nil, fmt.Errorf("user information does not exist")
}

func (config *JWTConfig) defaultParseToken(auth string, c *gin.Context) (any, error) {
	var token *jwt.Token
	var err error
	if _, ok := config.Claims.(jwt.MapClaims); ok {
		token, err = jwt.Parse(auth, config.KeyFunc)
	} else {
		t := reflect.ValueOf(config.Claims).Type().Elem()
		claims := reflect.New(t).Interface().(jwt.Claims)
		token, err = jwt.ParseWithClaims(auth, claims, config.KeyFunc)
	}
	if err != nil {
		return nil, err
	}
	if !token.Valid {
		return nil, errors.New("invalid token")
	}
	return token, nil
}

// defaultKeyFunc returns a signing key of the given token.
func (config *JWTConfig) defaultKeyFunc(t *jwt.Token) (any, error) {
	// Check the signing method
	if t.Method.Alg() != config.SigningMethod {
		return nil, fmt.Errorf("unexpected jwt signing method=%v", t.Header["alg"])
	}
	if len(config.SigningKeys) > 0 {
		if kid, ok := t.Header["kid"].(string); ok {
			if key, ok := config.SigningKeys[kid]; ok {
				return key, nil
			}
		}
		return nil, fmt.Errorf("unexpected jwt key id=%v", t.Header["kid"])
	}
	// key is of invalid type
	if signingKeyStr, ok := config.SigningKey.(string); ok {
		return []byte(signingKeyStr), nil
	}
	return config.SigningKey, nil
}

func Next(key any) gin.HandlerFunc {
	c := defaultConfig
	c.SigningKey = key
	return NextWithConfig(c)
}

// NextSuccessHandler SuccessHandler config
func NextSuccessHandler(key string, handler middleware.SuccessHandler) gin.HandlerFunc {
	c := defaultConfig
	c.SigningKey = key
	c.SuccessHandler = handler
	return NextWithConfig(c)
}

func NextWithConfig(config JWTConfig) gin.HandlerFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultConfig.Skipper
	}
	if config.SigningKey == nil && len(config.SigningKeys) == 0 && config.KeyFunc == nil && config.ParseTokenFunc == nil {
		panic("IGin: jwt middleware requires signing key")
	}
	if config.SigningMethod == "" {
		config.SigningMethod = defaultConfig.SigningMethod
	}
	if config.ContextKey == "" {
		config.ContextKey = defaultConfig.ContextKey
	}
	if config.Claims == nil {
		config.Claims = defaultConfig.Claims
	}
	if config.TokenLookup == "" {
		config.TokenLookup = defaultConfig.TokenLookup
	}
	if config.AuthScheme == "" {
		config.AuthScheme = defaultConfig.AuthScheme
	}
	if config.KeyFunc == nil {
		config.KeyFunc = config.defaultKeyFunc
	}
	if config.ParseTokenFunc == nil {
		config.ParseTokenFunc = config.defaultParseToken
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
				lastExtractorErr = ErrJWTMissing
				continue
			}
			for _, auth := range auths {
				token, err := config.ParseTokenFunc(auth, c)
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
			config.ErrorHandler(c, ErrJWTMissing, ErrJWTMissing.Code)
			return
		}
		if lastTokenErr != nil {
			config.ErrorHandler(c, ErrJWTInvalid, ErrJWTInvalid.Code)
			return
		}
	}
}
