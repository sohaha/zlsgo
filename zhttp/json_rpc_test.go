package zhttp

import (
	"errors"
	"testing"
	"time"

	"github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/zlog"
)

func TestNewJSONRPC(t *testing.T) {
	tt := zlsgo.NewTest(t)

	client, err := NewJSONRPC(testServerHost, "/__rpc", func(o *JSONRPCOptions) {
		o.Timeout = time.Second * 10
		o.Retry = true
		o.RetryDelay = time.Second * 6
		// o.TlsConfig = &tls.Config{}
	})
	tt.NoError(err)

	res, err := Get(testServerURL + "/__rpc")
	tt.NoError(err)
	tt.Equal([]interface{}{"int", "*zhttp.Result"}, res.JSON("Cal\\.Square").Value())
	zlog.Debug(res)

	var result Result
	err = client.Call("Cal.Square", 12, &result)
	tt.NoError(err)
	tt.Equal(144, result.Ans)
	tt.Equal(12, result.Num)
	t.Log(result, err)

	res, err = Post(testServerURL+"/__rpc", BodyJSON(map[string]interface{}{
		"method": "Cal.Square",
		"params": []int{12},
		"id":     1,
	}))
	tt.NoError(err)
	tt.Equal(144, res.JSON("result.Ans").Int())
	zlog.Debug(res, err)

	res, err = Put(testServerURL+"/__rpc", BodyJSON(map[string]interface{}{
		"method": "Cal.Square",
		"params": []int{12},
		"id":     1,
	}))
	t.Log(err)
	t.Log(res)

	_ = client.Close()
}

func TestJSONRPCNotConnected(t *testing.T) {
	tt := zlsgo.NewTest(t)

	client, err := NewJSONRPC("127.0.0.1:1", "/__rpc", func(o *JSONRPCOptions) {
		o.Timeout = time.Second
		o.Retry = false
	})
	tt.EqualTrue(err != nil)
	if client == nil {
		t.Fatal("NewJSONRPC must return a client even when connecting fails")
	}

	var result Result
	err = client.Call("Cal.Square", 12, &result)
	tt.EqualTrue(errors.Is(err, ErrJSONRPCNotConnected))
	tt.Equal(0, result.Ans)

	call := client.Go("Cal.Square", 12, &result, nil)
	tt.EqualTrue(errors.Is(call.Error, ErrJSONRPCNotConnected))
	tt.EqualTrue(call == <-call.Done)

	tt.EqualNil(client.Close())
}
