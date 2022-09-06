package zhttp

import (
	"testing"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zlog"
)

func TestJSONRPC(t *testing.T) {
	tt := zlsgo.NewTest(t)

	client, err := JSONRPC("18181", "/__rpc")
	tt.NoError(err)

	var result Result
	err = client.Call("Cal.Square", 12, &result)
	tt.NoError(err)
	tt.Equal(144, result.Ans)
	tt.Equal(12, result.Num)
	t.Log(result, err)

	res, err := Post("http://127.0.0.1:18181/__rpc", BodyJSON(map[string]interface{}{
		"method": "Cal.Square",
		"params": []int{12},
		"id":     1,
	}))
	tt.NoError(err)
	tt.Equal(144, res.JSON("result.Ans").Int())
	zlog.Debug(res, err)
}
