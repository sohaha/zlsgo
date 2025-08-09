package session_test

import (
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet/session"
	"github.com/sohaha/zlsgo/ztype"
)

func TestMemorySession(t *testing.T) {
	tt := zlsgo.NewTest(t)
	store := session.NewMemoryStore()

	tt.Run("New and Get", func(tt *zlsgo.TestUtil) {
		sessionID := "test-session"
		expiresAt := time.Now().Add(1 * time.Hour)

		s, err := store.New(sessionID, expiresAt)
		tt.NoError(err)
		tt.Equal(sessionID, s.ID())
		tt.Equal(expiresAt.Unix(), s.ExpiresAt().Unix())

		s2, err := store.Get(sessionID)
		tt.NoError(err)
		tt.Equal(sessionID, s2.ID())
	})

	tt.Run("Set and Get", func(tt *zlsgo.TestUtil) {
		sessionID := "test-session-2"
		s, _ := store.New(sessionID, time.Now().Add(time.Hour))

		key := "test-key"
		value := "test-value"
		s.Set(key, value)

		got := s.Get(key)
		tt.Equal(value, got.String())

		nonexistent := s.Get("nonexistent")
		tt.EqualFalse(nonexistent.Exists())
	})

	tt.Run("Delete", func(tt *zlsgo.TestUtil) {
		sessionID := "test-session-3"
		s, _ := store.New(sessionID, time.Now().Add(time.Hour))

		key := "to-delete"
		s.Set(key, "value")

		tt.EqualTrue(s.Get(key).Exists())
		err := s.Delete(key)
		tt.NoError(err, true)

		tt.EqualFalse(s.Get(key).Exists())
	})

	tt.Run("Destroy", func(tt *zlsgo.TestUtil) {
		sessionID := "test-session-4"
		s, _ := store.New(sessionID, time.Now().Add(time.Hour))

		s.Set("key1", "value1")
		s.Set("key2", "value2")

		tt.EqualTrue(s.Get("key1").Exists())
		tt.EqualTrue(s.Get("key2").Exists())

		err := s.Destroy()
		tt.NoError(err, true)

		tt.EqualFalse(s.Get("key1").Exists())
		tt.EqualFalse(s.Get("key2").Exists())
	})
}

func TestMemoryStore_Persistence_Sharded(t *testing.T) {
	tt := zlsgo.NewTest(t)
	dir := t.TempDir()

	store := session.NewMemoryStore(
		func(o *session.MemoryStoreOptions) {
			o.Dir = dir
			o.IntervalSec = 1
			o.Shards = 4
			o.FilenamePrefix = "sessions"
		},
	)

	aliveIDs := []string{"a1", "a2", "a3", "a4", "a5"}
	for _, id := range aliveIDs {
		s, err := store.New(id, time.Now().Add(30*time.Minute))
		tt.NoError(err, true)
		s.Set("k", id)
	}
	expiredIDs := []string{"e1", "e2"}
	for _, id := range expiredIDs {
		_, err := store.New(id, time.Now().Add(-1*time.Hour))
		tt.NoError(err, true)
	}

	time.Sleep(1500 * time.Millisecond)

	for i := 0; i < 4; i++ {
		fi, err := os.Stat(fmt.Sprintf("%s/sessions-%d.json", dir, i))
		tt.NoError(err, true)
		tt.EqualTrue(fi.Size() > 0)
	}

	store2 := session.NewMemoryStore(
		func(o *session.MemoryStoreOptions) {
			o.Dir = dir
			o.IntervalSec = 10
			o.Shards = 4
			o.FilenamePrefix = "sessions"
		},
	)

	for _, id := range aliveIDs {
		s2, err := store2.Get(id)
		tt.NoError(err, true)
		tt.Equal(id, s2.Get("k").String())
	}

	for _, id := range expiredIDs {
		_, err := store2.Get(id)
		tt.Equal(errors.New("session not found"), err)
	}
}

