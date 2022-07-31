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
	persistenceSt struct {
		Data             interface{}
		LifeSpan         time.Duration
		IntervalLifeSpan bool
	}
)

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
