package middleware

import (
	"errors"
	"fmt"
	"net/textproto"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

const (
	// extractorLimit is arbitrary number to limit values extractor can return. this limits possible resource exhaustion
	// attack vector
	extractorLimit = 20

	ExtractorMethodQuery  = "query"
	ExtractorMethodParam  = "param"
	ExtractorMethodCookie = "cookie"
	ExtractorMethodForm   = "form"
	ExtractorMethodHeader = "header"
)

var ErrHeaderExtractorValueMissing = errors.New("missing value in request header")
var ErrHeaderExtractorValueInvalid = errors.New("invalid value in request header")
var ErrQueryExtractorValueMissing = errors.New("missing value in the query string")
var ErrParamExtractorValueMissing = errors.New("missing value in path params")
var ErrCookieExtractorValueMissing = errors.New("missing value in cookies")
var ErrFormExtractorValueMissing = errors.New("missing value in the form")

type ValuesExtractor func(c *gin.Context) ([]string, error)

func CreateExtractors(lookups string, authScheme string) ([]ValuesExtractor, error) {
	if lookups == "" {
		return nil, nil
	}
	sources := strings.Split(lookups, ",")
	var extractors = make([]ValuesExtractor, 0)
	for _, source := range sources {
		parts := strings.Split(source, ":")
		if len(parts) < 2 {
			return nil, fmt.Errorf("extractor source for lookup could not be split into needed parts: %v", source)
		}
		switch parts[0] {
		case ExtractorMethodQuery:
			extractors = append(extractors, valuesFromQuery(parts[1]))
		case ExtractorMethodParam:
			extractors = append(extractors, valuesFromParam(parts[1]))
		case ExtractorMethodCookie:
			extractors = append(extractors, valuesFromCookie(parts[1]))
		case ExtractorMethodForm:
			extractors = append(extractors, valuesFromForm(parts[1]))
		case ExtractorMethodHeader:
			prefix := ""
			if len(parts) > 2 {
				prefix = parts[2]
			} else if authScheme != "" && parts[1] == igin.HeaderAuthorization {
				// backwards compatibility for JWT and KeyAuth:
				// * we only apply this fix to Authorization as header we use and uses prefixes like "Bearer <token-value>" etc
				// * previously header extractor assumed that auth-scheme/prefix had a space as suffix we need to retain that
				//   behaviour for default values and Authorization header.
				prefix = authScheme
				if !strings.HasSuffix(prefix, " ") {
					prefix += " "
				}
			}
			extractors = append(extractors, valuesFromHeader(parts[1], prefix))
		}
	}
	return extractors, nil
}

// valuesFromHeader returns a functions that extracts values from the request header.
// valuePrefix is parameter to remove first part (prefix) of the extracted value. This is useful if header value has static
// prefix like `Authorization: <auth-scheme> <authorisation-parameters>` where part that we want to remove is `<auth-scheme> `
// note the space at the end. In case of basic authentication `Authorization: Basic <credentials>` prefix we want to remove
// is `Basic `. In case of JWT tokens `Authorization: Bearer <token>` prefix is `Bearer `.
// If prefix is left empty the whole value is returned.
func valuesFromHeader(header string, valuePrefix string) ValuesExtractor {
	prefixLen := len(valuePrefix)
	// standard library parses http.Request header keys in canonical form but we may provide something else so fix this
	header = textproto.CanonicalMIMEHeaderKey(header)
	return func(c *gin.Context) ([]string, error) {
		values := c.Request.Header.Values(header)
		if len(values) == 0 {
			return nil, ErrHeaderExtractorValueMissing
		}
		result := make([]string, 0)
		for i, value := range values {
			if prefixLen == 0 {
				result = append(result, value)
				if i >= extractorLimit-1 {
					break
				}
				continue
			}
			if len(value) > prefixLen && strings.EqualFold(value[:prefixLen], valuePrefix) {
				result = append(result, value[prefixLen:])
				if i >= extractorLimit-1 {
					break
				}
			}
		}
		if len(result) == 0 {
			if prefixLen > 0 {
				return nil, ErrHeaderExtractorValueInvalid
			}
			return nil, ErrHeaderExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromQuery returns a function that extracts values from the query string.
func valuesFromQuery(param string) ValuesExtractor {
	return func(c *gin.Context) ([]string, error) {
		result := c.QueryArray(param)
		if len(result) == 0 {
			return nil, ErrQueryExtractorValueMissing
		} else if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}
		return result, nil
	}
}

// valuesFromParam returns a function that extracts values from the url param string.
func valuesFromParam(param string) ValuesExtractor {
	return func(c *gin.Context) ([]string, error) {
		result := c.PostFormArray(param)
		if len(result) == 0 {
			return nil, ErrParamExtractorValueMissing
		} else if len(result) > extractorLimit-1 {
			result = result[:extractorLimit]
		}
		return result, nil
	}
}

// valuesFromCookie returns a function that extracts values from the named cookie.
func valuesFromCookie(name string) ValuesExtractor {
	return func(c *gin.Context) ([]string, error) {
		cookies := c.Request.Cookies()
		if len(cookies) == 0 {
			return nil, ErrCookieExtractorValueMissing
		}
		result := make([]string, 0)
		for i, cookie := range cookies {
			if name == cookie.Name {
				result = append(result, cookie.Value)
				if i >= extractorLimit-1 {
					break
				}
			}
		}
		if len(result) == 0 {
			return nil, ErrCookieExtractorValueMissing
		}
		return result, nil
	}
}

// valuesFromForm returns a function that extracts values from the form field.
func valuesFromForm(name string) ValuesExtractor {
	return func(c *gin.Context) ([]string, error) {
		if c.Request.Form == nil {
			_ = c.Request.ParseMultipartForm(32 << 20) // same what `c.Request().FormValue(name)` does
		}
		values := c.Request.Form[name]
		if len(values) == 0 {
			return nil, ErrFormExtractorValueMissing
		}
		if len(values) > extractorLimit-1 {
			values = values[:extractorLimit]
		}
		result := append([]string{}, values...)
		return result, nil
	}
}
