package session

import (
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/sessions"
	"github.com/pkg6/igin/middleware"
)

const (
	sessionContextKey = "_session_store"
)

type (
	// SessionConfig defines the config for Session middleware.
	SessionConfig struct {
		// Skipper defines a function to skip middleware.
		Skipper middleware.Skipper
		// Session store.
		// Required.
		Store sessions.Store
	}
	Value map[interface{}]interface{}
)

var (
	hostname, _ = os.Hostname()
	// DefaultConfig is the default Session middleware config.
	defaultConfig = SessionConfig{
		Skipper: middleware.DefaultSkipper,
		Store:   sessions.NewCookieStore([]byte("https://github.com/pkg6/igin@" + hostname)),
	}
)

// Set session set
func Set(c *gin.Context, key string, value Value, options ...*sessions.Options) error {
	session, err := Session(c, key)
	if err != nil {
		return err
	}
	if len(options) > 0 {
		session.Options = options[0]
	}
	session.Values = value
	return session.Save(c.Request, c.Writer)
}

// Delete session delete
func Delete(c *gin.Context, key string) error {
	session, err := Session(c, key)
	if err != nil {
		return err
	}
	session.Values = map[interface{}]interface{}{}
	return session.Save(c.Request, c.Writer)
}

// Get session get
func Get(c *gin.Context, key string) (Value, error) {
	session, err := Session(c, key)
	if err != nil {
		return nil, err
	}
	return session.Values, nil
}

// Session Get returns a named session.
func Session(c *gin.Context, name string) (*sessions.Session, error) {
	session, exists := c.Get(sessionContextKey)
	if !exists {
		return nil, fmt.Errorf("%q session store not found", sessionContextKey)
	}
	return session.(sessions.Store).Get(c.Request, name)
}

// Next returns a Session middleware.
func Next() gin.HandlerFunc {
	return NextWithConfig(defaultConfig)
}

// NextWithStore returns a Session middleware.
func NextWithStore(store sessions.Store) gin.HandlerFunc {
	c := defaultConfig
	c.Store = store
	return NextWithConfig(c)
}

// NextWithConfig  with config a Session middleware.
func NextWithConfig(config SessionConfig) gin.HandlerFunc {
	// Defaults
	if config.Skipper == nil {
		config.Skipper = defaultConfig.Skipper
	}
	if config.Store == nil {
		panic("IGin: session middleware requires store")
	}
	return func(c *gin.Context) {
		if config.Skipper(c) {
			c.Next()
			return
		}
		c.Set(sessionContextKey, config.Store)
		c.Next()
	}
}
