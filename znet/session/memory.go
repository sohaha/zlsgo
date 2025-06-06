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
// It provides thread-safe access to session data with read-write mutex protection.
type Memory struct {
	id        string
	data      ztype.Map
	expiresAt time.Time
	mu        sync.RWMutex
}

var _ Session = (*Memory)(nil)

func (s *Memory) ID() string {
	return s.id
}

// Get retrieves a value from the session by key.
func (s *Memory) Get(key string) ztype.Type {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data.Get(key)
}

// Set stores a value in the session with the specified key.
func (s *Memory) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Delete removes a value from the session by key.
func (s *Memory) Delete(key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.data, key)

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

// Destroy removes all data from the session and resets its state..
// This method is thread-safe.
func (s *Memory) Destroy() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data = make(map[string]interface{})
	return nil
}

// MemoryStore implements the Store interface using an in-memory map.
// It provides a simple, non-persistent session storage solution
// suitable for development, testing, or single-instance applications.
type MemoryStore struct {
	sessions map[string]*Memory
	mu       sync.RWMutex
}

var _ Store = (*MemoryStore)(nil)

// NewMemoryStore creates and initializes a new in-memory session store.
func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		sessions: make(map[string]*Memory),
	}
}

// Get retrieves a session by its ID from the memory store.
func (store *MemoryStore) Get(sessionID string) (Session, error) {
	store.mu.RLock()
	defer store.mu.RUnlock()

	session, exists := store.sessions[sessionID]
	if !exists {
		return nil, errors.New("session not found")
	}

	if !session.expiresAt.IsZero() && time.Now().After(session.expiresAt) {
		go session.Delete(sessionID)
		return nil, errors.New("session expired")
	}

	return session, nil
}

// New creates a new session with the specified ID and expiration time.
func (store *MemoryStore) New(sessionID string, expiresAt time.Time) (Session, error) {
	session := &Memory{
		id:        sessionID,
		data:      make(map[string]interface{}),
		expiresAt: expiresAt,
	}

	store.mu.Lock()
	store.sessions[sessionID] = session
	store.mu.Unlock()

	return session, nil
}

// Save persists the session to the memory store.
func (store *MemoryStore) Save(session Session) error {
	return nil
}

// Delete removes a session from the memory store by its ID.
func (store *MemoryStore) Delete(sessionID string) error {
	store.mu.Lock()
	defer store.mu.Unlock()
	delete(store.sessions, sessionID)
	return nil
}

// Collect removes all expired sessions from the memory store.
func (store *MemoryStore) Collect() error {
	store.mu.Lock()
	defer store.mu.Unlock()

	now := ztime.Time()
	for id, session := range store.sessions {
		if !session.expiresAt.IsZero() && now.After(session.expiresAt) {
			delete(store.sessions, id)
		}
	}

	return nil
}

// Renew extends the expiration time of an existing session.
func (store *MemoryStore) Renew(sessionID string, expiresAt time.Time) error {
	store.mu.Lock()
	defer store.mu.Unlock()

	if session, exists := store.sessions[sessionID]; exists {
		session.expiresAt = expiresAt
		return nil
	}

	return errors.New("session not found")
}
