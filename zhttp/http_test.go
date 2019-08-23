package zhttp

import (
	"fmt"
	zls "github.com/sohaha/zlsgo"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"
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
	t.Log(res.GetCookie())
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
	EnableCookie(false)
	for _, v := range values {
		res, err := newMethod("GET", func(w http.ResponseWriter, _ *http.Request) {
			cookie := &http.Cookie{
				Name:     "c",
				Value:    v,
				Path:     "/",
				HttpOnly: true,
				MaxAge:   0,
			}
			w.Header().Add("Set-Cookie", cookie.String())
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
		t.Equal("GET", res.Request().Method)
		t.Log(res.GetCookie())
		t.Log(res.Body())
	}
	EnableCookie(true)
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
	curl := ts.URL
	switch method {
	case "Get":
		res, err = Get(curl, param...)
	case "Post":
		res, err = Post(curl, param...)
	case "Put":
		res, err = Put(curl, param...)
	case "Head":
		res, err = Head(curl, param...)
	case "Options":
		res, err = Options(curl, param...)
	case "Delete":
		res, err = Delete(curl, param...)
	case "Patch":
		res, err = Patch(curl, param...)
	default:
		method = strings.Title(method)
		res, err = Do(method, curl, param...)
		if err == nil {
			fmt.Println(res.Dump())
		}
	}

	return
}

func TestHttpProxy(T *testing.T) {
	t := zls.NewTest(T)
	err := SetProxy(func(r *http.Request) (*url.URL, error) {
		t.Log(r.URL.String())
		if strings.Contains(r.URL.String(), "qq.com") {
			return url.Parse("http://127.0.0.1:6666")
		}
		return nil, nil
	})

	if err != nil {
		t.T.Fatal(err)
	}

	SetTimeout(1 * time.Second)

	_, err = Get("http://www.qq.com")
	t.Equal(true, err != nil)
	t.Log(err)

	_, err = Get("http://baidu.com")
	t.Equal(false, err != nil)
	t.Log(err)
}

func TestHttpProxyUrl(T *testing.T) {
	err := SetProxyUrl("http://127.0.0.1:6666")
	t := zls.NewTest(T)
	if err != nil {
		t.T.Fatal(err)
	}

	SetTimeout(1 * time.Second)
	_, err = newMethod("GET", func(w http.ResponseWriter, _ *http.Request) {
	})
	t.Equal(true, err != nil)
	t.Log(err)
}
