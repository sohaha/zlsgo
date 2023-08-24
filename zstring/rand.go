package zstring

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"io"
	"sort"
	"strconv"
	"time"
)

// RandUint32Max returns pseudorandom uint32 in the range [0..max)
func RandUint32Max(max uint32) uint32 {
	x := RandUint32()
	return uint32((uint64(x) * uint64(max)) >> 32)
}

// RandInt random numbers in the specified range
func RandInt(min int, max int) int {
	if max < min {
		max = min
	}
	return min + int(RandUint32Max(uint32(max+1-min)))
}

// Rand random string of specified length, the second parameter limit can only appear the specified character
func Rand(n int, tpl ...string) string {
	var s string
	b := make([]byte, n)
	if len(tpl) > 0 {
		s = tpl[0]
	} else {
		s = letterBytes
	}
	l := len(s) - 1
	for i := n - 1; i >= 0; i-- {
		idx := RandInt(0, l)
		b[i] = s[idx]
	}
	return Bytes2String(b)
}

// UniqueID unique id minimum 6 digits
func UniqueID(n int) string {
	if n < 6 {
		n = 6
	}
	k := make([]byte, n)
	if _, err := io.ReadFull(rand.Reader, k); err != nil {
		return Rand(n-2) + strconv.Itoa(time.Now().Nanosecond()/10000000)
	}
	return hex.EncodeToString(k)
}

type (
	Weighteder struct {
		choices []interface{}
		totals  []uint32
		max     uint32
	}
	choice struct {
		Item   interface{}
		Weight uint32
	}
)

func WeightedRand(choices map[interface{}]uint32) (interface{}, error) {
	w, err := NewWeightedRand(choices)
	if err != nil {
		return nil, err
	}
	return w.Pick(), nil
}

func NewWeightedRand(choices map[interface{}]uint32) (*Weighteder, error) {
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

func (w *Weighteder) Pick() interface{} {
	return w.choices[w.weightedSearch()]
}

func (w *Weighteder) weightedSearch() int {
	x := RandUint32Max(w.max) + 1
	i, j := 0, len(w.totals)
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

var idWorkers, _ = NewIDWorker(0)

func UUID() int64 {
	id, _ := idWorkers.ID()
	return id
}
