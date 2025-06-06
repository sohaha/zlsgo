package zstring

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"math"
	"os"
	"sort"
	"strconv"
	"time"
	"unsafe"
)

// RandUint32Max returns a pseudorandom uint32 value in the range [0..max).
// It uses a fast algorithm to generate uniformly distributed random numbers
// within the specified range.
func RandUint32Max(max uint32) uint32 {
	x := RandUint32()
	return uint32((uint64(x) * uint64(max)) >> 32)
}

// RandInt generates a random integer within the specified inclusive range [min..max].
// If max is less than min, max will be set equal to min.
func RandInt(min int, max int) int {
	if max < min {
		max = min
	}

	if max == min {
		return min
	}

	range64 := int64(max) - int64(min) + 1

	if range64 > int64(^uint32(0)) {
		random := int64(RandUint32())<<31 | int64(RandUint32())
		return min + int(random%range64)
	}

	return min + int(RandUint32Max(uint32(range64)))
}

// Rand generates a random string of the specified length.
// By default, it uses alphanumeric characters (0-9, a-z, A-Z).
// An optional template string can be provided to limit the characters used.
func Rand(n int, tpl ...string) string {
	var s []rune
	b := make([]byte, n)
	if len(tpl) > 0 {
		s = []rune(tpl[0])
	} else {
		s = letterBytes
	}
	l := len(s) - 1
	for i := n - 1; i >= 0; i-- {
		idx := RandInt(0, l)
		b[i] = byte(s[idx])
	}
	return Bytes2String(b)
}

// UniqueID generates a cryptographically secure unique ID with at least the specified length.
// The ID is generated using crypto/rand for better security and uniqueness.
// If crypto/rand fails, it falls back to a more robust alternative method that
// combines multiple entropy sources to maintain uniqueness.
func UniqueID(n int) string {
	if n < 6 {
		n = 6
	}
	k := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		timestamp := time.Now().UnixNano()
		procPart := strconv.Itoa(os.Getpid() % 10000)
		random := Rand(n / 2)
		timePart := strconv.FormatInt(timestamp, 36)

		result := random + timePart + procPart
		if len(result) > n {
			return result[:n]
		}
		return result + Rand(n-len(result))
	}
	return hex.EncodeToString(k)
}

type (
	// Weighteder implements weighted random selection from a set of choices.
	// Each choice has an associated weight that determines its probability of being selected.
	Weighteder struct {
		// choices contains the items that can be selected
		choices []interface{}
		// totals contains the cumulative weights used for binary search
		totals []uint32
		// max is the sum of all weights
		max uint32
	}

	// choice represents an item with its associated weight for internal use
	choice struct {
		// Item is the value that can be selected
		Item interface{}
		// Weight determines the relative probability of this item being selected
		Weight uint32
	}
)

// WeightedRand performs a single weighted random selection from the provided choices.
// Each key in the map is an item that can be selected, and its value is the weight
// that determines its probability of being selected.
func WeightedRand(choices map[interface{}]uint32) (interface{}, error) {
	w, err := NewWeightedRand(choices)
	if err != nil {
		return nil, err
	}
	return w.Pick(), nil
}

// NewWeightedRand creates a new Weighteder for efficient repeated weighted random selections.
// It initializes the internal data structures needed for fast weighted selection.
func NewWeightedRand(choices map[interface{}]uint32) (*Weighteder, error) {
	if len(choices) == 0 {
		return nil, errors.New("choices is empty")
	}
	cs := make([]choice, 0, len(choices))
	for k, v := range choices {
		cs = append(cs, choice{Item: k, Weight: v})
	}

	sort.Slice(cs, func(i, j int) bool {
		return cs[i].Weight < cs[j].Weight
	})
	w := &Weighteder{
		totals:  make([]uint32, len(choices)),
		choices: make([]interface{}, len(choices)),
		max:     0,
	}
	for i := range cs {
		if cs[i].Weight < 0 {
			continue // ignore negative weights, can never be picked
		}

		if cs[i].Weight >= ^uint32(0) {
			return nil, errors.New("weight overflowed")
		}

		if (^uint32(0) - w.max) <= cs[i].Weight {
			return nil, errors.New("total weight overflowed")
		}

		w.max += cs[i].Weight
		w.totals[i] = w.max
		w.choices[i] = cs[i].Item
	}

	return w, nil
}

