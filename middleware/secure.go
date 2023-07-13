package middleware

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/pkg6/igin"
)

type SecureConfig struct {
	// Skipper defines a function to skip middleware.
	Skipper Skipper
	// XSSProtection provides protection against cross-site scripting attack (XSS)
	// by setting the `X-XSS-Protection` header.
	// Optional. Default value "1; mode=block".
	XSSProtection string

	// ContentTypeNosniff provides protection against overriding Content-Type
	// header by setting the `X-Content-Type-Options` header.
	// Optional. Default value "nosniff".
	ContentTypeNosniff string

	// XFrameOptions can be used to indicate whether or not a browser should
	// be allowed to render a page in a <frame>, <iframe> or <object> .
	// Sites can use this to avoid clickjacking attacks, by ensuring that their
	// content is not embedded into other sites.provides protection against
	// clickjacking.
	// Optional. Default value "SAMEORIGIN".
	// Possible values:
	// - "SAMEORIGIN" - The page can only be displayed in a frame on the same origin as the page itself.
	// - "DENY" - The page cannot be displayed in a frame, regardless of the site attempting to do so.
	// - "ALLOW-FROM uri" - The page can only be displayed in a frame on the specified origin.
	XFrameOptions string

	// HSTSMaxAge sets the `Strict-Transport-Security` header to indicate how
	// long (in seconds) browsers should remember that this site is only to
	// be accessed using HTTPS. This reduces your exposure to some SSL-stripping
	// man-in-the-middle (MITM) attacks.
	// Optional. Default value 0.
	HSTSMaxAge int

	// HSTSExcludeSubdomains won't include subdomains tag in the `Strict Transport Security`
	// header, excluding all subdomains from security policy. It has no effect
	// unless HSTSMaxAge is set to a non-zero value.
	// Optional. Default value false.
	HSTSExcludeSubdomains bool

	// ContentSecurityPolicy sets the `Content-Security-Policy` header providing
	// security against cross-site scripting (XSS), clickjacking and other code
	// injection attacks resulting from execution of malicious content in the
	// trusted web page context.
	// Optional. Default value "".
	ContentSecurityPolicy string

	// CSPReportOnly would use the `Content-Security-Policy-Report-Only` header instead
	// of the `Content-Security-Policy` header. This allows iterative updates of the
	// content security policy by only reporting the violations that would
	// have occurred instead of blocking the resource.
	// Optional. Default value false.
	CSPReportOnly bool

	// HSTSPreloadEnabled will add the preload tag in the `Strict Transport Security`
	// header, which enables the domain to be included in the HSTS preload list
	// maintained by Chrome (and used by Firefox and Safari): https://hstspreload.org/
	// Optional.  Default value false.
	HSTSPreloadEnabled bool

	// ReferrerPolicy sets the `Referrer-Policy` header providing security against
	// leaking potentially sensitive request paths to third parties.
	// Optional. Default value "".
	ReferrerPolicy string
}

var defaultSecureConfig = SecureConfig{
	Skipper:            DefaultSkipper,
	XSSProtection:      "1; mode=block",
	ContentTypeNosniff: "nosniff",
	XFrameOptions:      "SAMEORIGIN",
	HSTSPreloadEnabled: false,
}

func SecureNext() gin.HandlerFunc {
	return SecureNextWithConfig(defaultSecureConfig)
}

func SecureNextWithConfig(config SecureConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		if config.XSSProtection != "" {
			c.Header(igin.HeaderXXSSProtection, config.XSSProtection)
		}
		if config.ContentTypeNosniff != "" {
			c.Header(igin.HeaderXContentTypeOptions, config.ContentTypeNosniff)
		}
		if config.XFrameOptions != "" {
			c.Header(igin.HeaderXFrameOptions, config.XFrameOptions)
		}
		if (c.Request.Header.Get(igin.HeaderXForwardedProto) == "https") && config.HSTSMaxAge != 0 {
			subdomains := ""
			if !config.HSTSExcludeSubdomains {
				subdomains = "; includeSubdomains"
			}
			if config.HSTSPreloadEnabled {
				subdomains = fmt.Sprintf("%s; preload", subdomains)
			}
			c.Header(igin.HeaderStrictTransportSecurity, fmt.Sprintf("max-age=%d%s", config.HSTSMaxAge, subdomains))
		}
		if config.ContentSecurityPolicy != "" {
			if config.CSPReportOnly {
				c.Header(igin.HeaderContentSecurityPolicyReportOnly, config.ContentSecurityPolicy)
			} else {
				c.Header(igin.HeaderContentSecurityPolicy, config.ContentSecurityPolicy)
			}
		}
		if config.ReferrerPolicy != "" {
			c.Header(igin.HeaderReferrerPolicy, config.ReferrerPolicy)
		}
		c.Next()
	}
}
