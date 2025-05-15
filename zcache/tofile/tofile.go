// Package tofile provides functionality for persisting cache data to disk
// and restoring it when the application restarts
package tofile

import (
	"encoding/gob"
	"path/filepath"
	"time"

	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zfile"
	"github.com/sohaha/zlsgo/zjson"
	"github.com/sohaha/zlsgo/zstring"
)

type (
	// persistenceSt represents a cache item's data and metadata for serialization
	persistenceSt struct {
		// Data is the actual cached value
		Data             interface{}
		// LifeSpan is the duration for which this item will remain in the cache
		LifeSpan         time.Duration
		// IntervalLifeSpan indicates whether the expiration timer resets on access
		IntervalLifeSpan bool
	}
)

// PersistenceToFile sets up persistence for a cache table to a JSON file.
// It loads the cache data from the file if it exists and DisableAutoLoad is false.
// It returns a save function that can be called to persist the current cache state to disk.
//
// Parameters:
//   - table: The cache table to persist
//   - file: The path to the file where cache data will be stored
//   - DisableAutoLoad: If true, does not load cache data from the file on startup
//   - register: Types that need to be registered with gob for serialization
//
// Returns:
//   - save: A function that when called will save the current cache state to disk
//   - err: Any error that occurred during initial loading
func PersistenceToFile(table *zcache.Table, file string, DisableAutoLoad bool, register ...interface{}) (save func() error, err error) {
	gob.Register(&persistenceSt{})
	file = zfile.RealPath(file)

	for k := range register {
		gob.Register(register[k])
	}

	if !DisableAutoLoad && zfile.FileExist(file) {
		var content []byte
		content, err = zfile.ReadFile(file)
		if err != nil {
			return
		}
		zjson.ParseBytes(content).ForEach(func(k, v *zjson.Res) (b bool) {
			b = true
			key := k.String()
			base64 := zstring.String2Bytes(v.String())
			base64, err = zstring.Base64Decode(base64)
			if err != nil {
				return false
			}
			var value interface{}
			value, err = zstring.UnSerialize(base64)
			if err != nil {
				return false
			}

			if persistence, ok := value.(*persistenceSt); ok {
				table.SetRaw(key, persistence.Data, persistence.LifeSpan, persistence.IntervalLifeSpan)
			}
			return
		})
	}
	save = func() error {
		jsonData := exportJSON(table)
		_ = zfile.RealPathMkdir(filepath.Dir(file))

		return zfile.WriteFile(file, zstring.String2Bytes(jsonData))
	}
	return
}

// exportJSON serializes a cache table to a JSON string.
// Each cache item is serialized and base64 encoded to preserve binary data.
//
// Parameters:
//   - table: The cache table to export
//   - registers: Types that need to be registered with gob for serialization
//
// Returns:
//   - A JSON string representation of the cache table
func exportJSON(table *zcache.Table, registers ...interface{}) string {
	for i := range registers {
		gob.Register(registers[i])
	}
	jsonData := "{}"
	table.ForEachRaw(func(key string, item *zcache.Item) bool {
		item.RLock()
		v := &persistenceSt{
			Data:             item.Data(),
			LifeSpan:         item.LifeSpan(),
			IntervalLifeSpan: item.IntervalLifeSpan(),
		}
		item.RUnlock()
		value, err := zstring.Serialize(v)
		if err != nil {
			return true
		}
		jsonData, _ = zjson.Set(jsonData, key, zstring.Base64Encode(value))

		return true
	})
	return jsonData
}