// Pick selects and returns a random item from the Weighteder based on the weights.
// Items with higher weights are more likely to be selected.
func (w *Weighteder) Pick() interface{} {
	return w.choices[w.weightedSearch()]
}

// weightedSearch performs a binary search on the cumulative weights
// to find the index of the selected item based on a random value.
// This is an internal method used by Pick().
func (w *Weighteder) weightedSearch() int {
	x := RandUint32Max(w.max) + 1
	i, j := 0, len(w.totals)
	if i == j-1 {
		return 0
	}
	for i < j {
		h := int(uint(i+j) >> 1)
		if w.totals[h] < x {
			i = h + 1
		} else {
			j = h
		}
	}
	return i
}

// getMask calculates a bitmask for efficient random character selection.
// It finds the smallest mask that can cover the entire alphabet size.
func getMask(alphabetSize int) byte {
	for i := 1; i <= 8; i++ {
		mask := byte((2 << uint(i)) - 1)
		if int(mask) >= alphabetSize-1 {
			return mask
		}
	}
	return 0
}

// NewNanoID generates a secure unique string ID of the specified size.
// It's similar to UUID but more compact and customizable.
// By default, it uses alphanumeric characters, but a custom alphabet can be provided.
func NewNanoID(size int, alphabet ...string) (string, error) {
	var chars []rune
	if len(alphabet) == 0 {
		chars = letterBytes
	} else {
		chars = []rune(alphabet[0])
	}

	alphabetSize := len(chars)
	if alphabetSize == 0 || alphabetSize > 255 {
		return "", errors.New("alphabet must not be empty and contain no more than 255 chars")
	}
	if size <= 0 {
		return "", errors.New("size must be positive integer")
	}

	mask := getMask(alphabetSize)
	ceilArg := 1.6 * float64(int(mask)*size) / float64(alphabetSize)
	step := int(math.Ceil(ceilArg))

	id := make([]byte, 0, size)
	bytes := make([]byte, step)

	alphabetBytes := make([]byte, alphabetSize)
	for i, r := range chars {
		alphabetBytes[i] = byte(r)
	}

	isASCII := true
	for _, r := range chars {
		if r > 127 {
			isASCII = false
			break
		}
	}

	if isASCII {
		for {
			if _, err := rand.Read(bytes); err != nil {
				return "", err
			}

			for i := 0; i < step && len(id) < size; i++ {
				idx := bytes[i] & mask
				if idx < byte(alphabetSize) {
					id = append(id, alphabetBytes[idx])
				}
			}

			if len(id) >= size {
				return *(*string)(unsafe.Pointer(&id)), nil
			}
		}
	} else {
		result := make([]rune, size)
		j := 0

		for {
			if _, err := rand.Read(bytes); err != nil {
				return "", err
			}

			for i := 0; i < step && j < size; i++ {
				idx := bytes[i] & mask
				if idx < byte(alphabetSize) {
					result[j] = chars[idx]
					j++
				}
			}

			if j >= size {
				return string(result), nil
			}
		}
	}
}

// UUID generates UUID v4
func UUID() string {
	var buf [16]byte

	n, err := rand.Read(buf[:])
	if err != nil || n != 16 {
		return ""
	}

	buf[6] = (buf[6] & 0x0f) | 0x40 // UUID version 4
	buf[8] = (buf[8] & 0x3f) | 0x80 // RFC 4122 variant

	var dst [36]byte
	const hextable = "0123456789abcdef"
	for i, j := 0, 0; i < len(buf); {
		if j == 8 || j == 13 || j == 18 || j == 23 {
			dst[j] = '-'
			j++
		}
		dst[j] = hextable[buf[i]>>4]
		dst[j+1] = hextable[buf[i]&0x0f]
		j += 2
		i++
	}
	return Bytes2String(dst[:])
}
