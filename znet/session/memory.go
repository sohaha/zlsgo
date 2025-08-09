// Package session provides in-memory session storage implementation.
// This file contains the memory-based session store for development and testing.
package session

import (
	"encoding/json"
	"errors"
	"fmt"
	"hash/fnv"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/ztime"
	"github.com/sohaha/zlsgo/ztype"
	"github.com/sohaha/zlsgo/zutil"
)

// Memory implements the Session interface using an in-memory map.
type Memory struct {
	expiresAt time.Time
	data      sync.Map
	id        string
	mu        sync.RWMutex
	store     *MemoryStore
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
	s.data.Range(func(key, value interface{}) bool {
		s.data.Delete(key)
		return true
	})
	if s.store != nil {
		s.store.Delete(s.id)
	}
	return nil
}

// MemoryStore implements the Store interface using an in-memory map.
// It provides a simple, non-persistent session storage solution
// suitable for development, testing, or single-instance applications.
type MemoryStore struct {
	sessionPool sync.Pool
	sessions    sync.Map
	// optional persistence
	persist         *zfile.MemoryFile
	persistPath     string
	persistInterval int64
	persistStop     chan struct{}
	// sharding support
	shards       int
	persistFiles []*zfile.MemoryFile
	persistPaths []string
}

var _ Store = (*MemoryStore)(nil)

// MemoryStoreOptions config for optional persistence.
// Dir: directory to store snapshot file. IntervalSec: auto-flush seconds.
// Filename: optional filename (default: "sessions.json").
type MemoryStoreOptions struct {
	Dir         string
	IntervalSec int64
	Filename    string
	// Shards number of snapshot shards, default 1 (no sharding)
	Shards int
	// FilenamePrefix used when Shards > 1, default "sessions"
	FilenamePrefix string
}

