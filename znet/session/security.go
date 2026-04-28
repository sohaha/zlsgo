package session

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"strconv"
	"strings"
)

// errInvalidSessionID is returned when a session ID fails validation
var errInvalidSessionID = errors.New("invalid session ID format")

// ValidateSessionID checks if a session ID meets security requirements.
// It validates length, character set, and format to prevent session fixation attacks.
// A valid session ID should be at least 32 characters and contain only alphanumeric
// characters, hyphens, or underscores (safe for HTTP cookies and URLs).
func validateSessionID(sessionID string) error {
	if sessionID == "" {
		return errors.New("session ID cannot be empty")
	}

	if len(sessionID) < 32 {
		return errors.New("session ID too short (minimum 32 characters required for security)")
	}

	if len(sessionID) > 256 {
		return errors.New("session ID too long (maximum 256 characters)")
	}

	for i, c := range sessionID {
		if !((c >= 'a' && c <= 'z') ||
			(c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') ||
			c == '-' || c == '_') {
			return errors.New("invalid character at position " + strconv.Itoa(i) +
				": only alphanumeric, hyphen, and underscore allowed")
		}
	}

	var firstChar rune
	for i, c := range sessionID {
		if i == 0 {
			firstChar = c
		} else if c != firstChar {
			return nil
		}
	}
	return errors.New("session ID consists of repeated characters (weak entropy)")
}

// generateSessionID creates a cryptographically secure random session ID.
// It uses crypto/rand to generate 32 random bytes (256 bits of entropy)
// and encodes them using base64 URL encoding for safe HTTP transport.
// The resulting session ID is 44 characters long and provides strong security guarantees.
func generateSessionID() (string, error) {
	buf := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, buf); err != nil {
		return "", errors.New("failed to generate random session ID: " + err.Error())
	}

	sessionID := base64.RawURLEncoding.EncodeToString(buf)
	if err := validateSessionID(sessionID); err != nil {
		return "", errors.New("generated session ID failed validation: " + err.Error())
	}

	return sessionID, nil
}

// hashSessionID creates a SHA-256 hash of the session ID.
// This is useful for logging or indexing session IDs without exposing the actual values.
// The hash is returned as a hexadecimal string (64 characters).
func hashSessionID(sessionID string) string {
	h := sha256.New()
	h.Write([]byte(sessionID))
	return strings.ToLower(hex.EncodeToString(h.Sum(nil)))
}
