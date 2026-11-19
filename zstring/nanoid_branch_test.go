package zstring

import "testing"

// TestNewNanoIDCustomAlphabetBranch pins the ASCII/rune branch decision for a
// non-default alphabet. The byte path may only be taken for a pure-ASCII
// alphabet; any alphabet containing a non-ASCII rune must take the rune path so
// that multi-byte characters are never truncated. Every generated rune must come
// from the caller's alphabet, regardless of where the non-ASCII rune sits.
func TestNewNanoIDCustomAlphabetBranch(t *testing.T) {
	alphabets := []string{
		"0123456789abc", // pure ASCII -> byte path
		"abc你",          // non-ASCII at the end
		"你abc",          // non-ASCII at the front
		"你好世界",          // all non-ASCII
		"aé中9",          // mixed
		"abc你abc",       // non-ASCII in the middle
	}
	for _, alphabet := range alphabets {
		allowed := make(map[rune]bool)
		for _, r := range alphabet {
			allowed[r] = true
		}
		for i := 0; i < 200; i++ {
			id, err := NewNanoID(64, alphabet)
			if err != nil {
				t.Fatalf("alphabet %q: %v", alphabet, err)
			}
			for _, r := range id {
				if !allowed[r] {
					t.Fatalf("alphabet %q produced %q (U+%04X), not in alphabet", alphabet, r, r)
				}
			}
		}
	}
}
