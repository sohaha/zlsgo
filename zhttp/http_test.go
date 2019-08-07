package zhttp

import (
	"fmt"
	zls "github.com/sohaha/zlsgo"
	"github.com/sohaha/zlsgo/znet"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

var (
	one    sync.Once
	engine *znet.Engine
)

func TestHttp(T *testing.T) {
	t := zls.NewTest(T)
	var (
		res          *Res
		err          error
		data         string
		expectedText string
	)

	forMethod(t)

	GetMethod(t)

	// test post
	expectedText = "ok"
	res, err = newMethod("Post", func(w http.ResponseWriter, r *http.Request) {
		_ = r.ParseForm()
		id, ok := r.PostForm["id"]
		if !ok {
			t.T.Fatal("err")
		}
		_, _ = w.Write([]byte(expectedText + id[0]))
	}, Param{
		"id": "123",
	})
	if err != nil {
		t.T.Fatal(err)
	}
	data = res.String()
	t.Equal(expectedText+"123", data)

	// test post application/x-www-form-urlencoded
	expectedText = "ok"
	res, err = newMethod("Post", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query()
		id, ok := query["id"]
		if !ok {
			t.T.Fatal("err")
		}
		_, _ = w.Write([]byte(expectedText + id[0]))
	}, QueryParam{
		"id": "123",
	})
	if err != nil {
		t.T.Fatal(err)
	}
	data = res.String()
	t.Equal(expectedText+"123", data)

}

func GetMethod(t *zls.TestUtil) {
	jsonData := struct {
		Code int `json:"code"`
	}{}
	data := ""
	values := [...]string{
		"text",
		"{\"code\":200}",
	}
	for _, v := range values {
		res, err := newMethod("GET", func(w http.ResponseWriter, _ *http.Request) {
			_, _ = w.Write([]byte(v))
		})
		if err != nil {
			t.T.Fatal(err)
		}
		if err = res.ToJSON(&jsonData); err == nil {
			t.Equal(200, jsonData.Code)
			return
		}
		if data, err = res.ToString(); err == nil {
			t.Equal(v, data)
		}

	}
}

func forMethod(t *zls.TestUtil) {
	values := [...]string{"Get", "Put", "Head", "Options", "Delete", "Patch"}
	for _, v := range values {
		_, err := newMethod(v, func(_ http.ResponseWriter, _ *http.Request) {
		})
		if err != nil {
			t.T.Fatal(v, err)
		}
	}
}

func newMethod(method string, handler func(_ http.ResponseWriter, _ *http.Request), param ...interface{}) (res *Res, err error) {
	ts := httptest.NewServer(http.HandlerFunc(handler))
	url := ts.URL
	switch method {
	case "Get":
		res, err = Get(url, param...)
	case "Post":
		res, err = Post(url, param...)
	case "Put":
		res, err = Put(url, param...)
	case "Head":
		res, err = Head(url, param...)
	case "Options":
		res, err = Options(url, param...)
	case "Delete":
		res, err = Delete(url, param...)
	case "Patch":
		res, err = Patch(url, param...)
	default:
		method = strings.Title(method)
		res, err = Do(method, url, param...)
		if err == nil {
			fmt.Println(res.Dump())
		}
	}

	return
}
