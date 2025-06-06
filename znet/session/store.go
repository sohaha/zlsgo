// Package session provides session management for web applications.
// It includes interfaces and implementations for storing and retrieving
// session data with different storage backends.
package session

import (
	"time"

	"github.com/sohaha/zlsgo/ztype"
)

// Session represents a user session with key-value storage capabilities.
// It provides methods to manage session data and control session lifecycle.
type Session interface {
	ID() string
	Get(key string) ztype.Type
	Set(key string, value interface{})
	Delete(key string) error
	Save() error
	Destroy() error
	ExpiresAt() time.Time
}

// Store defines the interface for session storage backends.
// Implementations of Store are responsible for creating, retrieving,
// and managing session data in various storage systems.
type Store interface {
	New(sessionID string, expiresAt time.Time) (Session, error)
	Get(sessionID string) (Session, error)
	Save(session Session) error
	Delete(sessionID string) error
	Collect() error
	Renew(sessionID string, expiresAt time.Time) error
}