func TestMemoryStore(t *testing.T) {
	tt := zlsgo.NewTest(t)
	store := session.NewMemoryStore()

	tt.Run("Get non-existent session", func(tt *zlsgo.TestUtil) {
		_, err := store.Get("nonexistent")
		tt.Equal(errors.New("session not found"), err, true)
	})

	tt.Run("Session expiration", func(tt *zlsgo.TestUtil) {
		sessionID := "expiring-session"
		s, _ := store.New(sessionID, time.Now().Add(-time.Hour))
		s.Set("key", "value")

		_, err := store.Get(sessionID)
		tt.Equal(errors.New("session expired"), err)
	})

	tt.Run("Renew session", func(tt *zlsgo.TestUtil) {
		sessionID := "renewable-session"
		newExpiry := time.Now().Add(2 * time.Hour)

		_, _ = store.New(sessionID, time.Now().Add(-time.Hour))

		err := store.Renew(sessionID, newExpiry)
		tt.NoError(err, true)

		s2, err := store.Get(sessionID)
		tt.NoError(err, true)
		tt.Equal(newExpiry.Unix(), s2.ExpiresAt().Unix())
	})

	tt.Run("Renew non-existent session", func(tt *zlsgo.TestUtil) {
		err := store.Renew("nonexistent", time.Now().Add(time.Hour))
		tt.Equal(errors.New("session not found"), err)
	})

	tt.Run("Collect expired sessions", func(tt *zlsgo.TestUtil) {
		store.New("expired-1", time.Now().Add(-time.Hour))
		store.New("expired-2", time.Now().Add(-30*time.Minute))
		store.New("active-1", time.Now().Add(time.Hour))
		err := store.Collect()
		tt.NoError(err, true)

		_, err = store.Get("expired-1")
		tt.Equal(errors.New("session not found"), err)
		_, err = store.Get("expired-2")
		tt.Equal(errors.New("session not found"), err)
		_, err = store.Get("active-1")
		tt.NoError(err, true)
	})

	tt.Run("Concurrent access", func(tt *zlsgo.TestUtil) {
		sessionID := "concurrent-session"
		store.New(sessionID, time.Now().Add(time.Hour))

		var wg sync.WaitGroup
		numRoutines := 10
		wg.Add(numRoutines)

		for i := 0; i < numRoutines; i++ {
			go func(i int) {
				defer wg.Done()
				key := ztype.ToString(i)
				value := "value-" + key

				s, err := store.Get(sessionID)
				tt.NoError(err, true)

				s.Set(key, value)
				tt.Equal(value, s.Get(key).String())

				s.Delete(key)
				tt.EqualFalse(s.Get(key).Exists())
			}(i)
		}

		wg.Wait()

		s, err := store.Get(sessionID)
		tt.NoError(err, true)
		tt.Equal(sessionID, s.ID())
	})
}

func TestMemoryStore_Delete(t *testing.T) {
	tt := zlsgo.NewTest(t)
	store := session.NewMemoryStore()

	sessionID := "session-to-delete"
	s, _ := store.New(sessionID, time.Now().Add(time.Hour))
	s.Set("key", "value")

	err := store.Delete(sessionID)
	tt.NoError(err, true)

	_, err = store.Get(sessionID)
	tt.Equal(errors.New("session not found"), err)

	err = store.Delete("nonexistent")
	tt.NoError(err, true)
}

func TestMemoryStore_Save(t *testing.T) {
	tt := zlsgo.NewTest(t)
	store := session.NewMemoryStore()

	sessionID := "session-to-save"
	s, _ := store.New(sessionID, time.Now().Add(time.Hour))

	err := store.Save(s)
	tt.NoError(err, true)

	s2, err := store.Get(sessionID)
	tt.NoError(err, true)
	tt.Equal(sessionID, s2.ID())
}

func TestMemoryStore_Persistence(t *testing.T) {
	tt := zlsgo.NewTest(t)
	dir := t.TempDir()

	store := session.NewMemoryStore(
		func(o *session.MemoryStoreOptions) {
			o.Dir = dir
			o.IntervalSec = 1
			o.Filename = "sessions.json"
		},
	)

	activeID := "persist-active"
	s, err := store.New(activeID, time.Now().Add(1*time.Hour))
	tt.NoError(err, true)
	s.Set("k", "v")

	expiredID := "persist-expired"
	_, err = store.New(expiredID, time.Now().Add(-1*time.Hour))
	tt.NoError(err, true)

	time.Sleep(1500 * time.Millisecond)
	fi, err := os.Stat(dir + "/sessions.json")
	tt.NoError(err, true)
	tt.EqualTrue(fi.Size() > 0)

	store2 := session.NewMemoryStore(
		func(o *session.MemoryStoreOptions) {
			o.Dir = dir
			o.IntervalSec = 10
			o.Filename = "sessions.json"
		},
	)

	s2, err := store2.Get(activeID)
	tt.NoError(err, true)
	tt.Equal("v", s2.Get("k").String())

	_, err = store2.Get(expiredID)
	tt.Equal(errors.New("session not found"), err)
}
