// Package session provides HTTP session management for znet web applications.
// It supports configurable session storage backends and automatic session handling.
package session

import (
	"time"

	"github.com/sohaha/zlsgo/zdi"
	"github.com/sohaha/zlsgo/znet"
	"github.com/sohaha/zlsgo/zstring"
	"github.com/sohaha/zlsgo/zutil"
)

type (
	// Config holds the configuration for session management.
	// It allows customization of cookie settings and session behavior.
	Config struct {
		CookieName string
		ExpiresAt  time.Duration
		AutoRenew  bool
	}
)

// Default creates a new session handler with the default memory store.
// It accepts optional configuration functions to customize the session behavior.
// The returned handler can be used as middleware in znet applications.
func Default(opt ...func(*Config)) znet.Handler {
	return New(NewMemoryStore(), opt...)
}

// New creates a new session handler with the specified store implementation.
// It allows customizing session behavior through configuration options.
// The handler manages session lifecycle including creation, retrieval, and renewal.
func New(stores Store, opt ...func(*Config)) znet.Handler {
	conf := zutil.Optional(Config{
		CookieName: "session_id",
		ExpiresAt:  30 * time.Minute,
	}, opt...)

	return func(c *znet.Context) error {
		id := c.GetCookie(conf.CookieName)
		if id == "" {
			id = zstring.UUID()
		}

		s, err := stores.Get(id)
		if err != nil {
			expiresAt := time.Now().Add(conf.ExpiresAt)
			s, err = stores.New(id, expiresAt)
			if err != nil {
				return err
			}
			c.SetCookie(conf.CookieName, id, int(conf.ExpiresAt.Seconds()))
		}

		_ = c.Injector().Map(s)

		_ = c.Next()

		if conf.AutoRenew && time.Until(s.ExpiresAt()) < conf.ExpiresAt/2 {
			stores.Renew(id, time.Now().Add(conf.ExpiresAt))
		}

		return nil
	}
}

// Get retrieves the current session from the context.
// It returns the session if it exists, or an error if no session is found.
// This function is typically used within request handlers to access session data.
func Get(c *znet.Context) (s Session, err error) {
	err = c.Injector().(zdi.Invoker).Resolve(&s)
	return
}
