// Package session provides in-memory session storage implementation.
// This file contains the memory-based session store for development and testing.
package session

import (
	"errors"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztype"
)

// Memory implements the Session interface using an in-memory map.
type Memory struct {
	id        string
	data      sync.Map
	expiresAt time.Time
	mu        sync.RWMutex
}

// reset clears the session data and prepares it for reuse from the pool.
func (s *Memory) reset() {
	s.mu.Lock()
	s.id = ""
	s.expiresAt = time.Time{}
	s.mu.Unlock()

	s.data.Range(func(key, value interface{}) bool {
		s.data.Delete(key)
		return true
	})
}

var _ Session = (*Memory)(nil)

func (s *Memory) ID() string {
	return s.id
}

// Get retrieves a value from the session by key.
func (s *Memory) Get(key string) ztype.Type {
	if value, ok := s.data.Load(key); ok {
		return ztype.New(value)
	}
	return ztype.New(nil)
}

// Set stores a value in the session with the specified key.
func (s *Memory) Set(key string, value interface{}) {
	s.data.Store(key, value)
}

// Delete removes a value from the session by key.
func (s *Memory) Delete(key string) error {
	s.data.Delete(key)
	return nil
}

// Save persists the session data.
func (s *Memory) Save() error {
	return nil
}

// ExpiresAt returns the time when the session will expire.
func (s *Memory) ExpiresAt() time.Time {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.expiresAt
}

// Destroy removes all data from the session and resets its state.
// This method is thread-safe.
func (s *Memory) Destroy() error {
	// Clear all data from sync.Map
	s.data.Range(func(key, value interface{}) bool {
		s.data.Delete(key)
		return true
	})
	return nil
}

// MemoryStore implements the Store interface using an in-memory map.
// It provides a simple, non-persistent session storage solution
// suitable for development, testing, or single-instance applications.
type MemoryStore struct {
	sessions    sync.Map  // Lock-free concurrent map for session storage
	sessionPool sync.Pool // Pool for session object reuse
}

var _ Store = (*MemoryStore)(nil)

// NewMemoryStore creates and initializes a new in-memory session store.
func NewMemoryStore() *MemoryStore {
	store := &MemoryStore{}

	// Initialize session pool with factory function
	store.sessionPool.New = func() interface{} {
		return &Memory{}
	}

	return store
}

// Get retrieves a session by its ID from the memory store.
func (store *MemoryStore) Get(sessionID string) (Session, error) {
	value, exists := store.sessions.Load(sessionID)
	if !exists {
		return nil, errors.New("session not found")
	}

	session := value.(*Memory)
	if !session.expiresAt.IsZero() && time.Now().After(session.expiresAt) {
		go store.Delete(sessionID)
		return nil, errors.New("session expired")
	}

	return session, nil
}

// New creates a new session with the specified ID and expiration time.
func (store *MemoryStore) New(sessionID string, expiresAt time.Time) (Session, error) {
	// Get session from pool to reduce allocations
	session := store.sessionPool.Get().(*Memory)

	// Initialize session with provided data
	session.mu.Lock()
	session.id = sessionID
	session.expiresAt = expiresAt
	session.mu.Unlock()

	store.sessions.Store(sessionID, session)

	return session, nil
}

// Save persists the session to the memory store.
func (store *MemoryStore) Save(session Session) error {
	return nil
}

// Delete removes a session from the memory store by its ID.
func (store *MemoryStore) Delete(sessionID string) error {
	value, exists := store.sessions.LoadAndDelete(sessionID)

	// Return session to pool for reuse if it existed
	if exists {
		session := value.(*Memory)
		session.reset()
		store.sessionPool.Put(session)
	}

	return nil
}

// Collect removes all expired sessions from the memory store.
func (store *MemoryStore) Collect() error {
	var expiredSessions []*Memory
	var expiredIDs []string

	now := ztime.Time()

	// Find expired sessions
	store.sessions.Range(func(key, value interface{}) bool {
		sessionID := key.(string)
		session := value.(*Memory)

		if !session.expiresAt.IsZero() && now.After(session.expiresAt) {
			expiredSessions = append(expiredSessions, session)
			expiredIDs = append(expiredIDs, sessionID)
		}
		return true
	})

	// Remove expired sessions from store
	for _, id := range expiredIDs {
		store.sessions.Delete(id)
	}

	// Return expired sessions to pool for reuse
	for _, session := range expiredSessions {
		session.reset()
		store.sessionPool.Put(session)
	}

	return nil
}

// Renew extends the expiration time of an existing session.
func (store *MemoryStore) Renew(sessionID string, expiresAt time.Time) error {
	value, exists := store.sessions.Load(sessionID)
	if !exists {
		return errors.New("session not found")
	}

	session := value.(*Memory)
	session.mu.Lock()
	session.expiresAt = expiresAt
	session.mu.Unlock()

	return nil
}
