package session

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
)

func TestValidateSessionID(t *testing.T) {
	tt := zls.NewTest(t)

	validIDs := []string{
		"0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ", // 62 chars
		"secure-session-id-with-hyphens_12345",                           // 38 chars
		"Xm8sY9kL3nVp2qR5tW8uZ4aB6cD7eF9gH0jK1mN2oP3qR4sT5uV6w",          // 44 chars (base64)
		"abcdefghijklmnopqrstuvwxyz123456",                               // 32 chars (minimum)
	}

	for _, id := range validIDs {
		err := validateSessionID(id)
		tt.EqualNil(err)
	}

	invalidCases := []struct {
		id      string
		wantErr string
	}{
		{"", "session ID cannot be empty"},
		{"short", "session ID too short"},
		{"abc123", "session ID too short"},
		{"abc def with spaces to make it longer than 32 chars", "invalid character"},
		{"abc@123@invalid@characters@to@make@long@id", "invalid character"},
		{"abc/123/invalid/characters/to/make/long/id", "invalid character"},
		{strings.Repeat("a", 40), "repeated characters"}, // all 'a's
		{"long-id-" + strings.Repeat("a", 250), "session ID too long"},
	}

	for _, tc := range invalidCases {
		err := validateSessionID(tc.id)
		tt.NotNil(err)
		tt.EqualTrue(strings.Contains(err.Error(), tc.wantErr))
	}
}

func TestGenerateSessionID(t *testing.T) {
	tt := zls.NewTest(t)

	generatedIDs := make(map[string]bool)
	for i := 0; i < 100; i++ {
		id, err := generateSessionID()
		tt.EqualNil(err)
		tt.EqualTrue(len(id) >= 32)

		err = validateSessionID(id)
		tt.EqualNil(err)

		tt.EqualFalse(generatedIDs[id])
		generatedIDs[id] = true
	}
}

func TestHashSessionID(t *testing.T) {
	tt := zls.NewTest(t)

	id := "test-session-id-12345"
	hash1 := hashSessionID(id)
	hash2 := hashSessionID(id)

	tt.Equal(hash1, hash2)
	tt.Equal(64, len(hash1))

	id2 := "different-session-id-67890"
	hash3 := hashSessionID(id2)
	if hash1 == hash3 {
		t.Error("Different IDs should produce different hashes")
	}
}

func TestSessionMiddlewareValidation(t *testing.T) {
	tt := zls.NewTest(t)

	store := NewMemoryStore()
	defer store.Close()

	handler := New(store, func(conf *Config) {
		conf.CookieName = "session_id"
		conf.ExpiresAt = 30 * time.Minute
	})

	r := znet.New()
	r.Use(handler)

	r.GET("/test", func(c *znet.Context) error {
		s, err := Get(c)
		if err != nil {
			c.String(500, "No session")
			return nil
		}
		s.Set("key", "value")
		c.String(200, "OK")
		return nil
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)
	tt.Equal(200, w.Code, true)
}

func TestInvalidSessionIDRejection(t *testing.T) {
	tt := zls.NewTest(t)

	store := NewMemoryStore()
	defer store.Close()
	handler := New(store, func(conf *Config) {
		conf.CookieName = "session_id"
		conf.ExpiresAt = 30 * time.Minute
	})

	r := znet.New()
	r.Use(handler)
	r.GET("/test", func(c *znet.Context) error {
		s, err := Get(c)
		if err != nil {
			return err
		}
		c.String(200, s.ID())
		return nil
	})

	invalidIDs := []string{
		"short",
		"abc@123-with-invalid-characters-and-long-enough",
		"abc def with spaces and long enough for validation",
		strings.Repeat("a", 257),
	}

	for _, id := range invalidIDs {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: id})

		r.ServeHTTP(w, req)

		tt.Equal(200, w.Code, true)
		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatalf("expected replacement session cookie for %q", id)
		}
		newID := cookies[0].Value
		tt.EqualFalse(newID == id)
		tt.EqualNil(validateSessionID(newID))
	}
}

func TestSecureSessionIDGeneration(t *testing.T) {
	tt := zls.NewTest(t)

	store := NewMemoryStore()
	defer store.Close()

	generated := make(map[string]bool)
	for i := 0; i < 1000; i++ {
		id, err := generateSessionID()
		tt.EqualNil(err)
		tt.EqualFalse(generated[id])
		generated[id] = true

		sess, err := store.New(id, time.Now().Add(30*time.Minute))
		tt.EqualNil(err)
		tt.NotNil(sess)
	}
}

func TestSessionFixationPrevention(t *testing.T) {
	tt := zls.NewTest(t)
	weakIDs := []string{
		"password",      // dictionary word
		"12345678",      // all numbers
		"aaaaaaaa",      // repeated characters
		"abc123",        // too short
		"admin-session", // predictable
	}

	store := NewMemoryStore()
	defer store.Close()
	handler := New(store, func(conf *Config) {
		conf.CookieName = "session_id"
		conf.ExpiresAt = 30 * time.Minute
	})

	r := znet.New()
	r.Use(handler)
	r.GET("/test", func(c *znet.Context) error {
		s, err := Get(c)
		if err != nil {
			return err
		}
		c.String(200, s.ID())
		return nil
	})

	for _, weakID := range weakIDs {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/test", nil)
		req.AddCookie(&http.Cookie{Name: "session_id", Value: weakID})

		r.ServeHTTP(w, req)

		tt.Equal(200, w.Code, true)
		cookies := w.Result().Cookies()
		if len(cookies) == 0 {
			t.Fatalf("expected replacement session cookie for %q", weakID)
		}
		newID := cookies[0].Value
		tt.EqualFalse(newID == weakID)
		tt.EqualNil(validateSessionID(newID))
	}
}