// NewMemoryStore creates and initializes a new in-memory session store.
func NewMemoryStore(opt ...func(*MemoryStoreOptions)) *MemoryStore {
	store := &MemoryStore{}

	store.sessionPool.New = func() interface{} {
		return &Memory{}
	}

	cfg := zutil.Optional(MemoryStoreOptions{Filename: "sessions.json", Shards: 1, FilenamePrefix: "sessions"}, opt...)

	if cfg.Dir != "" && cfg.IntervalSec > 0 {
		dir := zfile.RealPathMkdir(cfg.Dir)
		store.persistInterval = cfg.IntervalSec
		if cfg.Shards <= 1 {
			cfg.Shards = 1
		}
		store.shards = cfg.Shards
		if store.shards == 1 {
			store.persistPath = filepath.Join(dir, cfg.Filename)
			store.persist = zfile.NewMemoryFile(store.persistPath)
		} else {
			prefix := cfg.FilenamePrefix
			if prefix == "" {
				prefix = "sessions"
			}
			store.persistFiles = make([]*zfile.MemoryFile, store.shards)
			store.persistPaths = make([]string, store.shards)
			for i := 0; i < store.shards; i++ {
				p := filepath.Join(dir, fmt.Sprintf("%s-%d.json", prefix, i))
				store.persistPaths[i] = p
				store.persistFiles[i] = zfile.NewMemoryFile(p)
			}
		}

		_ = store.loadFromDisk()

		store.persistStop = make(chan struct{})
		go store.persistLoop()
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
	session := store.sessionPool.Get().(*Memory)

	session.mu.Lock()
	session.id = sessionID
	session.expiresAt = expiresAt
	session.store = store
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

	store.sessions.Range(func(key, value interface{}) bool {
		sessionID := key.(string)
		session := value.(*Memory)

		if !session.expiresAt.IsZero() && now.After(session.expiresAt) {
			expiredSessions = append(expiredSessions, session)
			expiredIDs = append(expiredIDs, sessionID)
		}
		return true
	})

	for _, id := range expiredIDs {
		store.sessions.Delete(id)
	}

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

// persistedSession is the on-disk representation of one session.
type persistedSession struct {
	ExpiresAt time.Time              `json:"expires_at"`
	Data      map[string]interface{} `json:"data"`
}

// persistLoop periodically writes snapshot to memory file and syncs to disk.
func (store *MemoryStore) persistLoop() {
	ticker := time.NewTicker(time.Duration(store.persistInterval) * time.Second)
	for {
		select {
		case <-store.persistStop:
			ticker.Stop()
			return
		case <-ticker.C:
			_ = store.writeSnapshot()
		}
	}
}

// writeSnapshot serializes sessions and flushes to disk immediately.
func (store *MemoryStore) writeSnapshot() error {
	if store == nil || store.persist == nil {
		if store == nil || (store.shards <= 1 && store.persist == nil) {
			return nil
		}
	}
	if store.shards <= 1 {
		snapshot := make(map[string]persistedSession)
		now := time.Now()
		store.sessions.Range(func(key, value interface{}) bool {
			id, ok := key.(string)
			if !ok {
				return true
			}
			s := value.(*Memory)
			if !s.expiresAt.IsZero() && now.After(s.expiresAt) {
				return true
			}
			data := make(map[string]interface{})
			s.data.Range(func(k, v interface{}) bool {
				ks, ok := k.(string)
				if !ok {
					ks = ztype.ToString(k)
				}
				data[ks] = v
				return true
			})
			snapshot[id] = persistedSession{ExpiresAt: s.expiresAt, Data: data}
			return true
		})
		b, err := json.Marshal(snapshot)
		if err != nil {
			return err
		}
		if _, err = store.persist.Write(b); err != nil {
			return err
		}
		return store.persist.Sync()
	}

	if store.shards < 1 {
		store.shards = 1
	}
	parts := make([]map[string]persistedSession, store.shards)
	for i := 0; i < store.shards; i++ {
		parts[i] = make(map[string]persistedSession)
	}
	now := time.Now()
	store.sessions.Range(func(key, value interface{}) bool {
		id, ok := key.(string)
		if !ok {
			return true
		}
		s := value.(*Memory)
		if !s.expiresAt.IsZero() && now.After(s.expiresAt) {
			return true
		}
		data := make(map[string]interface{})
		s.data.Range(func(k, v interface{}) bool {
			ks, ok := k.(string)
			if !ok {
				ks = ztype.ToString(k)
			}
			data[ks] = v
			return true
		})
		idx := store.shardIndex(id)
		parts[idx][id] = persistedSession{ExpiresAt: s.expiresAt, Data: data}
		return true
	})

	for i := 0; i < store.shards; i++ {
		b, err := json.Marshal(parts[i])
		if err != nil {
			return err
		}
		if _, err = store.persistFiles[i].Write(b); err != nil {
			return err
		}
		if err = store.persistFiles[i].Sync(); err != nil {
			return err
		}
	}
	return nil
}

// loadFromDisk loads snapshot from disk on startup.
func (store *MemoryStore) loadFromDisk() error {
	now := time.Now()
	if store.shards <= 1 {
		if store.persistPath == "" {
			return nil
		}
		b, err := os.ReadFile(store.persistPath)
		if err != nil || len(b) == 0 {
			return nil
		}
		var snapshot map[string]persistedSession
		if err := json.Unmarshal(b, &snapshot); err != nil {
			return err
		}
		for id, ps := range snapshot {
			if !ps.ExpiresAt.IsZero() && now.After(ps.ExpiresAt) {
				continue
			}
			sess := store.sessionPool.Get().(*Memory)
			sess.mu.Lock()
			sess.id = id
			sess.expiresAt = ps.ExpiresAt
			sess.mu.Unlock()
			for k, v := range ps.Data {
				sess.data.Store(k, v)
			}
			store.sessions.Store(id, sess)
		}
		return nil
	}

	for i := 0; i < store.shards; i++ {
		path := store.persistPaths[i]
		if path == "" {
			continue
		}
		b, err := os.ReadFile(path)
		if err != nil || len(b) == 0 {
			continue
		}
		var snapshot map[string]persistedSession
		if err := json.Unmarshal(b, &snapshot); err != nil {
			return err
		}
		for id, ps := range snapshot {
			if !ps.ExpiresAt.IsZero() && now.After(ps.ExpiresAt) {
				continue
			}
			sess := store.sessionPool.Get().(*Memory)
			sess.mu.Lock()
			sess.id = id
			sess.expiresAt = ps.ExpiresAt
			sess.mu.Unlock()
			for k, v := range ps.Data {
				sess.data.Store(k, v)
			}
			store.sessions.Store(id, sess)
		}
	}
	return nil
}

// shardIndex returns shard index for a session id.
func (store *MemoryStore) shardIndex(id string) int {
	if store.shards <= 1 {
		return 0
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(id))
	return int(h.Sum32() % uint32(store.shards))
}
