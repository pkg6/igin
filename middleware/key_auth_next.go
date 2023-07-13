package middleware

import (
	"errors"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

type (
	// KeyAuthConfig defines the config for KeyAuth middleware.
	KeyAuthConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper Skipper
		// KeyLookup is a string in the form of "<source>:<name>" or "<source>:<name>,<source>:<name>" that is used
		// to extract key from the request.
		// Optional. Default value "header:Authorization".
		// Possible values:
		// - "header:<name>" or "header:<name>:<cut-prefix>"
		// 			`<cut-prefix>` is argument value to cut/trim prefix of the extracted value. This is useful if header
		//			value has static prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we
		//			want to cut is `<auth-scheme> ` note the space at the end.
		//			In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove is `Basic `.
		// - "query:<name>"
		// - "form:<name>"
		// - "cookie:<name>"
		// Multiple sources example:
		// - "header:Authorization,header:X-Api-Key"
		KeyLookup string

		// AuthScheme to be used in the Authorization header.
		// Optional. Default value "Bearer".
		AuthScheme string

		// Validator is a function to validate key.
		// Required.
		Validator KeyAuthValidator

		// ErrorHandler defines a function which is executed for an invalid key.
		// It may be used to define a custom error.
		ErrorHandler ErrorHandler

		// ContinueOnIgnoredError allows the next middleware/handler to be called when ErrorHandler decides to
		// ignore the error (by returning `nil`).
		// This is useful when parts of your site/api allow public access and some authorized routes provide extra functionality.
		// In that case you can use ErrorHandler to set a default public key auth value in the request context
		// and continue. Some logic down the remaining execution chain needs to check that (public) key auth value then.
		ContinueOnIgnoredError bool
	}

	// KeyAuthValidator defines a function to validate KeyAuth credentials.
	KeyAuthValidator func(auth string, c *gin.Context) (bool, error)
)

var defaultKeyAuthConfig = KeyAuthConfig{
	Skipper:      DefaultSkipper,
	KeyLookup:    ExtractorMethodHeader + ":" + igin.HeaderAuthorization,
	AuthScheme:   "Bearer",
	ErrorHandler: DefaultErrorHandler,
}

// KeyAuthNext returns an KeyAuth middleware.
//
// For valid key it calls the next handler.
// For invalid key, it sends "401 - Unauthorized" .
// For missing key, it sends "400 - Bad Request" .
func KeyAuthNext(fn KeyAuthValidator) gin.HandlerFunc {
	c := defaultKeyAuthConfig
	c.Validator = fn
	return KeyAuthNextWithConfig(c)
}

func KeyAuthNextWithConfig(config KeyAuthConfig) gin.HandlerFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultKeyAuthConfig.Skipper
	}
	// Defaults
	if config.AuthScheme == "" {
		config.AuthScheme = defaultKeyAuthConfig.AuthScheme
	}
	if config.KeyLookup == "" {
		config.KeyLookup = defaultKeyAuthConfig.KeyLookup
	}
	if config.ErrorHandler == nil {
		config.ErrorHandler = defaultKeyAuthConfig.ErrorHandler
	}
	if config.Validator == nil {
		panic("IGin: key-auth middleware requires a validator function")
	}
	extractors, cErr := CreateExtractors(config.KeyLookup, config.AuthScheme)
	if cErr != nil {
		panic(cErr)
	}
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		var lastExtractorErr error
		var lastValidatorErr error
		for _, extractor := range extractors {
			keys, err := extractor(c)
			if err != nil {
				lastExtractorErr = err
				continue
			}
			for _, key := range keys {
				valid, err := config.Validator(key, c)
				if err != nil {
					lastValidatorErr = err
					continue
				}
				if valid {
					c.Next()
					return
				}
				lastValidatorErr = errors.New("invalid key")
			}
		}
		// we are here only when we did not successfully extract and validate any of keys
		err := lastValidatorErr
		if err == nil { // prioritize validator errors over extracting errors
			// ugly part to preserve backwards compatible errors. someone could rely on them
			if lastExtractorErr == ErrQueryExtractorValueMissing {
				err = errors.New("missing key in the query string")
			} else if lastExtractorErr == ErrCookieExtractorValueMissing {
				err = errors.New("missing key in cookies")
			} else if lastExtractorErr == ErrFormExtractorValueMissing {
				err = errors.New("missing key in the form")
			} else if lastExtractorErr == ErrHeaderExtractorValueMissing {
				err = errors.New("missing key in request header")
			} else if lastExtractorErr == ErrHeaderExtractorValueInvalid {
				err = errors.New("invalid key in the request header")
			} else {
				err = lastExtractorErr
			}
			err = &ErrKeyAuthMissing{Err: err}
		}
		if lastValidatorErr != nil {
			// prioritize validator errors over extracting errors
			config.ErrorHandler(c, lastValidatorErr)
			return
		}
		config.ErrorHandler(c, err)
	}
}

// ErrKeyAuthMissing is error type when KeyAuth middleware is unable to extract value from lookups
type ErrKeyAuthMissing struct {
	Err error
}

// Error returns errors text
func (e *ErrKeyAuthMissing) Error() string {
	return e.Err.Error()
}

// Unwrap unwraps error
func (e *ErrKeyAuthMissing) Unwrap() error {
	return e.Err
}
