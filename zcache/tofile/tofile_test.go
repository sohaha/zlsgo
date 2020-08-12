package tofile

import (
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zcache"
	"github.com/sohaha/zlsgo/zfile"
)

type testSt struct {
	Name string
	Key  int
}

func TestToFile(t *testing.T) {
	tt := zlsgo.NewTest(t)
	cache := zcache.New("file")
	err := zfile.WriteFile("tmp.json", []byte(`{"tmp1":"XhAAFSp0b2ZpbGUucGVyc2lzdGVuY2VTdP+BAwEBDXBlcnNpc3RlbmNlU3QB/4IAAQMBBERhdGEBEAABCExpZmVTcGFuAQQAARBJbnRlcnZhbExpZmVTcGFuAQIAAABP/4I1AQ4qdG9maWxlLnRlc3RTdP+DAwEBBnRlc3RTdAH/hAABAgEETmFtZQEMAAEDS2V5AQQAAAAW/4QJAQZpc05hbWUAAfsFloLwAAEBAA==","tmp3":"XhAAFSp0b2ZpbGUucGVyc2lzdGVuY2VTdP+BAwEBDXBlcnNpc3RlbmNlU3QB/4IAAQMBBERhdGEBEAABCExpZmVTcGFuAQQAARBJbnRlcnZhbExpZmVTcGFuAQIAAAAg/4IdAQZzdHJpbmcMCwAJaXMgc3RyaW5nAfsBZaC8AAA=","tmp2":"XhAAFSp0b2ZpbGUucGVyc2lzdGVuY2VTdP+BAwEBDXBlcnNpc3RlbmNlU3QB/4IAAQMBBERhdGEBEAABCExpZmVTcGFuAQQAARBJbnRlcnZhbExpZmVTcGFuAQIAAAAW/4ITAQNpbnQEBAD+BTQB+wFloLwAAA=="}`))
	tt.EqualExit(nil, err)
	save, err := PersistenceToFile(cache, "tmp.json", false, &testSt{})
	tt.EqualExit(nil, err)

	cache.Set("tmp1", &testSt{
		Name: "isName",
		Key:  0,
	}, 3, true)
	cache.Set("tmp2", 666, 3)
	cache.Set("tmp3", "is string", 3)

	cache.ForEach(func(key string, value interface{}) bool {
		tt.Log(key)
		return true
	})
	tmp1, err := cache.Get("tmp1")
	tt.EqualExit(nil, err)
	tt.Equal("isName", tmp1.(*testSt).Name)
	tmp2, err := cache.GetInt("tmp2")
	tt.EqualNil(err)
	tt.Equal(666, tmp2)
	tmp3, err := cache.GetString("tmp3")
	tt.EqualNil(err)
	tt.Equal("is string", tmp3)
	go func() {
		time.Sleep(500 * time.Millisecond)
		_, _ = cache.Get("tmp1")
	}()
	time.Sleep(3*time.Second + 300*time.Millisecond)
	tmp1, err = cache.Get("tmp1")
	tt.EqualNil(err)
	t.Log(tmp1)
	_, err = cache.GetInt("tmp2")
	tt.EqualTrue(err != nil)
	_, err = cache.GetString("tmp3")
	tt.EqualTrue(err != nil)

	cache.Set("tmp0", 1, 2)
	tt.Equal(2, cache.Count())
	zfile.Rmdir("tmp.json")

	err = save()
	tt.EqualNil(err)
	tt.EqualTrue(zfile.FileExist("tmp.json"))

	_ = zfile.WriteFile("tmp.json", []byte(`{"tmp1":"XhAAFSp0b2ZpbGUucGVyc2lzdGVuY2VTdP+BAwEBDXBlcnNpc3RlbmNlU3QB/4IAAQMBBERhdGEBEAABCExpZmVTcGFuAQQAARBJbnRlcnZhbExpZmVTcGFuAQIAAABP/4I1AQ4qdG9maWxlLnRlc3RTdP+DAwEBBnRlc3RTdAH/hAABAgEETmFtZQEMAAEDS2V"}`))
	save, err = PersistenceToFile(cache, "tmp.json", false)
	tt.EqualTrue(err != nil)

	zfile.Rmdir("tmp.json")
}
