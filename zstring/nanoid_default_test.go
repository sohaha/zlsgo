package zstring

import "testing"

// TestDefaultAlphabetPrecomputed guards the init-time derivation of the byte
// alphabet and its ASCII flag: a non-ASCII default must not be forced onto the
// byte path.
func TestDefaultAlphabetPrecomputed(t *testing.T) {
	if !defaultAlphabetIsASCII {
		t.Fatalf("default alphabet must be ASCII")
	}
	if len(defaultAlphabetASCII) != len(letterBytes) {
		t.Fatalf("len = %d, want %d", len(defaultAlphabetASCII), len(letterBytes))
	}
	for i, r := range letterBytes {
		if defaultAlphabetASCII[i] != byte(r) {
			t.Fatalf("byte %d = %q, want %q", i, defaultAlphabetASCII[i], byte(r))
		}
	}
}

// TestNewNanoIDUsesDefaultAlphabet checks the default path only emits runes
// from letterBytes for several sizes.
func TestNewNanoIDUsesDefaultAlphabet(t *testing.T) {
	allowed := make(map[rune]bool, len(letterBytes))
	for _, r := range letterBytes {
		allowed[r] = true
	}
	for _, size := range []int{1, 10, 64} {
		id, err := NewNanoID(size)
		if err != nil {
			t.Fatalf("size %d: %v", size, err)
		}
		if len(id) != size {
			t.Fatalf("size %d: got len %d", size, len(id))
		}
		for _, r := range id {
			if !allowed[r] {
				t.Fatalf("size %d: unexpected rune %q in %q", size, r, id)
			}
		}
	}
}
